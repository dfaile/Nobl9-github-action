package resolver

import (
	"fmt"
	"testing"
	"time"

	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
)

func TestNewResolver(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)

	// Create a mock client for testing
	client := &nobl9.Client{}

	resolver := New(client, log)

	if resolver == nil {
		t.Fatal("expected resolver to be created")
	}

	if resolver.client != client {
		t.Error("expected client to be set")
	}

	if resolver.logger != log {
		t.Error("expected logger to be set")
	}

	if resolver.cache == nil {
		t.Error("expected cache to be created")
	}
}

func TestNewUserCache(t *testing.T) {
	ttl := 30 * time.Minute
	cache := NewUserCache(ttl)

	if cache == nil {
		t.Fatal("expected cache to be created")
	}

	if cache.ttl != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, cache.ttl)
	}

	if cache.users == nil {
		t.Error("expected users map to be initialized")
	}
}

func TestUserCacheOperations(t *testing.T) {
	cache := NewUserCache(30 * time.Minute)

	// Test Set and Get
	email := "test@example.com"
	userInfo := &UserInfo{
		Email:    email,
		UserID:   "user-123",
		Username: "testuser",
		Found:    true,
	}

	cache.Set(email, userInfo)

	retrieved := cache.Get(email)
	if retrieved == nil {
		t.Fatal("expected user to be retrieved from cache")
	}

	if retrieved.Email != email {
		t.Errorf("expected email %s, got %s", email, retrieved.Email)
	}

	if retrieved.UserID != "user-123" {
		t.Errorf("expected UserID user-123, got %s", retrieved.UserID)
	}

	// Test Get for non-existent user
	nonExistent := cache.Get("nonexistent@example.com")
	if nonExistent != nil {
		t.Error("expected nil for non-existent user")
	}

	// Test GetStats
	stats := cache.GetStats()
	if stats["size"] != 1 {
		t.Errorf("expected cache size 1, got %v", stats["size"])
	}

	if stats["ttl"] != "30m0s" {
		t.Errorf("expected TTL 30m0s, got %v", stats["ttl"])
	}

	// Test Clear
	cache.Clear()
	cleared := cache.Get(email)
	if cleared != nil {
		t.Error("expected user to be removed after clear")
	}

	statsAfterClear := cache.GetStats()
	if statsAfterClear["size"] != 0 {
		t.Errorf("expected cache size 0 after clear, got %v", statsAfterClear["size"])
	}
}

func TestIsValidEmail(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := New(client, log)

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "valid email with subdomain",
			email:    "user@sub.example.com",
			expected: true,
		},
		{
			name:     "valid email with plus",
			email:    "test+tag@example.com",
			expected: true,
		},
		{
			name:     "valid email with dots",
			email:    "test.user@example.com",
			expected: true,
		},
		{
			name:     "missing @",
			email:    "testexample.com",
			expected: false,
		},
		{
			name:     "missing domain",
			email:    "test@",
			expected: false,
		},
		{
			name:     "missing local part",
			email:    "@example.com",
			expected: false,
		},
		{
			name:     "multiple @",
			email:    "test@user@example.com",
			expected: false,
		},
		{
			name:     "no domain dot",
			email:    "test@example",
			expected: false,
		},
		{
			name:     "empty string",
			email:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			email:    "   ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("expected %v for email %s, got %v", tt.expected, tt.email, result)
			}
		})
	}
}

func TestValidateEmailFormat(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := New(client, log)

	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "invalid email",
			email:       "invalid-email",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resolver.ValidateEmailFormat(tt.email)

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

func TestValidateEmails(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := New(client, log)

	emails := []string{
		"valid@example.com",
		"invalid-email",
		"another@example.com",
		"also-invalid",
	}

	errors := resolver.ValidateEmails(emails)

	if len(errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d", len(errors))
	}
}

func TestExtractEmailsFromYAML(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := New(client, log)

	tests := []struct {
		name        string
		yamlContent []byte
		expected    []string
	}{
		{
			name: "YAML with emails",
			yamlContent: []byte(`apiVersion: n9/v1alpha
kind: RoleBinding
metadata:
  name: test-role-binding
  project: test-project
spec:
  users:
    - id: user1@example.com
    - id: user2@example.com
  roles:
    - project-owner`),
			expected: []string{"user1@example.com", "user2@example.com"},
		},
		{
			name: "YAML without emails",
			yamlContent: []byte(`apiVersion: n9/v1alpha
kind: Project
metadata:
  name: test-project
spec:
  displayName: Test Project`),
			expected: []string{},
		},
		{
			name: "YAML with invalid emails",
			yamlContent: []byte(`apiVersion: n9/v1alpha
kind: RoleBinding
spec:
  users:
    - id: invalid-email
    - id: another@example.com`),
			expected: []string{"another@example.com"},
		},
		{
			name:        "empty content",
			yamlContent: []byte{},
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emails, err := resolver.extractEmailsFromYAML(tt.yamlContent)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(emails) != len(tt.expected) {
				t.Errorf("expected %d emails, got %d", len(tt.expected), len(emails))
				return
			}

			// Check that all expected emails are present
			for _, expectedEmail := range tt.expected {
				found := false
				for _, email := range emails {
					if email == expectedEmail {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected email %s not found in results", expectedEmail)
				}
			}
		})
	}
}

func TestResolveEmails(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}

func TestResolveEmailsFromYAML(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}

func TestGetResolvedUserIDs(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := New(client, log)

	// Create test batch result
	batchResult := &BatchResolutionResult{
		Results: []*ResolutionResult{
			{
				Email:    "user1@example.com",
				UserID:   "user-123",
				Resolved: true,
			},
			{
				Email:    "user2@example.com",
				UserID:   "user-456",
				Resolved: true,
			},
			{
				Email:    "user3@example.com",
				Resolved: false,
				Error:    fmt.Errorf("user not found"),
			},
		},
	}

	emailToUserID := resolver.GetResolvedUserIDs(batchResult)

	if len(emailToUserID) != 2 {
		t.Errorf("expected 2 resolved users, got %d", len(emailToUserID))
	}

	if emailToUserID["user1@example.com"] != "user-123" {
		t.Errorf("expected user-123 for user1@example.com, got %s", emailToUserID["user1@example.com"])
	}

	if emailToUserID["user2@example.com"] != "user-456" {
		t.Errorf("expected user-456 for user2@example.com, got %s", emailToUserID["user2@example.com"])
	}
}

func TestGetUnresolvedEmails(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := New(client, log)

	// Create test batch result
	batchResult := &BatchResolutionResult{
		Results: []*ResolutionResult{
			{
				Email:    "user1@example.com",
				UserID:   "user-123",
				Resolved: true,
			},
			{
				Email:    "user2@example.com",
				Resolved: false,
				Error:    fmt.Errorf("user not found"),
			},
			{
				Email:    "user3@example.com",
				Resolved: false,
				Error:    fmt.Errorf("API error"),
			},
		},
	}

	unresolved := resolver.GetUnresolvedEmails(batchResult)

	if len(unresolved) != 2 {
		t.Errorf("expected 2 unresolved emails, got %d", len(unresolved))
	}

	// Check that the correct emails are in the unresolved list
	expectedUnresolved := map[string]bool{
		"user2@example.com": true,
		"user3@example.com": true,
	}

	for _, email := range unresolved {
		if !expectedUnresolved[email] {
			t.Errorf("unexpected unresolved email: %s", email)
		}
	}
}

func TestCacheOperations(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := New(client, log)

	// Test cache stats
	stats := resolver.GetCacheStats()
	if stats["size"] != 0 {
		t.Errorf("expected initial cache size 0, got %v", stats["size"])
	}

	// Test clear cache
	resolver.ClearCache()

	// Verify cache is cleared
	statsAfterClear := resolver.GetCacheStats()
	if statsAfterClear["size"] != 0 {
		t.Errorf("expected cache size 0 after clear, got %v", statsAfterClear["size"])
	}
}

func TestConcurrentResolution(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}
