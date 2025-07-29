# Retry Mechanism

This document describes the retry mechanism that provides comprehensive retry logic with exponential backoff and configurable retry policies for handling transient failures in API calls and operations.

## Overview

The retry mechanism provides:

- **Exponential Backoff** - Intelligent delay calculation with exponential backoff
- **Jitter** - Randomization to prevent thundering herd problems
- **Configurable Policies** - Customizable retry behavior for different scenarios
- **Error Classification** - Automatic classification of retryable vs non-retryable errors
- **Context Support** - Full context cancellation support
- **Detailed Logging** - Comprehensive logging of retry operations
- **Multiple Interfaces** - Simple and advanced retry interfaces

## Features

### Exponential Backoff

The retry mechanism implements exponential backoff:

- **Initial Delay** - Configurable initial delay between retries
- **Backoff Factor** - Exponential increase in delay (default: 2.0)
- **Maximum Delay** - Upper limit on delay to prevent excessive waiting
- **Jitter** - Randomization to prevent synchronized retries

### Error Classification

Automatic classification of errors:

- **Retryable Errors** - Transient failures that should trigger retries
- **Non-Retryable Errors** - Permanent failures that should not be retried
- **Pattern Matching** - Configurable patterns for error classification
- **Custom Patterns** - Support for custom retryable error patterns

### Policy Configuration

Flexible policy configuration:

- **Max Attempts** - Maximum number of retry attempts
- **Initial Delay** - Initial delay between retries
- **Max Delay** - Maximum delay limit
- **Backoff Factor** - Exponential backoff multiplier
- **Jitter Factor** - Randomization factor (0.0 to 1.0)
- **Retryable Patterns** - List of error patterns that trigger retries

### Context Support

Full context integration:

- **Cancellation** - Respect context cancellation during retries
- **Timeout** - Support for context-based timeouts
- **Graceful Shutdown** - Clean shutdown on context cancellation

## Usage

### Basic Retry

```go
import (
    "context"
    "github.com/your-org/nobl9-action/pkg/retry"
    "github.com/your-org/nobl9-action/pkg/logger"
)

// Create logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)

// Define retryable function
fn := func(ctx context.Context) (interface{}, error) {
    // Your operation here
    return "success", nil
}

// Use default retry policy
result, err := retry.RetrySimple(context.Background(), log, "my operation", fn)
if err != nil {
    log.Error("Operation failed", logger.Fields{"error": err})
    return
}

log.Info("Operation succeeded", logger.Fields{"result": result})
```

### Custom Retry Policy

```go
// Create custom retry policy
policy := retry.NewPolicy(
    5,                    // Max attempts
    2*time.Second,        // Initial delay
    60*time.Second,       // Max delay
    2.0,                  // Backoff factor
    0.1,                  // Jitter factor
)

// Use custom policy
result, err := retry.RetryWithResult(context.Background(), policy, log, "my operation", fn)
if err != nil {
    log.Error("Operation failed", logger.Fields{"error": err})
    return
}

log.Info("Operation succeeded", logger.Fields{"result": result})
```

### Detailed Retry Result

```go
// Get detailed retry information
result, err := retry.Retry(context.Background(), policy, log, "my operation", fn)
if err != nil {
    log.Error("Operation failed", logger.Fields{"error": err})
    return
}

log.Info("Retry operation completed", logger.Fields{
    "success":      result.Success,
    "attempts":     result.Attempts,
    "total_delay":  result.TotalDelay.String(),
    "final_result": result.FinalResult,
})
```

### Retryable API Operation

```go
// Create retryable API operation
operation := retry.NewRetryableAPIOperation(policy, log)

// Execute with retry logic
result, err := operation.Execute(context.Background(), "API call", fn)
if err != nil {
    log.Error("API call failed", logger.Fields{"error": err})
    return
}

log.Info("API call succeeded", logger.Fields{"result": result})
```

### Predefined Policies

```go
// API operations
apiPolicy := retry.CreatePolicyForAPI(5)

// Network operations
networkPolicy := retry.CreatePolicyForNetwork(3)

// Rate limiting
rateLimitPolicy := retry.CreatePolicyForRateLimit(4)

// Use predefined policy
result, err := retry.RetryWithResult(context.Background(), apiPolicy, log, "API operation", fn)
```

## Data Structures

### Policy

```go
type Policy struct {
    MaxAttempts     int           // Maximum number of retry attempts
    InitialDelay    time.Duration // Initial delay between retries
    MaxDelay        time.Duration // Maximum delay between retries
    BackoffFactor   float64       // Exponential backoff factor
    JitterFactor    float64       // Jitter factor for randomization (0.0 to 1.0)
    RetryableErrors []string      // List of error patterns that should trigger retries
}
```

### RetryResult

```go
type RetryResult struct {
    Attempts     int           // Number of attempts made
    Success      bool          // Whether the operation succeeded
    LastError    error         // Last error encountered
    TotalDelay   time.Duration // Total delay across all retries
    FinalResult  interface{}   // Final result of the operation
}
```

### RetryableError

```go
type RetryableError struct {
    Message string // Error message
    Err     error  // Underlying error
}
```

## Policy Configuration

### Default Policy

```go
policy := retry.DefaultPolicy()
// MaxAttempts: 3
// InitialDelay: 1s
// MaxDelay: 30s
// BackoffFactor: 2.0
// JitterFactor: 0.1
// RetryableErrors: ["timeout", "connection refused", "network error", "rate limit", "429", "503", "502", "500"]
```

### API Policy

```go
policy := retry.CreatePolicyForAPI(5)
// Optimized for API operations with longer delays and comprehensive error patterns
```

### Network Policy

```go
policy := retry.CreatePolicyForNetwork(3)
// Optimized for network operations with shorter delays and network-specific errors
```

### Rate Limit Policy

```go
policy := retry.CreatePolicyForRateLimit(4)
// Optimized for rate limiting with longer delays and rate limit specific errors
```

## Error Classification

### Retryable Error Patterns

```go
patterns := retry.RetryableErrorPatterns()
// Returns common retryable error patterns:
// - "timeout"
// - "connection refused"
// - "network error"
// - "rate limit"
// - "429"
// - "503"
// - "502"
// - "500"
// - "temporary failure"
// - "service unavailable"
// - "bad gateway"
// - "gateway timeout"
// - "too many requests"
// - "internal server error"
```

### Custom Error Patterns

```go
policy := &retry.Policy{
    MaxAttempts: 3,
    InitialDelay: 1 * time.Second,
    MaxDelay: 30 * time.Second,
    BackoffFactor: 2.0,
    JitterFactor: 0.1,
    RetryableErrors: []string{
        "custom error pattern",
        "another pattern",
        "timeout",
    },
}
```

## Delay Calculation

### Exponential Backoff

The delay is calculated using exponential backoff:

```
delay = initialDelay * (backoffFactor ^ (attempt - 1))
```

### Jitter

Jitter is applied to prevent synchronized retries:

```
jitter = delay * jitterFactor
finalDelay = delay + (random * 2 * jitter) - jitter
```

### Maximum Delay

The delay is capped at the maximum delay:

```
if delay > maxDelay {
    delay = maxDelay
}
```

## Context Integration

### Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Context cancellation is respected during retries
result, err := retry.Retry(ctx, policy, log, "operation", fn)
if err != nil {
    if err.Error() == "operation cancelled" {
        log.Info("Operation was cancelled")
    }
}
```

### Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Operation will be cancelled after timeout
result, err := retry.Retry(ctx, policy, log, "operation", fn)
```

## Logging

### Retry Logging

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "debug",
  "message": "Starting retry operation",
  "operation": "API call",
  "max_attempts": 3,
  "initial_delay": "1s"
}
```

### Attempt Logging

```json
{
  "timestamp": "2024-01-15T10:30:46Z",
  "level": "debug",
  "message": "Executing operation attempt",
  "operation": "API call",
  "attempt": 1
}
```

### Success Logging

```json
{
  "timestamp": "2024-01-15T10:30:47Z",
  "level": "info",
  "message": "Operation succeeded",
  "operation": "API call",
  "attempts": 2
}
```

### Error Logging

```json
{
  "timestamp": "2024-01-15T10:30:48Z",
  "level": "warn",
  "message": "Operation failed with retryable error",
  "operation": "API call",
  "attempt": 1,
  "error": "connection timeout"
}
```

### Final Failure Logging

```json
{
  "timestamp": "2024-01-15T10:30:49Z",
  "level": "error",
  "message": "Operation failed after all attempts",
  "operation": "API call",
  "attempts": 3,
  "total_delay": "7s",
  "final_error": "connection timeout"
}
```

## Integration with GitHub Action

### Nobl9 Client Integration

```go
// Create retryable API operation for Nobl9 client
policy := retry.CreatePolicyForAPI(3)
operation := retry.NewRetryableAPIOperation(policy, log)

// Wrap Nobl9 API calls with retry logic
fn := func(ctx context.Context) (interface{}, error) {
    return client.GetProject(ctx, "project-name")
}

result, err := operation.Execute(ctx, "get project", fn)
if err != nil {
    log.Error("Failed to get project", logger.Fields{"error": err})
    return
}

project := result.(*sdk.Project)
```

### Email Resolver Integration

```go
// Create retryable operation for email resolution
policy := retry.CreatePolicyForAPI(3)
operation := retry.NewRetryableAPIOperation(policy, log)

// Wrap email resolution with retry logic
fn := func(ctx context.Context) (interface{}, error) {
    return resolver.ResolveEmail(ctx, "user@example.com")
}

result, err := operation.Execute(ctx, "resolve email", fn)
if err != nil {
    log.Error("Failed to resolve email", logger.Fields{"error": err})
    return
}

resolutionResult := result.(*resolver.ResolutionResult)
```

### Manifest Application Integration

```go
// Create retryable operation for manifest application
policy := retry.CreatePolicyForAPI(3)
operation := retry.NewRetryableAPIOperation(policy, log)

// Wrap manifest application with retry logic
fn := func(ctx context.Context) (interface{}, error) {
    return nil, parser.ApplyManifest(ctx, manifestContent)
}

_, err := operation.Execute(ctx, "apply manifest", fn)
if err != nil {
    log.Error("Failed to apply manifest", logger.Fields{"error": err})
    return
}
```

## Best Practices

### Policy Selection

1. **API Operations** - Use `CreatePolicyForAPI()` for general API calls
2. **Network Operations** - Use `CreatePolicyForNetwork()` for network-specific operations
3. **Rate Limiting** - Use `CreatePolicyForRateLimit()` for rate-limited operations
4. **Custom Policies** - Create custom policies for specific requirements

### Error Classification

1. **Retryable Errors** - Only retry transient, recoverable errors
2. **Non-Retryable Errors** - Don't retry permanent failures (auth, validation, etc.)
3. **Pattern Matching** - Use specific error patterns for accurate classification
4. **Custom Patterns** - Add custom patterns for application-specific errors

### Context Usage

1. **Cancellation** - Always use context for cancellation support
2. **Timeouts** - Set appropriate timeouts for operations
3. **Graceful Shutdown** - Handle context cancellation gracefully
4. **Resource Cleanup** - Ensure proper cleanup on cancellation

### Logging

1. **Detailed Logging** - Log all retry attempts and results
2. **Error Context** - Include error context in logs
3. **Performance Metrics** - Log timing and attempt information
4. **Structured Logging** - Use structured logging for better analysis

### Performance

1. **Jitter** - Always use jitter to prevent thundering herd
2. **Max Delay** - Set reasonable maximum delays
3. **Max Attempts** - Limit retry attempts to prevent infinite loops
4. **Backoff Factor** - Use appropriate backoff factors (1.5-2.0)

### Error Handling

1. **Error Wrapping** - Wrap errors with context
2. **Error Types** - Use typed errors for better handling
3. **Error Recovery** - Implement error recovery strategies
4. **Error Reporting** - Report errors appropriately

### Security

1. **Error Sanitization** - Don't expose sensitive information in errors
2. **Rate Limiting** - Respect rate limits and back off appropriately
3. **Authentication** - Handle authentication errors appropriately
4. **Authorization** - Don't retry authorization failures

## Examples

### Complete Retry Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/your-org/nobl9-action/pkg/retry"
    "github.com/your-org/nobl9-action/pkg/logger"
)

func main() {
    log := logger.New(logger.LevelInfo, logger.FormatJSON)
    
    // Create retry policy
    policy := retry.CreatePolicyForAPI(3)
    
    // Define operation
    fn := func(ctx context.Context) (interface{}, error) {
        // Simulate API call
        if time.Now().Unix()%3 == 0 {
            return nil, fmt.Errorf("temporary error")
        }
        return "success", nil
    }
    
    // Execute with retry
    result, err := retry.RetryWithResult(context.Background(), policy, log, "API call", fn)
    if err != nil {
        log.Error("Operation failed", logger.Fields{"error": err})
        return
    }
    
    log.Info("Operation succeeded", logger.Fields{"result": result})
}
```

### Custom Error Handling

```go
// Create custom retryable error
err := retry.NewRetryableError("API call failed", originalErr)

// Check if error is retryable
if retry.IsRetryableError(err) {
    log.Info("Error is retryable", logger.Fields{"error": err})
}
```

### Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

policy := retry.CreatePolicyForAPI(5)
operation := retry.NewRetryableAPIOperation(policy, log)

result, err := operation.Execute(ctx, "long running operation", fn)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Error("Operation timed out", logger.Fields{"error": err})
    } else {
        log.Error("Operation failed", logger.Fields{"error": err})
    }
    return
}
``` 