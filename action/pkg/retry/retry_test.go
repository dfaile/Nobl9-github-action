package retry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/your-org/nobl9-action/pkg/logger"
)

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()

	if policy.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", policy.MaxAttempts)
	}

	if policy.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay 1s, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay 30s, got %v", policy.MaxDelay)
	}

	if policy.BackoffFactor != 2.0 {
		t.Errorf("expected BackoffFactor 2.0, got %f", policy.BackoffFactor)
	}

	if policy.JitterFactor != 0.1 {
		t.Errorf("expected JitterFactor 0.1, got %f", policy.JitterFactor)
	}

	if len(policy.RetryableErrors) == 0 {
		t.Error("expected RetryableErrors to be non-empty")
	}
}

func TestNewPolicy(t *testing.T) {
	policy := NewPolicy(5, 2*time.Second, 60*time.Second, 1.5, 0.2)

	if policy.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts 5, got %d", policy.MaxAttempts)
	}

	if policy.InitialDelay != 2*time.Second {
		t.Errorf("expected InitialDelay 2s, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 60*time.Second {
		t.Errorf("expected MaxDelay 60s, got %v", policy.MaxDelay)
	}

	if policy.BackoffFactor != 1.5 {
		t.Errorf("expected BackoffFactor 1.5, got %f", policy.BackoffFactor)
	}

	if policy.JitterFactor != 0.2 {
		t.Errorf("expected JitterFactor 0.2, got %f", policy.JitterFactor)
	}
}

func TestRetrySuccess(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	policy := DefaultPolicy()

	attempts := 0
	fn := func(ctx context.Context) (interface{}, error) {
		attempts++
		if attempts == 1 {
			return nil, fmt.Errorf("temporary error")
		}
		return "success", nil
	}

	result, err := Retry(context.Background(), policy, log, "test operation", fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if !result.Success {
		t.Error("expected success")
	}

	if result.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", result.Attempts)
	}

	if result.FinalResult != "success" {
		t.Errorf("expected result 'success', got %v", result.FinalResult)
	}
}

func TestRetryFailure(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	policy := DefaultPolicy()

	fn := func(ctx context.Context) (interface{}, error) {
		return nil, fmt.Errorf("permanent error")
	}

	result, err := Retry(context.Background(), policy, log, "test operation", fn)
	if err == nil {
		t.Error("expected error")
		return
	}

	if result.Success {
		t.Error("expected failure")
	}

	if result.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", result.Attempts)
	}

	if result.LastError == nil {
		t.Error("expected last error")
	}
}

func TestRetryWithNonRetryableError(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	policy := DefaultPolicy()

	fn := func(ctx context.Context) (interface{}, error) {
		return nil, fmt.Errorf("authentication failed")
	}

	result, err := Retry(context.Background(), policy, log, "test operation", fn)
	if err == nil {
		t.Error("expected error")
		return
	}

	if result.Success {
		t.Error("expected failure")
	}

	if result.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", result.Attempts)
	}
}

func TestRetryWithContextCancellation(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	policy := DefaultPolicy()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fn := func(ctx context.Context) (interface{}, error) {
		cancel() // Cancel context during first attempt
		return nil, fmt.Errorf("temporary error")
	}

	result, err := Retry(ctx, policy, log, "test operation", fn)
	if err == nil {
		t.Error("expected error")
		return
	}

	if !result.Success {
		t.Error("expected success (context cancellation should not be treated as failure)")
	}
}

func TestRetryWithResult(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	policy := DefaultPolicy()

	fn := func(ctx context.Context) (interface{}, error) {
		return "success", nil
	}

	result, err := RetryWithResult(context.Background(), policy, log, "test operation", fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}
}

func TestRetrySimple(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)

	fn := func(ctx context.Context) (interface{}, error) {
		return "success", nil
	}

	result, err := RetrySimple(context.Background(), log, "test operation", fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name              string
		errorMsg          string
		retryablePatterns []string
		expected          bool
	}{
		{
			name:              "retryable timeout",
			errorMsg:          "connection timeout",
			retryablePatterns: []string{"timeout"},
			expected:          true,
		},
		{
			name:              "retryable rate limit",
			errorMsg:          "rate limit exceeded",
			retryablePatterns: []string{"rate limit"},
			expected:          true,
		},
		{
			name:              "retryable HTTP 429",
			errorMsg:          "HTTP 429 Too Many Requests",
			retryablePatterns: []string{"429"},
			expected:          true,
		},
		{
			name:              "non-retryable auth error",
			errorMsg:          "authentication failed",
			retryablePatterns: []string{"timeout", "rate limit"},
			expected:          false,
		},
		{
			name:              "empty error",
			errorMsg:          "",
			retryablePatterns: []string{"timeout"},
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errorMsg)
			result := isRetryableError(err, tt.retryablePatterns)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	policy := &Policy{
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.0, // No jitter for deterministic testing
	}

	// Test first attempt
	delay := calculateDelay(1, policy)
	if delay != 1*time.Second {
		t.Errorf("expected 1s delay for first attempt, got %v", delay)
	}

	// Test second attempt
	delay = calculateDelay(2, policy)
	if delay != 2*time.Second {
		t.Errorf("expected 2s delay for second attempt, got %v", delay)
	}

	// Test third attempt
	delay = calculateDelay(3, policy)
	if delay != 4*time.Second {
		t.Errorf("expected 4s delay for third attempt, got %v", delay)
	}

	// Test max delay limit
	delay = calculateDelay(10, policy)
	if delay != 10*time.Second {
		t.Errorf("expected 10s delay (max), got %v", delay)
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	policy := &Policy{
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
	}

	// Test that jitter is applied (delay should vary)
	delays := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		delay := calculateDelay(2, policy)
		delays[delay] = true
	}

	// With jitter, we should see some variation
	if len(delays) < 2 {
		t.Error("expected jitter to create variation in delays")
	}
}

func TestRetryableAPIOperation(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	policy := DefaultPolicy()

	operation := NewRetryableAPIOperation(policy, log)

	fn := func(ctx context.Context) (interface{}, error) {
		return "success", nil
	}

	result, err := operation.Execute(context.Background(), "test operation", fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}
}

func TestRetryableAPIOperationWithCustomPolicy(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	defaultPolicy := DefaultPolicy()
	customPolicy := NewPolicy(5, 2*time.Second, 60*time.Second, 1.5, 0.2)

	operation := NewRetryableAPIOperation(defaultPolicy, log)

	fn := func(ctx context.Context) (interface{}, error) {
		return "success", nil
	}

	result, err := operation.ExecuteWithCustomPolicy(context.Background(), customPolicy, "test operation", fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}
}

func TestRetryableError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	retryableErr := NewRetryableError("retryable operation failed", originalErr)

	if retryableErr.Error() != "retryable operation failed: original error" {
		t.Errorf("unexpected error message: %s", retryableErr.Error())
	}

	if retryableErr.Unwrap() != originalErr {
		t.Errorf("unexpected unwrapped error")
	}

	if !IsRetryableError(retryableErr) {
		t.Error("expected IsRetryableError to return true")
	}

	// Test non-retryable error
	regularErr := fmt.Errorf("regular error")
	if IsRetryableError(regularErr) {
		t.Error("expected IsRetryableError to return false for regular error")
	}
}

func TestRetryableErrorPatterns(t *testing.T) {
	patterns := RetryableErrorPatterns()

	expectedPatterns := []string{
		"timeout",
		"connection refused",
		"network error",
		"rate limit",
		"429",
		"503",
		"502",
		"500",
	}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected pattern '%s' not found", expected)
		}
	}
}

func TestCreatePolicyForAPI(t *testing.T) {
	policy := CreatePolicyForAPI(5)

	if policy.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts 5, got %d", policy.MaxAttempts)
	}

	if policy.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay 1s, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay 30s, got %v", policy.MaxDelay)
	}

	if policy.BackoffFactor != 2.0 {
		t.Errorf("expected BackoffFactor 2.0, got %f", policy.BackoffFactor)
	}

	if policy.JitterFactor != 0.1 {
		t.Errorf("expected JitterFactor 0.1, got %f", policy.JitterFactor)
	}
}

func TestCreatePolicyForNetwork(t *testing.T) {
	policy := CreatePolicyForNetwork(3)

	if policy.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", policy.MaxAttempts)
	}

	if policy.InitialDelay != 500*time.Millisecond {
		t.Errorf("expected InitialDelay 500ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 10*time.Second {
		t.Errorf("expected MaxDelay 10s, got %v", policy.MaxDelay)
	}

	if policy.BackoffFactor != 1.5 {
		t.Errorf("expected BackoffFactor 1.5, got %f", policy.BackoffFactor)
	}

	if policy.JitterFactor != 0.2 {
		t.Errorf("expected JitterFactor 0.2, got %f", policy.JitterFactor)
	}
}

func TestCreatePolicyForRateLimit(t *testing.T) {
	policy := CreatePolicyForRateLimit(4)

	if policy.MaxAttempts != 4 {
		t.Errorf("expected MaxAttempts 4, got %d", policy.MaxAttempts)
	}

	if policy.InitialDelay != 2*time.Second {
		t.Errorf("expected InitialDelay 2s, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 60*time.Second {
		t.Errorf("expected MaxDelay 60s, got %v", policy.MaxDelay)
	}

	if policy.BackoffFactor != 2.0 {
		t.Errorf("expected BackoffFactor 2.0, got %f", policy.BackoffFactor)
	}

	if policy.JitterFactor != 0.1 {
		t.Errorf("expected JitterFactor 0.1, got %f", policy.JitterFactor)
	}
}

func TestRetryWithNilPolicy(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)

	fn := func(ctx context.Context) (interface{}, error) {
		return "success", nil
	}

	result, err := Retry(context.Background(), nil, log, "test operation", fn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if !result.Success {
		t.Error("expected success")
	}

	if result.FinalResult != "success" {
		t.Errorf("expected result 'success', got %v", result.FinalResult)
	}
}

func TestRetryableAPIOperationPolicyManagement(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	policy1 := DefaultPolicy()
	policy2 := NewPolicy(5, 2*time.Second, 60*time.Second, 1.5, 0.2)

	operation := NewRetryableAPIOperation(policy1, log)

	// Test GetPolicy
	if operation.GetPolicy() != policy1 {
		t.Error("expected GetPolicy to return the original policy")
	}

	// Test SetPolicy
	operation.SetPolicy(policy2)
	if operation.GetPolicy() != policy2 {
		t.Error("expected GetPolicy to return the new policy after SetPolicy")
	}
}
