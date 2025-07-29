package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/your-org/nobl9-action/pkg/logger"
)

// Scanner handles repository file scanning and processing
type Scanner struct {
	logger *logger.Logger
}

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

// ScanResult represents the result of a file scan
type ScanResult struct {
	Files        []*FileInfo
	TotalFiles   int
	YAMLFiles    int
	Nobl9Files   int
	Errors       []error
	ScanDuration string
}

// New creates a new scanner instance
func New(log *logger.Logger) *Scanner {
	return &Scanner{
		logger: log,
	}
}

// Scan scans the repository for files matching the pattern
func (s *Scanner) Scan(repoPath, filePattern string) (*ScanResult, error) {
	s.logger.Info("Starting repository file scan", logger.Fields{
		"repo_path":    repoPath,
		"file_pattern": filePattern,
	})

	result := &ScanResult{
		Files:  make([]*FileInfo, 0),
		Errors: make([]error, 0),
	}

	// Validate repository path
	if err := s.validateRepoPath(repoPath); err != nil {
		return nil, fmt.Errorf("invalid repository path: %w", err)
	}

	// Expand file pattern to absolute paths
	patterns, err := s.expandPatterns(repoPath, filePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to expand file patterns: %w", err)
	}

	// Scan each pattern
	for _, pattern := range patterns {
		if err := s.scanPattern(pattern, result); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to scan pattern %s: %w", pattern, err))
		}
	}

	// Update statistics
	result.TotalFiles = len(result.Files)
	result.YAMLFiles = s.countYAMLFiles(result.Files)
	result.Nobl9Files = s.countNobl9Files(result.Files)

	s.logger.Info("Repository file scan completed", logger.Fields{
		"total_files": result.TotalFiles,
		"yaml_files":  result.YAMLFiles,
		"nobl9_files": result.Nobl9Files,
		"errors":      len(result.Errors),
	})

	return result, nil
}

// validateRepoPath validates the repository path
func (s *Scanner) validateRepoPath(repoPath string) error {
	if repoPath == "" {
		return fmt.Errorf("repository path cannot be empty")
	}

	// Check if path exists
	info, err := os.Stat(repoPath)
	if err != nil {
		return fmt.Errorf("repository path does not exist: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("repository path is not a directory")
	}

	return nil
}

// expandPatterns expands file patterns to absolute paths
func (s *Scanner) expandPatterns(repoPath, filePattern string) ([]string, error) {
	patterns := make([]string, 0)

	// Handle multiple patterns separated by commas
	patternList := strings.Split(filePattern, ",")
	for _, pattern := range patternList {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// Convert relative pattern to absolute path
		absolutePattern := filepath.Join(repoPath, pattern)
		patterns = append(patterns, absolutePattern)
	}

	if len(patterns) == 0 {
		// Default pattern if none provided
		defaultPattern := filepath.Join(repoPath, "**/*.yaml")
		patterns = append(patterns, defaultPattern)
	}

	return patterns, nil
}

// scanPattern scans files matching a specific pattern
func (s *Scanner) scanPattern(pattern string, result *ScanResult) error {
	s.logger.Debug("Scanning pattern", logger.Fields{
		"pattern": pattern,
	})

	// Use filepath.Glob for simple patterns
	if !strings.Contains(pattern, "**") {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("failed to glob pattern: %w", err)
		}

		for _, match := range matches {
			if err := s.processFile(match, result); err != nil {
				result.Errors = append(result.Errors, err)
			}
		}
		return nil
	}

	// Handle recursive patterns with **
	return s.scanRecursivePattern(pattern, result)
}

// scanRecursivePattern scans files matching recursive patterns
func (s *Scanner) scanRecursivePattern(pattern string, result *ScanResult) error {
	// Extract base directory and file pattern
	baseDir, filePattern := s.extractRecursivePattern(pattern)

	s.logger.Debug("Scanning recursive pattern", logger.Fields{
		"base_dir":     baseDir,
		"file_pattern": filePattern,
	})

	return filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error accessing path %s: %w", path, err))
			return nil // Continue walking
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file matches pattern
		if s.matchesPattern(path, filePattern) {
			if err := s.processFile(path, result); err != nil {
				result.Errors = append(result.Errors, err)
			}
		}

		return nil
	})
}

// extractRecursivePattern extracts base directory and file pattern from recursive pattern
func (s *Scanner) extractRecursivePattern(pattern string) (string, string) {
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		// Fallback to simple pattern
		return filepath.Dir(pattern), filepath.Base(pattern)
	}

	baseDir := strings.TrimSuffix(parts[0], "/")
	filePattern := strings.TrimPrefix(parts[1], "/")

	if baseDir == "" {
		baseDir = "."
	}

	return baseDir, filePattern
}

// matchesPattern checks if a file path matches a pattern
func (s *Scanner) matchesPattern(path, pattern string) bool {
	// Simple pattern matching - can be enhanced with regex
	fileName := filepath.Base(path)

	// Handle wildcards
	if strings.Contains(pattern, "*") {
		return s.matchesWildcard(fileName, pattern)
	}

	// Exact match
	return fileName == pattern
}

// matchesWildcard checks if a filename matches a wildcard pattern
func (s *Scanner) matchesWildcard(fileName, pattern string) bool {
	// Simple wildcard matching - can be enhanced
	if pattern == "*" {
		return true
	}

	if strings.HasPrefix(pattern, "*.") {
		ext := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(fileName, ext)
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(fileName, prefix)
	}

	return fileName == pattern
}

// processFile processes a single file
func (s *Scanner) processFile(filePath string, result *ScanResult) error {
	s.logger.Debug("Processing file", logger.Fields{
		"file_path": filePath,
	})

	// Get file information
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Skip directories
	if info.IsDir() {
		return nil
	}

	// Create file info
	fileInfo := &FileInfo{
		Path:         filePath,
		RelativePath: s.getRelativePath(filePath),
		Size:         info.Size(),
		ModTime:      info,
		IsDir:        false,
		IsYAML:       s.isYAMLFile(filePath),
		IsNobl9:      false, // Will be determined after content analysis
	}

	// Read file content for YAML files
	if fileInfo.IsYAML {
		content, err := os.ReadFile(filePath)
		if err != nil {
			fileInfo.Error = fmt.Errorf("failed to read file: %w", err)
		} else {
			fileInfo.Content = content
			fileInfo.IsNobl9 = s.isNobl9File(content)
		}
	}

	result.Files = append(result.Files, fileInfo)

	s.logger.Debug("File processed", logger.Fields{
		"file_path": filePath,
		"is_yaml":   fileInfo.IsYAML,
		"is_nobl9":  fileInfo.IsNobl9,
		"size":      fileInfo.Size,
	})

	return nil
}

// getRelativePath gets the relative path from the repository root
func (s *Scanner) getRelativePath(filePath string) string {
	// This is a simplified implementation
	// In practice, you might want to store the repo path and calculate relative path
	return filepath.Base(filePath)
}

// isYAMLFile checks if a file is a YAML file
func (s *Scanner) isYAMLFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".yaml" || ext == ".yml"
}

// isNobl9File checks if file content contains Nobl9 configuration
func (s *Scanner) isNobl9File(content []byte) bool {
	contentStr := string(content)

	// Check for Nobl9-specific indicators
	nobl9Indicators := []string{
		"apiVersion: nobl9.com/",
		"kind: Project",
		"kind: RoleBinding",
		"kind: SLO",
		"kind: SLI",
		"kind: AlertPolicy",
		"kind: DataSource",
		"nobl9.io/",
	}

	for _, indicator := range nobl9Indicators {
		if strings.Contains(contentStr, indicator) {
			return true
		}
	}

	return false
}

// countYAMLFiles counts the number of YAML files
func (s *Scanner) countYAMLFiles(files []*FileInfo) int {
	count := 0
	for _, file := range files {
		if file.IsYAML {
			count++
		}
	}
	return count
}

// countNobl9Files counts the number of Nobl9 files
func (s *Scanner) countNobl9Files(files []*FileInfo) int {
	count := 0
	for _, file := range files {
		if file.IsNobl9 {
			count++
		}
	}
	return count
}

// GetNobl9Files returns only Nobl9 files from scan result
func (s *Scanner) GetNobl9Files(result *ScanResult) []*FileInfo {
	nobl9Files := make([]*FileInfo, 0)
	for _, file := range result.Files {
		if file.IsNobl9 {
			nobl9Files = append(nobl9Files, file)
		}
	}
	return nobl9Files
}

// GetYAMLFiles returns only YAML files from scan result
func (s *Scanner) GetYAMLFiles(result *ScanResult) []*FileInfo {
	yamlFiles := make([]*FileInfo, 0)
	for _, file := range result.Files {
		if file.IsYAML {
			yamlFiles = append(yamlFiles, file)
		}
	}
	return yamlFiles
}

// GetFilesWithErrors returns files that have errors
func (s *Scanner) GetFilesWithErrors(result *ScanResult) []*FileInfo {
	errorFiles := make([]*FileInfo, 0)
	for _, file := range result.Files {
		if file.Error != nil {
			errorFiles = append(errorFiles, file)
		}
	}
	return errorFiles
}

// ValidateFile validates a single file
func (s *Scanner) ValidateFile(filePath string) (*FileInfo, error) {
	s.logger.Debug("Validating file", logger.Fields{
		"file_path": filePath,
	})

	fileInfo := &FileInfo{
		Path:   filePath,
		IsYAML: s.isYAMLFile(filePath),
	}

	if !fileInfo.IsYAML {
		return fileInfo, fmt.Errorf("file is not a YAML file")
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		fileInfo.Error = err
		return fileInfo, fmt.Errorf("failed to read file: %w", err)
	}

	fileInfo.Content = content
	fileInfo.IsNobl9 = s.isNobl9File(content)

	if !fileInfo.IsNobl9 {
		return fileInfo, fmt.Errorf("file does not contain Nobl9 configuration")
	}

	s.logger.Debug("File validation completed", logger.Fields{
		"file_path": filePath,
		"is_nobl9":  fileInfo.IsNobl9,
		"size":      len(content),
	})

	return fileInfo, nil
}
