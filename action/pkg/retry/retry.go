package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/your-org/nobl9-action/pkg/errors"
	"github.com/your-org/nobl9-action/pkg/logger"
)

// Policy defines retry behavior
type Policy struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	BackoffFactor   float64       // Exponential backoff factor
	JitterFactor    float64       // Jitter factor for randomization (0.0 to 1.0)
	RetryableErrors []string      // List of error patterns that should trigger retries
}

// RetryResult represents the result of a retry operation
type RetryResult struct {
	Attempts    int           // Number of attempts made
	Success     bool          // Whether the operation succeeded
	LastError   error         // Last error encountered
	TotalDelay  time.Duration // Total delay across all retries
	FinalResult interface{}   // Final result of the operation
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) (interface{}, error)

// DefaultPolicy returns a default retry policy
func DefaultPolicy() *Policy {
	return &Policy{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"network error",
			"rate limit",
			"429",
			"503",
			"502",
			"500",
		},
	}
}

// NewPolicy creates a new retry policy with custom settings
func NewPolicy(maxAttempts int, initialDelay, maxDelay time.Duration, backoffFactor, jitterFactor float64) *Policy {
	return &Policy{
		MaxAttempts:   maxAttempts,
		InitialDelay:  initialDelay,
		MaxDelay:      maxDelay,
		BackoffFactor: backoffFactor,
		JitterFactor:  jitterFactor,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"network error",
			"rate limit",
			"429",
			"503",
			"502",
			"500",
		},
	}
}

// Retry executes a function with retry logic
func Retry(ctx context.Context, policy *Policy, log *logger.Logger, operation string, fn RetryableFunc) (*RetryResult, error) {
	if policy == nil {
		policy = DefaultPolicy()
	}

	result := &RetryResult{
		Attempts: 0,
		Success:  false,
	}

	var lastError error
	var finalResult interface{}

	log.Debug("Starting retry operation", logger.Fields{
		"operation":     operation,
		"max_attempts":  policy.MaxAttempts,
		"initial_delay": policy.InitialDelay.String(),
	})

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		result.Attempts = attempt

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			cancelErr := errors.NewTimeoutError("operation cancelled", ctx.Err())
			return result, cancelErr
		default:
		}

		// Execute the operation
		log.Debug("Executing operation attempt", logger.Fields{
			"operation": operation,
			"attempt":   attempt,
		})

		finalResult, lastError = fn(ctx)

		// If successful, return immediately
		if lastError == nil {
			result.Success = true
			result.FinalResult = finalResult
			result.LastError = nil

			log.Info("Operation succeeded", logger.Fields{
				"operation": operation,
				"attempts":  attempt,
			})

			return result, nil
		}

		// Check if error is retryable
		if !isRetryableError(lastError, policy.RetryableErrors) {
			result.LastError = lastError

			// Log non-retryable error with detailed information
			log.LogDetailedError(lastError, operation, map[string]interface{}{
				"attempt":        attempt,
				"max_attempts":   policy.MaxAttempts,
				"error_category": "non_retryable",
			}, logger.Fields{
				"operation": operation,
				"attempt":   attempt,
			})

			return result, errors.NewNonRetryableError(fmt.Sprintf("non-retryable error in %s", operation), lastError)
		}

		// Log the retryable error with detailed information
		log.LogDetailedError(lastError, operation, map[string]interface{}{
			"attempt":        attempt,
			"max_attempts":   policy.MaxAttempts,
			"error_category": "retryable",
		}, logger.Fields{
			"operation": operation,
			"attempt":   attempt,
		})

		// If this is the last attempt, don't wait
		if attempt == policy.MaxAttempts {
			break
		}

		// Calculate delay for next attempt
		delay := calculateDelay(attempt, policy)
		result.TotalDelay += delay

		log.Debug("Waiting before retry", logger.Fields{
			"operation": operation,
			"attempt":   attempt,
			"delay":     delay.String(),
		})

		// Wait for the delay or context cancellation
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			cancelErr := errors.NewTimeoutError("operation cancelled during retry", ctx.Err())
			return result, cancelErr
		}
	}

	// All attempts failed
	result.LastError = lastError
	result.FinalResult = finalResult

	// Log final failure with comprehensive error information
	log.LogDetailedError(lastError, operation, map[string]interface{}{
		"attempts":       result.Attempts,
		"total_delay":    result.TotalDelay.String(),
		"max_attempts":   policy.MaxAttempts,
		"error_category": "final_failure",
	}, logger.Fields{
		"operation":   operation,
		"attempts":    result.Attempts,
		"total_delay": result.TotalDelay.String(),
	})

	return result, errors.NewRetryableError(fmt.Sprintf("operation %s failed after %d attempts", operation, result.Attempts), lastError)
}

// RetryWithResult executes a function with retry logic and returns the result
func RetryWithResult(ctx context.Context, policy *Policy, log *logger.Logger, operation string, fn RetryableFunc) (interface{}, error) {
	result, err := Retry(ctx, policy, log, operation, fn)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("operation failed: %w", result.LastError)
	}

	return result.FinalResult, nil
}

// RetrySimple executes a function with simple retry logic
func RetrySimple(ctx context.Context, log *logger.Logger, operation string, fn RetryableFunc) (interface{}, error) {
	return RetryWithResult(ctx, DefaultPolicy(), log, operation, fn)
}

// isRetryableError checks if an error should trigger a retry
func isRetryableError(err error, retryablePatterns []string) bool {
	if err == nil {
		return false
	}

	// Check if it's a Nobl9 error first
	if errors.IsNobl9Error(err) {
		return errors.IsRetryableError(err)
	}

	errorMsg := err.Error()
	for _, pattern := range retryablePatterns {
		if containsIgnoreCase(errorMsg, pattern) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if a string contains another string (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive check
	// In a production environment, you might want to use a more robust approach
	return len(s) >= len(substr) &&
		(string(s[:len(substr)]) == substr ||
			string(s[len(s)-len(substr):]) == substr ||
			containsSubstring(s, substr))
}

// containsSubstring is a simple substring check
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// calculateDelay calculates the delay for a retry attempt
func calculateDelay(attempt int, policy *Policy) time.Duration {
	// Calculate exponential backoff
	delay := float64(policy.InitialDelay) * math.Pow(policy.BackoffFactor, float64(attempt-1))

	// Apply maximum delay limit
	if delay > float64(policy.MaxDelay) {
		delay = float64(policy.MaxDelay)
	}

	// Apply jitter for randomization
	if policy.JitterFactor > 0 {
		jitter := delay * policy.JitterFactor
		delay += (rand.Float64() * 2 * jitter) - jitter
	}

	// Ensure delay is not negative
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// RetryableAPIOperation represents an API operation that can be retried
type RetryableAPIOperation struct {
	policy *Policy
	logger *logger.Logger
}

// NewRetryableAPIOperation creates a new retryable API operation
func NewRetryableAPIOperation(policy *Policy, log *logger.Logger) *RetryableAPIOperation {
	return &RetryableAPIOperation{
		policy: policy,
		logger: log,
	}
}

// Execute executes an API operation with retry logic
func (r *RetryableAPIOperation) Execute(ctx context.Context, operation string, fn RetryableFunc) (interface{}, error) {
	return RetryWithResult(ctx, r.policy, r.logger, operation, fn)
}

// ExecuteWithCustomPolicy executes an API operation with a custom retry policy
func (r *RetryableAPIOperation) ExecuteWithCustomPolicy(ctx context.Context, policy *Policy, operation string, fn RetryableFunc) (interface{}, error) {
	if policy == nil {
		policy = r.policy
	}
	return RetryWithResult(ctx, policy, r.logger, operation, fn)
}

// GetPolicy returns the current retry policy
func (r *RetryableAPIOperation) GetPolicy() *Policy {
	return r.policy
}

// SetPolicy sets a new retry policy
func (r *RetryableAPIOperation) SetPolicy(policy *Policy) {
	r.policy = policy
}

// RetryableError represents a retryable error
type RetryableError struct {
	Message string
	Err     error
}

// Error implements the error interface
func (e *RetryableError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError creates a new retryable error
func NewRetryableError(message string, err error) *RetryableError {
	return &RetryableError{
		Message: message,
		Err:     err,
	}
}

// IsRetryableError checks if an error is a retryable error
func IsRetryableError(err error) bool {
	_, ok := err.(*RetryableError)
	return ok
}

// RetryableErrorPatterns returns common retryable error patterns
func RetryableErrorPatterns() []string {
	return []string{
		"timeout",
		"connection refused",
		"network error",
		"rate limit",
		"429",
		"503",
		"502",
		"500",
		"temporary failure",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
		"too many requests",
		"internal server error",
	}
}

// CreatePolicyForAPI creates a retry policy optimized for API operations
func CreatePolicyForAPI(maxAttempts int) *Policy {
	return &Policy{
		MaxAttempts:     maxAttempts,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		RetryableErrors: RetryableErrorPatterns(),
	}
}

// CreatePolicyForNetwork creates a retry policy optimized for network operations
func CreatePolicyForNetwork(maxAttempts int) *Policy {
	return &Policy{
		MaxAttempts:   maxAttempts,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 1.5,
		JitterFactor:  0.2,
		RetryableErrors: []string{
			"connection refused",
			"network error",
			"timeout",
			"connection reset",
			"no route to host",
		},
	}
}

// CreatePolicyForRateLimit creates a retry policy optimized for rate limiting
func CreatePolicyForRateLimit(maxAttempts int) *Policy {
	return &Policy{
		MaxAttempts:   maxAttempts,
		InitialDelay:  2 * time.Second,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
		RetryableErrors: []string{
			"rate limit",
			"429",
			"too many requests",
		},
	}
}
