# Logging Framework

This document describes the structured logging framework used by the Nobl9 GitHub Action, providing comprehensive logging capabilities for debugging, monitoring, and troubleshooting.

## Overview

The logging framework is built on top of [Logrus](https://github.com/sirupsen/logrus) and provides:

- **Structured Logging** - JSON and text formats with consistent field structure
- **Multiple Log Levels** - Debug, Info, Warn, Error with configurable verbosity
- **GitHub Actions Integration** - Automatic context and metadata injection
- **Domain-Specific Logging** - Specialized methods for Nobl9 operations
- **Context Support** - Request tracing and correlation IDs
- **Performance Monitoring** - API call timing and duration tracking

## Log Levels

The framework supports four log levels with increasing severity:

| Level | Description | Use Case |
|-------|-------------|----------|
| `debug` | Detailed debugging information | Development, troubleshooting |
| `info` | General operational information | Normal operation tracking |
| `warn` | Warning conditions | Non-critical issues |
| `error` | Error conditions | Failures and exceptions |

### Configuration

```yaml
# GitHub Actions workflow
- name: Sync Nobl9 Projects
  uses: docker://docker.io/your-dockerhub-username/nobl9-github-action:latest
  with:
    log-level: "info"    # debug, info, warn, error
    log-format: "json"   # json, text
```

## Log Formats

### JSON Format (Default)

Structured JSON output for machine processing and log aggregation:

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "info",
  "message": "Starting Nobl9 project processing",
  "event": "processing_start",
  "config": {
    "dry_run": false,
    "force": false
  },
  "github_actions": true,
  "workflow": "nobl9-sync",
  "run_id": "123456789",
  "actor": "github-actions[bot]",
  "repository": "org/repo",
  "sha": "abc123def456"
}
```

### Text Format

Human-readable text output for development and debugging:

```
time="2024-01-15T10:30:45Z" level=info msg="Starting Nobl9 project processing" event=processing_start dry_run=false force=false github_actions=true workflow=nobl9-sync run_id=123456789 actor=github-actions[bot] repository=org/repo sha=abc123def456
```

## Structured Fields

All log entries include consistent structured fields:

### Common Fields

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | ISO 8601 timestamp |
| `level` | string | Log level (debug, info, warn, error) |
| `message` | string | Human-readable message |
| `event` | string | Event type for categorization |

### GitHub Actions Fields

Automatically added when running in GitHub Actions:

| Field | Type | Description |
|-------|------|-------------|
| `github_actions` | boolean | Always true in GitHub Actions |
| `workflow` | string | GitHub workflow name |
| `run_id` | string | GitHub run ID |
| `actor` | string | GitHub actor (user/bot) |
| `repository` | string | GitHub repository (org/repo) |
| `sha` | string | Git commit SHA |

### Context Fields

Added when using context-aware logging:

| Field | Type | Description |
|-------|------|-------------|
| `request_id` | string | Request correlation ID |
| `correlation_id` | string | Cross-service correlation ID |

## Domain-Specific Logging

The framework provides specialized logging methods for Nobl9 operations:

### Processing Events

```go
// Start of processing
logger.LogProcessingStart(config)

// Completion with statistics
logger.LogProcessingComplete(stats)
```

**Example Output:**
```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "info",
  "message": "Starting Nobl9 project processing",
  "event": "processing_start",
  "config": {
    "dry_run": false,
    "force": false,
    "validate_only": false
  }
}
```

### File Processing

```go
// File processing results
logger.LogFileProcessed(filePath, fileType, success, additionalFields)
```

**Example Output:**
```json
{
  "timestamp": "2024-01-15T10:30:46Z",
  "level": "info",
  "message": "File processed successfully",
  "event": "file_processed",
  "file_path": "/workspace/projects/my-project.yaml",
  "file_type": "nobl9-project",
  "success": true,
  "processing_time_ms": 150
}
```

### Nobl9 API Calls

```go
// API call tracking with timing
logger.LogNobl9APICall(method, endpoint, success, duration, additionalFields)
```

**Example Output:**
```json
{
  "timestamp": "2024-01-15T10:30:47Z",
  "level": "debug",
  "message": "Nobl9 API call successful",
  "event": "nobl9_api_call",
  "method": "POST",
  "endpoint": "/api/v1/projects",
  "success": true,
  "duration": "150ms",
  "duration_ms": 150,
  "status_code": 201
}
```

### User Resolution

```go
// Email to UserID resolution
logger.LogUserResolution(email, userID, success, additionalFields)
```

**Example Output:**
```json
{
  "timestamp": "2024-01-15T10:30:48Z",
  "level": "info",
  "message": "User email resolved to UserID",
  "event": "user_resolution",
  "email": "user@example.com",
  "user_id": "okta-user-123",
  "success": true,
  "resolution_time_ms": 25
}
```

### Project Operations

```go
// Project creation/update
logger.LogProjectOperation(operation, projectName, success, additionalFields)

// Role binding operations
logger.LogRoleBindingOperation(operation, roleBindingName, projectName, success, additionalFields)
```

**Example Output:**
```json
{
  "timestamp": "2024-01-15T10:30:49Z",
  "level": "info",
  "message": "Project operation completed",
  "event": "project_operation",
  "operation": "create",
  "project_name": "my-project",
  "success": true,
  "project_id": "proj-123"
}
```

### Validation Results

```go
// File validation results
logger.LogValidationResult(filePath, valid, errors, warnings)
```

**Example Output:**
```json
{
  "timestamp": "2024-01-15T10:30:50Z",
  "level": "error",
  "message": "File validation failed",
  "event": "validation_result",
  "file_path": "/workspace/projects/invalid.yaml",
  "valid": false,
  "errors": [
    "missing required field: name",
    "invalid project name format"
  ],
  "warnings": [
    "deprecated field: old_field"
  ]
}
```

## Usage Examples

### Basic Logging

```go
import "github.com/your-org/nobl9-action/pkg/logger"

// Create logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)

// Basic messages
log.Info("Application started")
log.Warn("Configuration warning")
log.Error("Operation failed")

// With structured fields
log.Info("Processing file", logger.Fields{
    "file_path": "/path/to/file.yaml",
    "file_size": 1024,
})
```

### Context-Aware Logging

```go
import "context"

// Create context with correlation IDs
ctx := context.WithValue(context.Background(), "request_id", "req-123")
ctx = context.WithValue(ctx, "correlation_id", "corr-456")

// Use context-aware logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)
ctxLogger := log.WithContext(ctx)

ctxLogger.Info("Processing request")
```

### Domain-Specific Logging

```go
// Nobl9-specific operations
log := logger.New(logger.LevelInfo, logger.FormatJSON)

// Start processing
log.LogProcessingStart(map[string]interface{}{
    "dry_run": false,
    "force": false,
})

// Process file
log.LogFileProcessed("/path/to/project.yaml", "nobl9-project", true, logger.Fields{
    "processing_time_ms": 150,
})

// API call
start := time.Now()
// ... make API call ...
duration := time.Since(start)
log.LogNobl9APICall("POST", "/api/v1/projects", true, duration, logger.Fields{
    "status_code": 201,
})

// User resolution
log.LogUserResolution("user@example.com", "okta-user-123", true, logger.Fields{
    "resolution_time_ms": 25,
})

// Project operation
log.LogProjectOperation("create", "my-project", true, logger.Fields{
    "project_id": "proj-123",
})

// Complete processing
log.LogProcessingComplete(map[string]interface{}{
    "files_processed": 5,
    "projects_created": 2,
    "errors": 0,
})
```

## Error Handling

### Error Logging

```go
// Log errors with context
err := fmt.Errorf("API call failed")
log.ErrorWithErr("Failed to create project", err, logger.Fields{
    "project_name": "my-project",
    "attempt": 3,
})
```

### Fatal Logging

```go
// Fatal errors that terminate the application
log.FatalWithErr("Critical configuration error", err)
```

## GitHub Actions Integration

### Automatic Context

When running in GitHub Actions, the logger automatically includes:

- Workflow information
- Run details
- Actor information
- Repository context
- Commit SHA

### GitHub Actions Output

The logger integrates with GitHub Actions output for better visibility:

```yaml
# GitHub Actions workflow
- name: Sync Nobl9 Projects
  uses: docker://docker.io/your-dockerhub-username/nobl9-github-action:latest
  with:
    log-level: "info"
    log-format: "json"
  id: nobl9-sync

- name: Display logs
  run: |
    echo "Processing completed with status: ${{ steps.nobl9-sync.outputs.success }}"
```

### Log Filtering

Use GitHub Actions log filtering to focus on specific events:

```bash
# Filter for specific events
grep '"event":"processing_start"' logs.json

# Filter for errors
grep '"level":"error"' logs.json

# Filter for specific project
grep '"project_name":"my-project"' logs.json
```

## Performance Considerations

### Log Level Impact

- **Debug**: Most verbose, impacts performance
- **Info**: Balanced for production use
- **Warn**: Minimal performance impact
- **Error**: Minimal performance impact

### Output Format Impact

- **JSON**: Slightly slower due to serialization
- **Text**: Faster, human-readable

### Recommendations

- Use `info` level for production
- Use `debug` level for troubleshooting
- Use JSON format for log aggregation
- Use text format for development

## Troubleshooting

### Common Issues

**1. Missing Logs**
```
No log output visible
```
**Solution:** Check log level configuration and ensure output is not filtered.

**2. JSON Parsing Errors**
```
Invalid JSON in log output
```
**Solution:** Ensure proper field types and avoid circular references.

**3. Performance Issues**
```
Slow logging performance
```
**Solution:** Reduce log level, use text format, or filter unnecessary fields.

### Debug Mode

Enable debug logging for troubleshooting:

```yaml
- name: Debug Nobl9 Action
  uses: docker://docker.io/your-dockerhub-username/nobl9-github-action:latest
  with:
    log-level: "debug"
    log-format: "text"
    dry-run: true
```

### Log Analysis

Use tools to analyze structured logs:

```bash
# Count events by type
jq -r '.event' logs.json | sort | uniq -c

# Find errors
jq -r 'select(.level == "error") | .message' logs.json

# Performance analysis
jq -r 'select(.duration_ms) | .duration_ms' logs.json | awk '{sum+=$1} END {print "Average:", sum/NR}'
```

## Best Practices

### Logging Guidelines

1. **Use Appropriate Levels**
   - Debug: Detailed troubleshooting
   - Info: Normal operations
   - Warn: Non-critical issues
   - Error: Failures and exceptions

2. **Include Context**
   - Always include relevant fields
   - Use structured data when possible
   - Include correlation IDs for tracing

3. **Avoid Sensitive Data**
   - Never log credentials
   - Mask sensitive fields
   - Use placeholder values for secrets

4. **Consistent Naming**
   - Use consistent field names
   - Follow naming conventions
   - Document field meanings

5. **Performance Awareness**
   - Avoid expensive operations in logging
   - Use appropriate log levels
   - Consider log volume impact

### Field Naming Conventions

- Use snake_case for field names
- Be descriptive and specific
- Include units for measurements
- Use consistent terminology

### Event Naming

- Use descriptive event names
- Follow pattern: `noun_verb`
- Examples: `processing_start`, `file_processed`, `api_call` 