package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/your-org/nobl9-action/pkg/errors"
	"github.com/your-org/nobl9-action/version"
)

// Root command
var rootCmd = &cobra.Command{
	Use:     "nobl9-action",
	Short:   "Nobl9 GitHub Action for automated project and role management",
	Long:    `A GitHub Action that processes Nobl9 YAML configurations, resolves email addresses to Okta User IDs, and deploys projects and role bindings to Nobl9.`,
	Version: version.String(),
}

// Process command - main functionality
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process Nobl9 YAML files and deploy to Nobl9",
	Long:  `Read Nobl9 YAML configurations from a repository, validate them, resolve email addresses to Okta User IDs, and deploy projects and role bindings to Nobl9.`,
	RunE:  runProcess,
}

// Validate command - validation only
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate Nobl9 YAML files without deployment",
	Long:  `Validate Nobl9 YAML configurations for syntax and structure without deploying to Nobl9.`,
	RunE:  runValidate,
}

// Configuration flags
var (
	config struct {
		// Nobl9 credentials
		ClientID     string
		ClientSecret string

		// Repository configuration
		RepoPath    string
		FilePattern string

		// Logging
		LogLevel  string
		LogFormat string

		// Processing options
		DryRun bool
		Force  bool
	}
)

func init() {
	// Add commands to root
	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(validateCmd)

	// Process command flags
	processCmd.Flags().StringVar(&config.ClientID, "client-id", "", "Nobl9 API client ID (required)")
	processCmd.Flags().StringVar(&config.ClientSecret, "client-secret", "", "Nobl9 API client secret (required)")
	processCmd.Flags().StringVar(&config.RepoPath, "repo-path", ".", "Repository path to scan for YAML files")
	processCmd.Flags().StringVar(&config.FilePattern, "file-pattern", "**/*.yaml", "File pattern to match Nobl9 YAML files")
	processCmd.Flags().StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	processCmd.Flags().StringVar(&config.LogFormat, "log-format", "json", "Log format (json, text)")
	processCmd.Flags().BoolVar(&config.DryRun, "dry-run", false, "Perform dry run without making changes")
	processCmd.Flags().BoolVar(&config.Force, "force", false, "Force processing even if validation fails")

	// Validate command flags
	validateCmd.Flags().StringVar(&config.RepoPath, "repo-path", ".", "Repository path to scan for YAML files")
	validateCmd.Flags().StringVar(&config.FilePattern, "file-pattern", "**/*.yaml", "File pattern to match Nobl9 YAML files")
	validateCmd.Flags().StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	validateCmd.Flags().StringVar(&config.LogFormat, "log-format", "json", "Log format (json, text)")

	// Mark required flags
	if err := processCmd.MarkFlagRequired("client-id"); err != nil {
		logrus.WithError(err).Fatal("Failed to mark client-id as required")
	}
	if err := processCmd.MarkFlagRequired("client-secret"); err != nil {
		logrus.WithError(err).Fatal("Failed to mark client-secret as required")
	}
}

// setupLogging configures the logging system
func setupLogging() error {
	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return errors.NewConfigError("invalid log level", err)
	}
	logrus.SetLevel(level)

	// Set log format
	switch config.LogFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		return errors.NewConfigError(fmt.Sprintf("invalid log format: %s", config.LogFormat), nil)
	}

	return nil
}

// runProcess executes the main processing logic
func runProcess(cmd *cobra.Command, args []string) error {
	logrus.Info("Starting Nobl9 GitHub Action processing")

	// Setup logging
	if err := setupLogging(); err != nil {
		return errors.NewConfigError("failed to setup logging", err)
	}

	// Validate configuration
	if err := validateConfig(); err != nil {
		return errors.NewConfigError("configuration validation failed", err)
	}

	// TODO: Implement the main processing logic
	// This will be implemented in subsequent tasks:
	// - Scan repository for YAML files
	// - Parse and validate YAML
	// - Initialize Nobl9 client
	// - Process projects and role bindings
	// - Handle email-to-UserID resolution

	logrus.Info("Processing completed successfully")
	return nil
}

// runValidate executes validation logic
func runValidate(cmd *cobra.Command, args []string) error {
	logrus.Info("Starting Nobl9 YAML validation")

	// Setup logging
	if err := setupLogging(); err != nil {
		return errors.NewConfigError("failed to setup logging", err)
	}

	// TODO: Implement validation logic
	// This will be implemented in subsequent tasks:
	// - Scan repository for YAML files
	// - Parse and validate YAML structure
	// - Validate against Nobl9 schema
	// - Check for common issues

	logrus.Info("Validation completed successfully")
	return nil
}

// validateConfig validates the application configuration
func validateConfig() error {
	if config.ClientID == "" {
		return errors.NewConfigError("client-id is required", nil)
	}
	if config.ClientSecret == "" {
		return errors.NewConfigError("client-secret is required", nil)
	}
	if config.RepoPath == "" {
		return errors.NewConfigError("repo-path cannot be empty", nil)
	}

	return nil
}

// main function with proper error handling and exit codes
func main() {
	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		// Log the error with detailed information
		logrus.WithError(err).Error("Application failed")

		// Determine exit code based on error type
		exitCode := determineExitCode(err)

		// Exit with appropriate code
		os.Exit(exitCode)
	}
}

// determineExitCode determines the appropriate exit code based on error type
func determineExitCode(err error) int {
	// Check if it's a Nobl9 error
	if errors.IsNobl9Error(err) {
		nobl9Err := err.(*errors.Nobl9Error)

		switch nobl9Err.GetType() {
		case errors.ErrorTypeConfig:
			return 2
		case errors.ErrorTypeValidation:
			return 3
		case errors.ErrorTypeNobl9API:
			return 4
		case errors.ErrorTypeFileProcessing:
			return 5
		case errors.ErrorTypeAuth:
			return 6
		case errors.ErrorTypeNetwork:
			return 7
		case errors.ErrorTypeRateLimit:
			return 8
		case errors.ErrorTypeTimeout:
			return 9
		default:
			return 1
		}
	}

	// Check for specific error patterns in the error message
	switch {
	case errors.IsAuthError(err):
		return 6
	case errors.IsRateLimitError(err):
		return 8
	case errors.IsTimeoutError(err):
		return 9
	case errors.IsRetryableError(err):
		return 10
	default:
		return 1
	}
}
