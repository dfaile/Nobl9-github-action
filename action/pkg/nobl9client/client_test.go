package nobl9client

import (
	"context"
	"testing"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
)

func TestNewClient(t *testing.T) {
	// Test with valid credentials
	client, err := NewClient("test-client-id", "test-client-secret")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if client == nil {
		t.Error("expected client to be created")
	}

	if client.sdkClient == nil {
		t.Error("expected SDK client to be set")
	}

	if client.timeout != 60*time.Second {
		t.Errorf("expected timeout to be 60s, got %v", client.timeout)
	}
}

func TestProcessObjects(t *testing.T) {
	// Create a mock client for testing
	client := &Client{
		sdkClient: &sdk.Client{},
		timeout:   60 * time.Second,
	}

	ctx := context.Background()

	// Test with empty objects (dry run)
	objects := []ParsedObject{}
	result, err := client.ProcessObjects(ctx, objects, true)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("expected result to be created")
	}

	if result.Projects == nil {
		t.Error("expected projects slice to be initialized")
	}

	if result.RoleBindings == nil {
		t.Error("expected role bindings slice to be initialized")
	}

	if result.EmailsResolved == nil {
		t.Error("expected emails resolved map to be initialized")
	}

	if result.Errors == nil {
		t.Error("expected errors slice to be initialized")
	}
}

func TestGenerateSummary(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		result   *ProcessResult
		expected string
	}{
		{
			name: "empty result",
			result: &ProcessResult{
				Projects:       []ProcessedObject{},
				RoleBindings:   []ProcessedObject{},
				EmailsResolved: map[string]string{},
				Errors:         []error{},
			},
			expected: "Processing completed: 0 projects, 0 role bindings, 0 emails resolved, 0 errors",
		},
		{
			name: "with successful objects",
			result: &ProcessResult{
				Projects: []ProcessedObject{
					{Applied: true, Error: nil},
					{Applied: true, Error: nil},
				},
				RoleBindings: []ProcessedObject{
					{Applied: true, Error: nil},
				},
				EmailsResolved: map[string]string{
					"user1@example.com": "user1-id",
					"user2@example.com": "user2-id",
				},
				Errors: []error{},
			},
			expected: "Processing completed: 2 projects, 1 role bindings, 2 emails resolved, 0 errors",
		},
		{
			name: "with errors",
			result: &ProcessResult{
				Projects: []ProcessedObject{
					{Applied: true, Error: nil},
					{Applied: false, Error: &mockError{}},
				},
				RoleBindings: []ProcessedObject{
					{Applied: false, Error: &mockError{}},
				},
				EmailsResolved: map[string]string{},
				Errors:         []error{&mockError{}, &mockError{}},
			},
			expected: "Processing completed: 1 projects, 0 role bindings, 0 emails resolved, 2 errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := client.generateSummary(tt.result)

			if summary != tt.expected {
				t.Errorf("expected summary %q, got %q", tt.expected, summary)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid name",
			input:    "test-project",
			expected: "test-project",
		},
		{
			name:     "uppercase letters",
			input:    "Test-Project",
			expected: "test-project",
		},
		{
			name:     "invalid characters",
			input:    "test@project#123",
			expected: "test-project-123",
		},
		{
			name:     "spaces and special chars",
			input:    "Test Project 123!",
			expected: "test-project-123",
		},
		{
			name:     "leading and trailing hyphens",
			input:    "-test-project-",
			expected: "test-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeName(tt.input)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "test",
			maxLen:   10,
			expected: "test",
		},
		{
			name:     "string equal to max",
			input:    "test12345",
			maxLen:   9,
			expected: "test12345",
		},
		{
			name:     "string longer than max",
			input:    "test123456789",
			maxLen:   10,
			expected: "test123456",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Mock objects for testing
type mockObject struct {
	kind manifest.Kind
	name string
}

func (m *mockObject) GetKind() manifest.Kind {
	return m.kind
}

func (m *mockObject) GetName() string {
	return m.name
}

func (m *mockObject) GetVersion() manifest.Version {
	return manifest.Version("n9/v1alpha")
}

func (m *mockObject) Validate() error {
	return nil
}

func (m *mockObject) GetManifestSource() string {
	return "test"
}

func (m *mockObject) SetManifestSource(source string) manifest.Object {
	// Mock implementation - return self
	return m
}

type mockError struct{}

func (m *mockError) Error() string {
	return "mock error"
}
