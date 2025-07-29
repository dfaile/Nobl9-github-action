package logger

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
	fields logrus.Fields
}

// Fields type for structured logging
type Fields map[string]interface{}

// Level represents log levels
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Format represents log formats
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// New creates a new logger instance
func New(level Level, format Format) *Logger {
	logger := logrus.New()

	// Set log level
	switch level {
	case LevelDebug:
		logger.SetLevel(logrus.DebugLevel)
	case LevelInfo:
		logger.SetLevel(logrus.InfoLevel)
	case LevelWarn:
		logger.SetLevel(logrus.WarnLevel)
	case LevelError:
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	switch format {
	case FormatJSON:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case FormatText:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			DisableColors:   !isTerminal(),
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	return &Logger{
		Logger: logger,
		fields: make(logrus.Fields),
	}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields Fields) *Logger {
	newLogger := &Logger{
		Logger: l.Logger,
		fields: make(logrus.Fields),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithField creates a new logger with a single field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(Fields{key: value})
}

// WithContext creates a new logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := Fields{}

	// Extract useful context information
	if ctx != nil {
		// Add request ID if available
		if requestID := ctx.Value("request_id"); requestID != nil {
			fields["request_id"] = requestID
		}

		// Add correlation ID if available
		if correlationID := ctx.Value("correlation_id"); correlationID != nil {
			fields["correlation_id"] = correlationID
		}
	}

	return l.WithFields(fields)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Fields) {
	l.log(logrus.DebugLevel, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Fields) {
	l.log(logrus.InfoLevel, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Fields) {
	l.log(logrus.WarnLevel, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Fields) {
	l.log(logrus.ErrorLevel, msg, fields...)
}

// ErrorWithErr logs an error message with an error
func (l *Logger) ErrorWithErr(msg string, err error, fields ...Fields) {
	if err != nil {
		fields = append(fields, Fields{"error": err.Error()})
	}
	l.log(logrus.ErrorLevel, msg, fields...)
}

// LogError logs a structured error with detailed information
func (l *Logger) LogError(err error, fields ...Fields) {
	baseFields := Fields{
		"event": "error_occurred",
	}

	// Add error details
	if err != nil {
		baseFields["error_message"] = err.Error()
		baseFields["error_type"] = "unknown"
		baseFields["retryable"] = false
		baseFields["severity"] = "medium"
	}

	// Merge additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			baseFields[k] = v
		}
	}

	l.Error("Error occurred", baseFields)
}

// LogDetailedError logs a detailed error with comprehensive information
func (l *Logger) LogDetailedError(err error, operation string, context map[string]interface{}, fields ...Fields) {
	baseFields := Fields{
		"event":     "detailed_error",
		"operation": operation,
	}

	// Add error details
	if err != nil {
		baseFields["error_message"] = err.Error()
		baseFields["error_type"] = "unknown"
		baseFields["retryable"] = false
		baseFields["severity"] = "medium"
	}

	// Add context information
	if context != nil {
		for k, v := range context {
			baseFields["context_"+k] = v
		}
	}

	// Merge additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			baseFields[k] = v
		}
	}

	l.Error("Detailed error occurred", baseFields)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...Fields) {
	l.log(logrus.FatalLevel, msg, fields...)
	os.Exit(1)
}

// FatalWithErr logs a fatal message with an error and exits
func (l *Logger) FatalWithErr(msg string, err error, fields ...Fields) {
	if err != nil {
		fields = append(fields, Fields{"error": err.Error()})
	}
	l.log(logrus.FatalLevel, msg, fields...)
	os.Exit(1)
}

// log is the internal logging method
func (l *Logger) log(level logrus.Level, msg string, fields ...Fields) {
	entry := l.Logger.WithFields(l.fields)

	// Add additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			entry = entry.WithField(k, v)
		}
	}

	// Add GitHub Actions specific fields if running in GitHub Actions
	if isGitHubActions() {
		entry = entry.WithFields(logrus.Fields{
			"github_actions": true,
			"workflow":       getEnv("GITHUB_WORKFLOW", ""),
			"run_id":         getEnv("GITHUB_RUN_ID", ""),
			"actor":          getEnv("GITHUB_ACTOR", ""),
			"repository":     getEnv("GITHUB_REPOSITORY", ""),
			"sha":            getEnv("GITHUB_SHA", ""),
		})
	}

	entry.Log(level, msg)
}

// LogProcessingStart logs the start of processing
func (l *Logger) LogProcessingStart(config map[string]interface{}) {
	l.Info("Starting Nobl9 project processing", Fields{
		"event":  "processing_start",
		"config": config,
	})
}

// LogProcessingComplete logs the completion of processing
func (l *Logger) LogProcessingComplete(stats map[string]interface{}) {
	l.Info("Nobl9 project processing completed", Fields{
		"event": "processing_complete",
		"stats": stats,
	})
}

// LogFileProcessed logs when a file is processed
func (l *Logger) LogFileProcessed(filePath string, fileType string, success bool, fields ...Fields) {
	baseFields := Fields{
		"event":     "file_processed",
		"file_path": filePath,
		"file_type": fileType,
		"success":   success,
	}

	// Merge additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			baseFields[k] = v
		}
	}

	if success {
		l.Info("File processed successfully", baseFields)
	} else {
		l.Error("File processing failed", baseFields)
	}
}

// LogNobl9APICall logs Nobl9 API calls
func (l *Logger) LogNobl9APICall(method, endpoint string, success bool, duration time.Duration, fields ...Fields) {
	baseFields := Fields{
		"event":       "nobl9_api_call",
		"method":      method,
		"endpoint":    endpoint,
		"success":     success,
		"duration":    duration.String(),
		"duration_ms": duration.Milliseconds(),
	}

	// Merge additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			baseFields[k] = v
		}
	}

	if success {
		l.Debug("Nobl9 API call successful", baseFields)
	} else {
		l.Error("Nobl9 API call failed", baseFields)
	}
}

// LogUserResolution logs user email to UserID resolution
func (l *Logger) LogUserResolution(email, userID string, success bool, fields ...Fields) {
	baseFields := Fields{
		"event":   "user_resolution",
		"email":   email,
		"user_id": userID,
		"success": success,
	}

	// Merge additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			baseFields[k] = v
		}
	}

	if success {
		l.Info("User email resolved to UserID", baseFields)
	} else {
		l.Warn("Failed to resolve user email to UserID", baseFields)
	}
}

// LogProjectOperation logs project creation/update operations
func (l *Logger) LogProjectOperation(operation, projectName string, success bool, fields ...Fields) {
	baseFields := Fields{
		"event":        "project_operation",
		"operation":    operation,
		"project_name": projectName,
		"success":      success,
	}

	// Merge additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			baseFields[k] = v
		}
	}

	if success {
		l.Info("Project operation completed", baseFields)
	} else {
		l.Error("Project operation failed", baseFields)
	}
}

// LogRoleBindingOperation logs role binding creation/update operations
func (l *Logger) LogRoleBindingOperation(operation, roleBindingName, projectName string, success bool, fields ...Fields) {
	baseFields := Fields{
		"event":             "role_binding_operation",
		"operation":         operation,
		"role_binding_name": roleBindingName,
		"project_name":      projectName,
		"success":           success,
	}

	// Merge additional fields
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			baseFields[k] = v
		}
	}

	if success {
		l.Info("Role binding operation completed", baseFields)
	} else {
		l.Error("Role binding operation failed", baseFields)
	}
}

// LogValidationResult logs validation results
func (l *Logger) LogValidationResult(filePath string, valid bool, errors []string, warnings []string) {
	fields := Fields{
		"event":     "validation_result",
		"file_path": filePath,
		"valid":     valid,
	}

	if len(errors) > 0 {
		fields["errors"] = errors
	}

	if len(warnings) > 0 {
		fields["warnings"] = warnings
	}

	if valid {
		l.Info("File validation passed", fields)
	} else {
		l.Error("File validation failed", fields)
	}
}

// Helper functions

// isTerminal checks if output is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// isGitHubActions checks if running in GitHub Actions
func isGitHubActions() bool {
	return getEnv("GITHUB_ACTIONS", "") == "true"
}

// getEnv gets an environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
