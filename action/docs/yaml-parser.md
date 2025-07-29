# YAML Parser

This document describes the YAML parser that handles parsing and validation of Nobl9 configuration files using the official SDK manifest capabilities and schema validation.

## Overview

The YAML parser provides:

- **SDK Integration** - Uses the official Nobl9 Go SDK for parsing and validation
- **Schema Validation** - Validates Nobl9 objects against their schemas
- **Error Handling** - Comprehensive error reporting and recovery
- **Batch Processing** - Parse multiple files efficiently
- **Structured Logging** - Detailed logging of parsing operations

## Features

### SDK Integration

The parser leverages the official Nobl9 Go SDK:

- **Manifest Parsing** - Uses `sdk.DecodeObjects()` for YAML parsing
- **Object Validation** - Uses `manifest.Object.Validate()` for schema validation
- **Type Safety** - Leverages SDK's type system for object handling
- **Version Support** - Supports multiple Nobl9 API versions

### Schema Validation

The parser performs comprehensive validation:

- **Object Validation** - Validates each Nobl9 object against its schema
- **Required Fields** - Checks for required fields in objects
- **Field Types** - Validates field types and formats
- **Cross-References** - Validates references between objects

### Error Handling

Robust error handling with detailed reporting:

- **File-Level Errors** - Errors related to file access and format
- **Object-Level Errors** - Errors related to individual objects
- **Validation Errors** - Schema validation failures
- **Context Information** - Error context with file and object details

## Usage

### Basic Parsing

```go
import (
    "context"
    "github.com/your-org/nobl9-action/pkg/parser"
    "github.com/your-org/nobl9-action/pkg/nobl9"
    "github.com/your-org/nobl9-action/pkg/logger"
)

// Create parser with Nobl9 client and logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)
client := nobl9.New(config, log)
parser := parser.New(client, log)

// Parse a single file
ctx := context.Background()
result, err := parser.ParseFile(ctx, fileInfo)
if err != nil {
    log.Error("Failed to parse file", logger.Fields{"error": err})
    return
}

// Check parsing results
if result.IsValid {
    log.Info("File parsed successfully", logger.Fields{
        "valid_objects": len(result.ValidObjects),
    })
} else {
    log.Error("File has validation errors", logger.Fields{
        "errors": len(result.Errors),
    })
}
```

### Batch Parsing

```go
// Parse multiple files
results, err := parser.ParseFiles(ctx, fileInfos)
if err != nil {
    log.Error("Failed to parse files", logger.Fields{"error": err})
    return
}

// Process results
for _, result := range results {
    if result.IsValid {
        log.Info("File parsed successfully", logger.Fields{
            "file_path": result.FileInfo.Path,
            "valid_objects": len(result.ValidObjects),
        })
    } else {
        log.Error("File has errors", logger.Fields{
            "file_path": result.FileInfo.Path,
            "errors": len(result.Errors),
        })
    }
}
```

### Manifest Validation

```go
// Validate a manifest without parsing
content := []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project`)

err := parser.ValidateManifest(ctx, content)
if err != nil {
    log.Error("Manifest validation failed", logger.Fields{"error": err})
    return
}

log.Info("Manifest validation successful")
```

### Manifest Application

```go
// Apply a manifest to Nobl9
content := []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
spec:
  displayName: Test Project`)

err := parser.ApplyManifest(ctx, content)
if err != nil {
    log.Error("Manifest application failed", logger.Fields{"error": err})
    return
}

log.Info("Manifest applied successfully")
```

## Parse Results

### ParseResult Structure

```go
type ParseResult struct {
    FileInfo       *scanner.FileInfo    // Original file information
    Manifests      []manifest.Object    // All parsed objects
    ValidObjects   []manifest.Object    // Objects that passed validation
    InvalidObjects []InvalidObject      // Objects with validation errors
    Errors         []error              // All errors encountered
    Warnings       []string             // Warning messages
    IsValid        bool                 // Overall validity status
}
```

### InvalidObject Structure

```go
type InvalidObject struct {
    Object   manifest.Object // The invalid object
    Error    error          // Validation error
    Position string         // Position in file (e.g., "object 1")
}
```

### Result Processing

```go
// Get all valid objects
validObjects := parser.GetValidObjects(results)
log.Info("Total valid objects", logger.Fields{
    "count": len(validObjects),
})

// Get all invalid objects
invalidObjects := parser.GetInvalidObjects(results)
for _, invalid := range invalidObjects {
    log.Error("Invalid object", logger.Fields{
        "position": invalid.Position,
        "error":    invalid.Error,
    })
}

// Get all errors
errors := parser.GetErrors(results)
log.Error("Total errors", logger.Fields{
    "count": len(errors),
})

// Check overall validity
if parser.IsValid(results) {
    log.Info("All files are valid")
} else {
    log.Error("Some files have errors")
}
```

## Supported Nobl9 Objects

The parser supports all Nobl9 object types defined in the SDK:

### Core Objects
- **Project** - Nobl9 projects
- **RoleBinding** - User role assignments
- **SLO** - Service Level Objectives
- **SLI** - Service Level Indicators
- **AlertPolicy** - Alert policies
- **DataSource** - Data sources

### Object Validation

Each object type is validated according to its schema:

```go
// Object validation process
func (p *Parser) validateObject(ctx context.Context, obj manifest.Object) error {
    // Get object metadata
    kind := obj.GetKind()
    name := obj.GetName()
    version := obj.GetVersion()

    // Log validation start
    p.logger.Debug("Validating Nobl9 object", logger.Fields{
        "kind":        kind,
        "name":        name,
        "api_version": version,
    })

    // Validate using SDK
    if err := obj.Validate(); err != nil {
        return fmt.Errorf("object validation failed: %w", err)
    }

    // Log validation success
    p.logger.Debug("Object validation completed", logger.Fields{
        "kind": kind,
        "name": name,
    })

    return nil
}
```

## Error Handling

### Common Errors

#### File Errors
```
Error: file has errors: failed to read file
Error: file does not contain Nobl9 configuration
Error: failed to parse YAML: invalid YAML syntax
```

#### Validation Errors
```
Error: object validation failed: required field 'name' is missing
Error: object validation failed: invalid field type for 'spec'
Error: object validation failed: field 'metadata.project' is required
```

#### SDK Errors
```
Error: manifest validation failed: API error
Error: manifest application failed: authentication failed
Error: manifest application failed: resource not found
```

### Error Recovery

The parser implements graceful error recovery:

```go
// Parse with error recovery
results, err := parser.ParseFiles(ctx, files)
if err != nil {
    // Log overall error but continue processing
    log.Error("Parsing completed with errors", logger.Fields{"error": err})
}

// Process valid results even if some files failed
validObjects := parser.GetValidObjects(results)
if len(validObjects) > 0 {
    log.Info("Processing valid objects", logger.Fields{
        "count": len(validObjects),
    })
    // Continue with valid objects
}
```

## Logging

### Parse Operation Logging

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "info",
  "message": "Parsing Nobl9 YAML file",
  "file_path": "/path/to/project.yaml",
  "file_size": 1024
}
```

### Validation Logging

```json
{
  "timestamp": "2024-01-15T10:30:46Z",
  "level": "debug",
  "message": "Validating Nobl9 object",
  "kind": "Project",
  "name": "test-project",
  "api_version": "n9/v1alpha"
}
```

### Result Logging

```json
{
  "timestamp": "2024-01-15T10:30:47Z",
  "level": "info",
  "message": "YAML file parsing completed",
  "file_path": "/path/to/project.yaml",
  "total_objects": 2,
  "valid_objects": 2,
  "invalid_objects": 0,
  "errors": 0,
  "warnings": 0,
  "is_valid": true
}
```

## Performance Considerations

### Parsing Strategy

The parser uses efficient parsing strategies:

1. **Streaming Parsing** - Processes YAML content as streams
2. **Lazy Validation** - Validates objects only when needed
3. **Batch Processing** - Processes multiple files efficiently
4. **Memory Management** - Releases resources after processing

### Optimization Tips

- **File Size** - Large files are processed in chunks
- **Object Count** - Many objects are validated in parallel
- **Error Handling** - Errors don't stop processing of other objects
- **Caching** - Parsed objects are cached for reuse

## Integration with GitHub Action

### Configuration Integration

The parser integrates with the GitHub Action configuration:

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

// Create parser
parser := parser.New(client, log)
```

### Workflow Integration

```go
// Parse files from scanner
scanner := scanner.New(log)
scanResult, err := scanner.Scan(repoPath, filePattern)
if err != nil {
    log.Error("Failed to scan repository", logger.Fields{"error": err})
    return
}

// Parse Nobl9 files
nobl9Files := scanner.GetNobl9Files(scanResult)
results, err := parser.ParseFiles(ctx, nobl9Files)
if err != nil {
    log.Error("Failed to parse files", logger.Fields{"error": err})
    return
}

// Process results
if parser.IsValid(results) {
    log.Info("All files are valid")
    // Apply manifests
    for _, result := range results {
        for _, obj := range result.ValidObjects {
            // Process valid object
        }
    }
} else {
    log.Error("Some files have errors")
    // Handle errors
    errors := parser.GetErrors(results)
    for _, err := range errors {
        log.Error("Parsing error", logger.Fields{"error": err})
    }
}
```

## Best Practices

### File Organization

1. **Consistent Structure** - Use consistent YAML structure across files
2. **Clear Naming** - Use descriptive names for objects
3. **Version Management** - Specify API versions explicitly
4. **Documentation** - Include comments in YAML files

### Error Handling

1. **Graceful Degradation** - Continue processing despite individual failures
2. **Detailed Logging** - Log all errors with context
3. **User Feedback** - Provide clear error messages to users
4. **Recovery Strategies** - Implement retry and recovery mechanisms

### Performance

1. **Batch Processing** - Process multiple files together
2. **Parallel Validation** - Validate objects in parallel when possible
3. **Resource Management** - Release resources promptly
4. **Caching** - Cache parsed objects for reuse

### Validation

1. **Schema Compliance** - Ensure objects comply with Nobl9 schemas
2. **Cross-References** - Validate references between objects
3. **Required Fields** - Check for all required fields
4. **Type Safety** - Validate field types and formats 