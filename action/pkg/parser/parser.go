package parser

import (
	"context"
	"fmt"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
	"github.com/your-org/nobl9-action/pkg/scanner"
)

// Parser handles YAML parsing and validation for Nobl9 configuration files
type Parser struct {
	client *nobl9.Client
	logger *logger.Logger
}

// ParseResult represents the result of parsing a YAML file
type ParseResult struct {
	FileInfo       *scanner.FileInfo
	Manifests      []manifest.Object
	ValidObjects   []manifest.Object
	InvalidObjects []InvalidObject
	Errors         []error
	Warnings       []string
	IsValid        bool
}

// InvalidObject represents an invalid Nobl9 object
type InvalidObject struct {
	Object   manifest.Object
	Error    error
	Position string
}

// New creates a new parser instance
func New(client *nobl9.Client, log *logger.Logger) *Parser {
	return &Parser{
		client: client,
		logger: log,
	}
}

// ParseFile parses a single YAML file
func (p *Parser) ParseFile(ctx context.Context, fileInfo *scanner.FileInfo) (*ParseResult, error) {
	p.logger.Info("Parsing Nobl9 YAML file", logger.Fields{
		"file_path": fileInfo.Path,
		"file_size": fileInfo.Size,
	})

	result := &ParseResult{
		FileInfo:       fileInfo,
		Manifests:      make([]manifest.Object, 0),
		ValidObjects:   make([]manifest.Object, 0),
		InvalidObjects: make([]InvalidObject, 0),
		Errors:         make([]error, 0),
		Warnings:       make([]string, 0),
		IsValid:        true,
	}

	// Check if file has errors
	if fileInfo.Error != nil {
		result.Errors = append(result.Errors, fmt.Errorf("file has errors: %w", fileInfo.Error))
		result.IsValid = false
		return result, nil
	}

	// Check if file is a Nobl9 file
	if !fileInfo.IsNobl9 {
		result.Errors = append(result.Errors, fmt.Errorf("file does not contain Nobl9 configuration"))
		result.IsValid = false
		return result, nil
	}

	// Parse YAML content
	manifests, err := p.parseYAMLContent(fileInfo.Content)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to parse YAML: %w", err))
		result.IsValid = false
		return result, nil
	}

	result.Manifests = manifests

	// Validate each manifest
	for i, obj := range manifests {
		if err := p.validateObject(ctx, obj); err != nil {
			invalidObj := InvalidObject{
				Object:   obj,
				Error:    err,
				Position: fmt.Sprintf("object %d", i+1),
			}
			result.InvalidObjects = append(result.InvalidObjects, invalidObj)
			result.Errors = append(result.Errors, fmt.Errorf("object %d validation failed: %w", i+1, err))
			result.IsValid = false
		} else {
			result.ValidObjects = append(result.ValidObjects, obj)
		}
	}

	// Log parsing results
	p.logger.Info("YAML file parsing completed", logger.Fields{
		"file_path":       fileInfo.Path,
		"total_objects":   len(result.Manifests),
		"valid_objects":   len(result.ValidObjects),
		"invalid_objects": len(result.InvalidObjects),
		"errors":          len(result.Errors),
		"warnings":        len(result.Warnings),
		"is_valid":        result.IsValid,
	})

	return result, nil
}

// ParseFiles parses multiple YAML files
func (p *Parser) ParseFiles(ctx context.Context, files []*scanner.FileInfo) ([]*ParseResult, error) {
	p.logger.Info("Parsing multiple Nobl9 YAML files", logger.Fields{
		"file_count": len(files),
	})

	results := make([]*ParseResult, 0, len(files))
	var allErrors []error

	for _, fileInfo := range files {
		result, err := p.ParseFile(ctx, fileInfo)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to parse file %s: %w", fileInfo.Path, err))
			continue
		}

		results = append(results, result)

		// Collect errors from individual file parsing
		if !result.IsValid {
			allErrors = append(allErrors, result.Errors...)
		}
	}

	// Log overall parsing results
	validFiles := 0
	totalObjects := 0
	validObjects := 0
	invalidObjects := 0

	for _, result := range results {
		if result.IsValid {
			validFiles++
		}
		totalObjects += len(result.Manifests)
		validObjects += len(result.ValidObjects)
		invalidObjects += len(result.InvalidObjects)
	}

	p.logger.Info("Multiple YAML files parsing completed", logger.Fields{
		"total_files":     len(files),
		"valid_files":     validFiles,
		"invalid_files":   len(files) - validFiles,
		"total_objects":   totalObjects,
		"valid_objects":   validObjects,
		"invalid_objects": invalidObjects,
		"total_errors":    len(allErrors),
	})

	if len(allErrors) > 0 {
		return results, fmt.Errorf("parsing completed with %d errors", len(allErrors))
	}

	return results, nil
}

// parseYAMLContent parses YAML content into Nobl9 manifests
func (p *Parser) parseYAMLContent(content []byte) ([]manifest.Object, error) {
	// Parse YAML using Nobl9 SDK
	manifests, err := sdk.DecodeObjects(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML content: %w", err)
	}

	p.logger.Debug("YAML content parsed successfully", logger.Fields{
		"content_size": len(content),
		"object_count": len(manifests),
	})

	return manifests, nil
}

// validateObject validates a single Nobl9 object
func (p *Parser) validateObject(ctx context.Context, obj manifest.Object) error {
	// Get object metadata for logging
	kind := obj.GetKind()
	name := obj.GetName()
	version := obj.GetVersion()

	p.logger.Debug("Validating Nobl9 object", logger.Fields{
		"kind":        kind,
		"name":        name,
		"api_version": version,
	})

	// Validate object using Nobl9 SDK
	if err := obj.Validate(); err != nil {
		return fmt.Errorf("object validation failed: %w", err)
	}

	// Basic validation for all objects
	// Note: Type-specific validation would require additional SDK methods
	p.logger.Debug("Object validation completed", logger.Fields{
		"kind": kind,
		"name": name,
	})

	return nil
}

// ValidateManifest validates a Nobl9 manifest without parsing
func (p *Parser) ValidateManifest(ctx context.Context, content []byte) error {
	p.logger.Debug("Validating Nobl9 manifest", logger.Fields{
		"content_size": len(content),
	})

	// Use the Nobl9 client to validate the manifest
	err := p.client.ValidateManifest(ctx, content)
	if err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	p.logger.Debug("Manifest validation completed successfully")

	return nil
}

// ApplyManifest applies a Nobl9 manifest
func (p *Parser) ApplyManifest(ctx context.Context, content []byte) error {
	p.logger.Debug("Applying Nobl9 manifest", logger.Fields{
		"content_size": len(content),
	})

	// Use the Nobl9 client to apply the manifest
	err := p.client.ApplyManifest(ctx, content)
	if err != nil {
		return fmt.Errorf("manifest application failed: %w", err)
	}

	p.logger.Debug("Manifest application completed successfully")

	return nil
}

// GetValidObjects returns all valid objects from parse results
func (p *Parser) GetValidObjects(results []*ParseResult) []manifest.Object {
	var validObjects []manifest.Object

	for _, result := range results {
		validObjects = append(validObjects, result.ValidObjects...)
	}

	return validObjects
}

// GetInvalidObjects returns all invalid objects from parse results
func (p *Parser) GetInvalidObjects(results []*ParseResult) []InvalidObject {
	var invalidObjects []InvalidObject

	for _, result := range results {
		invalidObjects = append(invalidObjects, result.InvalidObjects...)
	}

	return invalidObjects
}

// GetErrors returns all errors from parse results
func (p *Parser) GetErrors(results []*ParseResult) []error {
	var errors []error

	for _, result := range results {
		errors = append(errors, result.Errors...)
	}

	return errors
}

// IsValid checks if all parse results are valid
func (p *Parser) IsValid(results []*ParseResult) bool {
	for _, result := range results {
		if !result.IsValid {
			return false
		}
	}
	return true
}
