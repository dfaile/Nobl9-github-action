package nobl9client

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/sirupsen/logrus"
)

// Client wraps the Nobl9 SDK client with additional functionality
type Client struct {
	sdkClient *sdk.Client
	timeout   time.Duration
}

// ProcessedObject represents a processed Nobl9 object
type ProcessedObject struct {
	Object      manifest.Object
	Kind        string
	Name        string
	Project     string
	UserEmails  []string
	ResolvedIDs map[string]string // email -> userID mapping
	Applied     bool
	Error       error
}

// ProcessResult represents the result of processing objects
type ProcessResult struct {
	Projects       []ProcessedObject
	RoleBindings   []ProcessedObject
	EmailsResolved map[string]string
	Errors         []error
	Summary        string
}

// Valid roles (from your lambda) - currently unused but kept for future validation
// var validRoles = map[string]bool{
//	"project-owner":  true,
//	"project-viewer": true,
//	"project-editor": true,
// }

// NewClient creates a new Nobl9 client
func NewClient(clientID, clientSecret string) (*Client, error) {
	// Set environment variables for the Nobl9 SDK (like your lambda)
	os.Setenv("NOBL9_SDK_CLIENT_ID", clientID)
	os.Setenv("NOBL9_SDK_CLIENT_SECRET", clientSecret)

	// Fix for environments where HOME is not set properly
	if os.Getenv("HOME") == "" {
		os.Setenv("HOME", "/tmp")
	}

	// Initialize the Nobl9 client using the same method as your lambda
	client, err := sdk.DefaultClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Nobl9 SDK client: %w", err)
	}

	return &Client{
		sdkClient: client,
		timeout:   60 * time.Second,
	}, nil
}

// ProcessObjects processes parsed objects (projects and role bindings)
func (c *Client) ProcessObjects(ctx context.Context, objects []ParsedObject, dryRun bool) (*ProcessResult, error) {
	// Create a context with timeout
	processCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	result := &ProcessResult{
		Projects:       make([]ProcessedObject, 0),
		RoleBindings:   make([]ProcessedObject, 0),
		EmailsResolved: make(map[string]string),
		Errors:         make([]error, 0),
	}

	logrus.WithFields(logrus.Fields{
		"total_objects": len(objects),
		"dry_run":       dryRun,
	}).Info("Starting Nobl9 object processing")

	// Step 1: Collect all emails that need resolution
	allEmails := make(map[string]bool)
	for _, obj := range objects {
		for _, email := range obj.UserEmails {
			allEmails[email] = true
		}
	}

	// Step 2: Resolve all emails to user IDs
	if len(allEmails) > 0 {
		logrus.WithField("email_count", len(allEmails)).Info("Resolving email addresses to user IDs")

		for email := range allEmails {
			userID, err := c.resolveEmailToUserID(processCtx, email)
			if err != nil {
				logrus.WithError(err).WithField("email", email).Error("Failed to resolve email")
				result.Errors = append(result.Errors, fmt.Errorf("failed to resolve email '%s': %w", email, err))
				continue
			}
			result.EmailsResolved[email] = userID
			logrus.WithFields(logrus.Fields{
				"email":   email,
				"user_id": userID,
			}).Debug("Email resolved to user ID")
		}
	}

	// Step 3: Process projects first
	var projectObjects []ParsedObject
	var roleBindingObjects []ParsedObject

	for _, obj := range objects {
		if obj.Kind == "Project" {
			projectObjects = append(projectObjects, obj)
		} else if obj.Kind == "RoleBinding" {
			roleBindingObjects = append(roleBindingObjects, obj)
		}
	}

	// Process projects
	for _, obj := range projectObjects {
		processed := c.processProject(processCtx, obj, dryRun)
		result.Projects = append(result.Projects, processed)
		if processed.Error != nil {
			result.Errors = append(result.Errors, processed.Error)
		}
	}

	// Process role bindings (with resolved emails)
	for _, obj := range roleBindingObjects {
		processed := c.processRoleBinding(processCtx, obj, result.EmailsResolved, dryRun)
		result.RoleBindings = append(result.RoleBindings, processed)
		if processed.Error != nil {
			result.Errors = append(result.Errors, processed.Error)
		}
	}

	// Generate summary
	result.Summary = c.generateSummary(result)

	logrus.WithFields(logrus.Fields{
		"projects_processed":      len(result.Projects),
		"role_bindings_processed": len(result.RoleBindings),
		"emails_resolved":         len(result.EmailsResolved),
		"errors":                  len(result.Errors),
		"dry_run":                 dryRun,
	}).Info("Nobl9 object processing completed")

	return result, nil
}

// resolveEmailToUserID resolves an email address to a user ID using Nobl9 API
func (c *Client) resolveEmailToUserID(ctx context.Context, email string) (string, error) {
	logrus.WithField("email", email).Debug("Resolving email to user ID")

	// Use Nobl9 SDK to get user by email (same as your lambda)
	user, err := c.sdkClient.Users().V2().GetUser(ctx, email)
	if err != nil {
		return "", fmt.Errorf("error retrieving user '%s': %w", email, err)
	}
	if user == nil {
		return "", fmt.Errorf("user with email '%s' not found in Nobl9", email)
	}

	return user.UserID, nil
}

// processProject processes a single project
func (c *Client) processProject(ctx context.Context, obj ParsedObject, dryRun bool) ProcessedObject {
	processed := ProcessedObject{
		Object:      obj.Object,
		Kind:        obj.Kind,
		Name:        obj.Name,
		Project:     obj.Project,
		UserEmails:  obj.UserEmails,
		ResolvedIDs: make(map[string]string),
		Applied:     false,
	}

	logrus.WithFields(logrus.Fields{
		"project_name": obj.Name,
		"dry_run":      dryRun,
	}).Info("Processing project")

	if dryRun {
		logrus.WithField("project_name", obj.Name).Info("DRY RUN: Would create project")
		processed.Applied = true
		return processed
	}

	// Apply the project using the SDK
	if err := c.sdkClient.Objects().V1().Apply(ctx, []manifest.Object{obj.Object}); err != nil {
		// Check if the error is because the project already exists
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "conflict") {
			logrus.WithField("project_name", obj.Name).Info("Project already exists")
			processed.Applied = true
			return processed
		}

		processed.Error = fmt.Errorf("failed to create project '%s': %w", obj.Name, err)
		logrus.WithError(processed.Error).Error("Failed to create project")
		return processed
	}

	processed.Applied = true
	logrus.WithField("project_name", obj.Name).Info("Project created successfully")
	return processed
}

// processRoleBinding processes a single role binding
func (c *Client) processRoleBinding(ctx context.Context, obj ParsedObject, emailResolution map[string]string, dryRun bool) ProcessedObject {
	processed := ProcessedObject{
		Object:      obj.Object,
		Kind:        obj.Kind,
		Name:        obj.Name,
		Project:     obj.Project,
		UserEmails:  obj.UserEmails,
		ResolvedIDs: make(map[string]string),
		Applied:     false,
	}

	logrus.WithFields(logrus.Fields{
		"role_binding_name": obj.Name,
		"project":           obj.Project,
		"dry_run":           dryRun,
	}).Info("Processing role binding")

	// Update user ID if it was an email that got resolved
	roleBinding, ok := obj.Object.(v1alphaRoleBinding.RoleBinding)
	if !ok {
		processed.Error = fmt.Errorf("object is not a RoleBinding")
		return processed
	}

	// Get the user from the role binding spec
	if roleBinding.Spec.User != nil {
		originalUser := *roleBinding.Spec.User

		// Check if this user is an email that was resolved
		if resolvedUserID, found := emailResolution[originalUser]; found {
			// Update the role binding with the resolved user ID
			roleBinding.Spec.User = &resolvedUserID
			processed.ResolvedIDs[originalUser] = resolvedUserID

			logrus.WithFields(logrus.Fields{
				"original_email": originalUser,
				"resolved_id":    resolvedUserID,
				"role_binding":   obj.Name,
			}).Info("Email resolved for role binding")
		}
	}

	if dryRun {
		logrus.WithFields(logrus.Fields{
			"role_binding_name": obj.Name,
			"project":           obj.Project,
		}).Info("DRY RUN: Would create role binding")
		processed.Applied = true
		return processed
	}

	// Apply the role binding using the SDK
	if err := c.sdkClient.Objects().V1().Apply(ctx, []manifest.Object{roleBinding}); err != nil {
		processed.Error = fmt.Errorf("failed to create role binding '%s': %w", obj.Name, err)
		logrus.WithError(processed.Error).Error("Failed to create role binding")
		return processed
	}

	processed.Applied = true
	logrus.WithFields(logrus.Fields{
		"role_binding_name": obj.Name,
		"project":           obj.Project,
	}).Info("Role binding created successfully")

	return processed
}

// generateSummary generates a summary of the processing results
func (c *Client) generateSummary(result *ProcessResult) string {
	successfulProjects := 0
	successfulRoleBindings := 0

	for _, proj := range result.Projects {
		if proj.Applied && proj.Error == nil {
			successfulProjects++
		}
	}

	for _, rb := range result.RoleBindings {
		if rb.Applied && rb.Error == nil {
			successfulRoleBindings++
		}
	}

	return fmt.Sprintf("Processing completed: %d projects, %d role bindings, %d emails resolved, %d errors",
		successfulProjects, successfulRoleBindings, len(result.EmailsResolved), len(result.Errors))
}

// ParsedObject represents a parsed object that needs processing
// This should match the structure from the parser
type ParsedObject struct {
	Object     manifest.Object
	Kind       string
	Name       string
	Project    string
	UserEmails []string
}

// Helper functions from your lambda

// ptr creates a pointer to a string
func ptr(s string) *string {
	return &s
}

// sanitizeName ensures the string is RFC-1123 compliant (from your lambda)
func sanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace non-alphanumeric characters (except hyphen) with a hyphen
	reg := regexp.MustCompile("[^a-z0-9-]+")
	name = reg.ReplaceAllString(name, "-")
	// Trim hyphens from the start and end
	name = strings.Trim(name, "-")
	return name
}

// truncate shortens a string to a max length (from your lambda)
func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
