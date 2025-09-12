package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Root command
var rootCmd = &cobra.Command{
	Use:     "nobl9-action",
	Short:   "Nobl9 GitHub Action for automated project and role management",
	Long:    `A GitHub Action that processes Nobl9 YAML configurations, resolves email addresses to Okta User IDs, and deploys projects and role bindings to Nobl9.`,
	Version: "1.0.0",
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
		return fmt.Errorf("invalid log level: %w", err)
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
		return fmt.Errorf("invalid log format: %s", config.LogFormat)
	}

	return nil
}

// runProcess executes the main processing logic
func runProcess(cmd *cobra.Command, args []string) error {
	logrus.Info("Starting Nobl9 GitHub Action processing")

	// Setup logging
	if err := setupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	// Validate configuration
	if err := validateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Step 1: Scan repository for YAML files
	logrus.WithFields(logrus.Fields{
		"repo_path":    config.RepoPath,
		"file_pattern": config.FilePattern,
	}).Info("Scanning for Nobl9 YAML files")

	files, err := scanFiles(config.RepoPath, config.FilePattern)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	if len(files) == 0 {
		logrus.Warn("No YAML files found matching pattern")
		return nil
	}

	logrus.WithField("file_count", len(files)).Info("Found YAML files to process")

	// Step 2: Initialize Nobl9 client
	nobl9Client, err := createNobl9Client(config.ClientID, config.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to create Nobl9 client: %w", err)
	}

	// Step 3: Process each file
	var totalProcessed, totalErrors, projectsCreated, roleBindingsCreated, emailsResolved int

	for _, filePath := range files {
		logrus.WithField("file", filePath).Info("Processing file")

		result, err := processFile(ctx, nobl9Client, filePath, config.DryRun)
		if err != nil {
			logrus.WithField("file", filePath).WithError(err).Error("Failed to process file")
			totalErrors++
			continue
		}

		totalProcessed++
		projectsCreated += result.ProjectsCreated
		roleBindingsCreated += result.RoleBindingsCreated
		emailsResolved += result.EmailsResolved

		logrus.WithFields(logrus.Fields{
			"file":            filePath,
			"projects":        result.ProjectsCreated,
			"role_bindings":   result.RoleBindingsCreated,
			"emails_resolved": result.EmailsResolved,
		}).Info("File processed successfully")
	}

	// Step 4: Log final summary
	logrus.WithFields(logrus.Fields{
		"total_files":           len(files),
		"files_processed":       totalProcessed,
		"files_with_errors":     totalErrors,
		"projects_created":      projectsCreated,
		"role_bindings_created": roleBindingsCreated,
		"emails_resolved":       emailsResolved,
		"dry_run":               config.DryRun,
	}).Info("Processing completed")

	// Set GitHub Action outputs if running in GitHub Actions
	setGitHubOutput("processed-files", fmt.Sprintf("%d", totalProcessed))
	setGitHubOutput("projects-created", fmt.Sprintf("%d", projectsCreated))
	setGitHubOutput("projects-updated", "0") // Not currently tracked
	setGitHubOutput("role-bindings-created", fmt.Sprintf("%d", roleBindingsCreated))
	setGitHubOutput("role-bindings-updated", "0") // Not currently tracked
	setGitHubOutput("users-resolved", fmt.Sprintf("%d", emailsResolved))
	setGitHubOutput("errors", fmt.Sprintf("%d", totalErrors))
	setGitHubOutput("success", fmt.Sprintf("%t", totalErrors == 0))

	if totalErrors > 0 {
		return fmt.Errorf("processing completed with %d errors", totalErrors)
	}

	return nil
}

// runValidate executes validation logic
func runValidate(cmd *cobra.Command, args []string) error {
	logrus.Info("Starting Nobl9 YAML validation")

	// Setup logging
	if err := setupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Step 1: Scan repository for YAML files
	logrus.WithFields(logrus.Fields{
		"repo_path":    config.RepoPath,
		"file_pattern": config.FilePattern,
	}).Info("Scanning for YAML files to validate")

	files, err := scanFiles(config.RepoPath, config.FilePattern)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	if len(files) == 0 {
		logrus.Warn("No YAML files found matching pattern")
		return nil
	}

	logrus.WithField("file_count", len(files)).Info("Found YAML files to validate")

	// Step 2: Validate each file
	var totalValidated, totalErrors int

	for _, filePath := range files {
		logrus.WithField("file", filePath).Info("Validating file")

		if err := validateFile(ctx, filePath); err != nil {
			logrus.WithField("file", filePath).WithError(err).Error("File validation failed")
			totalErrors++
		} else {
			logrus.WithField("file", filePath).Info("File validation passed")
			totalValidated++
		}
	}

	// Step 3: Log validation summary
	logrus.WithFields(logrus.Fields{
		"total_files":       len(files),
		"files_validated":   totalValidated,
		"files_with_errors": totalErrors,
	}).Info("Validation completed")

	// Set GitHub Action outputs for validation
	setGitHubOutput("processed-files", fmt.Sprintf("%d", totalValidated))
	setGitHubOutput("projects-created", "0") // Validation mode
	setGitHubOutput("projects-updated", "0") // Validation mode
	setGitHubOutput("role-bindings-created", "0") // Validation mode
	setGitHubOutput("role-bindings-updated", "0") // Validation mode
	setGitHubOutput("users-resolved", "0") // Validation mode
	setGitHubOutput("errors", fmt.Sprintf("%d", totalErrors))
	setGitHubOutput("success", fmt.Sprintf("%t", totalErrors == 0))

	if totalErrors > 0 {
		return fmt.Errorf("validation completed with %d errors", totalErrors)
	}

	return nil
}

// validateConfig validates the application configuration
func validateConfig() error {
	if config.ClientID == "" {
		return fmt.Errorf("client-id is required")
	}
	if config.ClientSecret == "" {
		return fmt.Errorf("client-secret is required")
	}
	if config.RepoPath == "" {
		return fmt.Errorf("repo-path cannot be empty")
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
	errStr := err.Error()

	// Check for specific error patterns in the error message
	switch {
	case contains(errStr, "configuration", "config"):
		return 2
	case contains(errStr, "validation", "invalid"):
		return 3
	case contains(errStr, "nobl9", "api"):
		return 4
	case contains(errStr, "file", "processing"):
		return 5
	case contains(errStr, "auth", "credentials"):
		return 6
	case contains(errStr, "network", "connection"):
		return 7
	case contains(errStr, "rate limit"):
		return 8
	case contains(errStr, "timeout"):
		return 9
	default:
		return 1
	}
}

// contains checks if error message contains any of the keywords
func contains(str string, keywords ...string) bool {
	lowerStr := strings.ToLower(str)
	for _, keyword := range keywords {
		if strings.Contains(lowerStr, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// Helper functions implementing the core functionality

// scanFiles scans for YAML files matching the given pattern
func scanFiles(repoPath, filePattern string) ([]string, error) {
	pattern := filepath.Join(repoPath, filePattern)

	// Use doublestar for glob pattern matching (supports **)
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob pattern %s: %w", pattern, err)
	}

	var files []string
	for _, match := range matches {
		// Check if it's a regular file
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if !info.IsDir() && isYAMLFile(match) {
			files = append(files, match)
		}
	}

	return files, nil
}

// isYAMLFile checks if the file has a YAML extension
func isYAMLFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".yaml" || ext == ".yml"
}

// createNobl9Client creates and initializes a Nobl9 SDK client
func createNobl9Client(clientID, clientSecret string) (*sdk.Client, error) {
	// Set environment variables for the Nobl9 SDK (like your lambda)
	os.Setenv("NOBL9_SDK_CLIENT_ID", clientID)
	os.Setenv("NOBL9_SDK_CLIENT_SECRET", clientSecret)

	// Fix for environments where HOME is not set properly
	if os.Getenv("HOME") == "" {
		os.Setenv("HOME", "/tmp")
	}

	// Initialize the Nobl9 client using the same method as your lambda
	client, err := sdk.DefaultClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Nobl9 SDK client: %w", err)
	}

	return client, nil
}

// ProcessResult represents the result of processing a single file
type ProcessResult struct {
	ProjectsCreated     int
	RoleBindingsCreated int
	EmailsResolved      int
}

// processFile processes a single YAML file using patterns from your lambda
func processFile(ctx context.Context, client *sdk.Client, filePath string, dryRun bool) (*ProcessResult, error) {
	result := &ProcessResult{}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if it contains Nobl9 configuration
	if !isNobl9File(content) {
		logrus.WithField("file", filePath).Debug("File does not contain Nobl9 configuration, skipping")
		return result, nil
	}

	// Parse YAML documents
	objects, emailsToResolve, err := parseYAMLContent(content, filePath)
	if err != nil {
		return result, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(objects) == 0 {
		logrus.WithField("file", filePath).Debug("No valid objects found in file")
		return result, nil
	}

	// Resolve emails to user IDs
	emailResolutions := make(map[string]string)
	if len(emailsToResolve) > 0 {
		logrus.WithField("email_count", len(emailsToResolve)).Debug("Resolving email addresses")

		for _, email := range emailsToResolve {
			userID, err := resolveEmailToUserID(ctx, client, email)
			if err != nil {
				logrus.WithField("email", email).WithError(err).Warn("Failed to resolve email")
				continue
			}
			emailResolutions[email] = userID
			result.EmailsResolved++
		}
	}

	// Update role bindings with resolved user IDs
	for _, obj := range objects {
		if rb, ok := obj.(v1alphaRoleBinding.RoleBinding); ok {
			if rb.Spec.User != nil {
				originalUser := *rb.Spec.User
				if resolvedID, found := emailResolutions[originalUser]; found {
					rb.Spec.User = &resolvedID
					logrus.WithFields(logrus.Fields{
						"email":   originalUser,
						"user_id": resolvedID,
					}).Debug("Email resolved for role binding")
				}
			}
		}
	}

	// Apply objects to Nobl9
	if !dryRun {
		logrus.WithField("object_count", len(objects)).Debug("Applying objects to Nobl9")

		if err := client.Objects().V1().Apply(ctx, objects); err != nil {
			// Check if the error is because objects already exist
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "conflict") {
				logrus.WithField("file", filePath).Info("Some objects already exist")
			} else {
				return result, fmt.Errorf("failed to apply objects: %w", err)
			}
		}
	} else {
		logrus.WithField("object_count", len(objects)).Info("DRY RUN: Would apply objects to Nobl9")
	}

	// Count created objects
	for _, obj := range objects {
		switch obj.GetKind() {
		case manifest.KindProject:
			result.ProjectsCreated++
		case manifest.KindRoleBinding:
			result.RoleBindingsCreated++
		}
	}

	return result, nil
}

// isNobl9File checks if file content contains Nobl9 configuration
func isNobl9File(content []byte) bool {
	contentStr := string(content)

	// Check for Nobl9-specific indicators based on the official YAML guide
	nobl9Indicators := []string{
		"apiVersion: n9/v1alpha",
		"kind: Agent",
		"kind: Alert",
		"kind: AlertMethod",
		"kind: AlertPolicy",
		"kind: AlertSilence",
		"kind: Annotation",
		"kind: BudgetAdjustment",
		"kind: DataExport",
		"kind: Direct",
		"kind: Objective",
		"kind: Project",
		"kind: Report",
		"kind: RoleBinding",
		"kind: Service",
		"kind: SLO",
		"kind: UserGroup",
		// Composite SLO indicators
		"composite:",
		"maxDelay:",
		"components:",
		"whenDelayed:",
	}

	for _, indicator := range nobl9Indicators {
		if strings.Contains(contentStr, indicator) {
			return true
		}
	}

	return false
}

// parseYAMLContent parses YAML content and extracts Nobl9 objects and emails
func parseYAMLContent(content []byte, source string) ([]manifest.Object, []string, error) {
	var emails []string

	// Parse using Nobl9 SDK first
	manifests, err := sdk.DecodeObjects(content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode objects: %w", err)
	}

	// Also manually parse to extract emails from role bindings
	documents := strings.Split(string(content), "---")
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		docEmails := extractEmailsFromDocument(doc)
		emails = append(emails, docEmails...)
	}

	// Remove duplicates from emails
	emailSet := make(map[string]bool)
	uniqueEmails := []string{}
	for _, email := range emails {
		if !emailSet[email] {
			emailSet[email] = true
			uniqueEmails = append(uniqueEmails, email)
		}
	}

	return manifests, uniqueEmails, nil
}

// extractEmailsFromDocument extracts email addresses from a YAML document
func extractEmailsFromDocument(docContent string) []string {
	var emails []string

	// Parse to find RoleBinding objects and extract user emails
	var doc map[string]interface{}
	if err := yaml.Unmarshal([]byte(docContent), &doc); err != nil {
		return emails
	}

	kind, ok := doc["kind"].(string)
	if !ok || kind != "RoleBinding" {
		return emails
	}

	spec, ok := doc["spec"].(map[string]interface{})
	if !ok {
		return emails
	}

	// Extract emails from different user fields
	if user, exists := spec["user"]; exists {
		if userStr, ok := user.(string); ok && isEmail(userStr) {
			emails = append(emails, userStr)
		}
	}

	if users, exists := spec["users"]; exists {
		if usersList, ok := users.([]interface{}); ok {
			for _, user := range usersList {
				if userStr, ok := user.(string); ok && isEmail(userStr) {
					emails = append(emails, userStr)
				}
			}
		}
	}

	if userIDs, exists := spec["userIds"]; exists {
		if userIDsStr, ok := userIDs.(string); ok {
			csvUsers := strings.Split(userIDsStr, ",")
			for _, user := range csvUsers {
				user = strings.TrimSpace(user)
				if user != "" && isEmail(user) {
					emails = append(emails, user)
				}
			}
		}
	}

	return emails
}

// isEmail checks if string is an email
func isEmail(s string) bool {
	return strings.Contains(s, "@")
}

// resolveEmailToUserID resolves an email address to a user ID using Nobl9 API
func resolveEmailToUserID(ctx context.Context, client *sdk.Client, email string) (string, error) {
	// Use Nobl9 SDK to get user by email (same as your lambda)
	user, err := client.Users().V2().GetUser(ctx, email)
	if err != nil {
		return "", fmt.Errorf("error retrieving user '%s': %w", email, err)
	}
	if user == nil {
		return "", fmt.Errorf("user with email '%s' not found in Nobl9", email)
	}

	return user.UserID, nil
}

// validateFile validates a single YAML file
func validateFile(ctx context.Context, filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check if it's a YAML file
	if !isYAMLFile(filePath) {
		return fmt.Errorf("file is not a YAML file")
	}

	// Check if it contains Nobl9 configuration
	if !isNobl9File(content) {
		return fmt.Errorf("file does not contain Nobl9 configuration")
	}

	// Parse and validate YAML structure
	_, err = sdk.DecodeObjects(content)
	if err != nil {
		return fmt.Errorf("invalid Nobl9 YAML: %w", err)
	}

	return nil
}

// setGitHubOutput sets a GitHub Action output variable
func setGitHubOutput(name, value string) {
	// Check if we're running in GitHub Actions
	githubOutputFile := os.Getenv("GITHUB_OUTPUT")
	if githubOutputFile == "" {
		// Not running in GitHub Actions, skip
		return
	}

	// Append to the GitHub output file
	file, err := os.OpenFile(githubOutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.WithField("error", err).Warn("Failed to open GitHub output file")
		return
	}
	defer file.Close()

	// Write the output in the format: name=value
	_, err = fmt.Fprintf(file, "%s=%s\n", name, value)
	if err != nil {
		logrus.WithField("error", err).Warn("Failed to write GitHub output")
	}
}
