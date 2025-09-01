package processor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
	"github.com/your-org/nobl9-action/pkg/scanner"
)

func TestNewProcessor(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}

	processor := New(client, log)

	if processor == nil {
		t.Fatal("expected processor to be created")
	}

	if processor.client != client {
		t.Error("expected client to be set")
	}

	if processor.logger != log {
		t.Error("expected logger to be set")
	}

	if processor.parser == nil {
		t.Error("expected parser to be created")
	}

	if processor.resolver == nil {
		t.Error("expected resolver to be created")
	}
}

func TestProcessFiles(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}

func TestProcessFile(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}

func TestProcessFileWithError(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	ctx := context.Background()

	// Create file with error
	fileInfo := &scanner.FileInfo{
		Path:  "error.yaml",
		Error: fmt.Errorf("file read error"),
	}

	result, err := processor.ProcessFile(ctx, fileInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("expected result but got nil")
		return
	}

	if result.IsSuccess {
		t.Error("expected processing to fail")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors to be present")
	}
}

func TestProcessWithDryRun(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}

func TestProcessFileWithDryRun(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}

func TestGetProcessingStats(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	result := &ProcessingResult{
		FilesProcessed:      2,
		FilesSkipped:        0,
		FilesWithErrors:     1,
		ProjectsCreated:     1,
		ProjectsUpdated:     0,
		RoleBindingsCreated: 1,
		RoleBindingsUpdated: 0,
		UsersResolved:       2,
		UsersUnresolved:     1,
		Errors:              []error{fmt.Errorf("test error")},
		Warnings:            []string{"test warning"},
		Duration:            5 * time.Second,
		IsSuccess:           false,
	}

	stats := processor.GetProcessingStats(result)

	if stats["files_processed"] != 2 {
		t.Errorf("expected files_processed 2, got %v", stats["files_processed"])
	}

	if stats["files_with_errors"] != 1 {
		t.Errorf("expected files_with_errors 1, got %v", stats["files_with_errors"])
	}

	if stats["projects_created"] != 1 {
		t.Errorf("expected projects_created 1, got %v", stats["projects_created"])
	}

	if stats["role_bindings_created"] != 1 {
		t.Errorf("expected role_bindings_created 1, got %v", stats["role_bindings_created"])
	}

	if stats["users_resolved"] != 2 {
		t.Errorf("expected users_resolved 2, got %v", stats["users_resolved"])
	}

	if stats["users_unresolved"] != 1 {
		t.Errorf("expected users_unresolved 1, got %v", stats["users_unresolved"])
	}

	if stats["errors"] != 1 {
		t.Errorf("expected errors 1, got %v", stats["errors"])
	}

	if stats["warnings"] != 1 {
		t.Errorf("expected warnings 1, got %v", stats["warnings"])
	}

	if stats["duration"] != "5s" {
		t.Errorf("expected duration 5s, got %v", stats["duration"])
	}

	if stats["is_success"] != false {
		t.Errorf("expected is_success false, got %v", stats["is_success"])
	}
}

func TestGetProcessingErrors(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	result := &ProcessingResult{
		Errors: []error{
			fmt.Errorf("error 1"),
			fmt.Errorf("error 2"),
		},
	}

	errors := processor.GetProcessingErrors(result)

	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}

	if errors[0] != "error 1" {
		t.Errorf("expected error 1, got %s", errors[0])
	}

	if errors[1] != "error 2" {
		t.Errorf("expected error 2, got %s", errors[1])
	}
}

func TestGetUnresolvedEmails(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	result := &ProcessingResult{}

	unresolved := processor.GetUnresolvedEmails(result)

	// Currently returns empty slice as per implementation
	if len(unresolved) != 0 {
		t.Errorf("expected 0 unresolved emails, got %d", len(unresolved))
	}
}

func TestProcessEmptyFiles(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	ctx := context.Background()

	// Test with empty files list
	result, err := processor.ProcessFiles(ctx, []*scanner.FileInfo{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("expected result but got nil")
		return
	}

	if result.FilesProcessed != 0 {
		t.Errorf("expected 0 files processed, got %d", result.FilesProcessed)
	}

	if !result.IsSuccess {
		t.Error("expected success for empty files")
	}
}

func TestProcessWithDryRunEmptyFiles(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	ctx := context.Background()

	// Test with empty files list
	result, err := processor.ProcessWithDryRun(ctx, []*scanner.FileInfo{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("expected result but got nil")
		return
	}

	if result.FilesProcessed != 0 {
		t.Errorf("expected 0 files processed, got %d", result.FilesProcessed)
	}

	if !result.IsSuccess {
		t.Error("expected success for empty files")
	}
}

func TestProcessFileWithInvalidYAML(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	ctx := context.Background()

	fileInfo := &scanner.FileInfo{
		Path:    "invalid.yaml",
		Size:    50,
		IsYAML:  true,
		IsNobl9: true,
		Content: []byte(`invalid: yaml: content: [{
  name: test-project
`),
	}

	result, err := processor.ProcessFile(ctx, fileInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("expected result but got nil")
		return
	}

	if result.IsSuccess {
		t.Error("expected processing to fail for invalid YAML")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors to be present for invalid YAML")
	}
}

func TestProcessFileWithNonNobl9YAML(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	processor := New(client, log)

	ctx := context.Background()

	fileInfo := &scanner.FileInfo{
		Path:    "config.yaml",
		Size:    80,
		IsYAML:  true,
		IsNobl9: false,
		Content: []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value`),
	}

	result, err := processor.ProcessFile(ctx, fileInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("expected result but got nil")
		return
	}

	if result.IsSuccess {
		t.Error("expected processing to fail for non-Nobl9 YAML")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors to be present for non-Nobl9 YAML")
	}
}
