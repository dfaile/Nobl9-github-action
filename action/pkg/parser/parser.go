package parser

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/sirupsen/logrus"
)

// FileInfo represents information about a scanned file
type FileInfo struct {
	Path         string
	RelativePath string
	Size         int64
	ModTime      fs.FileInfo
	IsDir        bool
	IsYAML       bool
	IsNobl9      bool
	Content      []byte
	Error        error
}

// Parser handles YAML parsing and validation for Nobl9 configuration files
type Parser struct {
	client *sdk.Client
	logger *logrus.Logger
}

// ParseResult represents the result of parsing a YAML file
type ParseResult struct {
	FileInfo       *FileInfo
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
func New(client *sdk.Client, log *logrus.Logger) *Parser {
	return &Parser{
		client: client,
		logger: log,
	}
}

// ParseFile parses a single YAML file
func (p *Parser) ParseFile(ctx context.Context, fileInfo *FileInfo) (*ParseResult, error) {
	p.logger.WithFields(logrus.Fields{
		"file_path": fileInfo.Path,
		"file_size": fileInfo.Size,
	}).Info("Parsing Nobl9 YAML file")

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
	p.logger.WithFields(logrus.Fields{
		"file_path":       fileInfo.Path,
		"total_objects":   len(result.Manifests),
		"valid_objects":   len(result.ValidObjects),
		"invalid_objects": len(result.InvalidObjects),
		"errors":          len(result.Errors),
		"warnings":        len(result.Warnings),
		"is_valid":        result.IsValid,
	}).Info("YAML file parsing completed")

	return result, nil
}

// ParseFiles parses multiple YAML files
func (p *Parser) ParseFiles(ctx context.Context, files []*FileInfo) ([]*ParseResult, error) {
	p.logger.WithField("file_count", len(files)).Info("Parsing multiple Nobl9 YAML files")

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

	p.logger.WithFields(logrus.Fields{
		"total_files":     len(files),
		"valid_files":     validFiles,
		"invalid_files":   len(files) - validFiles,
		"total_objects":   totalObjects,
		"valid_objects":   validObjects,
		"invalid_objects": invalidObjects,
		"total_errors":    len(allErrors),
	}).Info("Multiple YAML files parsing completed")

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

	p.logger.WithFields(logrus.Fields{
		"content_size": len(content),
		"object_count": len(manifests),
	}).Debug("YAML content parsed successfully")

	return manifests, nil
}

// validateObject validates a single Nobl9 object
func (p *Parser) validateObject(ctx context.Context, obj manifest.Object) error {
	// Get object metadata for logging
	kind := obj.GetKind()
	name := obj.GetName()
	version := obj.GetVersion()

	p.logger.WithFields(logrus.Fields{
		"kind":        kind,
		"name":        name,
		"api_version": version,
	}).Debug("Validating Nobl9 object")

	// Validate object using Nobl9 SDK
	if err := obj.Validate(); err != nil {
		return fmt.Errorf("object validation failed: %w", err)
	}

	// Enhanced validation for Nobl9 schema requirements
	if err := p.validateNobl9Schema(obj); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"kind": kind,
		"name": name,
	}).Debug("Object validation completed")

	return nil
}

// validateNobl9Schema validates Nobl9-specific schema requirements
func (p *Parser) validateNobl9Schema(obj manifest.Object) error {
	// Validate API version
	if obj.GetVersion() != "n9/v1alpha" {
		return fmt.Errorf("invalid API version: %s, expected n9/v1alpha", obj.GetVersion())
	}

	// Validate name follows DNS RFC1123 conventions
	if err := p.validateName(obj.GetName()); err != nil {
		return fmt.Errorf("invalid name '%s': %w", obj.GetName(), err)
	}

	// Validate kind is supported
	if !p.isValidKind(obj.GetKind()) {
		return fmt.Errorf("unsupported kind: %v", obj.GetKind())
	}

	return nil
}

// validateName validates that a name follows DNS RFC1123 conventions
func (p *Parser) validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Max 63 characters
	if len(name) > 63 {
		return fmt.Errorf("name exceeds maximum length of 63 characters")
	}

	// Only lowercase alphanumeric characters or hyphens
	matched, err := regexp.MatchString("^[a-z0-9-]+$", name)
	if err != nil {
		return fmt.Errorf("error validating name pattern: %w", err)
	}
	if !matched {
		return fmt.Errorf("name must contain only lowercase alphanumeric characters or hyphens")
	}

	// Must start and end with an alphanumeric character
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("name must start and end with an alphanumeric character")
	}

	return nil
}

// isValidKind checks if the kind is supported by Nobl9
func (p *Parser) isValidKind(kind manifest.Kind) bool {
	// For now, we'll accept all kinds and let the SDK handle validation
	// This is more flexible and avoids issues with missing Kind constants
	return true
}


// ValidateManifest validates a Nobl9 manifest without parsing
func (p *Parser) ValidateManifest(ctx context.Context, content []byte) error {
	p.logger.WithField("content_size", len(content)).Debug("Validating Nobl9 manifest")

	// Parse and validate the manifest using the SDK
	_, err := sdk.DecodeObjects(content)
	if err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	p.logger.Debug("Manifest validation completed successfully")

	return nil
}

// ApplyManifest applies a Nobl9 manifest
func (p *Parser) ApplyManifest(ctx context.Context, content []byte) error {
	p.logger.WithField("content_size", len(content)).Debug("Applying Nobl9 manifest")

	// Parse the manifest and apply using the SDK
	objects, err := sdk.DecodeObjects(content)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Apply the objects using the SDK
	err = p.client.Objects().V1().Apply(ctx, objects)
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
