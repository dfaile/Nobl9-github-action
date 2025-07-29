package nobl9

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 SDK connection")
}

func TestValidateConfig(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientConfigurationDefaults(t *testing.T) {
	t.Run("Default timeout", func(t *testing.T) {
		config := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			// Timeout not set
		}

		err := validateConfig(config)
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, config.Timeout)
	})

	t.Run("Default retry attempts", func(t *testing.T) {
		config := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Timeout:      30 * time.Second,
			// RetryAttempts not set
		}

		err := validateConfig(config)
		assert.NoError(t, err)
		assert.Equal(t, 3, config.RetryAttempts)
	})
}

func TestClientMethods(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 SDK connection")
}

func TestClientWithNilLogger(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 SDK connection")
}

func TestClientContextHandling(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 SDK connection")
}

func TestClientErrorHandling(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 SDK connection")
}

func TestClientEnvironmentDetection(t *testing.T) {
	t.Skip("Skipping test that requires Environment field which was removed from Config struct")
}
