# Nobl9 Client

This document describes the Nobl9 client wrapper that provides a high-level interface to the Nobl9 API using the official Go SDK.

## Overview

The Nobl9 client wrapper provides:

- **SDK Integration** - Seamless integration with the official Nobl9 Go SDK
- **Authentication Handling** - Automatic credential management and validation
- **Structured Logging** - Comprehensive API call logging and monitoring
- **Error Handling** - Robust error handling with detailed error messages
- **Context Support** - Full context support for timeouts and cancellation
- **Resource Management** - Project, RoleBinding, and User management

## Features

### SDK Integration

The client wraps the official [Nobl9 Go SDK](https://github.com/nobl9/nobl9-go) and provides:

- **Direct SDK Access** - Access to underlying SDK client for advanced usage
- **Simplified Interface** - High-level methods for common operations
- **Configuration Management** - Centralized configuration handling
- **Connection Testing** - Automatic connection validation on client creation

### Authentication

The client handles authentication automatically:

- **Credential Validation** - Validates client ID and secret on initialization
- **Connection Testing** - Tests connection to Nobl9 API on client creation
- **Environment Detection** - Auto-detects Nobl9 environment from credentials
- **Secure Storage** - Handles credentials securely without logging sensitive data

### Logging Integration

All API calls are automatically logged with structured information:

- **API Call Tracking** - Method, endpoint, duration, and status
- **Resource Operations** - Project and role binding operation logging
- **User Resolution** - Email to UserID resolution tracking
- **Error Context** - Detailed error information for troubleshooting

## Configuration

### Client Configuration

```go
type Config struct {
    ClientID     string        // Nobl9 API client ID
    ClientSecret string        // Nobl9 API client secret
    Environment  string        // Nobl9 environment (dev, staging, prod)
    Timeout      time.Duration // API call timeout
    RetryAttempts int          // Number of retry attempts
}
```

### Default Values

- **Timeout**: 30 seconds
- **Retry Attempts**: 3
- **Environment**: Auto-detected from client ID

### Environment Detection

The client automatically detects the Nobl9 environment from the client ID:

| Client ID Pattern | Detected Environment |
|------------------|---------------------|
| Contains "dev" | `dev` |
| Contains "staging" | `staging` |
| Contains "prod" | `prod` |
| Other | `unknown` |

## Usage

### Creating a Client

```go
import (
    "github.com/your-org/nobl9-action/pkg/nobl9"
    "github.com/your-org/nobl9-action/pkg/logger"
    "time"
)

// Create logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)

// Create client configuration
config := &nobl9.Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Environment:  "prod",
    Timeout:      30 * time.Second,
    RetryAttempts: 3,
}

// Create client
client, err := nobl9.New(config, log)
if err != nil {
    log.Fatal("Failed to create Nobl9 client", logger.Fields{"error": err})
}
defer client.Close()
```

### Organization Operations

```go
// Get organization information
ctx := context.Background()
org, err := client.GetOrganization(ctx)
if err != nil {
    log.Error("Failed to get organization", logger.Fields{"error": err})
    return
}

log.Info("Organization info", logger.Fields{
    "organization": org.Name,
    "environment":  client.GetConfig().Environment,
})
```

### Project Operations

```go
// Get a project
project, err := client.GetProject(ctx, "my-project")
if err != nil {
    log.Error("Failed to get project", logger.Fields{
        "project_name": "my-project",
        "error":        err,
    })
    return
}

// Create a new project
newProject := &sdk.Project{
    Metadata: sdk.Metadata{
        Name: "new-project",
    },
    Spec: sdk.ProjectSpec{
        DisplayName: "New Project",
        Description: "A new Nobl9 project",
    },
}

err = client.CreateProject(ctx, newProject)
if err != nil {
    log.Error("Failed to create project", logger.Fields{
        "project_name": newProject.Metadata.Name,
        "error":        err,
    })
    return
}

// Update a project
project.Spec.Description = "Updated description"
err = client.UpdateProject(ctx, project)
if err != nil {
    log.Error("Failed to update project", logger.Fields{
        "project_name": project.Metadata.Name,
        "error":        err,
    })
    return
}

// List all projects
projects, err := client.ListProjects(ctx)
if err != nil {
    log.Error("Failed to list projects", logger.Fields{"error": err})
    return
}

log.Info("Projects found", logger.Fields{
    "project_count": len(projects),
})

// Delete a project
err = client.DeleteProject(ctx, "old-project")
if err != nil {
    log.Error("Failed to delete project", logger.Fields{
        "project_name": "old-project",
        "error":        err,
    })
    return
}
```

### Role Binding Operations

```go
// Get a role binding
roleBinding, err := client.GetRoleBinding(ctx, "my-project", "my-role-binding")
if err != nil {
    log.Error("Failed to get role binding", logger.Fields{
        "project_name":     "my-project",
        "role_binding_name": "my-role-binding",
        "error":            err,
    })
    return
}

// Create a new role binding
newRoleBinding := &sdk.RoleBinding{
    Metadata: sdk.Metadata{
        Name:    "new-role-binding",
        Project: "my-project",
    },
    Spec: sdk.RoleBindingSpec{
        Users: []sdk.User{
            {ID: "user-123"},
            {ID: "user-456"},
        },
        Roles: []string{"project-owner", "project-editor"},
    },
}

err = client.CreateRoleBinding(ctx, newRoleBinding)
if err != nil {
    log.Error("Failed to create role binding", logger.Fields{
        "project_name":     newRoleBinding.Metadata.Project,
        "role_binding_name": newRoleBinding.Metadata.Name,
        "error":            err,
    })
    return
}

// Update a role binding
roleBinding.Spec.Users = append(roleBinding.Spec.Users, sdk.User{ID: "user-789"})
err = client.UpdateRoleBinding(ctx, roleBinding)
if err != nil {
    log.Error("Failed to update role binding", logger.Fields{
        "project_name":     roleBinding.Metadata.Project,
        "role_binding_name": roleBinding.Metadata.Name,
        "error":            err,
    })
    return
}

// List role bindings in a project
roleBindings, err := client.ListRoleBindings(ctx, "my-project")
if err != nil {
    log.Error("Failed to list role bindings", logger.Fields{
        "project_name": "my-project",
        "error":        err,
    })
    return
}

log.Info("Role bindings found", logger.Fields{
    "project_name":      "my-project",
    "role_binding_count": len(roleBindings),
})

// Delete a role binding
err = client.DeleteRoleBinding(ctx, "my-project", "old-role-binding")
if err != nil {
    log.Error("Failed to delete role binding", logger.Fields{
        "project_name":     "my-project",
        "role_binding_name": "old-role-binding",
        "error":            err,
    })
    return
}
```

### User Operations

```go
// Get a user by email
user, err := client.GetUser(ctx, "user@example.com")
if err != nil {
    log.Error("Failed to get user", logger.Fields{
        "email": user@example.com,
        "error": err,
    })
    return
}

log.Info("User found", logger.Fields{
    "email":   user.Email,
    "user_id": user.Metadata.ID,
})

// List all users
users, err := client.ListUsers(ctx)
if err != nil {
    log.Error("Failed to list users", logger.Fields{"error": err})
    return
}

log.Info("Users found", logger.Fields{
    "user_count": len(users),
})
```

### Manifest Operations

```go
// Apply a Nobl9 manifest
manifest := []byte(`
apiVersion: nobl9.com/v1
kind: Project
metadata:
  name: my-project
spec:
  displayName: My Project
  description: A test project
`)

err = client.ApplyManifest(ctx, manifest)
if err != nil {
    log.Error("Failed to apply manifest", logger.Fields{
        "manifest_size": len(manifest),
        "error":         err,
    })
    return
}

// Validate a manifest
err = client.ValidateManifest(ctx, manifest)
if err != nil {
    log.Error("Failed to validate manifest", logger.Fields{
        "manifest_size": len(manifest),
        "error":         err,
    })
    return
}
```

## Error Handling

### Common Errors

The client provides detailed error information for common scenarios:

#### Authentication Errors
```
Error: failed to create Nobl9 SDK client: invalid credentials
Error: failed to connect to Nobl9: authentication failed
```

#### Network Errors
```
Error: failed to get project my-project: network timeout
Error: failed to create role binding: connection refused
```

#### Resource Errors
```
Error: failed to get project non-existent: project not found
Error: failed to create project my-project: project already exists
Error: failed to get user user@example.com: user not found
```

### Error Recovery

The client includes built-in error recovery mechanisms:

```go
// Retry with exponential backoff
for attempt := 1; attempt <= client.GetConfig().RetryAttempts; attempt++ {
    err := client.CreateProject(ctx, project)
    if err == nil {
        break
    }
    
    if attempt == client.GetConfig().RetryAttempts {
        log.Error("Failed to create project after retries", logger.Fields{
            "project_name": project.Metadata.Name,
            "attempts":     attempt,
            "error":        err,
        })
        return err
    }
    
    // Wait before retry
    time.Sleep(time.Duration(attempt) * time.Second)
}
```

## Logging

### API Call Logging

All API calls are automatically logged with structured information:

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "debug",
  "message": "Nobl9 API call successful",
  "event": "nobl9_api_call",
  "method": "POST",
  "endpoint": "/projects",
  "success": true,
  "duration": "150ms",
  "duration_ms": 150,
  "project_name": "my-project",
  "project_id": "proj-123"
}
```

### Resource Operation Logging

Resource operations are logged with specific event types:

```json
{
  "timestamp": "2024-01-15T10:30:46Z",
  "level": "info",
  "message": "Project operation completed",
  "event": "project_operation",
  "operation": "create",
  "project_name": "my-project",
  "success": true,
  "project_id": "proj-123"
}
```

### User Resolution Logging

User email to UserID resolution is tracked:

```json
{
  "timestamp": "2024-01-15T10:30:47Z",
  "level": "info",
  "message": "User email resolved to UserID",
  "event": "user_resolution",
  "email": "user@example.com",
  "user_id": "okta-user-123",
  "success": true
}
```

## Context Support

### Timeout Handling

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// API calls will respect the timeout
project, err := client.GetProject(ctx, "my-project")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Error("API call timed out", logger.Fields{
            "project_name": "my-project",
            "timeout":      "30s",
        })
    }
    return err
}
```

### Cancellation Support

```go
// Create cancellable context
ctx, cancel := context.WithCancel(context.Background())

// Cancel after a delay
go func() {
    time.Sleep(5 * time.Second)
    cancel()
}()

// API calls will be cancelled
project, err := client.GetProject(ctx, "my-project")
if err != nil {
    if ctx.Err() == context.Canceled {
        log.Info("API call was cancelled")
    }
    return err
}
```

## Advanced Usage

### Direct SDK Access

For advanced usage, you can access the underlying SDK client:

```go
// Get the underlying SDK client
sdkClient := client.GetSDKClient()

// Use SDK methods directly
slo, err := sdkClient.SLOs.Get(ctx, "my-project", "my-slo")
if err != nil {
    log.Error("Failed to get SLO", logger.Fields{"error": err})
    return
}
```

### Custom Configuration

```go
// Create custom configuration
config := &nobl9.Config{
    ClientID:     "custom-client-id",
    ClientSecret: "custom-client-secret",
    Environment:  "custom-env",
    Timeout:      60 * time.Second,  // Custom timeout
    RetryAttempts: 5,                // Custom retry attempts
}

// Create client with custom config
client, err := nobl9.New(config, log)
if err != nil {
    log.Fatal("Failed to create client", logger.Fields{"error": err})
}
```

## Integration with GitHub Action

### Configuration Integration

The client integrates with the GitHub Action configuration:

```go
// Get configuration from GitHub Action
actionConfig, err := config.Load()
if err != nil {
    log.Fatal("Failed to load configuration", logger.Fields{"error": err})
}

// Create Nobl9 client configuration
nobl9Config := &nobl9.Config{
    ClientID:     actionConfig.Nobl9.ClientID,
    ClientSecret: actionConfig.Nobl9.ClientSecret,
    Environment:  actionConfig.Nobl9.Environment,
    Timeout:      30 * time.Second,
    RetryAttempts: 3,
}

// Create client
client, err := nobl9.New(nobl9Config, log)
if err != nil {
    log.Fatal("Failed to create Nobl9 client", logger.Fields{"error": err})
}
```

### Logging Integration

The client uses the same logging framework as the GitHub Action:

```go
// Create logger with action configuration
log := logger.New(
    logger.Level(actionConfig.Logging.Level),
    logger.Format(actionConfig.Logging.Format),
)

// Create client with integrated logging
client, err := nobl9.New(nobl9Config, log)
if err != nil {
    log.Fatal("Failed to create Nobl9 client", logger.Fields{"error": err})
}
```

## Best Practices

### Client Management

1. **Reuse Clients** - Create clients once and reuse them for multiple operations
2. **Proper Cleanup** - Always call `client.Close()` when done
3. **Context Usage** - Use contexts for timeouts and cancellation
4. **Error Handling** - Handle errors appropriately and log context

### Performance Optimization

1. **Connection Pooling** - The SDK handles connection pooling automatically
2. **Request Batching** - Batch operations when possible
3. **Caching** - Cache frequently accessed resources
4. **Retry Logic** - Use appropriate retry strategies for transient failures

### Security

1. **Credential Management** - Never log or expose credentials
2. **Environment Separation** - Use different credentials for different environments
3. **Access Control** - Use least privilege credentials
4. **Audit Logging** - Enable comprehensive logging for audit purposes

### Error Handling

1. **Graceful Degradation** - Handle errors gracefully and continue processing
2. **Retry Strategies** - Implement appropriate retry logic for transient failures
3. **Error Classification** - Distinguish between transient and permanent errors
4. **User Feedback** - Provide meaningful error messages to users 