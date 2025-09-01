package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/your-org/nobl9-action/pkg/logger"
)

func TestNewScanner(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	scanner := New(log)

	if scanner == nil {
		t.Fatal("expected scanner to be created")
	}

	if scanner.logger != log {
		t.Error("expected scanner to be set")
	}
}

func TestValidateRepoPath(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "scanner-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		repoPath    string
		expectError bool
	}{
		{
			name:        "valid directory",
			repoPath:    tempDir,
			expectError: false,
		},
		{
			name:        "empty path",
			repoPath:    "",
			expectError: true,
		},
		{
			name:        "non-existent path",
			repoPath:    "/non/existent/path",
			expectError: true,
		},
		{
			name:        "file instead of directory",
			repoPath:    tempDir + "/file",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a file for the file test case
			if tt.name == "file instead of directory" {
				file, err := os.Create(tt.repoPath)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				file.Close()
				defer os.Remove(tt.repoPath)
			}

			err := scanner.validateRepoPath(tt.repoPath)

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

func TestIsYAMLFile(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	tests := []struct {
		filePath string
		expected bool
	}{
		{"file.yaml", true},
		{"file.yml", true},
		{"file.YAML", true},
		{"file.YML", true},
		{"file.txt", false},
		{"file.json", false},
		{"file", false},
		{"yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := scanner.isYAMLFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("expected %v for %s, got %v", tt.expected, tt.filePath, result)
			}
		})
	}
}

func TestIsNobl9File(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name: "Project kind",
			content: `apiVersion: nobl9.com/v1
kind: Project
metadata:
  name: my-project`,
			expected: true,
		},
		{
			name: "RoleBinding kind",
			content: `apiVersion: nobl9.com/v1
kind: RoleBinding
metadata:
  name: my-role-binding`,
			expected: true,
		},
		{
			name: "SLO kind",
			content: `apiVersion: nobl9.com/v1
kind: SLO
metadata:
  name: my-slo`,
			expected: true,
		},
		{
			name: "SLI kind",
			content: `apiVersion: nobl9.com/v1
kind: SLI
metadata:
  name: my-sli`,
			expected: true,
		},
		{
			name: "AlertPolicy kind",
			content: `apiVersion: nobl9.com/v1
kind: AlertPolicy
metadata:
  name: my-alert`,
			expected: true,
		},
		{
			name: "DataSource kind",
			content: `apiVersion: nobl9.com/v1
kind: DataSource
metadata:
  name: my-datasource`,
			expected: true,
		},
		{
			name: "Nobl9 annotation",
			content: `apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    nobl9.io/project: my-project`,
			expected: true,
		},
		{
			name: "Regular YAML",
			content: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config`,
			expected: false,
		},
		{
			name:     "Empty content",
			content:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.isNobl9File([]byte(tt.content))
			if result != tt.expected {
				t.Errorf("expected %v for %s, got %v", tt.expected, tt.name, result)
			}
		})
	}
}

func TestExpandPatterns(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	tests := []struct {
		name         string
		repoPath     string
		filePattern  string
		expectedLen  int
		expectedPath string
	}{
		{
			name:         "single pattern",
			repoPath:     "/repo",
			filePattern:  "*.yaml",
			expectedLen:  1,
			expectedPath: "/repo/*.yaml",
		},
		{
			name:         "multiple patterns",
			repoPath:     "/repo",
			filePattern:  "*.yaml,*.yml",
			expectedLen:  2,
			expectedPath: "/repo/*.yaml",
		},
		{
			name:         "empty pattern",
			repoPath:     "/repo",
			filePattern:  "",
			expectedLen:  1,
			expectedPath: "/repo/**/*.yaml",
		},
		{
			name:         "whitespace patterns",
			repoPath:     "/repo",
			filePattern:  "  *.yaml  ,  *.yml  ",
			expectedLen:  2,
			expectedPath: "/repo/*.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := scanner.expandPatterns(tt.repoPath, tt.filePattern)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(patterns) != tt.expectedLen {
				t.Errorf("expected %d patterns, got %d", tt.expectedLen, len(patterns))
			}

			if len(patterns) > 0 && patterns[0] != tt.expectedPath {
				t.Errorf("expected first pattern to be %s, got %s", tt.expectedPath, patterns[0])
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			path:     "/path/to/file.yaml",
			pattern:  "file.yaml",
			expected: true,
		},
		{
			name:     "wildcard extension",
			path:     "/path/to/file.yaml",
			pattern:  "*.yaml",
			expected: true,
		},
		{
			name:     "wildcard prefix",
			path:     "/path/to/file.yaml",
			pattern:  "file*",
			expected: true,
		},
		{
			name:     "wildcard all",
			path:     "/path/to/file.yaml",
			pattern:  "*",
			expected: true,
		},
		{
			name:     "no match",
			path:     "/path/to/file.yaml",
			pattern:  "other.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.matchesPattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("expected %v for path %s and pattern %s, got %v", tt.expected, tt.path, tt.pattern, result)
			}
		})
	}
}

func TestScanWithTestFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "scanner-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"project.yaml": `apiVersion: nobl9.com/v1
kind: Project
metadata:
  name: test-project`,
		"role-binding.yaml": `apiVersion: nobl9.com/v1
kind: RoleBinding
metadata:
  name: test-role-binding`,
		"config.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config`,
		"readme.md": `# Test Readme`,
	}

	for fileName, content := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", fileName, err)
		}
	}

	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	// Test scanning with different patterns
	tests := []struct {
		name          string
		pattern       string
		expectedTotal int
		expectedYAML  int
		expectedNobl9 int
	}{
		{
			name:          "all YAML files",
			pattern:       "*.yaml",
			expectedTotal: 3,
			expectedYAML:  3,
			expectedNobl9: 2,
		},
		{
			name:          "all files",
			pattern:       "*",
			expectedTotal: 4,
			expectedYAML:  3,
			expectedNobl9: 2,
		},
		{
			name:          "specific Nobl9 file",
			pattern:       "project.yaml",
			expectedTotal: 1,
			expectedYAML:  1,
			expectedNobl9: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scanner.Scan(tempDir, tt.pattern)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.TotalFiles != tt.expectedTotal {
				t.Errorf("expected %d total files, got %d", tt.expectedTotal, result.TotalFiles)
			}

			if result.YAMLFiles != tt.expectedYAML {
				t.Errorf("expected %d YAML files, got %d", tt.expectedYAML, result.YAMLFiles)
			}

			if result.Nobl9Files != tt.expectedNobl9 {
				t.Errorf("expected %d Nobl9 files, got %d", tt.expectedNobl9, result.Nobl9Files)
			}
		})
	}
}

func TestGetNobl9Files(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	// Create test scan result
	result := &ScanResult{
		Files: []*FileInfo{
			{
				Path:    "project.yaml",
				IsYAML:  true,
				IsNobl9: true,
			},
			{
				Path:    "config.yaml",
				IsYAML:  true,
				IsNobl9: false,
			},
			{
				Path:    "role-binding.yaml",
				IsYAML:  true,
				IsNobl9: true,
			},
			{
				Path:    "readme.md",
				IsYAML:  false,
				IsNobl9: false,
			},
		},
	}

	nobl9Files := scanner.GetNobl9Files(result)

	if len(nobl9Files) != 2 {
		t.Errorf("expected 2 Nobl9 files, got %d", len(nobl9Files))
	}

	// Check that only Nobl9 files are returned
	for _, file := range nobl9Files {
		if !file.IsNobl9 {
			t.Errorf("expected file %s to be a Nobl9 file", file.Path)
		}
	}
}

func TestGetYAMLFiles(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	// Create test scan result
	result := &ScanResult{
		Files: []*FileInfo{
			{
				Path:    "project.yaml",
				IsYAML:  true,
				IsNobl9: true,
			},
			{
				Path:    "config.yaml",
				IsYAML:  true,
				IsNobl9: false,
			},
			{
				Path:    "readme.md",
				IsYAML:  false,
				IsNobl9: false,
			},
		},
	}

	yamlFiles := scanner.GetYAMLFiles(result)

	if len(yamlFiles) != 2 {
		t.Errorf("expected 2 YAML files, got %d", len(yamlFiles))
	}

	// Check that only YAML files are returned
	for _, file := range yamlFiles {
		if !file.IsYAML {
			t.Errorf("expected file %s to be a YAML file", file.Path)
		}
	}
}

func TestGetFilesWithErrors(t *testing.T) {
	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	// Create test scan result with errors
	result := &ScanResult{
		Files: []*FileInfo{
			{
				Path:   "valid.yaml",
				IsYAML: true,
				Error:  nil,
			},
			{
				Path:   "error.yaml",
				IsYAML: true,
				Error:  fmt.Errorf("test error"),
			},
			{
				Path:   "another-error.yaml",
				IsYAML: true,
				Error:  fmt.Errorf("another error"),
			},
		},
	}

	errorFiles := scanner.GetFilesWithErrors(result)

	if len(errorFiles) != 2 {
		t.Errorf("expected 2 files with errors, got %d", len(errorFiles))
	}

	// Check that only files with errors are returned
	for _, file := range errorFiles {
		if file.Error == nil {
			t.Errorf("expected file %s to have an error", file.Path)
		}
	}
}

func TestValidateFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "scanner-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	scanner := New(logger.New(logger.LevelInfo, logger.FormatJSON))

	// Create test files
	nobl9File := filepath.Join(tempDir, "project.yaml")
	nobl9Content := `apiVersion: nobl9.com/v1
kind: Project
metadata:
  name: test-project`
	err = os.WriteFile(nobl9File, []byte(nobl9Content), 0644)
	if err != nil {
		t.Fatalf("failed to create Nobl9 test file: %v", err)
	}

	regularYAMLFile := filepath.Join(tempDir, "config.yaml")
	regularContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config`
	err = os.WriteFile(regularYAMLFile, []byte(regularContent), 0644)
	if err != nil {
		t.Fatalf("failed to create regular YAML test file: %v", err)
	}

	nonYAMLFile := filepath.Join(tempDir, "readme.md")
	err = os.WriteFile(nonYAMLFile, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("failed to create non-YAML test file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		expectNobl9 bool
	}{
		{
			name:        "valid Nobl9 file",
			filePath:    nobl9File,
			expectError: false,
			expectNobl9: true,
		},
		{
			name:        "regular YAML file",
			filePath:    regularYAMLFile,
			expectError: true,
			expectNobl9: false,
		},
		{
			name:        "non-YAML file",
			filePath:    nonYAMLFile,
			expectError: true,
			expectNobl9: false,
		},
		{
			name:        "non-existent file",
			filePath:    filepath.Join(tempDir, "nonexistent.yaml"),
			expectError: true,
			expectNobl9: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo, err := scanner.ValidateFile(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if fileInfo != nil {
				if fileInfo.IsNobl9 != tt.expectNobl9 {
					t.Errorf("expected IsNobl9 to be %v, got %v", tt.expectNobl9, fileInfo.IsNobl9)
				}
			}
		})
	}
}
