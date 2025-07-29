package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all configuration for the Nobl9 action
type Config struct {
	// Nobl9 API credentials
	Nobl9 struct {
		ClientID     string
		ClientSecret string
		Environment  string // Auto-detected from credentials
	}

	// Repository configuration
	Repository struct {
		Path        string
		FilePattern string
	}

	// Processing options
	Processing struct {
		DryRun       bool
		Force        bool
		ValidateOnly bool
	}

	// Logging configuration
	Logging struct {
		Level  string
		Format string
	}

	// GitHub Actions specific
	GitHub struct {
		Workspace string
		EventPath string
		Token     string
	}
}

// Load loads configuration from environment variables and GitHub Actions context
func Load() (*Config, error) {
	config := &Config{}

	// Load Nobl9 credentials
	if err := config.loadNobl9Config(); err != nil {
		return nil, fmt.Errorf("failed to load Nobl9 configuration: %w", err)
	}

	// Load repository configuration
	if err := config.loadRepositoryConfig(); err != nil {
		return nil, fmt.Errorf("failed to load repository configuration: %w", err)
	}

	// Load processing options
	if err := config.loadProcessingConfig(); err != nil {
		return nil, fmt.Errorf("failed to load processing configuration: %w", err)
	}

	// Load logging configuration
	if err := config.loadLoggingConfig(); err != nil {
		return nil, fmt.Errorf("failed to load logging configuration: %w", err)
	}

	// Load GitHub Actions configuration
	if err := config.loadGitHubConfig(); err != nil {
		return nil, fmt.Errorf("failed to load GitHub configuration: %w", err)
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// loadNobl9Config loads Nobl9 API configuration
func (c *Config) loadNobl9Config() error {
	// Load credentials from environment variables (GitHub Actions inputs)
	c.Nobl9.ClientID = getEnv("INPUT_CLIENT_ID", "")
	c.Nobl9.ClientSecret = getEnv("INPUT_CLIENT_SECRET", "")

	// Fallback to direct environment variables for local development
	if c.Nobl9.ClientID == "" {
		c.Nobl9.ClientID = getEnv("NOBL9_CLIENT_ID", "")
	}
	if c.Nobl9.ClientSecret == "" {
		c.Nobl9.ClientSecret = getEnv("NOBL9_CLIENT_SECRET", "")
	}

	// Auto-detect environment from credentials
	c.Nobl9.Environment = c.detectEnvironment()

	return nil
}

// loadRepositoryConfig loads repository configuration
func (c *Config) loadRepositoryConfig() error {
	c.Repository.Path = getEnv("INPUT_REPO_PATH", ".")
	c.Repository.FilePattern = getEnv("INPUT_FILE_PATTERN", "**/*.yaml")

	return nil
}

// loadProcessingConfig loads processing options
func (c *Config) loadProcessingConfig() error {
	var err error

	c.Processing.DryRun, err = parseBool(getEnv("INPUT_DRY_RUN", "false"))
	if err != nil {
		return fmt.Errorf("invalid dry-run value: %w", err)
	}

	c.Processing.Force, err = parseBool(getEnv("INPUT_FORCE", "false"))
	if err != nil {
		return fmt.Errorf("invalid force value: %w", err)
	}

	c.Processing.ValidateOnly, err = parseBool(getEnv("INPUT_VALIDATE_ONLY", "false"))
	if err != nil {
		return fmt.Errorf("invalid validate-only value: %w", err)
	}

	return nil
}

// loadLoggingConfig loads logging configuration
func (c *Config) loadLoggingConfig() error {
	c.Logging.Level = getEnv("INPUT_LOG_LEVEL", "info")
	c.Logging.Format = getEnv("INPUT_LOG_FORMAT", "json")

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s (valid: %v)", c.Logging.Level, validLevels)
	}

	// Validate log format
	validFormats := []string{"json", "text"}
	if !contains(validFormats, c.Logging.Format) {
		return fmt.Errorf("invalid log format: %s (valid: %v)", c.Logging.Format, validFormats)
	}

	return nil
}

// loadGitHubConfig loads GitHub Actions specific configuration
func (c *Config) loadGitHubConfig() error {
	c.GitHub.Workspace = getEnv("GITHUB_WORKSPACE", "")
	c.GitHub.EventPath = getEnv("GITHUB_EVENT_PATH", "")
	c.GitHub.Token = getEnv("GITHUB_TOKEN", "")

	return nil
}

// validate validates the configuration
func (c *Config) validate() error {
	// Validate required Nobl9 credentials
	if c.Nobl9.ClientID == "" {
		return fmt.Errorf("Nobl9 client ID is required")
	}
	if c.Nobl9.ClientSecret == "" {
		return fmt.Errorf("Nobl9 client secret is required")
	}

	// Validate repository path
	if c.Repository.Path == "" {
		return fmt.Errorf("repository path cannot be empty")
	}

	// Validate file pattern
	if c.Repository.FilePattern == "" {
		return fmt.Errorf("file pattern cannot be empty")
	}

	// Validate GitHub workspace
	if c.GitHub.Workspace == "" {
		return fmt.Errorf("GitHub workspace is required")
	}

	return nil
}

// detectEnvironment detects Nobl9 environment from credentials
func (c *Config) detectEnvironment() string {
	// This is a simplified detection - in practice, you might need to
	// make an API call to determine the environment
	clientID := strings.ToLower(c.Nobl9.ClientID)

	switch {
	case strings.Contains(clientID, "dev"):
		return "dev"
	case strings.Contains(clientID, "staging"):
		return "staging"
	case strings.Contains(clientID, "prod"):
		return "prod"
	default:
		return "unknown"
	}
}

// IsGitHubActions returns true if running in GitHub Actions
func (c *Config) IsGitHubActions() bool {
	return getEnv("GITHUB_ACTIONS", "") == "true"
}

// IsDryRun returns true if dry run mode is enabled
func (c *Config) IsDryRun() bool {
	return c.Processing.DryRun
}

// IsValidateOnly returns true if validate-only mode is enabled
func (c *Config) IsValidateOnly() bool {
	return c.Processing.ValidateOnly
}

// IsForce returns true if force mode is enabled
func (c *Config) IsForce() bool {
	return c.Processing.Force
}

// GetNobl9Credentials returns Nobl9 credentials for API calls
func (c *Config) GetNobl9Credentials() (string, string) {
	return c.Nobl9.ClientID, c.Nobl9.ClientSecret
}

// GetRepositoryPath returns the full repository path
func (c *Config) GetRepositoryPath() string {
	if c.IsGitHubActions() {
		return c.GitHub.Workspace
	}
	return c.Repository.Path
}

// Helper functions

// getEnv gets an environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseBool parses a boolean string
func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
