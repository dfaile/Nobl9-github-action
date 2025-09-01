package parser

import (
	"context"
	"fmt"
	"testing"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
	"github.com/your-org/nobl9-action/pkg/scanner"
)

func TestNewParser(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}

	parser := New(client, log)

	if parser == nil {
		t.Fatal("expected parser to be created")
	}

	if parser.client != client {
		t.Error("expected client to be set")
	}

	if parser.logger != log {
		t.Error("expected logger to be set")
	}
}

func TestParseYAMLContent(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	parser := New(client, log)

	tests := []struct {
		name        string
		content     []byte
		expectError bool
	}{
		{
			name: "valid Nobl9 YAML",
			content: []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
spec:
  displayName: Test Project
  description: A test project`),
			expectError: false,
		},
		{
			name: "invalid YAML",
			content: []byte(`invalid: yaml: content: [{
  name: test-project
`),
			expectError: true,
		},
		{
			name:        "empty content",
			content:     []byte{},
			expectError: true,
		},
		{
			name: "multiple objects",
			content: []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
spec:
  displayName: Test Project
---
apiVersion: n9/v1alpha
kind: RoleBinding
metadata:
  name: test-role-binding
  project: test-project
spec:
  users:
    - id: user@example.com
  roles:
    - project-owner`),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifests, err := parser.parseYAMLContent(tt.content)

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

			if len(manifests) == 0 {
				t.Error("expected manifests but got none")
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	parser := New(client, log)

	ctx := context.Background()

	tests := []struct {
		name        string
		fileInfo    *scanner.FileInfo
		expectError bool
		expectValid bool
	}{
		{
			name: "valid Nobl9 file",
			fileInfo: &scanner.FileInfo{
				Path:    "test.yaml",
				Size:    100,
				IsYAML:  true,
				IsNobl9: true,
				Content: []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
spec:
  displayName: Test Project`),
			},
			expectError: false,
			expectValid: true,
		},
		{
			name: "file with error",
			fileInfo: &scanner.FileInfo{
				Path:  "error.yaml",
				Error: fmt.Errorf("file read error"),
			},
			expectError: false, // ParseFile handles errors gracefully
			expectValid: false,
		},
		{
			name: "non-Nobl9 file",
			fileInfo: &scanner.FileInfo{
				Path:    "config.yaml",
				Size:    50,
				IsYAML:  true,
				IsNobl9: false,
				Content: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config`),
			},
			expectError: false, // ParseFile handles this gracefully
			expectValid: false,
		},
		{
			name: "invalid YAML content",
			fileInfo: &scanner.FileInfo{
				Path:    "invalid.yaml",
				Size:    30,
				IsYAML:  true,
				IsNobl9: true,
				Content: []byte(`invalid: yaml: content: [{
  name: test-project
`),
			},
			expectError: false, // ParseFile handles this gracefully
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseFile(ctx, tt.fileInfo)

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

			if result == nil {
				t.Error("expected result but got nil")
				return
			}

			if result.IsValid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, result.IsValid)
			}

			if result.FileInfo != tt.fileInfo {
				t.Error("expected file info to be set")
			}
		})
	}
}

func TestParseFiles(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	parser := New(client, log)

	ctx := context.Background()

	validFile := &scanner.FileInfo{
		Path:    "valid.yaml",
		Size:    100,
		IsYAML:  true,
		IsNobl9: true,
		Content: []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
spec:
  displayName: Test Project`),
	}

	invalidFile := &scanner.FileInfo{
		Path:    "invalid.yaml",
		Size:    30,
		IsYAML:  true,
		IsNobl9: true,
		Content: []byte(`invalid: yaml: content: [{
  name: test-project
`),
	}

	tests := []struct {
		name        string
		files       []*scanner.FileInfo
		expectError bool
		expectValid bool
	}{
		{
			name:        "no files",
			files:       []*scanner.FileInfo{},
			expectError: false,
			expectValid: true,
		},
		{
			name:        "single valid file",
			files:       []*scanner.FileInfo{validFile},
			expectError: false,
			expectValid: true,
		},
		{
			name:        "single invalid file",
			files:       []*scanner.FileInfo{invalidFile},
			expectError: true, // ParseFiles returns error when any file is invalid
			expectValid: false,
		},
		{
			name:        "mixed files",
			files:       []*scanner.FileInfo{validFile, invalidFile},
			expectError: true, // ParseFiles returns error when any file is invalid
			expectValid: true, // But there is at least one valid file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := parser.ParseFiles(ctx, tt.files)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if len(results) != len(tt.files) {
				t.Errorf("expected %d results, got %d", len(tt.files), len(results))
			}

			// Check if any results are valid
			anyValid := false
			if len(results) == 0 {
				// If no files, consider it valid (no invalid files)
				anyValid = true
			} else {
				for _, result := range results {
					if result.IsValid {
						anyValid = true
						break
					}
				}
			}

			if anyValid != tt.expectValid {
				t.Errorf("expected any valid=%v, got any valid=%v", tt.expectValid, anyValid)
			}
		})
	}
}

func TestGetValidObjects(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	parser := New(client, log)

	// Create test results
	results := []*ParseResult{
		{
			ValidObjects: []manifest.Object{}, // Would be actual objects
		},
		{
			ValidObjects: []manifest.Object{}, // Would be actual objects
		},
	}

	validObjects := parser.GetValidObjects(results)

	if len(validObjects) != 0 {
		t.Errorf("expected 0 valid objects, got %d", len(validObjects))
	}
}

func TestGetInvalidObjects(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	parser := New(client, log)

	// Create test results
	results := []*ParseResult{
		{
			InvalidObjects: []InvalidObject{}, // Would be actual invalid objects
		},
		{
			InvalidObjects: []InvalidObject{}, // Would be actual invalid objects
		},
	}

	invalidObjects := parser.GetInvalidObjects(results)

	if len(invalidObjects) != 0 {
		t.Errorf("expected 0 invalid objects, got %d", len(invalidObjects))
	}
}

func TestGetErrors(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	parser := New(client, log)

	// Create test results with errors
	results := []*ParseResult{
		{
			Errors: []error{fmt.Errorf("test error 1")},
		},
		{
			Errors: []error{fmt.Errorf("test error 2")},
		},
	}

	errors := parser.GetErrors(results)

	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}
}

func TestIsValid(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	parser := New(client, log)

	tests := []struct {
		name        string
		results     []*ParseResult
		expectValid bool
	}{
		{
			name:        "no results",
			results:     []*ParseResult{},
			expectValid: true,
		},
		{
			name: "all valid results",
			results: []*ParseResult{
				{IsValid: true},
				{IsValid: true},
			},
			expectValid: true,
		},
		{
			name: "mixed results",
			results: []*ParseResult{
				{IsValid: true},
				{IsValid: false},
			},
			expectValid: false,
		},
		{
			name: "all invalid results",
			results: []*ParseResult{
				{IsValid: false},
				{IsValid: false},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := parser.IsValid(tt.results)

			if isValid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, isValid)
			}
		})
	}
}
