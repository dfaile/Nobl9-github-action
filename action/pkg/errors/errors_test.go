package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		severity  ErrorSeverity
		message   string
		err       error
		expected  *Nobl9Error
	}{
		{
			name:      "config error",
			errorType: ErrorTypeConfig,
			severity:  SeverityHigh,
			message:   "configuration failed",
			err:       fmt.Errorf("invalid config"),
			expected: &Nobl9Error{
				Type:      ErrorTypeConfig,
				Severity:  SeverityHigh,
				Message:   "configuration failed",
				Err:       fmt.Errorf("invalid config"),
				Retryable: false,
			},
		},
		{
			name:      "retryable error",
			errorType: ErrorTypeNetwork,
			severity:  SeverityMedium,
			message:   "network timeout",
			err:       fmt.Errorf("connection timeout"),
			expected: &Nobl9Error{
				Type:      ErrorTypeNetwork,
				Severity:  SeverityMedium,
				Message:   "network timeout",
				Err:       fmt.Errorf("connection timeout"),
				Retryable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.errorType, tt.severity, tt.message, tt.err)

			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Severity, result.Severity)
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.Retryable, result.Retryable)
			assert.Equal(t, tt.expected.Err, result.Err)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

func TestNewWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field": "client_id",
		"value": "invalid",
		"code":  400,
	}

	nobl9Err := NewWithDetails(ErrorTypeConfig, SeverityHigh, "invalid configuration", fmt.Errorf("bad config"), details)

	assert.Equal(t, ErrorTypeConfig, nobl9Err.Type)
	assert.Equal(t, SeverityHigh, nobl9Err.Severity)
	assert.Equal(t, "invalid configuration", nobl9Err.Message)
	assert.Equal(t, details, nobl9Err.Details)
	assert.False(t, nobl9Err.Timestamp.IsZero())
}

func TestNewWithContext(t *testing.T) {
	context := map[string]interface{}{
		"operation": "create_project",
		"project":   "test-project",
		"user":      "test@example.com",
	}

	nobl9Err := NewWithContext(ErrorTypeNobl9API, SeverityMedium, "API call failed", fmt.Errorf("timeout"), context)

	assert.Equal(t, ErrorTypeNobl9API, nobl9Err.Type)
	assert.Equal(t, SeverityMedium, nobl9Err.Severity)
	assert.Equal(t, "API call failed", nobl9Err.Message)
	assert.Equal(t, context, nobl9Err.Context)
	assert.False(t, nobl9Err.Timestamp.IsZero())
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrap(originalErr, ErrorTypeValidation, SeverityLow, "wrapped message")

	assert.Equal(t, ErrorTypeValidation, wrappedErr.Type)
	assert.Equal(t, SeverityLow, wrappedErr.Severity)
	assert.Equal(t, "wrapped message", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Err)
	assert.False(t, wrappedErr.Retryable)
}

func TestWrapWithDetails(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	details := map[string]interface{}{
		"field": "email",
		"value": "invalid@",
	}

	wrappedErr := WrapWithDetails(originalErr, ErrorTypeValidation, SeverityMedium, "validation failed", details)

	assert.Equal(t, ErrorTypeValidation, wrappedErr.Type)
	assert.Equal(t, SeverityMedium, wrappedErr.Severity)
	assert.Equal(t, "validation failed", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Err)
	assert.Equal(t, details, wrappedErr.Details)
}

func TestWrapWithContext(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	context := map[string]interface{}{
		"file": "config.yaml",
		"line": 10,
	}

	wrappedErr := WrapWithContext(originalErr, ErrorTypeFileProcessing, SeverityHigh, "file processing failed", context)

	assert.Equal(t, ErrorTypeFileProcessing, wrappedErr.Type)
	assert.Equal(t, SeverityHigh, wrappedErr.Severity)
	assert.Equal(t, "file processing failed", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Err)
	assert.Equal(t, context, wrappedErr.Context)
}

func TestSpecificErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		err      error
		expected ErrorType
	}{
		{"config error", "config failed", fmt.Errorf("bad config"), ErrorTypeConfig},
		{"validation error", "validation failed", fmt.Errorf("invalid input"), ErrorTypeValidation},
		{"nobl9 api error", "API failed", fmt.Errorf("timeout"), ErrorTypeNobl9API},
		{"file processing error", "file failed", fmt.Errorf("read error"), ErrorTypeFileProcessing},
		{"network error", "network failed", fmt.Errorf("connection lost"), ErrorTypeNetwork},
		{"auth error", "auth failed", fmt.Errorf("invalid token"), ErrorTypeAuth},
		{"rate limit error", "rate limited", fmt.Errorf("429"), ErrorTypeRateLimit},
		{"timeout error", "timeout", fmt.Errorf("deadline exceeded"), ErrorTypeTimeout},
		{"user resolution error", "user not found", fmt.Errorf("404"), ErrorTypeUserResolution},
		{"manifest error", "manifest invalid", fmt.Errorf("parse error"), ErrorTypeManifest},
		{"retryable error", "retryable", fmt.Errorf("temporary"), ErrorTypeRetryable},
		{"non retryable error", "permanent", fmt.Errorf("fatal"), ErrorTypeNonRetryable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nobl9Err *Nobl9Error

			switch tt.expected {
			case ErrorTypeConfig:
				nobl9Err = NewConfigError(tt.message, tt.err)
			case ErrorTypeValidation:
				nobl9Err = NewValidationError(tt.message, tt.err)
			case ErrorTypeNobl9API:
				nobl9Err = NewNobl9APIError(tt.message, tt.err)
			case ErrorTypeFileProcessing:
				nobl9Err = NewFileProcessingError(tt.message, tt.err)
			case ErrorTypeNetwork:
				nobl9Err = NewNetworkError(tt.message, tt.err)
			case ErrorTypeAuth:
				nobl9Err = NewAuthError(tt.message, tt.err)
			case ErrorTypeRateLimit:
				nobl9Err = NewRateLimitError(tt.message, tt.err)
			case ErrorTypeTimeout:
				nobl9Err = NewTimeoutError(tt.message, tt.err)
			case ErrorTypeUserResolution:
				nobl9Err = NewUserResolutionError(tt.message, tt.err)
			case ErrorTypeManifest:
				nobl9Err = NewManifestError(tt.message, tt.err)
			case ErrorTypeRetryable:
				nobl9Err = NewRetryableError(tt.message, tt.err)
			case ErrorTypeNonRetryable:
				nobl9Err = NewNonRetryableError(tt.message, tt.err)
			}

			assert.Equal(t, tt.expected, nobl9Err.Type)
			assert.Equal(t, tt.message, nobl9Err.Message)
			assert.Equal(t, tt.err, nobl9Err.Err)
		})
	}
}

func TestNobl9Error_Error(t *testing.T) {
	originalErr := fmt.Errorf("underlying error")
	nobl9Err := New(ErrorTypeConfig, SeverityHigh, "configuration failed", originalErr)

	errorString := nobl9Err.Error()

	assert.Contains(t, errorString, "configuration failed")
	assert.Contains(t, errorString, "underlying error")
	assert.Contains(t, errorString, string(ErrorTypeConfig))
}

func TestNobl9Error_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	nobl9Err := New(ErrorTypeValidation, SeverityMedium, "validation failed", originalErr)

	unwrapped := nobl9Err.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
}

func TestNobl9Error_IsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		{"retryable error", ErrorTypeRetryable, true},
		{"network error", ErrorTypeNetwork, true},
		{"rate limit error", ErrorTypeRateLimit, true},
		{"timeout error", ErrorTypeTimeout, true},
		{"config error", ErrorTypeConfig, false},
		{"validation error", ErrorTypeValidation, false},
		{"auth error", ErrorTypeAuth, false},
		{"non retryable error", ErrorTypeNonRetryable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nobl9Err := New(tt.errorType, SeverityMedium, "test error", fmt.Errorf("test"))
			assert.Equal(t, tt.expected, nobl9Err.IsRetryable())
		})
	}
}

func TestNobl9Error_Getters(t *testing.T) {
	details := map[string]interface{}{"key": "value"}
	context := map[string]interface{}{"ctx": "data"}

	nobl9Err := NewWithDetails(ErrorTypeConfig, SeverityHigh, "test", fmt.Errorf("test"), details)
	nobl9Err.Context = context

	assert.Equal(t, ErrorTypeConfig, nobl9Err.GetType())
	assert.Equal(t, SeverityHigh, nobl9Err.GetSeverity())
	assert.Equal(t, details, nobl9Err.GetDetails())
	assert.Equal(t, context, nobl9Err.GetContext())
}

func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nobl9 error", New(ErrorTypeConfig, SeverityMedium, "test", nil), true},
		{"standard error", fmt.Errorf("standard error"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNobl9Error(tt.err))
		})
	}
}

func TestRetryableErrorDetection(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"retryable error", NewRetryableError("retryable", nil), true},
		{"network error", NewNetworkError("network", nil), true},
		{"rate limit error", NewRateLimitError("rate limit", nil), true},
		{"timeout error", NewTimeoutError("timeout", nil), true},
		{"non retryable error", NewNonRetryableError("permanent", nil), false},
		{"config error", NewConfigError("config", nil), false},
		{"standard error", fmt.Errorf("standard"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRetryableError(tt.err))
		})
	}
}

func TestAuthErrorDetection(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"auth error", NewAuthError("auth failed", nil), true},
		{"config error", NewConfigError("config failed", nil), false},
		{"standard error", fmt.Errorf("standard"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAuthError(tt.err))
		})
	}
}

func TestRateLimitErrorDetection(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"rate limit error", NewRateLimitError("rate limited", nil), true},
		{"timeout error", NewTimeoutError("timeout", nil), false},
		{"standard error", fmt.Errorf("standard"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRateLimitError(tt.err))
		})
	}
}

func TestTimeoutErrorDetection(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"timeout error", NewTimeoutError("timeout", nil), true},
		{"rate limit error", NewRateLimitError("rate limited", nil), false},
		{"standard error", fmt.Errorf("standard"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsTimeoutError(tt.err))
		})
	}
}

func TestErrorAggregator(t *testing.T) {
	aggregator := NewErrorAggregator()

	// Test initial state
	assert.False(t, aggregator.HasErrors())
	assert.False(t, aggregator.HasCriticalErrors())
	assert.Empty(t, aggregator.GetErrors())

	// Add errors
	err1 := NewConfigError("config error", nil)
	err2 := NewValidationError("validation error", nil)
	err3 := New(ErrorTypeNobl9API, SeverityCritical, "critical error", nil)

	aggregator.AddError(err1)
	aggregator.AddError(err2)
	aggregator.AddError(err3)

	// Test error collection
	assert.True(t, aggregator.HasErrors())
	assert.True(t, aggregator.HasCriticalErrors())
	assert.Len(t, aggregator.GetErrors(), 3)

	// Test error filtering
	configErrors := aggregator.GetErrorsByType(ErrorTypeConfig)
	assert.Len(t, configErrors, 1)
	assert.Equal(t, ErrorTypeConfig, configErrors[0].Type)

	criticalErrors := aggregator.GetErrorsBySeverity(SeverityCritical)
	assert.Len(t, criticalErrors, 1)
	assert.Equal(t, SeverityCritical, criticalErrors[0].Severity)

	retryableErrors := aggregator.GetRetryableErrors()
	assert.Len(t, retryableErrors, 0) // None of our test errors are retryable

	// Test error summary
	summary := aggregator.GetErrorSummary()
	assert.Equal(t, 3, summary["total_errors"])
	assert.Equal(t, 1, summary["critical_errors"])
	assert.Equal(t, 0, summary["retryable_errors"])

	// Check error types map
	errorTypes := summary["error_types"].(map[string]int)
	assert.Equal(t, 1, errorTypes["configuration"])
	assert.Equal(t, 1, errorTypes["validation"])
	assert.Equal(t, 1, errorTypes["nobl9_api"])
}

func TestErrorAggregator_AddErrorFromErr(t *testing.T) {
	aggregator := NewErrorAggregator()

	standardErr := fmt.Errorf("standard error")
	aggregator.AddErrorFromErr(standardErr, ErrorTypeFileProcessing, SeverityMedium, "file processing failed")

	assert.True(t, aggregator.HasErrors())
	errors := aggregator.GetErrors()
	assert.Len(t, errors, 1)
	assert.Equal(t, ErrorTypeFileProcessing, errors[0].Type)
	assert.Equal(t, SeverityMedium, errors[0].Severity)
	assert.Equal(t, "file processing failed", errors[0].Message)
	assert.Equal(t, standardErr, errors[0].Err)
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nobl9 error",
			err:      New(ErrorTypeConfig, SeverityHigh, "config failed", fmt.Errorf("bad config")),
			expected: "config failed: bad config",
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: "standard error",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatError(tt.err)
			if tt.err == nil {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.expected)
			}
		})
	}
}

func TestIsRetryableErrorType(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		{"retryable", ErrorTypeRetryable, true},
		{"network", ErrorTypeNetwork, true},
		{"rate limit", ErrorTypeRateLimit, true},
		{"timeout", ErrorTypeTimeout, true},
		{"config", ErrorTypeConfig, false},
		{"validation", ErrorTypeValidation, false},
		{"auth", ErrorTypeAuth, false},
		{"non retryable", ErrorTypeNonRetryable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isRetryableErrorType(tt.errorType))
		})
	}
}

func TestErrorSeverityLevels(t *testing.T) {
	// Test that severity levels are properly defined
	assert.Equal(t, ErrorSeverity("low"), SeverityLow)
	assert.Equal(t, ErrorSeverity("medium"), SeverityMedium)
	assert.Equal(t, ErrorSeverity("high"), SeverityHigh)
	assert.Equal(t, ErrorSeverity("critical"), SeverityCritical)
}

func TestErrorTypes(t *testing.T) {
	// Test that error types are properly defined
	assert.Equal(t, ErrorType("configuration"), ErrorTypeConfig)
	assert.Equal(t, ErrorType("validation"), ErrorTypeValidation)
	assert.Equal(t, ErrorType("nobl9_api"), ErrorTypeNobl9API)
	assert.Equal(t, ErrorType("file_processing"), ErrorTypeFileProcessing)
	assert.Equal(t, ErrorType("network"), ErrorTypeNetwork)
	assert.Equal(t, ErrorType("authentication"), ErrorTypeAuth)
	assert.Equal(t, ErrorType("rate_limit"), ErrorTypeRateLimit)
	assert.Equal(t, ErrorType("timeout"), ErrorTypeTimeout)
	assert.Equal(t, ErrorType("user_resolution"), ErrorTypeUserResolution)
	assert.Equal(t, ErrorType("manifest"), ErrorTypeManifest)
	assert.Equal(t, ErrorType("retryable"), ErrorTypeRetryable)
	assert.Equal(t, ErrorType("non_retryable"), ErrorTypeNonRetryable)
}

func TestErrorAggregator_EmptyState(t *testing.T) {
	aggregator := NewErrorAggregator()

	// Test empty state methods
	assert.Empty(t, aggregator.GetErrors())
	assert.Empty(t, aggregator.GetErrorsByType(ErrorTypeConfig))
	assert.Empty(t, aggregator.GetErrorsBySeverity(SeverityHigh))
	assert.Empty(t, aggregator.GetRetryableErrors())

	summary := aggregator.GetErrorSummary()
	assert.Equal(t, 0, summary["total_errors"])
	assert.Equal(t, 0, summary["critical_errors"])
	assert.Equal(t, 0, summary["retryable_errors"])
}

func TestErrorAggregator_ComplexScenario(t *testing.T) {
	aggregator := NewErrorAggregator()

	// Add various types of errors
	aggregator.AddError(NewConfigError("config 1", nil))
	aggregator.AddError(NewConfigError("config 2", nil))
	aggregator.AddError(NewValidationError("validation 1", nil))
	aggregator.AddError(NewNetworkError("network 1", nil))
	aggregator.AddError(New(ErrorTypeNobl9API, SeverityCritical, "critical 1", nil))
	aggregator.AddError(New(ErrorTypeNobl9API, SeverityCritical, "critical 2", nil))

	// Test filtering
	configErrors := aggregator.GetErrorsByType(ErrorTypeConfig)
	assert.Len(t, configErrors, 2)

	criticalErrors := aggregator.GetErrorsBySeverity(SeverityCritical)
	assert.Len(t, criticalErrors, 2)

	retryableErrors := aggregator.GetRetryableErrors()
	assert.Len(t, retryableErrors, 1) // Only network error is retryable

	// Test summary
	summary := aggregator.GetErrorSummary()
	assert.Equal(t, 6, summary["total_errors"])
	assert.Equal(t, 2, summary["critical_errors"])
	assert.Equal(t, 1, summary["retryable_errors"])

	// Check error types map
	errorTypes := summary["error_types"].(map[string]int)
	assert.Equal(t, 2, errorTypes["configuration"])
	assert.Equal(t, 1, errorTypes["validation"])
	assert.Equal(t, 1, errorTypes["network"])
	assert.Equal(t, 2, errorTypes["nobl9_api"])
}
