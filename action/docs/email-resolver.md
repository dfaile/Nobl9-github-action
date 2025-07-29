# Email Resolver

This document describes the email resolver that handles email-to-UserID resolution using the Nobl9 API to convert email addresses to Okta User IDs for role binding creation.

## Overview

The email resolver provides:

- **API Integration** - Uses the Nobl9 API to resolve email addresses to UserIDs
- **Caching** - Intelligent caching to reduce API calls and improve performance
- **Batch Processing** - Efficient batch resolution of multiple email addresses
- **Concurrent Processing** - Parallel resolution with rate limiting
- **Error Handling** - Comprehensive error handling and recovery
- **YAML Integration** - Extract emails from YAML content for resolution

## Features

### API Integration

The resolver integrates with the Nobl9 API:

- **User Lookup** - Uses `client.GetUser()` to resolve emails to UserIDs
- **Error Handling** - Handles various API error scenarios gracefully
- **Rate Limiting** - Implements concurrent request limiting
- **Authentication** - Uses the same authentication as the Nobl9 client

### Caching

Intelligent caching system for performance optimization:

- **TTL-based Cache** - Configurable time-to-live for cached entries
- **Thread-safe** - Concurrent access with read-write mutex
- **Hit/Miss Tracking** - Track cache performance metrics
- **Negative Caching** - Cache "not found" results to avoid repeated API calls

### Batch Processing

Efficient batch resolution capabilities:

- **Concurrent Resolution** - Process multiple emails in parallel
- **Semaphore Limiting** - Limit concurrent API calls (default: 10)
- **Result Aggregation** - Aggregate results with statistics
- **Error Collection** - Collect and categorize all errors

### YAML Integration

Direct integration with YAML content:

- **Email Extraction** - Extract email addresses from YAML content
- **Pattern Matching** - Identify email patterns in YAML structures
- **Validation** - Validate extracted email addresses
- **Deduplication** - Remove duplicate email addresses

## Usage

### Basic Email Resolution

```go
import (
    "context"
    "github.com/your-org/nobl9-action/pkg/resolver"
    "github.com/your-org/nobl9-action/pkg/nobl9"
    "github.com/your-org/nobl9-action/pkg/logger"
)

// Create resolver with Nobl9 client and logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)
client := nobl9.New(config, log)
resolver := resolver.New(client, log)

// Resolve a single email
ctx := context.Background()
result, err := resolver.ResolveEmail(ctx, "user@example.com")
if err != nil {
    log.Error("Failed to resolve email", logger.Fields{"error": err})
    return
}

if result.Resolved {
    log.Info("Email resolved successfully", logger.Fields{
        "email":   result.Email,
        "user_id": result.UserID,
        "cached":  result.FromCache,
    })
} else {
    log.Error("Email resolution failed", logger.Fields{
        "email": result.Email,
        "error": result.Error,
    })
}
```

### Batch Email Resolution

```go
// Resolve multiple emails
emails := []string{
    "user1@example.com",
    "user2@example.com",
    "user3@example.com",
}

batchResult, err := resolver.ResolveEmails(ctx, emails)
if err != nil {
    log.Error("Failed to resolve emails", logger.Fields{"error": err})
    return
}

log.Info("Batch resolution completed", logger.Fields{
    "total_emails":   batchResult.TotalEmails,
    "resolved_count": batchResult.ResolvedCount,
    "error_count":    batchResult.ErrorCount,
    "cache_hits":     batchResult.CacheHits,
    "duration":       batchResult.Duration.String(),
})

// Process individual results
for _, result := range batchResult.Results {
    if result.Resolved {
        log.Info("Email resolved", logger.Fields{
            "email":   result.Email,
            "user_id": result.UserID,
        })
    } else {
        log.Error("Email resolution failed", logger.Fields{
            "email": result.Email,
            "error": result.Error,
        })
    }
}
```

### YAML Email Resolution

```go
// Extract and resolve emails from YAML content
yamlContent := []byte(`apiVersion: n9/v1alpha
kind: RoleBinding
metadata:
  name: test-role-binding
  project: test-project
spec:
  users:
    - id: user1@example.com
    - id: user2@example.com
  roles:
    - project-owner`)

batchResult, err := resolver.ResolveEmailsFromYAML(ctx, yamlContent)
if err != nil {
    log.Error("Failed to resolve emails from YAML", logger.Fields{"error": err})
    return
}

log.Info("YAML email resolution completed", logger.Fields{
    "total_emails":   batchResult.TotalEmails,
    "resolved_count": batchResult.ResolvedCount,
})
```

### Result Processing

```go
// Get resolved UserIDs as a map
emailToUserID := resolver.GetResolvedUserIDs(batchResult)
for email, userID := range emailToUserID {
    log.Info("Resolved user", logger.Fields{
        "email":   email,
        "user_id": userID,
    })
}

// Get unresolved emails
unresolved := resolver.GetUnresolvedEmails(batchResult)
if len(unresolved) > 0 {
    log.Error("Unresolved emails", logger.Fields{
        "emails": unresolved,
    })
}
```

## Data Structures

### ResolutionResult

```go
type ResolutionResult struct {
    Email     string        // Normalized email address
    UserID    string        // Resolved UserID (if successful)
    Resolved  bool          // Whether resolution was successful
    Error     error         // Error details (if failed)
    Duration  time.Duration // Resolution duration
    FromCache bool          // Whether result came from cache
}
```

### BatchResolutionResult

```go
type BatchResolutionResult struct {
    Results       []*ResolutionResult // Individual resolution results
    TotalEmails   int                 // Total number of emails processed
    ResolvedCount int                 // Number of successfully resolved emails
    ErrorCount    int                 // Number of failed resolutions
    CacheHits     int                 // Number of cache hits
    Duration      time.Duration       // Total batch processing duration
    Errors        []error             // Collection of all errors
}
```

### UserInfo

```go
type UserInfo struct {
    Email    string // User's email address
    UserID   string // Nobl9 UserID
    Username string // User's username
    FullName string // User's full name
    Active   bool   // Whether user is active
    Found    bool   // Whether user was found
    Error    error  // Error details (if not found)
}
```

## Caching

### Cache Configuration

```go
// Create resolver with custom cache TTL
resolver := resolver.New(client, log)

// Get cache statistics
stats := resolver.GetCacheStats()
log.Info("Cache statistics", logger.Fields{
    "size": stats["size"],
    "ttl":  stats["ttl"],
})

// Clear cache
resolver.ClearCache()
```

### Cache Behavior

The cache provides several benefits:

1. **Performance** - Avoid repeated API calls for the same email
2. **Rate Limiting** - Reduce API rate limit impact
3. **Error Reduction** - Cache successful resolutions
4. **Negative Caching** - Cache "not found" results

### Cache TTL

Default cache TTL is 30 minutes, which can be configured:

```go
// Create cache with custom TTL
cache := resolver.NewUserCache(60 * time.Minute) // 1 hour TTL
```

## Error Handling

### Common Errors

#### API Errors
```
Error: failed to resolve user: API authentication failed
Error: failed to resolve user: rate limit exceeded
Error: failed to resolve user: network timeout
```

#### User Not Found
```
Error: user not found: user does not exist in Nobl9
Error: user not found: user is inactive
```

#### Validation Errors
```
Error: invalid email format: malformed-email
Error: invalid email format: missing-domain
```

### Error Recovery

The resolver implements graceful error recovery:

```go
// Process results with error handling
for _, result := range batchResult.Results {
    if result.Resolved {
        // Process successful resolution
        log.Info("User resolved", logger.Fields{
            "email":   result.Email,
            "user_id": result.UserID,
        })
    } else {
        // Handle resolution failure
        log.Error("User resolution failed", logger.Fields{
            "email": result.Email,
            "error": result.Error,
        })
        
        // Continue processing other users
    }
}
```

## Performance Optimization

### Concurrent Processing

The resolver uses concurrent processing for batch operations:

```go
// Configure concurrent processing
// Default: 10 concurrent requests
// Can be adjusted based on API rate limits

emails := []string{"user1@example.com", "user2@example.com", "..."}
batchResult, err := resolver.ResolveEmails(ctx, emails)
```

### Caching Strategy

Optimal caching configuration:

```go
// Use appropriate TTL based on user data volatility
// Short TTL (5-15 minutes): Frequently changing user data
// Medium TTL (30-60 minutes): Standard user data
// Long TTL (2-4 hours): Stable user data

cache := resolver.NewUserCache(30 * time.Minute)
```

### Rate Limiting

The resolver implements rate limiting to respect API limits:

- **Semaphore-based** - Limits concurrent API calls
- **Configurable** - Adjust based on API rate limits
- **Graceful degradation** - Continues processing despite rate limits

## Logging

### Resolution Logging

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "debug",
  "message": "Resolving email to UserID",
  "email": "user@example.com"
}
```

### Success Logging

```json
{
  "timestamp": "2024-01-15T10:30:46Z",
  "level": "info",
  "message": "User email resolved to UserID",
  "event": "user_resolution",
  "email": "user@example.com",
  "user_id": "okta-user-123",
  "success": true,
  "user_id": "okta-user-123",
  "username": "user",
  "full_name": "Test User",
  "active": true,
  "duration": "150ms"
}
```

### Batch Logging

```json
{
  "timestamp": "2024-01-15T10:30:47Z",
  "level": "info",
  "message": "Batch email resolution completed",
  "total_emails": 5,
  "resolved_count": 4,
  "error_count": 1,
  "cache_hits": 2,
  "duration": "2.5s"
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

// Create resolver
resolver := resolver.New(client, log)
```

### Workflow Integration

```go
// Parse YAML files
parser := parser.New(client, log)
results, err := parser.ParseFiles(ctx, fileInfos)
if err != nil {
    log.Error("Failed to parse files", logger.Fields{"error": err})
    return
}

// Extract and resolve emails from all files
for _, result := range results {
    if result.IsValid {
        // Resolve emails from YAML content
        batchResult, err := resolver.ResolveEmailsFromYAML(ctx, result.FileInfo.Content)
        if err != nil {
            log.Error("Failed to resolve emails", logger.Fields{
                "file":  result.FileInfo.Path,
                "error": err,
            })
            continue
        }

        // Process resolved UserIDs
        emailToUserID := resolver.GetResolvedUserIDs(batchResult)
        log.Info("Resolved users for file", logger.Fields{
            "file":           result.FileInfo.Path,
            "resolved_count": len(emailToUserID),
        })

        // Update YAML with resolved UserIDs
        // (Implementation depends on specific requirements)
    }
}
```

### Error Reporting

```go
// Report resolution errors to GitHub Actions
if len(batchResult.Errors) > 0 {
    log.Error("Email resolution errors", logger.Fields{
        "error_count": len(batchResult.Errors),
        "errors":      batchResult.Errors,
    })
    
    // Set GitHub Actions output
    fmt.Printf("::set-output name=errors::%d\n", len(batchResult.Errors))
    fmt.Printf("::set-output name=unresolved_emails::%s\n", 
        strings.Join(resolver.GetUnresolvedEmails(batchResult), ","))
}
```

## Best Practices

### Email Validation

1. **Pre-validation** - Validate email format before resolution
2. **Normalization** - Normalize email addresses (lowercase, trim)
3. **Deduplication** - Remove duplicate email addresses
4. **Format Checking** - Ensure proper email format

### Caching Strategy

1. **Appropriate TTL** - Set TTL based on data volatility
2. **Monitor Performance** - Track cache hit rates
3. **Clear When Needed** - Clear cache when user data changes
4. **Size Management** - Monitor cache size and memory usage

### Error Handling

1. **Graceful Degradation** - Continue processing despite individual failures
2. **Detailed Logging** - Log all errors with context
3. **User Feedback** - Provide clear error messages
4. **Retry Logic** - Implement retry for transient failures

### Performance

1. **Batch Processing** - Process multiple emails together
2. **Concurrent Resolution** - Use parallel processing
3. **Rate Limiting** - Respect API rate limits
4. **Caching** - Cache results to reduce API calls

### Security

1. **Input Validation** - Validate all email inputs
2. **Error Sanitization** - Don't expose sensitive information in errors
3. **Rate Limiting** - Prevent abuse through rate limiting
4. **Audit Logging** - Log all resolution attempts for audit purposes 