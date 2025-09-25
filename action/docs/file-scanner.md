# File Scanner

This document describes the repository file scanner used by the Nobl9 GitHub Action to detect and process Nobl9 YAML configuration files.

## Overview

The file scanner provides comprehensive repository scanning capabilities:

- **Pattern-Based Scanning** - Support for glob patterns and recursive directory scanning
- **Nobl9 File Detection** - Automatic identification of Nobl9 configuration files
- **YAML Validation** - File type validation and content analysis
- **Error Handling** - Comprehensive error reporting and recovery
- **Performance Optimization** - Efficient file processing and memory management

## Features

### Pattern Matching

The scanner supports various file pattern matching strategies:

#### Simple Patterns
```
*.yaml          # All YAML files in current directory
*.yml           # All YML files in current directory
project.yaml    # Specific file
```

#### Recursive Patterns
```
**/*.yaml       # All YAML files in all subdirectories
**/nobl9-*.yaml # All Nobl9 YAML files in all subdirectories
projects/**/*.yaml # All YAML files in projects directory and subdirectories
```

#### Multiple Patterns
```
*.yaml,*.yml    # Both YAML and YML files
**/project.yaml,**/role-binding.yaml # Specific files in any directory
```

### Nobl9 File Detection

The scanner automatically identifies Nobl9 configuration files by checking for:

#### Nobl9 API Versions
- `apiVersion: n9/v1alpha`

#### Nobl9 Resource Kinds
- `kind: Project`
- `kind: RoleBinding`
- `kind: SLO`
- `kind: SLI`
- `kind: AlertPolicy`
- `kind: DataSource`

#### Nobl9 Annotations
- `nobl9.io/` prefix in annotations

### File Information

For each scanned file, the scanner provides:

```go
type FileInfo struct {
    Path         string      // Absolute file path
    RelativePath string      // Relative path from repository root
    Size         int64       // File size in bytes
    ModTime      fs.FileInfo // File modification time
    IsDir        bool        // Whether file is a directory
    IsYAML       bool        // Whether file is a YAML file
    IsNobl9      bool        // Whether file contains Nobl9 configuration
    Content      []byte      // File content (for YAML files)
    Error        error       // Any error encountered during processing
}
```

## Usage

### Basic Scanning

```go
import "github.com/your-org/nobl9-action/pkg/scanner"

// Create scanner with logger
log := logger.New(logger.LevelInfo, logger.FormatJSON)
scanner := scanner.New(log)

// Scan repository
result, err := scanner.Scan("/path/to/repo", "**/*.yaml")
if err != nil {
    log.Error("Failed to scan repository", logger.Fields{"error": err})
    return
}

// Process results
log.Info("Scan completed", logger.Fields{
    "total_files": result.TotalFiles,
    "yaml_files":  result.YAMLFiles,
    "nobl9_files": result.Nobl9Files,
})
```

### Filtering Results

```go
// Get only Nobl9 files
nobl9Files := scanner.GetNobl9Files(result)
for _, file := range nobl9Files {
    log.Info("Processing Nobl9 file", logger.Fields{
        "file_path": file.Path,
        "file_size": file.Size,
    })
}

// Get only YAML files
yamlFiles := scanner.GetYAMLFiles(result)
for _, file := range yamlFiles {
    log.Info("Processing YAML file", logger.Fields{
        "file_path": file.Path,
        "is_nobl9":  file.IsNobl9,
    })
}

// Get files with errors
errorFiles := scanner.GetFilesWithErrors(result)
for _, file := range errorFiles {
    log.Error("File processing error", logger.Fields{
        "file_path": file.Path,
        "error":     file.Error,
    })
}
```

### Individual File Validation

```go
// Validate a specific file
fileInfo, err := scanner.ValidateFile("/path/to/project.yaml")
if err != nil {
    log.Error("File validation failed", logger.Fields{
        "file_path": "/path/to/project.yaml",
        "error":     err,
    })
    return
}

if fileInfo.IsNobl9 {
    log.Info("Valid Nobl9 file", logger.Fields{
        "file_path": fileInfo.Path,
        "file_size": fileInfo.Size,
    })
}
```

## Configuration

### GitHub Actions Integration

The scanner integrates seamlessly with GitHub Actions:

```yaml
# GitHub Actions workflow
- name: Scan Nobl9 Files
  uses: ./
  with:
    repo-path: "."                    # Repository root
    file-pattern: "**/nobl9-*.yaml"   # Custom pattern
    log-level: "info"
```

### Default Patterns

If no file pattern is specified, the scanner uses:

```
**/*.yaml
```

This pattern scans all YAML files in the repository and subdirectories.

### Custom Patterns

You can specify custom patterns for different use cases:

```yaml
# Scan specific directories
file-pattern: "projects/**/*.yaml"

# Scan multiple file types
file-pattern: "*.yaml,*.yml"

# Scan specific files
file-pattern: "nobl9-project.yaml,nobl9-role-binding.yaml"

# Scan with naming conventions
file-pattern: "**/nobl9-*.yaml"
```

## File Processing

### YAML File Detection

The scanner identifies YAML files by extension:

- `.yaml` (case-insensitive)
- `.yml` (case-insensitive)

### Nobl9 File Detection

The scanner analyzes file content to identify Nobl9 configuration:

1. **File Extension Check** - Must be `.yaml` or `.yml`
2. **Content Analysis** - Searches for Nobl9-specific indicators
3. **Validation** - Ensures file contains valid Nobl9 configuration

### Content Analysis

The scanner performs lightweight content analysis:

```go
// Check for Nobl9 indicators
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
```

## Error Handling

### Common Errors

The scanner handles various error scenarios:

#### Repository Path Errors
```
Error: repository path does not exist
Error: repository path is not a directory
Error: repository path cannot be empty
```

#### File Access Errors
```
Error: failed to stat file
Error: failed to read file
Error: permission denied
```

#### Pattern Errors
```
Error: failed to glob pattern
Error: invalid pattern syntax
```

### Error Recovery

The scanner continues processing even when individual files fail:

```go
// Scan with error handling
result, err := scanner.Scan("/path/to/repo", "**/*.yaml")
if err != nil {
    // Handle scan-level errors
    log.Error("Scan failed", logger.Fields{"error": err})
    return
}

// Check for file-level errors
if len(result.Errors) > 0 {
    log.Warn("Some files had errors", logger.Fields{
        "error_count": len(result.Errors),
    })
    
    for _, err := range result.Errors {
        log.Error("File processing error", logger.Fields{"error": err})
    }
}

// Continue with successful files
nobl9Files := scanner.GetNobl9Files(result)
```

## Performance Considerations

### Scanning Strategy

The scanner uses optimized scanning strategies:

1. **Simple Patterns** - Uses `filepath.Glob` for direct pattern matching
2. **Recursive Patterns** - Uses `filepath.WalkDir` for directory traversal
3. **Content Analysis** - Only reads content for YAML files
4. **Memory Management** - Processes files one at a time

### Optimization Tips

- **Use Specific Patterns** - Avoid scanning unnecessary directories
- **Limit File Types** - Only scan YAML files when possible
- **Exclude Directories** - Skip large directories that don't contain configs
- **Batch Processing** - Process files in batches for large repositories

### Memory Usage

The scanner is designed for memory efficiency:

- **Streaming Processing** - Files are processed one at a time
- **Content Loading** - Only YAML files have content loaded into memory
- **Garbage Collection** - File content is released after processing

## Examples

### Basic Repository Scan

```go
package main

import (
    "github.com/your-org/nobl9-action/pkg/scanner"
    "github.com/your-org/nobl9-action/pkg/logger"
)

func main() {
    // Setup logging
    log := logger.New(logger.LevelInfo, logger.FormatJSON)
    
    // Create scanner
    scanner := scanner.New(log)
    
    // Scan repository
    result, err := scanner.Scan(".", "**/*.yaml")
    if err != nil {
        log.Fatal("Failed to scan repository", logger.Fields{"error": err})
    }
    
    // Process Nobl9 files
    nobl9Files := scanner.GetNobl9Files(result)
    for _, file := range nobl9Files {
        log.Info("Found Nobl9 file", logger.Fields{
            "file_path": file.Path,
            "file_size": file.Size,
        })
        
        // Process file content
        // ... process file.Content ...
    }
}
```

### Advanced Scanning

```go
// Scan with custom patterns
patterns := []string{
    "projects/**/*.yaml",
    "configs/nobl9-*.yaml",
    "templates/*.yml",
}

for _, pattern := range patterns {
    result, err := scanner.Scan(".", pattern)
    if err != nil {
        log.Error("Pattern scan failed", logger.Fields{
            "pattern": pattern,
            "error":   err,
        })
        continue
    }
    
    log.Info("Pattern scan completed", logger.Fields{
        "pattern":       pattern,
        "total_files":   result.TotalFiles,
        "nobl9_files":   result.Nobl9Files,
    })
}
```

### Error Handling

```go
// Comprehensive error handling
result, err := scanner.Scan(".", "**/*.yaml")
if err != nil {
    log.Error("Scan failed", logger.Fields{"error": err})
    return
}

// Check for file-level errors
errorFiles := scanner.GetFilesWithErrors(result)
if len(errorFiles) > 0 {
    log.Warn("Files with errors found", logger.Fields{
        "error_count": len(errorFiles),
    })
    
    for _, file := range errorFiles {
        log.Error("File error", logger.Fields{
            "file_path": file.Path,
            "error":     file.Error,
        })
    }
}

// Continue with valid files
nobl9Files := scanner.GetNobl9Files(result)
log.Info("Processing valid files", logger.Fields{
    "valid_files": len(nobl9Files),
})
```

## Integration with GitHub Action

### Action Configuration

The scanner is automatically used by the GitHub Action:

```yaml
# action.yml
inputs:
  repo-path:
    description: 'Repository path to scan for Nobl9 YAML files'
    required: false
    default: '.'
  
  file-pattern:
    description: 'File pattern to match Nobl9 YAML files (glob pattern)'
    required: false
    default: '**/*.yaml'
```

### Workflow Usage

```yaml
# .github/workflows/nobl9-sync.yml
- name: Sync Nobl9 Projects
  uses: ./
  with:
    client-id: ${{ secrets.NOBL9_CLIENT_ID }}
    client-secret: ${{ secrets.NOBL9_CLIENT_SECRET }}
    repo-path: "."
    file-pattern: "**/nobl9-*.yaml"
    log-level: "info"
```

### Output Integration

The scanner results are integrated with GitHub Actions outputs:

```yaml
outputs:
  processed-files:
    description: 'Number of Nobl9 YAML files processed'
  
  nobl9-files:
    description: 'Number of Nobl9 configuration files found'
```

## Troubleshooting

### Common Issues

#### No Files Found
```
Warning: No Nobl9 files found in repository
```
**Solution:** Check file pattern and ensure files exist in expected locations.

#### Permission Errors
```
Error: permission denied
```
**Solution:** Ensure GitHub Actions has read access to repository files.

#### Pattern Not Matching
```
Warning: Pattern did not match any files
```
**Solution:** Verify pattern syntax and file locations.

### Debug Mode

Enable debug logging for troubleshooting:

```yaml
- name: Debug File Scan
  uses: ./
  with:
    log-level: "debug"
    file-pattern: "**/*.yaml"
```

### Pattern Testing

Test patterns locally before using in GitHub Actions:

```bash
# Test pattern matching
find . -name "*.yaml" -type f

# Test specific pattern
find . -path "**/nobl9-*.yaml" -type f
```

## Best Practices

### Pattern Design

1. **Be Specific** - Use specific patterns to avoid scanning unnecessary files
2. **Use Conventions** - Follow consistent naming conventions for Nobl9 files
3. **Test Patterns** - Test patterns locally before deploying
4. **Document Patterns** - Document expected file locations and naming

### File Organization

1. **Consistent Structure** - Organize Nobl9 files in consistent directory structure
2. **Clear Naming** - Use descriptive names for Nobl9 configuration files
3. **Separation** - Separate Nobl9 configs from other YAML files
4. **Versioning** - Include version information in file names when appropriate

### Performance Optimization

1. **Limit Scope** - Only scan directories that contain Nobl9 files
2. **Use Caching** - Cache scan results when possible
3. **Batch Processing** - Process files in batches for large repositories
4. **Monitor Performance** - Track scan performance and optimize patterns 