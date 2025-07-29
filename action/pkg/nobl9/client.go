package nobl9

import (
	"context"
	"fmt"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/users/v2"
	"github.com/your-org/nobl9-action/pkg/errors"
	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/retry"
)

// Client wraps the Nobl9 SDK client with additional functionality
type Client struct {
	sdkClient *sdk.Client
	logger    *logger.Logger
	config    *Config
	retryOp   *retry.RetryableAPIOperation
}

// Config holds Nobl9 client configuration
type Config struct {
	ClientID      string
	ClientSecret  string
	Timeout       time.Duration
	RetryAttempts int
}

// New creates a new Nobl9 client
func New(config *Config, log *logger.Logger) (*Client, error) {
	if config == nil {
		return nil, errors.NewConfigError("config cannot be nil", nil)
	}

	if log == nil {
		return nil, errors.NewConfigError("logger cannot be nil", nil)
	}

	// Validate required configuration
	if err := validateConfig(config); err != nil {
		return nil, errors.NewConfigError("invalid configuration", err)
	}

	// Create SDK client configuration
	sdkConfig := &sdk.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Timeout:      config.Timeout,
	}

	// Create SDK client
	sdkClient, err := sdk.NewClient(sdkConfig)
	if err != nil {
		return nil, errors.NewConfigError("failed to create Nobl9 SDK client", err)
	}

	// Create retry policy for API operations
	retryPolicy := retry.CreatePolicyForAPI(config.RetryAttempts)
	retryOp := retry.NewRetryableAPIOperation(retryPolicy, log)

	client := &Client{
		sdkClient: sdkClient,
		logger:    log,
		config:    config,
		retryOp:   retryOp,
	}

	// Test connection
	if err := client.testConnection(); err != nil {
		return nil, errors.NewNobl9APIError("failed to connect to Nobl9", err)
	}

	log.Info("Nobl9 client created successfully", logger.Fields{
		"timeout":        config.Timeout.String(),
		"retry_attempts": config.RetryAttempts,
	})

	return client, nil
}

// validateConfig validates the client configuration
func validateConfig(config *Config) error {
	if config == nil {
		return errors.NewConfigError("config cannot be nil", nil)
	}

	if config.ClientID == "" {
		return errors.NewConfigError("client ID is required", nil)
	}

	if config.ClientSecret == "" {
		return errors.NewConfigError("client secret is required", nil)
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryAttempts <= 0 {
		config.RetryAttempts = 3
	}

	return nil
}

// testConnection tests the connection to Nobl9
func (c *Client) testConnection() error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	// Test connection by getting organization info with retry logic
	fn := func(ctx context.Context) (interface{}, error) {
		return c.sdkClient.GetOrganization(ctx)
	}

	result, err := c.retryOp.Execute(ctx, "test connection", fn)
	if err != nil {
		c.logger.LogDetailedError(err, "test connection", map[string]interface{}{
			"endpoint": "/organizations",
			"method":   "GET",
			"duration": time.Since(start).String(),
		}, logger.Fields{
			"error": err.Error(),
		})
		return errors.NewNobl9APIError("failed to connect to Nobl9", err)
	}

	orgName := result.(string)

	c.logger.LogNobl9APICall("GET", "/organizations", true, time.Since(start), logger.Fields{
		"organization": orgName,
	})

	c.logger.Info("Successfully connected to Nobl9", logger.Fields{
		"organization": orgName,
	})

	return nil
}

// GetOrganization returns the current organization information
func (c *Client) GetOrganization(ctx context.Context) (string, error) {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		return c.sdkClient.GetOrganization(ctx)
	}

	result, err := c.retryOp.Execute(ctx, "get organization", fn)
	if err != nil {
		c.logger.LogDetailedError(err, "get organization", map[string]interface{}{
			"endpoint": "/organizations",
			"method":   "GET",
			"duration": time.Since(start).String(),
		}, logger.Fields{
			"error": err.Error(),
		})
		return "", errors.NewNobl9APIError("failed to get organization", err)
	}

	orgName := result.(string)

	c.logger.LogNobl9APICall("GET", "/organizations", true, time.Since(start), logger.Fields{
		"organization": orgName,
	})

	return orgName, nil
}

// GetProject retrieves a project by name
func (c *Client) GetProject(ctx context.Context, name string) (*project.Project, error) {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Get projects with the specific name
		params := v1.GetProjectsRequest{
			Names: []string{name},
		}
		return c.sdkClient.Objects().V1().GetV1alphaProjects(ctx, params)
	}

	result, err := c.retryOp.Execute(ctx, fmt.Sprintf("get project %s", name), fn)
	if err != nil {
		c.logger.LogDetailedError(err, "get project", map[string]interface{}{
			"endpoint":     "/projects/" + name,
			"method":       "GET",
			"project_name": name,
			"duration":     time.Since(start).String(),
		}, logger.Fields{
			"error": err.Error(),
		})
		return nil, errors.NewNobl9APIError(fmt.Sprintf("failed to get project %s", name), err)
	}

	projects := result.([]project.Project)
	if len(projects) == 0 {
		c.logger.LogDetailedError(fmt.Errorf("project not found"), "get project", map[string]interface{}{
			"endpoint":     "/projects/" + name,
			"method":       "GET",
			"project_name": name,
			"duration":     time.Since(start).String(),
		}, logger.Fields{
			"error": "project not found",
		})
		return nil, errors.NewNobl9APIError(fmt.Sprintf("project %s not found", name), fmt.Errorf("project not found"))
	}

	project := &projects[0]

	c.logger.LogNobl9APICall("GET", "/projects/"+name, true, time.Since(start), logger.Fields{
		"project_name": name,
		"project_id":   project.Metadata.Name,
	})

	return project, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, projectObj *project.Project) error {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Convert project to manifest.Object and apply
		objects := []manifest.Object{projectObj}
		return nil, c.sdkClient.Objects().V1().Apply(ctx, objects)
	}

	_, err := c.retryOp.Execute(ctx, fmt.Sprintf("create project %s", projectObj.Metadata.Name), fn)
	if err != nil {
		c.logger.LogNobl9APICall("POST", "/projects", false, time.Since(start), logger.Fields{
			"project_name": projectObj.Metadata.Name,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to create project %s: %w", projectObj.Metadata.Name, err)
	}

	c.logger.LogNobl9APICall("POST", "/projects", true, time.Since(start), logger.Fields{
		"project_name": projectObj.Metadata.Name,
		"project_id":   projectObj.Metadata.Name,
	})

	c.logger.LogProjectOperation("create", projectObj.Metadata.Name, true, logger.Fields{
		"project_id": projectObj.Metadata.Name,
	})

	return nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, projectObj *project.Project) error {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Convert project to manifest.Object and apply
		objects := []manifest.Object{projectObj}
		return nil, c.sdkClient.Objects().V1().Apply(ctx, objects)
	}

	_, err := c.retryOp.Execute(ctx, fmt.Sprintf("update project %s", projectObj.Metadata.Name), fn)
	if err != nil {
		c.logger.LogNobl9APICall("PUT", "/projects/"+projectObj.Metadata.Name, false, time.Since(start), logger.Fields{
			"project_name": projectObj.Metadata.Name,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to update project %s: %w", projectObj.Metadata.Name, err)
	}

	c.logger.LogNobl9APICall("PUT", "/projects/"+projectObj.Metadata.Name, true, time.Since(start), logger.Fields{
		"project_name": projectObj.Metadata.Name,
		"project_id":   projectObj.Metadata.Name,
	})

	c.logger.LogProjectOperation("update", projectObj.Metadata.Name, true, logger.Fields{
		"project_id": projectObj.Metadata.Name,
	})

	return nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, name string) error {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Delete project by name
		return nil, c.sdkClient.Objects().V1().DeleteByName(ctx, manifest.KindProject, "", name)
	}

	_, err := c.retryOp.Execute(ctx, fmt.Sprintf("delete project %s", name), fn)
	if err != nil {
		c.logger.LogNobl9APICall("DELETE", "/projects/"+name, false, time.Since(start), logger.Fields{
			"project_name": name,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to delete project %s: %w", name, err)
	}

	c.logger.LogNobl9APICall("DELETE", "/projects/"+name, true, time.Since(start), logger.Fields{
		"project_name": name,
	})

	c.logger.LogProjectOperation("delete", name, true)

	return nil
}

// ListProjects lists all projects
func (c *Client) ListProjects(ctx context.Context) ([]project.Project, error) {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Get all projects
		params := v1.GetProjectsRequest{}
		return c.sdkClient.Objects().V1().GetV1alphaProjects(ctx, params)
	}

	result, err := c.retryOp.Execute(ctx, "list projects", fn)
	if err != nil {
		c.logger.LogNobl9APICall("GET", "/projects", false, time.Since(start), logger.Fields{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := result.([]project.Project)

	c.logger.LogNobl9APICall("GET", "/projects", true, time.Since(start), logger.Fields{
		"project_count": len(projects),
	})

	return projects, nil
}

// GetRoleBinding retrieves a role binding by name and project
func (c *Client) GetRoleBinding(ctx context.Context, projectName, name string) (*rolebinding.RoleBinding, error) {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Get role bindings with the specific name in the project
		params := v1.GetRoleBindingsRequest{
			Project: projectName,
			Names:   []string{name},
		}
		return c.sdkClient.Objects().V1().GetV1alphaRoleBindings(ctx, params)
	}

	result, err := c.retryOp.Execute(ctx, fmt.Sprintf("get role binding %s in project %s", name, projectName), fn)
	if err != nil {
		c.logger.LogNobl9APICall("GET", "/projects/"+projectName+"/rolebindings/"+name, false, time.Since(start), logger.Fields{
			"project_name":      projectName,
			"role_binding_name": name,
			"error":             err.Error(),
		})
		return nil, fmt.Errorf("failed to get role binding %s in project %s: %w", name, projectName, err)
	}

	roleBindings := result.([]rolebinding.RoleBinding)
	if len(roleBindings) == 0 {
		c.logger.LogNobl9APICall("GET", "/projects/"+projectName+"/rolebindings/"+name, false, time.Since(start), logger.Fields{
			"project_name":      projectName,
			"role_binding_name": name,
			"error":             "role binding not found",
		})
		return nil, fmt.Errorf("role binding %s not found in project %s", name, projectName)
	}

	roleBinding := &roleBindings[0]

	c.logger.LogNobl9APICall("GET", "/projects/"+projectName+"/rolebindings/"+name, true, time.Since(start), logger.Fields{
		"project_name":      projectName,
		"role_binding_name": name,
		"role_binding_id":   roleBinding.Metadata.Name,
	})

	return roleBinding, nil
}

// CreateRoleBinding creates a new role binding
func (c *Client) CreateRoleBinding(ctx context.Context, roleBindingObj *rolebinding.RoleBinding) error {
	start := time.Now()

	projectName := roleBindingObj.Spec.ProjectRef

	fn := func(ctx context.Context) (interface{}, error) {
		// Convert role binding to manifest.Object and apply
		objects := []manifest.Object{roleBindingObj}
		return nil, c.sdkClient.Objects().V1().Apply(ctx, objects)
	}

	_, err := c.retryOp.Execute(ctx, fmt.Sprintf("create role binding %s in project %s", roleBindingObj.Metadata.Name, projectName), fn)
	if err != nil {
		c.logger.LogNobl9APICall("POST", "/projects/"+projectName+"/rolebindings", false, time.Since(start), logger.Fields{
			"project_name":      projectName,
			"role_binding_name": roleBindingObj.Metadata.Name,
			"error":             err.Error(),
		})
		return fmt.Errorf("failed to create role binding %s in project %s: %w", roleBindingObj.Metadata.Name, projectName, err)
	}

	c.logger.LogNobl9APICall("POST", "/projects/"+projectName+"/rolebindings", true, time.Since(start), logger.Fields{
		"project_name":      projectName,
		"role_binding_name": roleBindingObj.Metadata.Name,
		"role_binding_id":   roleBindingObj.Metadata.Name,
	})

	c.logger.LogRoleBindingOperation("create", roleBindingObj.Metadata.Name, projectName, true, logger.Fields{
		"role_binding_id": roleBindingObj.Metadata.Name,
	})

	return nil
}

// UpdateRoleBinding updates an existing role binding
func (c *Client) UpdateRoleBinding(ctx context.Context, roleBindingObj *rolebinding.RoleBinding) error {
	start := time.Now()

	projectName := roleBindingObj.Spec.ProjectRef

	fn := func(ctx context.Context) (interface{}, error) {
		// Convert role binding to manifest.Object and apply
		objects := []manifest.Object{roleBindingObj}
		return nil, c.sdkClient.Objects().V1().Apply(ctx, objects)
	}

	_, err := c.retryOp.Execute(ctx, fmt.Sprintf("update role binding %s in project %s", roleBindingObj.Metadata.Name, projectName), fn)
	if err != nil {
		c.logger.LogNobl9APICall("PUT", "/projects/"+projectName+"/rolebindings/"+roleBindingObj.Metadata.Name, false, time.Since(start), logger.Fields{
			"project_name":      projectName,
			"role_binding_name": roleBindingObj.Metadata.Name,
			"error":             err.Error(),
		})
		return fmt.Errorf("failed to update role binding %s in project %s: %w", roleBindingObj.Metadata.Name, projectName, err)
	}

	c.logger.LogNobl9APICall("PUT", "/projects/"+projectName+"/rolebindings/"+roleBindingObj.Metadata.Name, true, time.Since(start), logger.Fields{
		"project_name":      projectName,
		"role_binding_name": roleBindingObj.Metadata.Name,
		"role_binding_id":   roleBindingObj.Metadata.Name,
	})

	c.logger.LogRoleBindingOperation("update", roleBindingObj.Metadata.Name, projectName, true, logger.Fields{
		"role_binding_id": roleBindingObj.Metadata.Name,
	})

	return nil
}

// DeleteRoleBinding deletes a role binding
func (c *Client) DeleteRoleBinding(ctx context.Context, projectName, name string) error {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Delete role binding by name in the project
		return nil, c.sdkClient.Objects().V1().DeleteByName(ctx, manifest.KindRoleBinding, projectName, name)
	}

	_, err := c.retryOp.Execute(ctx, fmt.Sprintf("delete role binding %s in project %s", name, projectName), fn)
	if err != nil {
		c.logger.LogNobl9APICall("DELETE", "/projects/"+projectName+"/rolebindings/"+name, false, time.Since(start), logger.Fields{
			"project_name":      projectName,
			"role_binding_name": name,
			"error":             err.Error(),
		})
		return fmt.Errorf("failed to delete role binding %s in project %s: %w", name, projectName, err)
	}

	c.logger.LogNobl9APICall("DELETE", "/projects/"+projectName+"/rolebindings/"+name, true, time.Since(start), logger.Fields{
		"project_name":      projectName,
		"role_binding_name": name,
	})

	c.logger.LogRoleBindingOperation("delete", name, projectName, true)

	return nil
}

// ListRoleBindings lists all role bindings in a project
func (c *Client) ListRoleBindings(ctx context.Context, projectName string) ([]rolebinding.RoleBinding, error) {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Get all role bindings in the project
		params := v1.GetRoleBindingsRequest{
			Project: projectName,
		}
		return c.sdkClient.Objects().V1().GetV1alphaRoleBindings(ctx, params)
	}

	result, err := c.retryOp.Execute(ctx, fmt.Sprintf("list role bindings in project %s", projectName), fn)
	if err != nil {
		c.logger.LogNobl9APICall("GET", "/projects/"+projectName+"/rolebindings", false, time.Since(start), logger.Fields{
			"project_name": projectName,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to list role bindings in project %s: %w", projectName, err)
	}

	roleBindings := result.([]rolebinding.RoleBinding)

	c.logger.LogNobl9APICall("GET", "/projects/"+projectName+"/rolebindings", true, time.Since(start), logger.Fields{
		"project_name":       projectName,
		"role_binding_count": len(roleBindings),
	})

	return roleBindings, nil
}

// GetUser retrieves a user by email
func (c *Client) GetUser(ctx context.Context, email string) (*v2.User, error) {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Get user by email using the users API
		return c.sdkClient.Users().V2().GetUser(ctx, email)
	}

	result, err := c.retryOp.Execute(ctx, fmt.Sprintf("get user %s", email), fn)
	if err != nil {
		c.logger.LogNobl9APICall("GET", "/users/"+email, false, time.Since(start), logger.Fields{
			"email": email,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get user %s: %w", email, err)
	}

	user := result.(*v2.User)

	c.logger.LogNobl9APICall("GET", "/users/"+email, true, time.Since(start), logger.Fields{
		"email":   email,
		"user_id": user.UserID,
	})

	c.logger.LogUserResolution(email, user.UserID, true, logger.Fields{
		"user_id": user.UserID,
	})

	return user, nil
}

// ListUsers lists all users (Note: This might not be available in the current SDK)
func (c *Client) ListUsers(ctx context.Context) ([]*v2.User, error) {
	start := time.Now()

	// Note: The current SDK doesn't seem to have a list users endpoint
	// This is a placeholder implementation
	c.logger.Warn("ListUsers not implemented in current SDK version", logger.Fields{
		"sdk_version": "v0.111.0",
	})

	c.logger.LogNobl9APICall("GET", "/users", true, time.Since(start), logger.Fields{
		"user_count": 0,
		"note":       "ListUsers not implemented in current SDK version",
	})

	return []*v2.User{}, nil
}

// ApplyManifest applies a Nobl9 manifest
func (c *Client) ApplyManifest(ctx context.Context, manifest []byte) error {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Decode manifest objects and apply them
		objects, err := sdk.DecodeObjects(manifest)
		if err != nil {
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}
		return nil, c.sdkClient.Objects().V1().Apply(ctx, objects)
	}

	_, err := c.retryOp.Execute(ctx, "apply manifest", fn)
	if err != nil {
		c.logger.LogNobl9APICall("POST", "/manifests", false, time.Since(start), logger.Fields{
			"manifest_size": len(manifest),
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to apply manifest: %w", err)
	}

	c.logger.LogNobl9APICall("POST", "/manifests", true, time.Since(start), logger.Fields{
		"manifest_size": len(manifest),
	})

	return nil
}

// ValidateManifest validates a Nobl9 manifest
func (c *Client) ValidateManifest(ctx context.Context, manifest []byte) error {
	start := time.Now()

	fn := func(ctx context.Context) (interface{}, error) {
		// Decode manifest objects and validate them
		objects, err := sdk.DecodeObjects(manifest)
		if err != nil {
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}

		// Validate each object
		validationErrors := make([]error, 0)
		for _, obj := range objects {
			if err := obj.Validate(); err != nil {
				validationErrors = append(validationErrors, err)
			}
		}
		if len(validationErrors) > 0 {
			return nil, fmt.Errorf("manifest validation failed: %v", validationErrors)
		}

		return nil, nil
	}

	_, err := c.retryOp.Execute(ctx, "validate manifest", fn)
	if err != nil {
		c.logger.LogNobl9APICall("POST", "/manifests/validate", false, time.Since(start), logger.Fields{
			"manifest_size": len(manifest),
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to validate manifest: %w", err)
	}

	c.logger.LogNobl9APICall("POST", "/manifests/validate", true, time.Since(start), logger.Fields{
		"manifest_size": len(manifest),
	})

	return nil
}

// Close closes the client connection
func (c *Client) Close() error {
	c.logger.Info("Closing Nobl9 client connection")
	return nil
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetSDKClient returns the underlying SDK client (for advanced usage)
func (c *Client) GetSDKClient() *sdk.Client {
	return c.sdkClient
}

// GetRetryPolicy returns the current retry policy
func (c *Client) GetRetryPolicy() *retry.Policy {
	return c.retryOp.GetPolicy()
}

// SetRetryPolicy sets a new retry policy
func (c *Client) SetRetryPolicy(policy *retry.Policy) {
	c.retryOp.SetPolicy(policy)
}
