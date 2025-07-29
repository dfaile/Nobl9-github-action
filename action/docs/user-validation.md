# User Validation and Permission System

This document describes the comprehensive user validation and permission system implemented in the Nobl9 Backstage Action. The system ensures that users exist, are active, and have the necessary permissions before creating role bindings.

## Overview

The user validation system provides multiple layers of validation to ensure the integrity and security of role binding operations:

1. **User Existence Validation** - Verifies that users exist in the Nobl9 system
2. **User Status Validation** - Checks if users are active and available
3. **Permission Validation** - Validates user permissions for specific roles
4. **Role Binding Requirements** - Ensures role bindings meet organizational requirements
5. **Conflict Detection** - Identifies existing role assignments to prevent conflicts

## Components

### Validator Package

The `validator` package provides the core validation functionality:

```go
package validator

type Validator struct {
    client   *nobl9.Client
    resolver *resolver.Resolver
    logger   *logger.Logger
}
```

### Key Types

#### UserValidation

Represents validation results for a single user:

```go
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
```

#### RoleBindingValidation

Represents validation results for a role binding:

```go
type RoleBindingValidation struct {
    Name            string
    ProjectName     string
    Role            string
    Users           []*UserValidation
    IsValid         bool
    Errors          []error
    Warnings        []string
    Requirements    *RoleBindingRequirements
    Duration        time.Duration
}
```

#### RoleBindingRequirements

Defines requirements for role bindings:

```go
type RoleBindingRequirements struct {
    MinUsers        int
    MaxUsers        int
    RequiredRoles   []string
    AllowedRoles    []string
    ProjectRequired bool
}
```

## Validation Process

### 1. Role Binding Structure Validation

Validates the basic structure of role bindings:

- **Name Validation**: Ensures role binding names follow DNS RFC1123 standards
- **Project Reference**: Validates project name format and existence
- **Role Reference**: Ensures role references are valid

```go
func (v *Validator) validateRoleBindingStructure(roleBindingObj *rolebinding.RoleBinding) error
```

### 2. Project Existence Validation

Verifies that the target project exists:

```go
func (v *Validator) validateProjectExists(ctx context.Context, projectName string) error
```

### 3. User Validation

Comprehensive user validation process:

#### Email Format Validation

```go
func (v *Validator) validateUser(ctx context.Context, user *UserValidation, emailToUserID map[string]string) error
```

Validates:
- Email format compliance
- User existence in Nobl9
- User active status
- User permissions for the target role

#### User Existence Verification

```go
func (v *Validator) verifyUserExists(ctx context.Context, user *UserValidation) error
```

#### User Status Check

```go
func (v *Validator) checkUserActive(ctx context.Context, user *UserValidation) error
```

#### Permission Validation

```go
func (v *Validator) checkUserPermissions(ctx context.Context, user *UserValidation) error
```

### 4. Role Binding Requirements Validation

Validates role binding requirements:

```go
func (v *Validator) validateRoleBindingRequirements(validation *RoleBindingValidation) error
```

Checks:
- Minimum user requirements
- Maximum user limits
- Valid user count for assignment

### 5. Conflict Detection

Identifies potential conflicts:

```go
func (v *Validator) checkRoleBindingConflicts(ctx context.Context, validation *RoleBindingValidation) error
```

```go
func (v *Validator) checkUserRoleConflict(ctx context.Context, user *UserValidation, projectName, role string) error
```

## Role-Specific Requirements

### Project Owner Role

```go
case "project-owner":
    return &RoleBindingRequirements{
        MinUsers:        1,
        MaxUsers:        10,
        RequiredRoles:   []string{"project-owner"},
        AllowedRoles:    []string{"project-owner"},
        ProjectRequired: true,
    }
```

### Project Editor Role

```go
case "project-editor":
    return &RoleBindingRequirements{
        MinUsers:        0,
        MaxUsers:        50,
        RequiredRoles:   []string{"project-editor"},
        AllowedRoles:    []string{"project-editor"},
        ProjectRequired: true,
    }
```

### Project Viewer Role

```go
case "project-viewer":
    return &RoleBindingRequirements{
        MinUsers:        0,
        MaxUsers:        100,
        RequiredRoles:   []string{"project-viewer"},
        AllowedRoles:    []string{"project-viewer"},
        ProjectRequired: true,
    }
```

## Integration with Processor

The validator is integrated into the processor workflow:

```go
// In processRoleBinding method
if roleBindingObj, ok := obj.(*rolebinding.RoleBinding); ok {
    validation, err := p.validator.ValidateRoleBinding(ctx, roleBindingObj, emailToUserID)
    if err != nil {
        return errors.NewValidationError("failed to validate role binding", err)
    }

    if !validation.IsValid {
        // Handle validation errors
        return validation.Errors[0]
    }
}
```

## Error Handling

### Validation Error Types

The system uses structured error types:

- **ValidationError**: General validation failures
- **UserResolutionError**: User lookup and resolution issues
- **ConfigError**: Configuration-related errors

### Error Aggregation

Errors are collected and aggregated for comprehensive reporting:

```go
type ValidationResult struct {
    IsValid        bool
    Errors         []error
    Warnings       []string
    ValidatedUsers []*UserValidation
    Duration       time.Duration
}
```

## Logging and Monitoring

### Structured Logging

All validation operations are logged with structured fields:

```go
p.logger.Info("Role binding validation completed", logger.Fields{
    "role_binding_name": summary["role_binding_name"],
    "project_name":      summary["project_name"],
    "role":              summary["role"],
    "is_valid":          summary["is_valid"],
    "total_users":       summary["total_users"],
    "valid_users":       summary["valid_users"],
    "invalid_users":     summary["invalid_users"],
    "error_count":       summary["error_count"],
    "warning_count":     summary["warning_count"],
    "duration":          summary["duration"],
})
```

### Validation Summary

Provides comprehensive validation statistics:

```go
func (v *Validator) GetValidationSummary(validation *RoleBindingValidation) map[string]interface{}
```

## Testing

### Unit Tests

Comprehensive unit tests cover:

- Role binding name validation
- Project name validation
- Role binding requirements
- User validation
- Conflict detection

### Mock Testing

Uses mock implementations for testing:

```go
type MockClient struct {
    mock.Mock
}

type MockResolver struct {
    mock.Mock
}
```

## Usage Examples

### Basic Validation

```go
validator := validator.New(client, resolver, log)

// Validate a role binding
validation, err := validator.ValidateRoleBinding(ctx, roleBindingObj, emailToUserID)
if err != nil {
    log.Error("Validation failed", logger.Fields{"error": err})
    return err
}

if !validation.IsValid {
    log.Error("Role binding validation failed", logger.Fields{
        "errors": validation.Errors,
        "warnings": validation.Warnings,
    })
    return validation.Errors[0]
}
```

### User Validation

```go
// Validate multiple users
emails := []string{"user1@example.com", "user2@example.com"}
validations, err := validator.ValidateUsers(ctx, emails, emailToUserID)
if err != nil {
    log.Error("User validation failed", logger.Fields{"error": err})
    return err
}

// Check validation results
for _, userValidation := range validations {
    if !userValidation.CanBeAssigned {
        log.Warn("User cannot be assigned", logger.Fields{
            "email": userValidation.Email,
            "error": userValidation.ValidationError,
        })
    }
}
```

## Best Practices

### 1. Pre-validation

Always validate role bindings before processing:

```go
// Validate before processing
validation, err := validator.ValidateRoleBinding(ctx, roleBindingObj, emailToUserID)
if err != nil || !validation.IsValid {
    return err
}
```

### 2. Error Handling

Handle validation errors gracefully:

```go
if !validation.IsValid {
    for _, err := range validation.Errors {
        log.Error("Validation error", logger.Fields{"error": err})
    }
    return validation.Errors[0]
}
```

### 3. Monitoring

Monitor validation metrics:

```go
summary := validator.GetValidationSummary(validation)
log.Info("Validation metrics", logger.Fields{
    "valid_users": summary["valid_users"],
    "invalid_users": summary["invalid_users"],
    "duration": summary["duration"],
})
```

### 4. Caching

Use caching for user lookups to improve performance:

```go
// The resolver already includes caching
result, err := resolver.ResolveEmail(ctx, email)
```

## Security Considerations

### 1. Permission Validation

Always validate user permissions before role assignment:

```go
if err := validator.checkUserPermissions(ctx, user); err != nil {
    return errors.NewValidationError("user lacks required permissions", err)
}
```

### 2. Conflict Detection

Prevent duplicate role assignments:

```go
if err := validator.checkUserRoleConflict(ctx, user, projectName, role); err != nil {
    return err
}
```

### 3. Input Validation

Validate all inputs thoroughly:

```go
if err := validator.validateRoleBindingName(name); err != nil {
    return errors.NewValidationError("invalid role binding name", err)
}
```

## Future Enhancements

### 1. Advanced Permission Models

- Role hierarchy validation
- Conditional permission checks
- Time-based permission validation

### 2. Performance Optimizations

- Batch validation operations
- Parallel user validation
- Enhanced caching strategies

### 3. Enhanced Monitoring

- Validation metrics collection
- Performance monitoring
- Alerting on validation failures

### 4. Configuration Management

- Configurable validation rules
- Environment-specific requirements
- Dynamic requirement updates 