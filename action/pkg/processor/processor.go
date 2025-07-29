package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/your-org/nobl9-action/pkg/errors"
	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
	"github.com/your-org/nobl9-action/pkg/parser"
	"github.com/your-org/nobl9-action/pkg/resolver"
	"github.com/your-org/nobl9-action/pkg/scanner"
	"github.com/your-org/nobl9-action/pkg/validator"
)

// Processor handles the processing of Nobl9 configuration files
type Processor struct {
	client    *nobl9.Client
	parser    *parser.Parser
	resolver  *resolver.Resolver
	validator *validator.Validator
	logger    *logger.Logger
}

// ProcessingResult represents the result of processing files
type ProcessingResult struct {
	FilesProcessed      int
	FilesSkipped        int
	FilesWithErrors     int
	ProjectsCreated     int
	ProjectsUpdated     int
	RoleBindingsCreated int
	RoleBindingsUpdated int
	UsersResolved       int
	UsersUnresolved     int
	Errors              []error
	Warnings            []string
	Duration            time.Duration
	IsSuccess           bool
}

// FileProcessingResult represents the result of processing a single file
type FileProcessingResult struct {
	FileInfo            *scanner.FileInfo
	ParseResult         *parser.ParseResult
	ResolutionResult    *resolver.BatchResolutionResult
	ProjectsCreated     int
	ProjectsUpdated     int
	RoleBindingsCreated int
	RoleBindingsUpdated int
	UsersResolved       int
	UsersUnresolved     int
	Errors              []error
	Warnings            []string
	Duration            time.Duration
	IsSuccess           bool
}

// New creates a new processor instance
func New(client *nobl9.Client, log *logger.Logger) *Processor {
	parser := parser.New(client, log)
	resolver := resolver.New(client, log)
	validator := validator.New(client, resolver, log)

	return &Processor{
		client:    client,
		parser:    parser,
		resolver:  resolver,
		validator: validator,
		logger:    log,
	}
}

// ProcessFiles processes multiple Nobl9 configuration files
func (p *Processor) ProcessFiles(ctx context.Context, files []*scanner.FileInfo) (*ProcessingResult, error) {
	start := time.Now()

	p.logger.Info("Starting file processing", logger.Fields{
		"file_count": len(files),
	})

	result := &ProcessingResult{
		FilesProcessed:      0,
		FilesSkipped:        0,
		FilesWithErrors:     0,
		ProjectsCreated:     0,
		ProjectsUpdated:     0,
		RoleBindingsCreated: 0,
		RoleBindingsUpdated: 0,
		UsersResolved:       0,
		UsersUnresolved:     0,
		Errors:              make([]error, 0),
		Warnings:            make([]string, 0),
		IsSuccess:           true,
	}

	// Create error aggregator for comprehensive error tracking
	errorAggregator := errors.NewErrorAggregator()

	for _, fileInfo := range files {
		fileResult, err := p.ProcessFile(ctx, fileInfo)
		if err != nil {
			processingErr := errors.NewFileProcessingError(fmt.Sprintf("failed to process file %s", fileInfo.Path), err)
			errorAggregator.AddError(processingErr)
			result.Errors = append(result.Errors, processingErr)
			result.FilesWithErrors++
			result.IsSuccess = false
			continue
		}

		// Aggregate results
		result.FilesProcessed++
		result.ProjectsCreated += fileResult.ProjectsCreated
		result.ProjectsUpdated += fileResult.ProjectsUpdated
		result.RoleBindingsCreated += fileResult.RoleBindingsCreated
		result.RoleBindingsUpdated += fileResult.RoleBindingsUpdated
		result.UsersResolved += fileResult.UsersResolved
		result.UsersUnresolved += fileResult.UsersUnresolved
		result.Errors = append(result.Errors, fileResult.Errors...)
		result.Warnings = append(result.Warnings, fileResult.Warnings...)

		if !fileResult.IsSuccess {
			result.FilesWithErrors++
			result.IsSuccess = false
		}
	}

	result.Duration = time.Since(start)

	// Log comprehensive processing summary with error details
	errorSummary := errorAggregator.GetErrorSummary()
	p.logger.Info("File processing completed", logger.Fields{
		"files_processed":       result.FilesProcessed,
		"files_skipped":         result.FilesSkipped,
		"files_with_errors":     result.FilesWithErrors,
		"projects_created":      result.ProjectsCreated,
		"projects_updated":      result.ProjectsUpdated,
		"role_bindings_created": result.RoleBindingsCreated,
		"role_bindings_updated": result.RoleBindingsUpdated,
		"users_resolved":        result.UsersResolved,
		"users_unresolved":      result.UsersUnresolved,
		"errors":                len(result.Errors),
		"warnings":              len(result.Warnings),
		"duration":              result.Duration.String(),
		"is_success":            result.IsSuccess,
		"error_summary":         errorSummary,
	})

	// Log detailed error information if there are errors
	if errorAggregator.HasErrors() {
		p.logger.LogDetailedError(fmt.Errorf("processing completed with errors"), "file processing", map[string]interface{}{
			"total_files":       len(files),
			"files_processed":   result.FilesProcessed,
			"files_with_errors": result.FilesWithErrors,
			"error_summary":     errorSummary,
		}, logger.Fields{
			"error_count": len(result.Errors),
		})
	}

	return result, nil
}

// ProcessFile processes a single Nobl9 configuration file
func (p *Processor) ProcessFile(ctx context.Context, fileInfo *scanner.FileInfo) (*FileProcessingResult, error) {
	start := time.Now()

	p.logger.Info("Processing Nobl9 configuration file", logger.Fields{
		"file_path": fileInfo.Path,
		"file_size": fileInfo.Size,
	})

	result := &FileProcessingResult{
		FileInfo:            fileInfo,
		ProjectsCreated:     0,
		ProjectsUpdated:     0,
		RoleBindingsCreated: 0,
		RoleBindingsUpdated: 0,
		UsersResolved:       0,
		UsersUnresolved:     0,
		Errors:              make([]error, 0),
		Warnings:            make([]string, 0),
		IsSuccess:           true,
	}

	// Step 1: Parse the file
	parseResult, err := p.parser.ParseFile(ctx, fileInfo)
	if err != nil {
		parseErr := errors.NewFileProcessingError("failed to parse file", err)
		result.Errors = append(result.Errors, parseErr)
		result.IsSuccess = false

		p.logger.LogDetailedError(parseErr, "file parsing", map[string]interface{}{
			"file_path": fileInfo.Path,
			"file_size": fileInfo.Size,
		}, logger.Fields{
			"error": err.Error(),
		})

		return result, nil
	}

	result.ParseResult = parseResult

	if !parseResult.IsValid {
		for _, parseErr := range parseResult.Errors {
			validationErr := errors.NewValidationError("file validation failed", parseErr)
			result.Errors = append(result.Errors, validationErr)
		}
		result.IsSuccess = false

		p.logger.LogDetailedError(fmt.Errorf("file validation failed"), "file validation", map[string]interface{}{
			"file_path":   fileInfo.Path,
			"error_count": len(parseResult.Errors),
		}, logger.Fields{
			"errors": parseResult.Errors,
		})

		return result, nil
	}

	// Step 2: Resolve emails to UserIDs
	resolutionResult, err := p.resolver.ResolveEmailsFromYAML(ctx, fileInfo.Content)
	if err != nil {
		resolutionErr := errors.NewUserResolutionError("failed to resolve emails", err)
		result.Errors = append(result.Errors, resolutionErr)
		result.IsSuccess = false

		p.logger.LogDetailedError(resolutionErr, "email resolution", map[string]interface{}{
			"file_path": fileInfo.Path,
		}, logger.Fields{
			"error": err.Error(),
		})

		return result, nil
	}

	result.ResolutionResult = resolutionResult
	result.UsersResolved = resolutionResult.ResolvedCount
	result.UsersUnresolved = resolutionResult.ErrorCount

	// Step 3: Process valid objects
	for _, obj := range parseResult.ValidObjects {
		if err := p.processObject(ctx, obj, resolutionResult); err != nil {
			processingErr := errors.NewFileProcessingError(fmt.Sprintf("failed to process object %s", obj.GetName()), err)
			result.Errors = append(result.Errors, processingErr)
			result.IsSuccess = false

			p.logger.LogDetailedError(processingErr, "object processing", map[string]interface{}{
				"file_path":   fileInfo.Path,
				"object_kind": obj.GetKind(),
				"object_name": obj.GetName(),
			}, logger.Fields{
				"error": err.Error(),
			})
		}
	}

	// Step 4: Apply manifests if any valid objects exist
	if len(parseResult.ValidObjects) > 0 {
		if err := p.applyManifest(ctx, fileInfo.Content); err != nil {
			manifestErr := errors.NewManifestError("failed to apply manifest", err)
			result.Errors = append(result.Errors, manifestErr)
			result.IsSuccess = false

			p.logger.LogDetailedError(manifestErr, "manifest application", map[string]interface{}{
				"file_path":    fileInfo.Path,
				"object_count": len(parseResult.ValidObjects),
			}, logger.Fields{
				"error": err.Error(),
			})
		}
	}

	result.Duration = time.Since(start)

	p.logger.Info("File processing completed", logger.Fields{
		"file_path":             fileInfo.Path,
		"projects_created":      result.ProjectsCreated,
		"projects_updated":      result.ProjectsUpdated,
		"role_bindings_created": result.RoleBindingsCreated,
		"role_bindings_updated": result.RoleBindingsUpdated,
		"users_resolved":        result.UsersResolved,
		"users_unresolved":      result.UsersUnresolved,
		"errors":                len(result.Errors),
		"warnings":              len(result.Warnings),
		"duration":              result.Duration.String(),
		"is_success":            result.IsSuccess,
	})

	return result, nil
}

// processObject processes a single Nobl9 object
func (p *Processor) processObject(ctx context.Context, obj manifest.Object, resolutionResult *resolver.BatchResolutionResult) error {
	kind := obj.GetKind()
	name := obj.GetName()

	p.logger.Debug("Processing Nobl9 object", logger.Fields{
		"kind": kind,
		"name": name,
	})

	// Get resolved UserIDs for role bindings
	emailToUserID := p.resolver.GetResolvedUserIDs(resolutionResult)

	switch kind {
	case manifest.KindProject:
		return p.processProject(ctx, obj)
	case manifest.KindRoleBinding:
		return p.processRoleBinding(ctx, obj, emailToUserID)
	default:
		p.logger.Debug("Skipping object type", logger.Fields{
			"kind": kind,
			"name": name,
		})
		return nil
	}
}

// processProject processes a Project object
func (p *Processor) processProject(ctx context.Context, obj manifest.Object) error {
	name := obj.GetName()

	p.logger.Debug("Processing project", logger.Fields{
		"project_name": name,
	})

	// Check if project exists
	existingProject, err := p.client.GetProject(ctx, name)
	if err != nil {
		// Project doesn't exist, will be created
		p.logger.Info("Project will be created", logger.Fields{
			"project_name": name,
		})
		return nil
	}

	// Project exists, will be updated
	p.logger.Info("Project will be updated", logger.Fields{
		"project_name": name,
		"project_id":   existingProject.Metadata.Name,
	})

	return nil
}

// processRoleBinding processes a RoleBinding object
func (p *Processor) processRoleBinding(ctx context.Context, obj manifest.Object, emailToUserID map[string]string) error {
	name := obj.GetName()

	p.logger.Debug("Processing role binding", logger.Fields{
		"role_binding_name": name,
	})

	// Validate role binding before processing
	if roleBindingObj, ok := obj.(*rolebinding.RoleBinding); ok {
		validation, err := p.validator.ValidateRoleBinding(ctx, roleBindingObj, emailToUserID)
		if err != nil {
			return errors.NewValidationError("failed to validate role binding", err)
		}

		if !validation.IsValid {
			p.logger.LogDetailedError(fmt.Errorf("role binding validation failed"), "role binding validation", map[string]interface{}{
				"role_binding_name": name,
				"error_count":       len(validation.Errors),
				"warning_count":     len(validation.Warnings),
			}, logger.Fields{
				"errors":   validation.Errors,
				"warnings": validation.Warnings,
			})

			// Return the first validation error
			if len(validation.Errors) > 0 {
				return validation.Errors[0]
			}
		}

		// Log validation summary
		summary := p.validator.GetValidationSummary(validation)
		p.logger.Info("Role binding validation completed", logger.Fields{
			"role_binding_name": summary["role_binding_name"],
			"project_name":      summary["project_name"],
			"role":              summary["role"],
			"is_valid":          summary["is_valid"],
			"total_users":       summary["total_users"],
			"valid_users":       summary["valid_users"],
			"invalid_users":     summary["invalid_users"],
			"error_count":       summary["error_count"],
			"warning_count":     summary["warning_count"],
			"duration":          summary["duration"],
		})
	}

	p.logger.Info("Role binding will be processed", logger.Fields{
		"role_binding_name": name,
		"resolved_users":    len(emailToUserID),
	})

	return nil
}

// applyManifest applies a Nobl9 manifest
func (p *Processor) applyManifest(ctx context.Context, content []byte) error {
	p.logger.Debug("Applying Nobl9 manifest", logger.Fields{
		"content_size": len(content),
	})

	// Use the parser to apply the manifest
	err := p.parser.ApplyManifest(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to apply manifest: %w", err)
	}

	p.logger.Info("Manifest applied successfully")

	return nil
}

// ProcessWithDryRun processes files in dry-run mode (validation only)
func (p *Processor) ProcessWithDryRun(ctx context.Context, files []*scanner.FileInfo) (*ProcessingResult, error) {
	start := time.Now()

	p.logger.Info("Starting dry-run processing", logger.Fields{
		"file_count": len(files),
	})

	result := &ProcessingResult{
		FilesProcessed:      0,
		FilesSkipped:        0,
		FilesWithErrors:     0,
		ProjectsCreated:     0,
		ProjectsUpdated:     0,
		RoleBindingsCreated: 0,
		RoleBindingsUpdated: 0,
		UsersResolved:       0,
		UsersUnresolved:     0,
		Errors:              make([]error, 0),
		Warnings:            make([]string, 0),
		IsSuccess:           true,
	}

	for _, fileInfo := range files {
		fileResult, err := p.ProcessFileWithDryRun(ctx, fileInfo)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to process file %s: %w", fileInfo.Path, err))
			result.FilesWithErrors++
			result.IsSuccess = false
			continue
		}

		// Aggregate results
		result.FilesProcessed++
		result.ProjectsCreated += fileResult.ProjectsCreated
		result.ProjectsUpdated += fileResult.ProjectsUpdated
		result.RoleBindingsCreated += fileResult.RoleBindingsCreated
		result.RoleBindingsUpdated += fileResult.RoleBindingsUpdated
		result.UsersResolved += fileResult.UsersResolved
		result.UsersUnresolved += fileResult.UsersUnresolved
		result.Errors = append(result.Errors, fileResult.Errors...)
		result.Warnings = append(result.Warnings, fileResult.Warnings...)

		if !fileResult.IsSuccess {
			result.FilesWithErrors++
			result.IsSuccess = false
		}
	}

	result.Duration = time.Since(start)

	p.logger.Info("Dry-run processing completed", logger.Fields{
		"files_processed":       result.FilesProcessed,
		"files_skipped":         result.FilesSkipped,
		"files_with_errors":     result.FilesWithErrors,
		"projects_created":      result.ProjectsCreated,
		"projects_updated":      result.ProjectsUpdated,
		"role_bindings_created": result.RoleBindingsCreated,
		"role_bindings_updated": result.RoleBindingsUpdated,
		"users_resolved":        result.UsersResolved,
		"users_unresolved":      result.UsersUnresolved,
		"errors":                len(result.Errors),
		"warnings":              len(result.Warnings),
		"duration":              result.Duration.String(),
		"is_success":            result.IsSuccess,
	})

	return result, nil
}

// ProcessFileWithDryRun processes a single file in dry-run mode
func (p *Processor) ProcessFileWithDryRun(ctx context.Context, fileInfo *scanner.FileInfo) (*FileProcessingResult, error) {
	start := time.Now()

	p.logger.Info("Processing file in dry-run mode", logger.Fields{
		"file_path": fileInfo.Path,
		"file_size": fileInfo.Size,
	})

	result := &FileProcessingResult{
		FileInfo:            fileInfo,
		ProjectsCreated:     0,
		ProjectsUpdated:     0,
		RoleBindingsCreated: 0,
		RoleBindingsUpdated: 0,
		UsersResolved:       0,
		UsersUnresolved:     0,
		Errors:              make([]error, 0),
		Warnings:            make([]string, 0),
		IsSuccess:           true,
	}

	// Step 1: Parse the file
	parseResult, err := p.parser.ParseFile(ctx, fileInfo)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to parse file: %w", err))
		result.IsSuccess = false
		return result, nil
	}

	result.ParseResult = parseResult

	if !parseResult.IsValid {
		result.Errors = append(result.Errors, parseResult.Errors...)
		result.IsSuccess = false
		return result, nil
	}

	// Step 2: Resolve emails to UserIDs
	resolutionResult, err := p.resolver.ResolveEmailsFromYAML(ctx, fileInfo.Content)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to resolve emails: %w", err))
		result.IsSuccess = false
		return result, nil
	}

	result.ResolutionResult = resolutionResult
	result.UsersResolved = resolutionResult.ResolvedCount
	result.UsersUnresolved = resolutionResult.ErrorCount

	// Step 3: Validate manifest (dry-run)
	if len(parseResult.ValidObjects) > 0 {
		if err := p.validateManifest(ctx, fileInfo.Content); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("manifest validation failed: %w", err))
			result.IsSuccess = false
		}
	}

	// Step 4: Simulate processing (dry-run)
	for _, obj := range parseResult.ValidObjects {
		if err := p.simulateObjectProcessing(ctx, obj, resolutionResult); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to simulate processing object %s: %w", obj.GetName(), err))
			result.IsSuccess = false
		}
	}

	result.Duration = time.Since(start)

	p.logger.Info("Dry-run file processing completed", logger.Fields{
		"file_path":             fileInfo.Path,
		"projects_created":      result.ProjectsCreated,
		"projects_updated":      result.ProjectsUpdated,
		"role_bindings_created": result.RoleBindingsCreated,
		"role_bindings_updated": result.RoleBindingsUpdated,
		"users_resolved":        result.UsersResolved,
		"users_unresolved":      result.UsersUnresolved,
		"errors":                len(result.Errors),
		"warnings":              len(result.Warnings),
		"duration":              result.Duration.String(),
		"is_success":            result.IsSuccess,
	})

	return result, nil
}

// validateManifest validates a Nobl9 manifest
func (p *Processor) validateManifest(ctx context.Context, content []byte) error {
	p.logger.Debug("Validating Nobl9 manifest", logger.Fields{
		"content_size": len(content),
	})

	// Use the parser to validate the manifest
	err := p.parser.ValidateManifest(ctx, content)
	if err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	p.logger.Info("Manifest validation successful")

	return nil
}

// simulateObjectProcessing simulates processing an object in dry-run mode
func (p *Processor) simulateObjectProcessing(ctx context.Context, obj manifest.Object, resolutionResult *resolver.BatchResolutionResult) error {
	kind := obj.GetKind()
	name := obj.GetName()

	p.logger.Debug("Simulating object processing", logger.Fields{
		"kind": kind,
		"name": name,
	})

	// Get resolved UserIDs for role bindings
	emailToUserID := p.resolver.GetResolvedUserIDs(resolutionResult)

	switch kind {
	case manifest.KindProject:
		return p.simulateProjectProcessing(ctx, obj)
	case manifest.KindRoleBinding:
		return p.simulateRoleBindingProcessing(ctx, obj, emailToUserID)
	default:
		p.logger.Debug("Skipping object type in simulation", logger.Fields{
			"kind": kind,
			"name": name,
		})
		return nil
	}
}

// simulateProjectProcessing simulates processing a Project object
func (p *Processor) simulateProjectProcessing(ctx context.Context, obj manifest.Object) error {
	name := obj.GetName()

	p.logger.Debug("Simulating project processing", logger.Fields{
		"project_name": name,
	})

	// Check if project exists
	existingProject, err := p.client.GetProject(ctx, name)
	if err != nil {
		// Project doesn't exist, would be created
		p.logger.Info("Project would be created (dry-run)", logger.Fields{
			"project_name": name,
		})
		return nil
	}

	// Project exists, would be updated
	p.logger.Info("Project would be updated (dry-run)", logger.Fields{
		"project_name": name,
		"project_id":   existingProject.Metadata.Name,
	})

	return nil
}

// simulateRoleBindingProcessing simulates processing a RoleBinding object
func (p *Processor) simulateRoleBindingProcessing(ctx context.Context, obj manifest.Object, emailToUserID map[string]string) error {
	name := obj.GetName()

	p.logger.Debug("Simulating role binding processing", logger.Fields{
		"role_binding_name": name,
	})

	p.logger.Info("Role binding would be processed (dry-run)", logger.Fields{
		"role_binding_name": name,
		"resolved_users":    len(emailToUserID),
	})

	return nil
}

// GetProcessingStats returns processing statistics
func (p *Processor) GetProcessingStats(result *ProcessingResult) map[string]interface{} {
	return map[string]interface{}{
		"files_processed":       result.FilesProcessed,
		"files_skipped":         result.FilesSkipped,
		"files_with_errors":     result.FilesWithErrors,
		"projects_created":      result.ProjectsCreated,
		"projects_updated":      result.ProjectsUpdated,
		"role_bindings_created": result.RoleBindingsCreated,
		"role_bindings_updated": result.RoleBindingsUpdated,
		"users_resolved":        result.UsersResolved,
		"users_unresolved":      result.UsersUnresolved,
		"errors":                len(result.Errors),
		"warnings":              len(result.Warnings),
		"duration":              result.Duration.String(),
		"is_success":            result.IsSuccess,
	}
}

// GetUnresolvedEmails returns all unresolved emails from processing
func (p *Processor) GetUnresolvedEmails(result *ProcessingResult) []string {
	// This would need to be implemented based on how unresolved emails are tracked
	// For now, return an empty slice
	return []string{}
}

// GetProcessingErrors returns all processing errors
func (p *Processor) GetProcessingErrors(result *ProcessingResult) []string {
	errors := make([]string, len(result.Errors))
	for i, err := range result.Errors {
		errors[i] = err.Error()
	}
	return errors
}
