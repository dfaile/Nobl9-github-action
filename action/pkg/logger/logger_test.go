package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		format   Format
		expected logrus.Level
	}{
		{
			name:     "debug level",
			level:    LevelDebug,
			format:   FormatJSON,
			expected: logrus.DebugLevel,
		},
		{
			name:     "info level",
			level:    LevelInfo,
			format:   FormatJSON,
			expected: logrus.InfoLevel,
		},
		{
			name:     "warn level",
			level:    LevelWarn,
			format:   FormatJSON,
			expected: logrus.WarnLevel,
		},
		{
			name:     "error level",
			level:    LevelError,
			format:   FormatJSON,
			expected: logrus.ErrorLevel,
		},
		{
			name:     "invalid level defaults to info",
			level:    "invalid",
			format:   FormatJSON,
			expected: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level, tt.format)

			if logger.GetLevel() != tt.expected {
				t.Errorf("expected level %v, got %v", tt.expected, logger.GetLevel())
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	fields := Fields{
		"key1": "value1",
		"key2": "value2",
	}

	newLogger := logger.WithFields(fields)

	// Capture output
	var buf bytes.Buffer
	newLogger.SetOutput(&buf)

	newLogger.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Check that fields are present
	if logEntry["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %v", logEntry["key1"])
	}
	if logEntry["key2"] != "value2" {
		t.Errorf("expected key2=value2, got %v", logEntry["key2"])
	}
}

func TestLoggerWithField(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	newLogger := logger.WithField("test_key", "test_value")

	// Capture output
	var buf bytes.Buffer
	newLogger.SetOutput(&buf)

	newLogger.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry["test_key"] != "test_value" {
		t.Errorf("expected test_key=test_value, got %v", logEntry["test_key"])
	}
}

func TestLoggerWithContext(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	// Use the same context key types defined in the logger package
	ctx := context.WithValue(context.Background(), RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, CorrelationIDKey, "corr-456")

	newLogger := logger.WithContext(ctx)

	// Capture output
	var buf bytes.Buffer
	newLogger.SetOutput(&buf)

	newLogger.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry["request_id"] != "req-123" {
		t.Errorf("expected request_id=req-123, got %v", logEntry["request_id"])
	}
	if logEntry["correlation_id"] != "corr-456" {
		t.Errorf("expected correlation_id=corr-456, got %v", logEntry["correlation_id"])
	}
}

func TestLoggerLevels(t *testing.T) {
	logger := New(LevelDebug, FormatJSON)

	// Capture output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 log lines, got %d", len(lines))
	}

	// Check that all levels are present
	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("debug message not found in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("info message not found in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("warn message not found in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("error message not found in output")
	}
}

func TestLoggerErrorWithErr(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	// Capture output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	err := fmt.Errorf("test error")
	logger.ErrorWithErr("error occurred", err)

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry["error"] != "test error" {
		t.Errorf("expected error=test error, got %v", logEntry["error"])
	}
}

func TestLogProcessingStart(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	// Capture output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	config := map[string]interface{}{
		"dry_run": true,
		"force":   false,
	}

	logger.LogProcessingStart(config)

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry["event"] != "processing_start" {
		t.Errorf("expected event=processing_start, got %v", logEntry["event"])
	}

	configData, ok := logEntry["config"].(map[string]interface{})
	if !ok {
		t.Fatal("config field is not a map")
	}

	if configData["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", configData["dry_run"])
	}
}

func TestLogProcessingComplete(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	// Capture output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	stats := map[string]interface{}{
		"files_processed":  5,
		"projects_created": 2,
		"errors":           0,
	}

	logger.LogProcessingComplete(stats)

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry["event"] != "processing_complete" {
		t.Errorf("expected event=processing_complete, got %v", logEntry["event"])
	}

	statsData, ok := logEntry["stats"].(map[string]interface{})
	if !ok {
		t.Fatal("stats field is not a map")
	}

	if statsData["files_processed"] != float64(5) {
		t.Errorf("expected files_processed=5, got %v", statsData["files_processed"])
	}
}

func TestLogFileProcessed(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	tests := []struct {
		name     string
		filePath string
		fileType string
		success  bool
		expected string
	}{
		{
			name:     "successful file processing",
			filePath: "/path/to/file.yaml",
			fileType: "nobl9-project",
			success:  true,
			expected: "File processed successfully",
		},
		{
			name:     "failed file processing",
			filePath: "/path/to/file.yaml",
			fileType: "nobl9-project",
			success:  false,
			expected: "File processing failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer
			logger.SetOutput(&buf)

			logger.LogFileProcessed(tt.filePath, tt.fileType, tt.success)

			// Parse JSON output
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			if logEntry["event"] != "file_processed" {
				t.Errorf("expected event=file_processed, got %v", logEntry["event"])
			}

			if logEntry["file_path"] != tt.filePath {
				t.Errorf("expected file_path=%s, got %v", tt.filePath, logEntry["file_path"])
			}

			if logEntry["file_type"] != tt.fileType {
				t.Errorf("expected file_type=%s, got %v", tt.fileType, logEntry["file_type"])
			}

			if logEntry["success"] != tt.success {
				t.Errorf("expected success=%v, got %v", tt.success, logEntry["success"])
			}

			if !strings.Contains(logEntry["message"].(string), tt.expected) {
				t.Errorf("expected message to contain '%s', got '%s'", tt.expected, logEntry["message"])
			}
		})
	}
}

func TestLogNobl9APICall(t *testing.T) {
	logger := New(LevelDebug, FormatJSON)

	// Capture output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	duration := 150 * time.Millisecond
	logger.LogNobl9APICall("POST", "/api/v1/projects", true, duration)

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry["event"] != "nobl9_api_call" {
		t.Errorf("expected event=nobl9_api_call, got %v", logEntry["event"])
	}

	if logEntry["method"] != "POST" {
		t.Errorf("expected method=POST, got %v", logEntry["method"])
	}

	if logEntry["endpoint"] != "/api/v1/projects" {
		t.Errorf("expected endpoint=/api/v1/projects, got %v", logEntry["endpoint"])
	}

	if logEntry["success"] != true {
		t.Errorf("expected success=true, got %v", logEntry["success"])
	}

	if logEntry["duration"] != "150ms" {
		t.Errorf("expected duration=150ms, got %v", logEntry["duration"])
	}

	if logEntry["duration_ms"] != float64(150) {
		t.Errorf("expected duration_ms=150, got %v", logEntry["duration_ms"])
	}
}

func TestLogUserResolution(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	tests := []struct {
		name     string
		email    string
		userID   string
		success  bool
		expected string
	}{
		{
			name:     "successful user resolution",
			email:    "user@example.com",
			userID:   "user-123",
			success:  true,
			expected: "User email resolved to UserID",
		},
		{
			name:     "failed user resolution",
			email:    "user@example.com",
			userID:   "",
			success:  false,
			expected: "Failed to resolve user email to UserID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer
			logger.SetOutput(&buf)

			logger.LogUserResolution(tt.email, tt.userID, tt.success)

			// Parse JSON output
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			if logEntry["event"] != "user_resolution" {
				t.Errorf("expected event=user_resolution, got %v", logEntry["event"])
			}

			if logEntry["email"] != tt.email {
				t.Errorf("expected email=%s, got %v", tt.email, logEntry["email"])
			}

			if logEntry["user_id"] != tt.userID {
				t.Errorf("expected user_id=%s, got %v", tt.userID, logEntry["user_id"])
			}

			if logEntry["success"] != tt.success {
				t.Errorf("expected success=%v, got %v", tt.success, logEntry["success"])
			}

			if !strings.Contains(logEntry["message"].(string), tt.expected) {
				t.Errorf("expected message to contain '%s', got '%s'", tt.expected, logEntry["message"])
			}
		})
	}
}

func TestLogValidationResult(t *testing.T) {
	logger := New(LevelInfo, FormatJSON)

	tests := []struct {
		name     string
		filePath string
		valid    bool
		errors   []string
		warnings []string
		expected string
	}{
		{
			name:     "valid file",
			filePath: "/path/to/file.yaml",
			valid:    true,
			errors:   []string{},
			warnings: []string{},
			expected: "File validation passed",
		},
		{
			name:     "invalid file with errors",
			filePath: "/path/to/file.yaml",
			valid:    false,
			errors:   []string{"missing required field", "invalid format"},
			warnings: []string{"deprecated field"},
			expected: "File validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer
			logger.SetOutput(&buf)

			logger.LogValidationResult(tt.filePath, tt.valid, tt.errors, tt.warnings)

			// Parse JSON output
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			if logEntry["event"] != "validation_result" {
				t.Errorf("expected event=validation_result, got %v", logEntry["event"])
			}

			if logEntry["file_path"] != tt.filePath {
				t.Errorf("expected file_path=%s, got %v", tt.filePath, logEntry["file_path"])
			}

			if logEntry["valid"] != tt.valid {
				t.Errorf("expected valid=%v, got %v", tt.valid, logEntry["valid"])
			}

			if !strings.Contains(logEntry["message"].(string), tt.expected) {
				t.Errorf("expected message to contain '%s', got '%s'", tt.expected, logEntry["message"])
			}
		})
	}
}

func TestTextFormat(t *testing.T) {
	logger := New(LevelInfo, FormatText)

	// Capture output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Info("test message", Fields{"key": "value"})

	output := buf.String()

	// Check for text format indicators
	if !strings.Contains(output, "time=") {
		t.Error("text format should contain timestamp")
	}

	if !strings.Contains(output, "level=info") {
		t.Error("text format should contain log level")
	}

	if !strings.Contains(output, "msg=\"test message\"") {
		t.Error("text format should contain message")
	}

	if !strings.Contains(output, "key=value") {
		t.Error("text format should contain structured fields")
	}
}

func TestGitHubActionsFields(t *testing.T) {
	// Set GitHub Actions environment variables
	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("GITHUB_WORKFLOW", "test-workflow")
	os.Setenv("GITHUB_RUN_ID", "123456789")
	os.Setenv("GITHUB_ACTOR", "test-user")
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	os.Setenv("GITHUB_SHA", "abc123")

	defer func() {
		os.Unsetenv("GITHUB_ACTIONS")
		os.Unsetenv("GITHUB_WORKFLOW")
		os.Unsetenv("GITHUB_RUN_ID")
		os.Unsetenv("GITHUB_ACTOR")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GITHUB_SHA")
	}()

	logger := New(LevelInfo, FormatJSON)

	// Capture output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry["github_actions"] != true {
		t.Error("expected github_actions=true")
	}

	if logEntry["workflow"] != "test-workflow" {
		t.Errorf("expected workflow=test-workflow, got %v", logEntry["workflow"])
	}

	if logEntry["run_id"] != "123456789" {
		t.Errorf("expected run_id=123456789, got %v", logEntry["run_id"])
	}

	if logEntry["actor"] != "test-user" {
		t.Errorf("expected actor=test-user, got %v", logEntry["actor"])
	}

	if logEntry["repository"] != "test/repo" {
		t.Errorf("expected repository=test/repo, got %v", logEntry["repository"])
	}

	if logEntry["sha"] != "abc123" {
		t.Errorf("expected sha=abc123, got %v", logEntry["sha"])
	}
}
