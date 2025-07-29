package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"INPUT_CLIENT_ID":     "test-client-id",
				"INPUT_CLIENT_SECRET": "test-client-secret",
				"GITHUB_WORKSPACE":    "/workspace",
			},
			expectError: false,
		},
		{
			name: "missing client ID",
			envVars: map[string]string{
				"INPUT_CLIENT_SECRET": "test-client-secret",
				"GITHUB_WORKSPACE":    "/workspace",
			},
			expectError: true,
			errorMsg:    "Nobl9 client ID is required",
		},
		{
			name: "missing client secret",
			envVars: map[string]string{
				"INPUT_CLIENT_ID":  "test-client-id",
				"GITHUB_WORKSPACE": "/workspace",
			},
			expectError: true,
			errorMsg:    "Nobl9 client secret is required",
		},
		{
			name: "missing GitHub workspace",
			envVars: map[string]string{
				"INPUT_CLIENT_ID":     "test-client-id",
				"INPUT_CLIENT_SECRET": "test-client-secret",
			},
			expectError: true,
			errorMsg:    "GitHub workspace is required",
		},
		{
			name: "fallback to direct environment variables",
			envVars: map[string]string{
				"NOBL9_CLIENT_ID":     "fallback-client-id",
				"NOBL9_CLIENT_SECRET": "fallback-client-secret",
				"GITHUB_WORKSPACE":    "/workspace",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Load configuration
			config, err := Load()

			// Check error expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate configuration
			if config == nil {
				t.Error("expected configuration but got nil")
				return
			}

			// Check that credentials are loaded
			clientID, clientSecret := config.GetNobl9Credentials()
			if clientID == "" {
				t.Error("expected client ID to be loaded")
			}
			if clientSecret == "" {
				t.Error("expected client secret to be loaded")
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input       string
		expected    bool
		expectError bool
	}{
		{"true", true, false},
		{"TRUE", true, false},
		{"True", true, false},
		{"1", true, false},
		{"yes", true, false},
		{"on", true, false},
		{"false", false, false},
		{"FALSE", false, false},
		{"False", false, false},
		{"0", false, false},
		{"no", false, false},
		{"off", false, false},
		{"invalid", false, true},
		{"", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseBool(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input '%s'", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for input '%s': %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %v for input '%s', got %v", tt.expected, tt.input, result)
			}
		})
	}
}

func TestDetectEnvironment(t *testing.T) {
	tests := []struct {
		clientID string
		expected string
	}{
		{"dev-client-id", "dev"},
		{"staging-client-id", "staging"},
		{"prod-client-id", "prod"},
		{"unknown-client-id", "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.clientID, func(t *testing.T) {
			config := &Config{}
			config.Nobl9.ClientID = tt.clientID

			result := config.detectEnvironment()
			if result != tt.expected {
				t.Errorf("expected environment '%s' for client ID '%s', got '%s'", tt.expected, tt.clientID, result)
			}
		})
	}
}

func TestIsGitHubActions(t *testing.T) {
	tests := []struct {
		envValue string
		expected bool
	}{
		{"true", true},
		{"TRUE", false}, // Only "true" (lowercase) is supported
		{"false", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("GITHUB_ACTIONS", tt.envValue)
			defer os.Unsetenv("GITHUB_ACTIONS")

			config := &Config{}
			result := config.IsGitHubActions()

			if result != tt.expected {
				t.Errorf("expected %v for GITHUB_ACTIONS='%s', got %v", tt.expected, tt.envValue, result)
			}
		})
	}
}

func TestGetRepositoryPath(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		repoPath  string
		expected  string
	}{
		{
			name:      "Local development",
			workspace: "",
			repoPath:  "/local/path",
			expected:  "/local/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			config.GitHub.Workspace = tt.workspace
			config.Repository.Path = tt.repoPath

			result := config.GetRepositoryPath()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
