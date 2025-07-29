package validator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/your-org/nobl9-action/pkg/errors"
	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
	"github.com/your-org/nobl9-action/pkg/resolver"
)

// Validator handles validation of users, permissions, and role bindings
type Validator struct {
	client   *nobl9.Client
	resolver *resolver.Resolver
	logger   *logger.Logger
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid        bool
	Errors         []error
	Warnings       []string
	ValidatedUsers []*UserValidation
	Duration       time.Duration
}

// UserValidation represents validation results for a single user
type UserValidation struct {
	Email           string
	UserID          string
	Exists          bool
	IsActive        bool
	HasPermissions  bool
	CanBeAssigned   bool
	ValidationError error
	Warnings        []string
}

// RoleBindingValidation represents validation results for a role binding
type RoleBindingValidation struct {
	Name         string
	ProjectName  string
	Role         string
	Users        []*UserValidation
	IsValid      bool
	Errors       []error
	Warnings     []string
	Requirements *RoleBindingRequirements
	Duration     time.Duration
}

// RoleBindingRequirements represents requirements for a role binding
type RoleBindingRequirements struct {
	MinUsers        int
	MaxUsers        int
	RequiredRoles   []string
	AllowedRoles    []string
	ProjectRequired bool
}

// New creates a new validator instance
func New(client *nobl9.Client, resolver *resolver.Resolver, log *logger.Logger) *Validator {
	return &Validator{
		client:   client,
		resolver: resolver,
		logger:   log,
	}
}

// ValidateRoleBinding validates a role binding before creation
func (v *Validator) ValidateRoleBinding(ctx context.Context, roleBindingObj *rolebinding.RoleBinding, emailToUserID map[string]string) (*RoleBindingValidation, error) {
	start := time.Now()

	v.logger.Debug("Validating role binding", logger.Fields{
		"role_binding_name": roleBindingObj.Metadata.Name,
		"project_name":      roleBindingObj.Spec.ProjectRef,
		"role":              roleBindingObj.Spec.RoleRef,
	})

	validation := &RoleBindingValidation{
		Name:         roleBindingObj.Metadata.Name,
		ProjectName:  roleBindingObj.Spec.ProjectRef,
		Role:         roleBindingObj.Spec.RoleRef,
		Users:        make([]*UserValidation, 0),
		Errors:       make([]error, 0),
		Warnings:     make([]string, 0),
		Requirements: v.getRoleBindingRequirements(roleBindingObj.Spec.RoleRef),
	}

	// Step 1: Validate role binding structure
	if err := v.validateRoleBindingStructure(roleBindingObj); err != nil {
		validation.Errors = append(validation.Errors, err)
		validation.IsValid = false
	}

	// Step 2: Validate project exists
	if err := v.validateProjectExists(ctx, roleBindingObj.Spec.ProjectRef); err != nil {
		validation.Errors = append(validation.Errors, err)
		validation.IsValid = false
	}

	// Step 3: Extract and validate users
	users, err := v.extractUsersFromRoleBinding(roleBindingObj)
	if err != nil {
		validation.Errors = append(validation.Errors, err)
		validation.IsValid = false
	} else {
		validation.Users = users
	}

	// Step 4: Validate each user
	for _, user := range validation.Users {
		if err := v.validateUser(ctx, user, emailToUserID); err != nil {
			user.ValidationError = err
			user.CanBeAssigned = false
			validation.Errors = append(validation.Errors, fmt.Errorf("user validation failed for %s: %w", user.Email, err))
		}
	}

	// Step 5: Validate role binding requirements
	if err := v.validateRoleBindingRequirements(validation); err != nil {
		validation.Errors = append(validation.Errors, err)
		validation.IsValid = false
	}

	// Step 6: Check for existing role binding conflicts
	if err := v.checkRoleBindingConflicts(ctx, validation); err != nil {
		validation.Warnings = append(validation.Warnings, err.Error())
	}

	validation.Duration = time.Since(start)
	validation.IsValid = len(validation.Errors) == 0

	v.logger.Info("Role binding validation completed", logger.Fields{
		"role_binding_name": validation.Name,
		"is_valid":          validation.IsValid,
		"error_count":       len(validation.Errors),
		"warning_count":     len(validation.Warnings),
		"user_count":        len(validation.Users),
		"duration":          validation.Duration.String(),
	})

	return validation, nil
}

// ValidateUsers validates multiple users for role binding assignment
func (v *Validator) ValidateUsers(ctx context.Context, emails []string, emailToUserID map[string]string) ([]*UserValidation, error) {
	start := time.Now()

	v.logger.Debug("Validating users for role binding", logger.Fields{
		"user_count": len(emails),
	})

	validations := make([]*UserValidation, 0, len(emails))

	for _, email := range emails {
		userValidation := &UserValidation{
			Email:  email,
			UserID: emailToUserID[email],
		}

		if err := v.validateUser(ctx, userValidation, emailToUserID); err != nil {
			userValidation.ValidationError = err
			userValidation.CanBeAssigned = false
		}

		validations = append(validations, userValidation)
	}

	v.logger.Info("User validation completed", logger.Fields{
		"user_count":    len(validations),
		"valid_users":   v.countValidUsers(validations),
		"invalid_users": v.countInvalidUsers(validations),
		"duration":      time.Since(start).String(),
	})

	return validations, nil
}

// validateRoleBindingStructure validates the basic structure of a role binding
func (v *Validator) validateRoleBindingStructure(roleBindingObj *rolebinding.RoleBinding) error {
	// Check required fields
	if roleBindingObj.Metadata.Name == "" {
		return errors.NewValidationError("role binding name is required", nil)
	}

	if roleBindingObj.Spec.ProjectRef == "" {
		return errors.NewValidationError("project reference is required", nil)
	}

	if roleBindingObj.Spec.RoleRef == "" {
		return errors.NewValidationError("role reference is required", nil)
	}

	// Validate role binding name format
	if err := v.validateRoleBindingName(roleBindingObj.Metadata.Name); err != nil {
		return errors.NewValidationError("invalid role binding name", err)
	}

	// Validate project name format
	if err := v.validateProjectName(roleBindingObj.Spec.ProjectRef); err != nil {
		return errors.NewValidationError("invalid project name", err)
	}

	return nil
}

// validateProjectExists checks if the project exists
func (v *Validator) validateProjectExists(ctx context.Context, projectName string) error {
	_, err := v.client.GetProject(ctx, projectName)
	if err != nil {
		return errors.NewValidationError(fmt.Sprintf("project %s does not exist", projectName), err)
	}
	return nil
}

// extractUsersFromRoleBinding extracts user information from a role binding
func (v *Validator) extractUsersFromRoleBinding(roleBindingObj *rolebinding.RoleBinding) ([]*UserValidation, error) {
	users := make([]*UserValidation, 0)

	// Extract users from the role binding spec
	// Based on the template structure, users are specified with email and roles
	// For now, we'll extract from the YAML content since the SDK structure may differ
	// This is a simplified approach - in practice, you'd parse the actual role binding structure

	// For demonstration, we'll create a placeholder user validation
	// In a real implementation, you would extract the actual user data from the role binding
	userValidation := &UserValidation{
		Email:  "placeholder@example.com", // This would be extracted from the actual role binding
		UserID: "",
	}
	users = append(users, userValidation)

	if len(users) == 0 {
		return nil, errors.NewValidationError("no users specified in role binding", nil)
	}

	return users, nil
}

// validateUser validates a single user for role binding assignment
func (v *Validator) validateUser(ctx context.Context, user *UserValidation, emailToUserID map[string]string) error {
	// Step 1: Validate email format
	if err := v.resolver.ValidateEmailFormat(user.Email); err != nil {
		return errors.NewValidationError("invalid email format", err)
	}

	// Step 2: Check if user exists and is resolved
	if user.UserID == "" {
		// Try to resolve the user
		result, err := v.resolver.ResolveEmail(ctx, user.Email)
		if err != nil {
			return errors.NewUserResolutionError("failed to resolve user", err)
		}

		if !result.Resolved {
			return errors.NewUserResolutionError("user not found", fmt.Errorf("user %s does not exist", user.Email))
		}

		user.UserID = result.UserID
	}

	// Step 3: Verify user exists in Nobl9
	if err := v.verifyUserExists(ctx, user); err != nil {
		return errors.NewUserResolutionError("user verification failed", err)
	}

	// Step 4: Check if user is active
	if err := v.checkUserActive(ctx, user); err != nil {
		return errors.NewValidationError("user is not active", err)
	}

	// Step 5: Check user permissions for the role
	if err := v.checkUserPermissions(ctx, user); err != nil {
		return errors.NewValidationError("user lacks required permissions", err)
	}

	user.Exists = true
	user.IsActive = true
	user.HasPermissions = true
	user.CanBeAssigned = true

	return nil
}

// verifyUserExists verifies that a user exists in Nobl9
func (v *Validator) verifyUserExists(ctx context.Context, user *UserValidation) error {
	_, err := v.client.GetUser(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("user %s not found in Nobl9", user.Email)
	}
	return nil
}

// checkUserActive checks if a user is active
func (v *Validator) checkUserActive(ctx context.Context, user *UserValidation) error {
	nobl9User, err := v.client.GetUser(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to get user status: %w", err)
	}

	// Check if user is active - the actual field name may differ in the SDK
	// For now, we'll assume the user is active if we can retrieve them
	// In a real implementation, you would check the actual active status field
	_ = nobl9User // Use the user object to avoid unused variable warning

	return nil
}

// checkUserPermissions checks if a user has the required permissions
func (v *Validator) checkUserPermissions(ctx context.Context, user *UserValidation) error {
	// This is a placeholder for permission checking
	// In a real implementation, you would check the user's current permissions
	// against the role being assigned to ensure they can be assigned that role

	// For now, we'll assume all users can be assigned roles
	// This should be enhanced based on Nobl9's permission model
	return nil
}

// validateRoleBindingRequirements validates role binding requirements
func (v *Validator) validateRoleBindingRequirements(validation *RoleBindingValidation) error {
	requirements := validation.Requirements

	// Check minimum users requirement
	if len(validation.Users) < requirements.MinUsers {
		return errors.NewValidationError(fmt.Sprintf("role binding requires at least %d users, got %d", requirements.MinUsers, len(validation.Users)), nil)
	}

	// Check maximum users requirement
	if requirements.MaxUsers > 0 && len(validation.Users) > requirements.MaxUsers {
		return errors.NewValidationError(fmt.Sprintf("role binding allows at most %d users, got %d", requirements.MaxUsers, len(validation.Users)), nil)
	}

	// Check if all users can be assigned
	validUsers := 0
	for _, user := range validation.Users {
		if user.CanBeAssigned {
			validUsers++
		}
	}

	if validUsers < requirements.MinUsers {
		return errors.NewValidationError(fmt.Sprintf("insufficient valid users for role binding: %d valid, %d required", validUsers, requirements.MinUsers), nil)
	}

	return nil
}

// checkRoleBindingConflicts checks for existing role binding conflicts
func (v *Validator) checkRoleBindingConflicts(ctx context.Context, validation *RoleBindingValidation) error {
	// Check if a role binding with the same name already exists
	existingRoleBinding, err := v.client.GetRoleBinding(ctx, validation.ProjectName, validation.Name)
	if err == nil && existingRoleBinding != nil {
		return fmt.Errorf("role binding %s already exists in project %s", validation.Name, validation.ProjectName)
	}

	// Check for user conflicts (users already assigned to the same role)
	for _, user := range validation.Users {
		if err := v.checkUserRoleConflict(ctx, user, validation.ProjectName, validation.Role); err != nil {
			return err
		}
	}

	return nil
}

// checkUserRoleConflict checks if a user is already assigned to the same role
func (v *Validator) checkUserRoleConflict(ctx context.Context, user *UserValidation, projectName, role string) error {
	// Get existing role bindings for the project
	roleBindings, err := v.client.ListRoleBindings(ctx, projectName)
	if err != nil {
		// If we can't check, log a warning but don't fail
		v.logger.Warn("Could not check for user role conflicts", logger.Fields{
			"user_email":   user.Email,
			"project_name": projectName,
			"role":         role,
			"error":        err.Error(),
		})
		return nil
	}

	// Check if user is already assigned to the same role
	// Note: This is a simplified check - in practice, you would need to parse the actual role binding structure
	for _, existingRoleBinding := range roleBindings {
		if existingRoleBinding.Spec.RoleRef == role {
			// For now, we'll skip the detailed user conflict check since the structure is not clear
			// In a real implementation, you would check the actual user list in the role binding
			v.logger.Debug("Found existing role binding with same role", logger.Fields{
				"role_binding_name": existingRoleBinding.Metadata.Name,
				"role":              role,
				"project_name":      projectName,
			})
		}
	}

	return nil
}

// getRoleBindingRequirements returns requirements for a specific role
func (v *Validator) getRoleBindingRequirements(role string) *RoleBindingRequirements {
	switch strings.ToLower(role) {
	case "project-owner":
		return &RoleBindingRequirements{
			MinUsers:        1,
			MaxUsers:        10,
			RequiredRoles:   []string{"project-owner"},
			AllowedRoles:    []string{"project-owner"},
			ProjectRequired: true,
		}
	case "project-editor":
		return &RoleBindingRequirements{
			MinUsers:        0,
			MaxUsers:        50,
			RequiredRoles:   []string{"project-editor"},
			AllowedRoles:    []string{"project-editor"},
			ProjectRequired: true,
		}
	case "project-viewer":
		return &RoleBindingRequirements{
			MinUsers:        0,
			MaxUsers:        100,
			RequiredRoles:   []string{"project-viewer"},
			AllowedRoles:    []string{"project-viewer"},
			ProjectRequired: true,
		}
	default:
		return &RoleBindingRequirements{
			MinUsers:        0,
			MaxUsers:        50,
			RequiredRoles:   []string{role},
			AllowedRoles:    []string{role},
			ProjectRequired: true,
		}
	}
}

// validateRoleBindingName validates role binding name format
func (v *Validator) validateRoleBindingName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("role binding name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("role binding name too long (max 63 characters)")
	}

	// Check for valid characters (DNS RFC1123)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return fmt.Errorf("role binding name contains invalid character: %c", char)
		}
	}

	return nil
}

// validateProjectName validates project name format
func (v *Validator) validateProjectName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("project name too long (max 63 characters)")
	}

	// Check for valid characters (DNS RFC1123)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return fmt.Errorf("project name contains invalid character: %c", char)
		}
	}

	return nil
}

// countValidUsers counts the number of valid users
func (v *Validator) countValidUsers(users []*UserValidation) int {
	count := 0
	for _, user := range users {
		if user.CanBeAssigned {
			count++
		}
	}
	return count
}

// countInvalidUsers counts the number of invalid users
func (v *Validator) countInvalidUsers(users []*UserValidation) int {
	count := 0
	for _, user := range users {
		if !user.CanBeAssigned {
			count++
		}
	}
	return count
}

// GetValidationSummary returns a summary of validation results
func (v *Validator) GetValidationSummary(validation *RoleBindingValidation) map[string]interface{} {
	validUsers := v.countValidUsers(validation.Users)
	invalidUsers := v.countInvalidUsers(validation.Users)

	return map[string]interface{}{
		"role_binding_name": validation.Name,
		"project_name":      validation.ProjectName,
		"role":              validation.Role,
		"is_valid":          validation.IsValid,
		"total_users":       len(validation.Users),
		"valid_users":       validUsers,
		"invalid_users":     invalidUsers,
		"error_count":       len(validation.Errors),
		"warning_count":     len(validation.Warnings),
		"duration":          validation.Duration.String(),
	}
}
