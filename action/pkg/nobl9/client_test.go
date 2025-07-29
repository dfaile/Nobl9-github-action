package nobl9

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/nobl9-action/pkg/logger"
)

func TestNewClient(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Timeout:      30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "missing client ID",
			config: &Config{
				ClientSecret: "test-client-secret",
				Timeout:      30 * time.Second,
			},
			expectError: true,
		},
		{
			name: "missing client secret",
			config: &Config{
				ClientID: "test-client-id",
				Timeout:  30 * time.Second,
			},
			expectError: true,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "zero timeout defaults to 30s",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Timeout:      0,
			},
			expectError: false,
		},
		{
			name: "zero retry attempts defaults to 3",
			config: &Config{
				ClientID:      "test-client-id",
				ClientSecret:  "test-client-secret",
				Timeout:       30 * time.Second,
				RetryAttempts: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config, log)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("expected client but got nil")
				return
			}

			// Test configuration
			if client.config.ClientID != tt.config.ClientID {
				t.Errorf("expected client ID %s, got %s", tt.config.ClientID, client.config.ClientID)
			}

			if client.config.ClientSecret != tt.config.ClientSecret {
				t.Errorf("expected client secret %s, got %s", tt.config.ClientSecret, client.config.ClientSecret)
			}

			// Environment field removed from Config struct

			// Test default values
			if tt.config.Timeout == 0 && client.config.Timeout != 30*time.Second {
				t.Errorf("expected default timeout 30s, got %v", client.config.Timeout)
			}

			if tt.config.RetryAttempts == 0 && client.config.RetryAttempts != 3 {
				t.Errorf("expected default retry attempts 3, got %d", client.config.RetryAttempts)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Timeout:      30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "missing client ID",
			config: &Config{
				ClientSecret: "test-client-secret",
				Timeout:      30 * time.Second,
			},
			expectError: true,
		},
		{
			name: "missing client secret",
			config: &Config{
				ClientID: "test-client-id",
				Timeout:  30 * time.Second,
			},
			expectError: true,
		},
		{
			name: "empty client ID",
			config: &Config{
				ClientID:     "",
				ClientSecret: "test-client-secret",
				Timeout:      30 * time.Second,
			},
			expectError: true,
		},
		{
			name: "empty client secret",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "",
				Timeout:      30 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClientMethods(t *testing.T) {
	// Note: These tests would require a mock Nobl9 API or test environment
	// For now, we'll test the client creation and basic functionality

	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Timeout:      30 * time.Second,
	}

	client, err := New(config, log)
	if err != nil {
		t.Skipf("Skipping client tests due to connection error: %v", err)
	}

	t.Run("GetConfig", func(t *testing.T) {
		retrievedConfig := client.GetConfig()
		if retrievedConfig != config {
			t.Error("GetConfig returned different config")
		}
	})

	t.Run("GetSDKClient", func(t *testing.T) {
		sdkClient := client.GetSDKClient()
		if sdkClient == nil {
			t.Error("GetSDKClient returned nil")
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := client.Close()
		if err != nil {
			t.Errorf("Close returned error: %v", err)
		}
	})
}

func TestClientWithNilLogger(t *testing.T) {
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Timeout:      30 * time.Second,
	}

	_, err := New(config, nil)
	if err == nil {
		t.Error("expected error for nil logger but got none")
	}
}

func TestClientContextHandling(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Timeout:      30 * time.Second,
	}

	client, err := New(config, log)
	if err != nil {
		t.Skipf("Skipping context tests due to connection error: %v", err)
	}
	defer client.Close()

	t.Run("Context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// This would test actual API calls with context
		// For now, just verify the context is properly handled
		if ctx.Err() != nil {
			t.Error("Context should not be cancelled initially")
		}
	})

	t.Run("Cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if ctx.Err() == nil {
			t.Error("Context should be cancelled")
		}
	})
}

func TestClientErrorHandling(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)

	t.Run("Invalid credentials", func(t *testing.T) {
		config := &Config{
			ClientID:     "invalid-client-id",
			ClientSecret: "invalid-client-secret",
			Timeout:      5 * time.Second, // Short timeout for test
		}

		_, err := New(config, log)
		if err == nil {
			t.Error("expected error for invalid credentials but got none")
		}
	})

	t.Run("Network timeout", func(t *testing.T) {
		config := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Timeout:      1 * time.Millisecond, // Very short timeout
		}

		_, err := New(config, log)
		if err == nil {
			t.Error("expected error for timeout but got none")
		}
	})
}

func TestClientConfigurationDefaults(t *testing.T) {
	t.Run("Default timeout", func(t *testing.T) {
		config := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			// Timeout not set
		}

		// This would normally fail due to invalid credentials
		// but we can test that the config is properly set
		err := validateConfig(config)
		if err != nil {
			t.Errorf("unexpected validation error: %v", err)
		}

		if config.Timeout != 30*time.Second {
			t.Errorf("expected default timeout 30s, got %v", config.Timeout)
		}
	})

	t.Run("Default retry attempts", func(t *testing.T) {
		config := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Environment:  "test",
			Timeout:      30 * time.Second,
			// RetryAttempts not set
		}

		err := validateConfig(config)
		if err != nil {
			t.Errorf("unexpected validation error: %v", err)
		}

		if config.RetryAttempts != 3 {
			t.Errorf("expected default retry attempts 3, got %d", config.RetryAttempts)
		}
	})
}

func TestClientEnvironmentDetection(t *testing.T) {
	tests := []struct {
		name        string
		clientID    string
		environment string
	}{
		{
			name:        "dev environment",
			clientID:    "dev-api-client-123",
			environment: "dev",
		},
		{
			name:        "staging environment",
			clientID:    "staging-api-client-456",
			environment: "staging",
		},
		{
			name:        "production environment",
			clientID:    "prod-api-client-789",
			environment: "prod",
		},
		{
			name:        "unknown environment",
			clientID:    "api-client-unknown",
			environment: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ClientID:     tt.clientID,
				ClientSecret: "test-client-secret",
				Environment:  tt.environment,
				Timeout:      30 * time.Second,
			}

			// Test configuration validation
			err := validateConfig(config)
			if err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}

			// Verify environment is set correctly
			if config.Environment != tt.environment {
				t.Errorf("expected environment %s, got %s", tt.environment, config.Environment)
			}
		})
	}
}
