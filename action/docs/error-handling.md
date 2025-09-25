# Error Handling Documentation

## Overview

The Nobl9 GitHub Action implements a comprehensive error handling system that provides detailed error categorization, structured logging, and intelligent retry logic. This system ensures that errors are properly tracked, logged, and handled appropriately based on their type and severity.

## Error Types

### Configuration Errors (`ErrorTypeConfig`)
- **Severity**: High
- **Retryable**: No
- **Description**: Errors related to application configuration
- **Examples**: Missing required environment variables, invalid configuration values
- **Exit Code**: 2

### Validation Errors (`ErrorTypeValidation`)
- **Severity**: Medium
- **Retryable**: No
- **Description**: Errors related to YAML validation and schema compliance
- **Examples**: Invalid YAML syntax, missing required fields, schema violations
- **Exit Code**: 3

### Nobl9 API Errors (`ErrorTypeNobl9API`)
- **Severity**: High
- **Retryable**: Depends on specific error
- **Description**: Errors from Nobl9 API calls
- **Examples**: API timeouts, invalid responses, server errors
- **Exit Code**: 4

### File Processing Errors (`ErrorTypeFileProcessing`)
- **Severity**: Medium
- **Retryable**: No
- **Description**: Errors during file processing operations
- **Examples**: File read failures, parsing errors, file system issues
- **Exit Code**: 5

### Authentication Errors (`ErrorTypeAuth`)
- **Severity**: Critical
- **Retryable**: No
- **Description**: Authentication and authorization errors
- **Examples**: Invalid credentials, expired tokens, insufficient permissions
- **Exit Code**: 6

### Network Errors (`ErrorTypeNetwork`)
- **Severity**: Medium
- **Retryable**: Yes
- **Description**: Network connectivity issues
- **Examples**: Connection refused, DNS resolution failures, network timeouts
- **Exit Code**: 7

### Rate Limiting Errors (`ErrorTypeRateLimit`)
- **Severity**: Medium
- **Retryable**: Yes
- **Description**: API rate limiting and quota exceeded errors
- **Examples**: 429 Too Many Requests, quota exceeded
- **Exit Code**: 8

### Timeout Errors (`ErrorTypeTimeout`)
- **Severity**: Medium
- **Retryable**: Yes
- **Description**: Operation timeout errors
- **Examples**: Context deadline exceeded, request timeouts
- **Exit Code**: 9

### User Resolution Errors (`ErrorTypeUserResolution`)
- **Severity**: Medium
- **Retryable**: No
- **Description**: Errors during email to UserID resolution
- **Examples**: User not found, invalid email format
- **Exit Code**: 10

### Manifest Errors (`ErrorTypeManifest`)
- **Severity**: High
- **Retryable**: No
- **Description**: Errors related to Nobl9 manifest processing
- **Examples**: Invalid manifest structure, validation failures
- **Exit Code**: 11

## Error Severity Levels

### Critical (`SeverityCritical`)
- **Description**: Errors that prevent the application from functioning
- **Action**: Immediate termination, no retry
- **Examples**: Authentication failures, critical configuration errors

### High (`SeverityHigh`)
- **Description**: Errors that significantly impact functionality
- **Action**: Log error, may retry depending on type
- **Examples**: API errors, configuration issues

### Medium (`SeverityMedium`)
- **Description**: Errors that affect specific operations but not overall functionality
- **Action**: Log error, retry if appropriate
- **Examples**: Network timeouts, validation errors

### Low (`SeverityLow`)
- **Description**: Minor errors that don't affect core functionality
- **Action**: Log warning, continue processing
- **Examples**: Non-critical validation warnings

## Error Handling Components

### 1. Error Aggregator

The `ErrorAggregator` provides comprehensive error tracking and analysis:

```go
// Create error aggregator
errorAggregator := errors.NewErrorAggregator()

// Add errors
errorAggregator.AddError(err)

// Get error summary
summary := errorAggregator.GetErrorSummary()

// Check for critical errors
if errorAggregator.HasCriticalErrors() {
    // Handle critical errors
}
```

### 2. Structured Logging

All errors are logged with structured information:

```go
// Log detailed error
logger.LogDetailedError(err, "operation_name", map[string]interface{}{
    "file_path": "/path/to/file",
    "attempt":   3,
    "duration":  "5s",
}, logger.Fields{
    "error": err.Error(),
})
```

### 3. Retry Logic

The retry system automatically handles retryable errors:

```go
// Retry with custom policy
policy := retry.CreatePolicyForAPI(3)
result, err := retry.Retry(ctx, policy, logger, "operation_name", func(ctx context.Context) (interface{}, error) {
    return performOperation(ctx)
})
```

## Error Patterns and Detection

### Retryable Error Patterns
The system automatically detects retryable errors based on common patterns:

- `timeout`
- `connection refused`
- `network error`
- `rate limit`
- `429` (Too Many Requests)
- `503` (Service Unavailable)
- `502` (Bad Gateway)
- `500` (Internal Server Error)

### Authentication Error Patterns
- `unauthorized`
- `forbidden`
- `invalid credentials`
- `authentication failed`
- `401` (Unauthorized)
- `403` (Forbidden)

### Rate Limiting Error Patterns
- `rate limit`
- `429` (Too Many Requests)
- `too many requests`
- `quota exceeded`

## Exit Codes

The application uses specific exit codes to indicate different types of failures:

| Exit Code | Error Type | Description |
|-----------|------------|-------------|
| 0 | Success | Operation completed successfully |
| 1 | General Error | Unspecified error |
| 2 | Configuration Error | Configuration validation failed |
| 3 | Validation Error | YAML validation failed |
| 4 | Nobl9 API Error | API call failed |
| 5 | File Processing Error | File processing failed |
| 6 | Authentication Error | Authentication failed |
| 7 | Network Error | Network connectivity issue |
| 8 | Rate Limit Error | API rate limit exceeded |
| 9 | Timeout Error | Operation timed out |
| 10 | Retryable Error | Retryable operation failed |

## Best Practices

### 1. Error Wrapping
Always wrap errors with context:

```go
// Good
return errors.NewNobl9APIError("failed to create project", err)

// Better
return errors.NewNobl9APIErrorWithDetails("failed to create project", err, map[string]interface{}{
    "project_name": projectName,
    "attempt":      attempt,
})
```

### 2. Error Logging
Log errors with sufficient context:

```go
logger.LogDetailedError(err, "project_creation", map[string]interface{}{
    "project_name": projectName,
    "user_id":      userID,
    "duration":     duration.String(),
})
```

### 3. Error Aggregation
Use error aggregators for batch operations:

```go
errorAggregator := errors.NewErrorAggregator()

for _, file := range files {
    if err := processFile(file); err != nil {
        errorAggregator.AddError(err)
    }
}

if errorAggregator.HasErrors() {
    summary := errorAggregator.GetErrorSummary()
    logger.Error("Batch processing completed with errors", logger.Fields{
        "error_summary": summary,
    })
}
```

### 4. Retry Configuration
Configure appropriate retry policies:

```go
// For API operations
policy := retry.CreatePolicyForAPI(3)

// For network operations
policy := retry.CreatePolicyForNetwork(5)

// For rate limiting
policy := retry.CreatePolicyForRateLimit(3)
```

## Monitoring and Alerting

### Error Metrics
The system provides comprehensive error metrics:

- Total error count
- Error count by type
- Error count by severity
- Retryable vs non-retryable errors
- Error duration and frequency

### Log Analysis
Structured logs enable easy analysis:

```json
{
  "level": "error",
  "event": "detailed_error",
  "operation": "project_creation",
  "error_message": "failed to create project",
  "error_type": "nobl9_api",
  "retryable": true,
  "severity": "medium",
  "context_project_name": "my-project",
  "context_attempt": 2,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Troubleshooting

### Common Error Scenarios

1. **Authentication Failures**
   - Check client credentials
   - Verify token expiration
   - Ensure proper permissions

2. **Rate Limiting**
   - Implement exponential backoff
   - Reduce request frequency
   - Check API quotas

3. **Network Issues**
   - Verify network connectivity
   - Check firewall settings
   - Validate DNS resolution

4. **Validation Errors**
   - Review YAML syntax
   - Check required fields
   - Validate against schema

### Debug Mode
Enable debug logging for detailed error information:

```bash
./nobl9-action --log-level=debug process --client-id=xxx --client-secret=xxx
```

## Integration with CI/CD

### GitHub Actions
The error handling system integrates seamlessly with GitHub Actions:

```yaml
- name: Process Nobl9 Configuration
  uses: docker://docker.io/dfaile/nobl9-github-action:latest
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
  continue-on-error: false
```

### Exit Code Handling
GitHub Actions can handle different exit codes:

```yaml
- name: Process with fallback
  id: process
  run: |
    ./nobl9-action process --client-id=${{ secrets.NOBL9_CLIENT_ID }}
    echo "exit_code=$?" >> $GITHUB_OUTPUT
  
- name: Handle configuration errors
  if: steps.process.outputs.exit_code == '2'
  run: |
    echo "Configuration error detected"
    # Handle configuration issues
```

This comprehensive error handling system ensures that the Nobl9 GitHub Action provides reliable, observable, and maintainable error management throughout all operations. 