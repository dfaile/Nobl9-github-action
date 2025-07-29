package errors

import (
	"fmt"
	"strings"
	"time"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Configuration errors
	ErrorTypeConfig ErrorType = "configuration"

	// Validation errors
	ErrorTypeValidation ErrorType = "validation"

	// Nobl9 API errors
	ErrorTypeNobl9API ErrorType = "nobl9_api"

	// File processing errors
	ErrorTypeFileProcessing ErrorType = "file_processing"

	// Network errors
	ErrorTypeNetwork ErrorType = "network"

	// Authentication errors
	ErrorTypeAuth ErrorType = "authentication"

	// Rate limiting errors
	ErrorTypeRateLimit ErrorType = "rate_limit"

	// Timeout errors
	ErrorTypeTimeout ErrorType = "timeout"

	// User resolution errors
	ErrorTypeUserResolution ErrorType = "user_resolution"

	// Manifest errors
	ErrorTypeManifest ErrorType = "manifest"

	// Retryable errors
	ErrorTypeRetryable ErrorType = "retryable"

	// Non-retryable errors
	ErrorTypeNonRetryable ErrorType = "non_retryable"
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// Nobl9Error represents a comprehensive error with additional context
type Nobl9Error struct {
	Type      ErrorType
	Severity  ErrorSeverity
	Message   string
	Details   map[string]interface{}
	Timestamp time.Time
	Retryable bool
	Err       error
	Context   map[string]interface{}
}

// Error implements the error interface
func (e *Nobl9Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *Nobl9Error) Unwrap() error {
	return e.Err
}

// IsRetryable returns whether the error is retryable
func (e *Nobl9Error) IsRetryable() bool {
	return e.Retryable
}

// GetType returns the error type
func (e *Nobl9Error) GetType() ErrorType {
	return e.Type
}

// GetSeverity returns the error severity
func (e *Nobl9Error) GetSeverity() ErrorSeverity {
	return e.Severity
}

// GetDetails returns the error details
func (e *Nobl9Error) GetDetails() map[string]interface{} {
	return e.Details
}

// GetContext returns the error context
func (e *Nobl9Error) GetContext() map[string]interface{} {
	return e.Context
}

// New creates a new Nobl9Error
func New(errorType ErrorType, severity ErrorSeverity, message string, err error) *Nobl9Error {
	return &Nobl9Error{
		Type:      errorType,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Retryable: isRetryableErrorType(errorType),
		Err:       err,
		Details:   make(map[string]interface{}),
		Context:   make(map[string]interface{}),
	}
}

// NewWithDetails creates a new Nobl9Error with details
func NewWithDetails(errorType ErrorType, severity ErrorSeverity, message string, err error, details map[string]interface{}) *Nobl9Error {
	e := New(errorType, severity, message, err)
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// NewWithContext creates a new Nobl9Error with context
func NewWithContext(errorType ErrorType, severity ErrorSeverity, message string, err error, context map[string]interface{}) *Nobl9Error {
	e := New(errorType, severity, message, err)
	for k, v := range context {
		e.Context[k] = v
	}
	return e
}

// Wrap wraps an existing error with Nobl9 error context
func Wrap(err error, errorType ErrorType, severity ErrorSeverity, message string) *Nobl9Error {
	return New(errorType, severity, message, err)
}

// WrapWithDetails wraps an existing error with Nobl9 error context and details
func WrapWithDetails(err error, errorType ErrorType, severity ErrorSeverity, message string, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(errorType, severity, message, err, details)
}

// WrapWithContext wraps an existing error with Nobl9 error context and context
func WrapWithContext(err error, errorType ErrorType, severity ErrorSeverity, message string, context map[string]interface{}) *Nobl9Error {
	return NewWithContext(errorType, severity, message, err, context)
}

// Configuration errors
func NewConfigError(message string, err error) *Nobl9Error {
	return New(ErrorTypeConfig, SeverityHigh, message, err)
}

func NewConfigErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeConfig, SeverityHigh, message, err, details)
}

// Validation errors
func NewValidationError(message string, err error) *Nobl9Error {
	return New(ErrorTypeValidation, SeverityMedium, message, err)
}

func NewValidationErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeValidation, SeverityMedium, message, err, details)
}

// Nobl9 API errors
func NewNobl9APIError(message string, err error) *Nobl9Error {
	return New(ErrorTypeNobl9API, SeverityHigh, message, err)
}

func NewNobl9APIErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeNobl9API, SeverityHigh, message, err, details)
}

// File processing errors
func NewFileProcessingError(message string, err error) *Nobl9Error {
	return New(ErrorTypeFileProcessing, SeverityMedium, message, err)
}

func NewFileProcessingErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeFileProcessing, SeverityMedium, message, err, details)
}

// Network errors
func NewNetworkError(message string, err error) *Nobl9Error {
	return New(ErrorTypeNetwork, SeverityMedium, message, err)
}

func NewNetworkErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeNetwork, SeverityMedium, message, err, details)
}

// Authentication errors
func NewAuthError(message string, err error) *Nobl9Error {
	return New(ErrorTypeAuth, SeverityCritical, message, err)
}

func NewAuthErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeAuth, SeverityCritical, message, err, details)
}

// Rate limiting errors
func NewRateLimitError(message string, err error) *Nobl9Error {
	return New(ErrorTypeRateLimit, SeverityMedium, message, err)
}

func NewRateLimitErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeRateLimit, SeverityMedium, message, err, details)
}

// Timeout errors
func NewTimeoutError(message string, err error) *Nobl9Error {
	return New(ErrorTypeTimeout, SeverityMedium, message, err)
}

func NewTimeoutErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeTimeout, SeverityMedium, message, err, details)
}

// User resolution errors
func NewUserResolutionError(message string, err error) *Nobl9Error {
	return New(ErrorTypeUserResolution, SeverityMedium, message, err)
}

func NewUserResolutionErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeUserResolution, SeverityMedium, message, err, details)
}

// Manifest errors
func NewManifestError(message string, err error) *Nobl9Error {
	return New(ErrorTypeManifest, SeverityHigh, message, err)
}

func NewManifestErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeManifest, SeverityHigh, message, err, details)
}

// Retryable errors
func NewRetryableError(message string, err error) *Nobl9Error {
	return New(ErrorTypeRetryable, SeverityMedium, message, err)
}

func NewRetryableErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeRetryable, SeverityMedium, message, err, details)
}

// Non-retryable errors
func NewNonRetryableError(message string, err error) *Nobl9Error {
	return New(ErrorTypeNonRetryable, SeverityHigh, message, err)
}

func NewNonRetryableErrorWithDetails(message string, err error, details map[string]interface{}) *Nobl9Error {
	return NewWithDetails(ErrorTypeNonRetryable, SeverityHigh, message, err, details)
}

// Error categorization functions
func IsNobl9Error(err error) bool {
	_, ok := err.(*Nobl9Error)
	return ok
}

func IsRetryableError(err error) bool {
	if nobl9Err, ok := err.(*Nobl9Error); ok {
		return nobl9Err.IsRetryable()
	}

	// Check for common retryable error patterns
	errorMsg := strings.ToLower(err.Error())
	retryablePatterns := []string{
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

	for _, pattern := range retryablePatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

func IsAuthError(err error) bool {
	if nobl9Err, ok := err.(*Nobl9Error); ok {
		return nobl9Err.GetType() == ErrorTypeAuth
	}

	errorMsg := strings.ToLower(err.Error())
	authPatterns := []string{
		"unauthorized",
		"forbidden",
		"invalid credentials",
		"authentication failed",
		"401",
		"403",
	}

	for _, pattern := range authPatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

func IsRateLimitError(err error) bool {
	if nobl9Err, ok := err.(*Nobl9Error); ok {
		return nobl9Err.GetType() == ErrorTypeRateLimit
	}

	errorMsg := strings.ToLower(err.Error())
	rateLimitPatterns := []string{
		"rate limit",
		"429",
		"too many requests",
		"quota exceeded",
	}

	for _, pattern := range rateLimitPatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

func IsTimeoutError(err error) bool {
	if nobl9Err, ok := err.(*Nobl9Error); ok {
		return nobl9Err.GetType() == ErrorTypeTimeout
	}

	errorMsg := strings.ToLower(err.Error())
	timeoutPatterns := []string{
		"timeout",
		"deadline exceeded",
		"context deadline exceeded",
		"408",
		"504",
	}

	for _, pattern := range timeoutPatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

// Error aggregation
type ErrorAggregator struct {
	errors []*Nobl9Error
}

func NewErrorAggregator() *ErrorAggregator {
	return &ErrorAggregator{
		errors: make([]*Nobl9Error, 0),
	}
}

func (ea *ErrorAggregator) AddError(err *Nobl9Error) {
	ea.errors = append(ea.errors, err)
}

func (ea *ErrorAggregator) AddErrorFromErr(err error, errorType ErrorType, severity ErrorSeverity, message string) {
	if nobl9Err, ok := err.(*Nobl9Error); ok {
		ea.errors = append(ea.errors, nobl9Err)
	} else {
		ea.errors = append(ea.errors, New(errorType, severity, message, err))
	}
}

func (ea *ErrorAggregator) GetErrors() []*Nobl9Error {
	return ea.errors
}

func (ea *ErrorAggregator) GetErrorsByType(errorType ErrorType) []*Nobl9Error {
	var filtered []*Nobl9Error
	for _, err := range ea.errors {
		if err.GetType() == errorType {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

func (ea *ErrorAggregator) GetErrorsBySeverity(severity ErrorSeverity) []*Nobl9Error {
	var filtered []*Nobl9Error
	for _, err := range ea.errors {
		if err.GetSeverity() == severity {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

func (ea *ErrorAggregator) GetRetryableErrors() []*Nobl9Error {
	var filtered []*Nobl9Error
	for _, err := range ea.errors {
		if err.IsRetryable() {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

func (ea *ErrorAggregator) HasErrors() bool {
	return len(ea.errors) > 0
}

func (ea *ErrorAggregator) HasCriticalErrors() bool {
	for _, err := range ea.errors {
		if err.GetSeverity() == SeverityCritical {
			return true
		}
	}
	return false
}

func (ea *ErrorAggregator) GetErrorSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"total_errors":     len(ea.errors),
		"critical_errors":  0,
		"high_errors":      0,
		"medium_errors":    0,
		"low_errors":       0,
		"retryable_errors": 0,
		"error_types":      make(map[string]int),
	}

	for _, err := range ea.errors {
		// Count by severity
		switch err.GetSeverity() {
		case SeverityCritical:
			summary["critical_errors"] = summary["critical_errors"].(int) + 1
		case SeverityHigh:
			summary["high_errors"] = summary["high_errors"].(int) + 1
		case SeverityMedium:
			summary["medium_errors"] = summary["medium_errors"].(int) + 1
		case SeverityLow:
			summary["low_errors"] = summary["low_errors"].(int) + 1
		}

		// Count retryable errors
		if err.IsRetryable() {
			summary["retryable_errors"] = summary["retryable_errors"].(int) + 1
		}

		// Count by type
		errorType := string(err.GetType())
		summary["error_types"].(map[string]int)[errorType]++
	}

	return summary
}

// Helper functions
func isRetryableErrorType(errorType ErrorType) bool {
	retryableTypes := []ErrorType{
		ErrorTypeNetwork,
		ErrorTypeRateLimit,
		ErrorTypeTimeout,
		ErrorTypeRetryable,
	}

	for _, t := range retryableTypes {
		if errorType == t {
			return true
		}
	}

	return false
}

// Error formatting
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	if nobl9Err, ok := err.(*Nobl9Error); ok {
		return formatNobl9Error(nobl9Err)
	}
	return err.Error()
}

func formatNobl9Error(err *Nobl9Error) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s] %s", err.GetType(), err.Message))

	if err.Err != nil {
		sb.WriteString(fmt.Sprintf(": %v", err.Err))
	}

	if len(err.Details) > 0 {
		sb.WriteString(" | Details: ")
		detailParts := make([]string, 0, len(err.Details))
		for k, v := range err.Details {
			detailParts = append(detailParts, fmt.Sprintf("%s=%v", k, v))
		}
		sb.WriteString(strings.Join(detailParts, ", "))
	}

	return sb.String()
}
