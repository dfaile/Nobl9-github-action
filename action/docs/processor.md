# Processor

This document describes the processor that handles project and role binding creation/update logic using the Nobl9 SDK's Apply method, integrating with the parser and resolver.

## Overview

The processor provides:

- **File Processing** - Process multiple Nobl9 configuration files
- **Object Processing** - Handle Project and RoleBinding objects
- **Manifest Application** - Apply manifests using the Nobl9 SDK
- **Dry-Run Mode** - Validate and simulate processing without making changes
- **Result Aggregation** - Collect and aggregate processing results
- **Error Handling** - Comprehensive error handling and recovery

## Features

### File Processing

The processor handles multiple file processing scenarios:

- **Batch Processing** - Process multiple files efficiently
- **Individual Processing** - Process single files with detailed results
- **Error Recovery** - Continue processing despite individual file failures
- **Result Aggregation** - Aggregate results across multiple files

### Object Processing

Specialized processing for different Nobl9 object types:

- **Project Objects** - Handle project creation and updates
- **RoleBinding Objects** - Handle role binding creation and updates
- **Email Resolution** - Integrate with email resolver for UserID conversion
- **Object Validation** - Validate objects before processing

### Manifest Application

Direct integration with Nobl9 SDK:

- **SDK Integration** - Use the official Nobl9 SDK for manifest application
- **Apply Method** - Use the SDK's Apply method for creating/updating resources
- **Validation** - Validate manifests before application
- **Error Handling** - Handle application errors gracefully

### Dry-Run Mode

Safe validation and simulation:

- **Validation Only** - Validate manifests without making changes
- **Simulation** - Simulate processing operations
- **Error Detection** - Detect potential issues before actual processing
- **Safe Testing** - Test configurations safely

## Usage

### Basic File Processing

```go
import (
    "context"
    "github.com/your-org/nobl9-action/pkg/processor"
    "github.com/your-org/nobl9-action/pkg/nobl9"
    "github.com/your-org/nobl9-action/pkg/logger"
)

// Create processor with Nobl9 client and logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)
client := nobl9.New(config, log)
processor := processor.New(client, log)

// Process multiple files
ctx := context.Background()
result, err := processor.ProcessFiles(ctx, fileInfos)
if err != nil {
    log.Error("Failed to process files", logger.Fields{"error": err})
    return
}

// Check processing results
if result.IsSuccess {
    log.Info("Processing completed successfully", logger.Fields{
        "files_processed": result.FilesProcessed,
        "projects_created": result.ProjectsCreated,
        "role_bindings_created": result.RoleBindingsCreated,
    })
} else {
    log.Error("Processing completed with errors", logger.Fields{
        "files_with_errors": result.FilesWithErrors,
        "errors": len(result.Errors),
    })
}
```

### Single File Processing

```go
// Process a single file
fileInfo := &scanner.FileInfo{
    Path:    "project.yaml",
    Size:    100,
    IsYAML:  true,
    IsNobl9: true,
    Content: []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
spec:
  displayName: Test Project`),
}

fileResult, err := processor.ProcessFile(ctx, fileInfo)
if err != nil {
    log.Error("Failed to process file", logger.Fields{"error": err})
    return
}

log.Info("File processing completed", logger.Fields{
    "file_path": fileResult.FileInfo.Path,
    "is_success": fileResult.IsSuccess,
    "users_resolved": fileResult.UsersResolved,
    "users_unresolved": fileResult.UsersUnresolved,
})
```

### Dry-Run Processing

```go
// Process files in dry-run mode
result, err := processor.ProcessWithDryRun(ctx, fileInfos)
if err != nil {
    log.Error("Failed to process files in dry-run mode", logger.Fields{"error": err})
    return
}

log.Info("Dry-run processing completed", logger.Fields{
    "files_processed": result.FilesProcessed,
    "is_success": result.IsSuccess,
    "warnings": len(result.Warnings),
})

// Check for potential issues
if len(result.Warnings) > 0 {
    log.Warn("Processing warnings detected", logger.Fields{
        "warnings": result.Warnings,
    })
}
```

### Single File Dry-Run

```go
// Process single file in dry-run mode
fileResult, err := processor.ProcessFileWithDryRun(ctx, fileInfo)
if err != nil {
    log.Error("Failed to process file in dry-run mode", logger.Fields{"error": err})
    return
}

log.Info("Dry-run file processing completed", logger.Fields{
    "file_path": fileResult.FileInfo.Path,
    "is_success": fileResult.IsSuccess,
    "duration": fileResult.Duration.String(),
})
```

### Result Analysis

```go
// Get processing statistics
stats := processor.GetProcessingStats(result)
log.Info("Processing statistics", logger.Fields{
    "files_processed": stats["files_processed"],
    "projects_created": stats["projects_created"],
    "role_bindings_created": stats["role_bindings_created"],
    "users_resolved": stats["users_resolved"],
    "duration": stats["duration"],
})

// Get processing errors
errors := processor.GetProcessingErrors(result)
if len(errors) > 0 {
    log.Error("Processing errors", logger.Fields{
        "errors": errors,
    })
}

// Get unresolved emails
unresolved := processor.GetUnresolvedEmails(result)
if len(unresolved) > 0 {
    log.Warn("Unresolved emails", logger.Fields{
        "emails": unresolved,
    })
}
```

## Data Structures

### ProcessingResult

```go
type ProcessingResult struct {
    FilesProcessed      int           // Number of files processed
    FilesSkipped        int           // Number of files skipped
    FilesWithErrors     int           // Number of files with errors
    ProjectsCreated     int           // Number of projects created
    ProjectsUpdated     int           // Number of projects updated
    RoleBindingsCreated int           // Number of role bindings created
    RoleBindingsUpdated int           // Number of role bindings updated
    UsersResolved       int           // Number of users resolved
    UsersUnresolved     int           // Number of users unresolved
    Errors              []error       // Collection of all errors
    Warnings            []string      // Collection of all warnings
    Duration            time.Duration // Total processing duration
    IsSuccess           bool          // Overall success status
}
```

### FileProcessingResult

```go
type FileProcessingResult struct {
    FileInfo            *scanner.FileInfo              // Original file information
    ParseResult         *parser.ParseResult            // Parsing results
    ResolutionResult    *resolver.BatchResolutionResult // Email resolution results
    ProjectsCreated     int                            // Projects created in this file
    ProjectsUpdated     int                            // Projects updated in this file
    RoleBindingsCreated int                            // Role bindings created in this file
    RoleBindingsUpdated int                            // Role bindings updated in this file
    UsersResolved       int                            // Users resolved in this file
    UsersUnresolved     int                            // Users unresolved in this file
    Errors              []error                        // Errors for this file
    Warnings            []string                       // Warnings for this file
    Duration            time.Duration                  // Processing duration for this file
    IsSuccess           bool                           // Success status for this file
}
```

## Processing Workflow

### Standard Processing

1. **File Parsing** - Parse YAML content into Nobl9 objects
2. **Email Resolution** - Resolve email addresses to UserIDs
3. **Object Processing** - Process individual objects (Project, RoleBinding)
4. **Manifest Application** - Apply manifests using the Nobl9 SDK
5. **Result Aggregation** - Collect and aggregate results

### Dry-Run Processing

1. **File Parsing** - Parse YAML content into Nobl9 objects
2. **Email Resolution** - Resolve email addresses to UserIDs
3. **Manifest Validation** - Validate manifests without applying
4. **Simulation** - Simulate object processing operations
5. **Result Aggregation** - Collect and aggregate results

### Object Processing

#### Project Objects

```go
func (p *Processor) processProject(ctx context.Context, obj manifest.Object) error {
    name := obj.GetName()
    
    // Check if project exists
    existingProject, err := p.client.GetProject(ctx, name)
    if err != nil {
        // Project doesn't exist, will be created
        p.logger.Info("Project will be created", logger.Fields{
            "project_name": name,
        })
        return nil
    }
    
    // Project exists, will be updated
    p.logger.Info("Project will be updated", logger.Fields{
        "project_name": name,
        "project_id":   existingProject.Metadata.ID,
    })
    
    return nil
}
```

#### RoleBinding Objects

```go
func (p *Processor) processRoleBinding(ctx context.Context, obj manifest.Object, emailToUserID map[string]string) error {
    name := obj.GetName()
    
    // For role bindings, we need to ensure all users are resolved
    // This is handled by the manifest application process
    // The SDK will handle the conversion of emails to UserIDs during application
    
    p.logger.Info("Role binding will be processed", logger.Fields{
        "role_binding_name": name,
        "resolved_users":    len(emailToUserID),
    })
    
    return nil
}
```

## Manifest Application

### Apply Manifest

```go
func (p *Processor) applyManifest(ctx context.Context, content []byte) error {
    p.logger.Debug("Applying Nobl9 manifest", logger.Fields{
        "content_size": len(content),
    })
    
    // Use the parser to apply the manifest
    err := p.parser.ApplyManifest(ctx, content)
    if err != nil {
        return fmt.Errorf("failed to apply manifest: %w", err)
    }
    
    p.logger.Info("Manifest applied successfully")
    
    return nil
}
```

### Validate Manifest

```go
func (p *Processor) validateManifest(ctx context.Context, content []byte) error {
    p.logger.Debug("Validating Nobl9 manifest", logger.Fields{
        "content_size": len(content),
    })
    
    // Use the parser to validate the manifest
    err := p.parser.ValidateManifest(ctx, content)
    if err != nil {
        return fmt.Errorf("manifest validation failed: %w", err)
    }
    
    p.logger.Info("Manifest validation successful")
    
    return nil
}
```

## Error Handling

### Common Errors

#### File Processing Errors
```
Error: failed to process file test.yaml: failed to parse file
Error: failed to process file test.yaml: failed to resolve emails
Error: failed to process file test.yaml: failed to apply manifest
```

#### Object Processing Errors
```
Error: failed to process object test-project: project validation failed
Error: failed to process object test-role-binding: role binding validation failed
```

#### Manifest Application Errors
```
Error: failed to apply manifest: API authentication failed
Error: failed to apply manifest: resource already exists
Error: failed to apply manifest: invalid manifest format
```

### Error Recovery

The processor implements graceful error recovery:

```go
// Process files with error recovery
for _, fileInfo := range files {
    fileResult, err := p.ProcessFile(ctx, fileInfo)
    if err != nil {
        // Log error but continue processing other files
        result.Errors = append(result.Errors, fmt.Errorf("failed to process file %s: %w", fileInfo.Path, err))
        result.FilesWithErrors++
        result.IsSuccess = false
        continue
    }
    
    // Aggregate results even if some files failed
    result.FilesProcessed++
    result.ProjectsCreated += fileResult.ProjectsCreated
    result.RoleBindingsCreated += fileResult.RoleBindingsCreated
}
```

## Logging

### Processing Logging

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "info",
  "message": "Starting file processing",
  "file_count": 3
}
```

### File Processing Logging

```json
{
  "timestamp": "2024-01-15T10:30:46Z",
  "level": "info",
  "message": "Processing Nobl9 configuration file",
  "file_path": "/path/to/project.yaml",
  "file_size": 1024
}
```

### Object Processing Logging

```json
{
  "timestamp": "2024-01-15T10:30:47Z",
  "level": "debug",
  "message": "Processing Nobl9 object",
  "kind": "Project",
  "name": "test-project"
}
```

### Result Logging

```json
{
  "timestamp": "2024-01-15T10:30:48Z",
  "level": "info",
  "message": "File processing completed",
  "files_processed": 3,
  "files_skipped": 0,
  "files_with_errors": 1,
  "projects_created": 2,
  "projects_updated": 0,
  "role_bindings_created": 3,
  "role_bindings_updated": 0,
  "users_resolved": 5,
  "users_unresolved": 1,
  "errors": 2,
  "warnings": 0,
  "duration": "2.5s",
  "is_success": false
}
```

## Integration with GitHub Action

### Configuration Integration

```go
// Get configuration from GitHub Action
actionConfig, err := config.Load()
if err != nil {
    log.Fatal("Failed to load configuration", logger.Fields{"error": err})
}

// Create Nobl9 client
nobl9Config := &nobl9.Config{
    ClientID:     actionConfig.Nobl9.ClientID,
    ClientSecret: actionConfig.Nobl9.ClientSecret,
    Environment:  actionConfig.Nobl9.Environment,
    Timeout:      30 * time.Second,
}

client, err := nobl9.New(nobl9Config, log)
if err != nil {
    log.Fatal("Failed to create Nobl9 client", logger.Fields{"error": err})
}

// Create processor
processor := processor.New(client, log)
```

### Workflow Integration

```go
// Scan for Nobl9 files
scanner := scanner.New(log)
scanResult, err := scanner.Scan(repoPath, filePattern)
if err != nil {
    log.Error("Failed to scan repository", logger.Fields{"error": err})
    return
}

// Get Nobl9 files
nobl9Files := scanner.GetNobl9Files(scanResult)

// Process files based on configuration
var result *processor.ProcessingResult
if actionConfig.IsDryRun() {
    result, err = processor.ProcessWithDryRun(ctx, nobl9Files)
} else {
    result, err = processor.ProcessFiles(ctx, nobl9Files)
}

if err != nil {
    log.Error("Failed to process files", logger.Fields{"error": err})
    return
}

// Report results
log.Info("Processing completed", logger.Fields{
    "files_processed": result.FilesProcessed,
    "is_success": result.IsSuccess,
})
```

### Error Reporting

```go
// Report processing errors to GitHub Actions
if !result.IsSuccess {
    log.Error("Processing completed with errors", logger.Fields{
        "files_with_errors": result.FilesWithErrors,
        "errors": len(result.Errors),
    })
    
    // Set GitHub Actions output
    fmt.Printf("::set-output name=success::false\n")
    fmt.Printf("::set-output name=files_with_errors::%d\n", result.FilesWithErrors)
    fmt.Printf("::set-output name=error_count::%d\n", len(result.Errors))
} else {
    log.Info("Processing completed successfully", logger.Fields{
        "files_processed": result.FilesProcessed,
        "projects_created": result.ProjectsCreated,
        "role_bindings_created": result.RoleBindingsCreated,
    })
    
    // Set GitHub Actions output
    fmt.Printf("::set-output name=success::true\n")
    fmt.Printf("::set-output name=files_processed::%d\n", result.FilesProcessed)
    fmt.Printf("::set-output name=projects_created::%d\n", result.ProjectsCreated)
    fmt.Printf("::set-output name=role_bindings_created::%d\n", result.RoleBindingsCreated)
}
```

## Best Practices

### File Processing

1. **Batch Processing** - Process multiple files together for efficiency
2. **Error Recovery** - Continue processing despite individual failures
3. **Result Aggregation** - Aggregate results across all files
4. **Progress Tracking** - Track processing progress for large file sets

### Object Processing

1. **Type-Specific Handling** - Handle different object types appropriately
2. **Validation** - Validate objects before processing
3. **Email Resolution** - Ensure all emails are resolved before processing
4. **Error Context** - Provide context for object processing errors

### Manifest Application

1. **SDK Integration** - Use the official Nobl9 SDK for manifest application
2. **Validation** - Validate manifests before application
3. **Error Handling** - Handle application errors gracefully
4. **Rollback Strategy** - Consider rollback strategies for failed applications

### Dry-Run Mode

1. **Safe Testing** - Use dry-run mode for testing configurations
2. **Validation** - Validate all aspects without making changes
3. **Simulation** - Simulate processing operations
4. **Error Detection** - Detect potential issues before actual processing

### Error Handling

1. **Graceful Degradation** - Continue processing despite individual failures
2. **Detailed Logging** - Log all errors with context
3. **Error Aggregation** - Aggregate errors for comprehensive reporting
4. **Recovery Strategies** - Implement recovery strategies for common errors

### Performance

1. **Batch Processing** - Process multiple files together
2. **Concurrent Processing** - Use concurrent processing where appropriate
3. **Resource Management** - Manage resources efficiently
4. **Caching** - Cache results to reduce redundant operations

### Security

1. **Input Validation** - Validate all inputs before processing
2. **Error Sanitization** - Don't expose sensitive information in errors
3. **Access Control** - Ensure proper access control for operations
4. **Audit Logging** - Log all processing operations for audit purposes 