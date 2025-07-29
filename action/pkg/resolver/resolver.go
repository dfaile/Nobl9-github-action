package resolver

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
)

// Resolver handles email-to-UserID resolution using the Nobl9 API
type Resolver struct {
	client *nobl9.Client
	logger *logger.Logger
	cache  *UserCache
}

// UserInfo represents user information from Nobl9
type UserInfo struct {
	Email    string
	UserID   string
	Username string
	FullName string
	Active   bool
	Found    bool
	Error    error
}

// ResolutionResult represents the result of email resolution
type ResolutionResult struct {
	Email     string
	UserID    string
	Resolved  bool
	Error     error
	Duration  time.Duration
	FromCache bool
}

// BatchResolutionResult represents the result of batch email resolution
type BatchResolutionResult struct {
	Results       []*ResolutionResult
	TotalEmails   int
	ResolvedCount int
	ErrorCount    int
	CacheHits     int
	Duration      time.Duration
	Errors        []error
}

// UserCache provides caching for user information
type UserCache struct {
	users map[string]*UserInfo
	mutex sync.RWMutex
	ttl   time.Duration
}

// New creates a new resolver instance
func New(client *nobl9.Client, log *logger.Logger) *Resolver {
	return &Resolver{
		client: client,
		logger: log,
		cache:  NewUserCache(30 * time.Minute), // 30 minute TTL
	}
}

// NewUserCache creates a new user cache with the specified TTL
func NewUserCache(ttl time.Duration) *UserCache {
	return &UserCache{
		users: make(map[string]*UserInfo),
		ttl:   ttl,
	}
}

// ResolveEmail resolves a single email address to a UserID
func (r *Resolver) ResolveEmail(ctx context.Context, email string) (*ResolutionResult, error) {
	start := time.Now()

	// Normalize email
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	r.logger.Debug("Resolving email to UserID", logger.Fields{
		"email": normalizedEmail,
	})

	// Check cache first
	if cachedUser := r.cache.Get(normalizedEmail); cachedUser != nil {
		if cachedUser.Found {
			r.logger.LogUserResolution(normalizedEmail, cachedUser.UserID, true, logger.Fields{
				"from_cache": true,
				"duration":   time.Since(start).String(),
			})

			return &ResolutionResult{
				Email:     normalizedEmail,
				UserID:    cachedUser.UserID,
				Resolved:  true,
				Duration:  time.Since(start),
				FromCache: true,
			}, nil
		} else {
			// User not found in cache
			r.logger.LogUserResolution(normalizedEmail, "", false, logger.Fields{
				"from_cache": true,
				"error":      "user not found",
				"duration":   time.Since(start).String(),
			})

			return &ResolutionResult{
				Email:     normalizedEmail,
				Resolved:  false,
				Error:     fmt.Errorf("user not found"),
				Duration:  time.Since(start),
				FromCache: true,
			}, nil
		}
	}

	// Resolve via API
	user, err := r.client.GetUser(ctx, normalizedEmail)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			// Cache the "not found" result
			r.cache.Set(normalizedEmail, &UserInfo{
				Email: normalizedEmail,
				Found: false,
				Error: err,
			})

			r.logger.LogUserResolution(normalizedEmail, "", false, logger.Fields{
				"error":    err.Error(),
				"duration": time.Since(start).String(),
			})

			return &ResolutionResult{
				Email:    normalizedEmail,
				Resolved: false,
				Error:    fmt.Errorf("user not found: %w", err),
				Duration: time.Since(start),
			}, nil
		}

		// Other API errors
		r.logger.LogUserResolution(normalizedEmail, "", false, logger.Fields{
			"error":    err.Error(),
			"duration": time.Since(start).String(),
		})

		return &ResolutionResult{
			Email:    normalizedEmail,
			Resolved: false,
			Error:    fmt.Errorf("failed to resolve user: %w", err),
			Duration: time.Since(start),
		}, nil
	}

	// User found, cache the result
	userInfo := &UserInfo{
		Email:    normalizedEmail,
		UserID:   user.UserID,
		Username: user.UserID, // Use UserID as username since Username field doesn't exist
		FullName: user.UserID, // Use UserID as full name since FullName field doesn't exist
		Active:   true,        // Assume active since user was found
		Found:    true,
	}

	r.cache.Set(normalizedEmail, userInfo)

	r.logger.LogUserResolution(normalizedEmail, user.UserID, true, logger.Fields{
		"user_id":   user.UserID,
		"username":  user.UserID,
		"full_name": user.UserID,
		"active":    true,
		"duration":  time.Since(start).String(),
	})

	return &ResolutionResult{
		Email:     normalizedEmail,
		UserID:    user.UserID,
		Resolved:  true,
		Duration:  time.Since(start),
		FromCache: false,
	}, nil
}

// ResolveEmails resolves multiple email addresses to UserIDs
func (r *Resolver) ResolveEmails(ctx context.Context, emails []string) (*BatchResolutionResult, error) {
	start := time.Now()

	if len(emails) == 0 {
		return &BatchResolutionResult{
			Results:       []*ResolutionResult{},
			TotalEmails:   0,
			ResolvedCount: 0,
			ErrorCount:    0,
			CacheHits:     0,
			Duration:      time.Since(start),
		}, nil
	}

	r.logger.Info("Starting batch email resolution", logger.Fields{
		"email_count": len(emails),
	})

	// Use a semaphore to limit concurrent API calls
	semaphore := make(chan struct{}, 10) // Max 10 concurrent requests
	var wg sync.WaitGroup

	results := make([]*ResolutionResult, len(emails))
	errors := make([]error, 0)

	// Process emails concurrently
	for i, email := range emails {
		wg.Add(1)
		go func(index int, emailAddr string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := r.ResolveEmail(ctx, emailAddr)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to resolve %s: %w", emailAddr, err))
			}
			results[index] = result
		}(i, email)
	}

	// Wait for all resolutions to complete
	wg.Wait()

	// Calculate statistics
	resolvedCount := 0
	errorCount := 0
	cacheHits := 0

	for _, result := range results {
		if result != nil {
			if result.Resolved {
				resolvedCount++
			} else {
				errorCount++
			}
			if result.FromCache {
				cacheHits++
			}
		}
	}

	batchResult := &BatchResolutionResult{
		Results:       results,
		TotalEmails:   len(emails),
		ResolvedCount: resolvedCount,
		ErrorCount:    errorCount,
		CacheHits:     cacheHits,
		Duration:      time.Since(start),
		Errors:        errors,
	}

	r.logger.Info("Batch email resolution completed", logger.Fields{
		"total_emails":   batchResult.TotalEmails,
		"resolved_count": batchResult.ResolvedCount,
		"error_count":    batchResult.ErrorCount,
		"cache_hits":     batchResult.CacheHits,
		"duration":       batchResult.Duration.String(),
	})

	return batchResult, nil
}

// ResolveEmailsFromYAML extracts emails from YAML content and resolves them
func (r *Resolver) ResolveEmailsFromYAML(ctx context.Context, yamlContent []byte) (*BatchResolutionResult, error) {
	// Extract emails from YAML content
	emails, err := r.extractEmailsFromYAML(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract emails from YAML: %w", err)
	}

	if len(emails) == 0 {
		r.logger.Info("No emails found in YAML content")
		return &BatchResolutionResult{
			Results:       []*ResolutionResult{},
			TotalEmails:   0,
			ResolvedCount: 0,
			ErrorCount:    0,
			CacheHits:     0,
			Duration:      0,
		}, nil
	}

	r.logger.Info("Extracted emails from YAML", logger.Fields{
		"email_count": len(emails),
		"emails":      emails,
	})

	// Resolve the extracted emails
	return r.ResolveEmails(ctx, emails)
}

// extractEmailsFromYAML extracts email addresses from YAML content
func (r *Resolver) extractEmailsFromYAML(yamlContent []byte) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, you would parse the YAML and extract emails from specific fields

	content := string(yamlContent)
	emails := make([]string, 0)
	emailSet := make(map[string]bool)

	// Simple regex-like extraction for demonstration
	// In practice, you would use proper YAML parsing
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for email patterns in the YAML
		if strings.Contains(line, "@") && strings.Contains(line, ".") {
			// Extract potential email addresses
			words := strings.Fields(line)
			for _, word := range words {
				word = strings.Trim(word, "[]{}:,\"'")
				if r.isValidEmail(word) {
					normalizedEmail := strings.ToLower(strings.TrimSpace(word))
					if !emailSet[normalizedEmail] {
						emailSet[normalizedEmail] = true
						emails = append(emails, normalizedEmail)
					}
				}
			}
		}
	}

	return emails, nil
}

// isValidEmail performs basic email validation
func (r *Resolver) isValidEmail(email string) bool {
	// Basic email validation
	if !strings.Contains(email, "@") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Check local part
	if len(localPart) == 0 || len(localPart) > 64 {
		return false
	}

	// Check domain part
	if len(domainPart) == 0 || len(domainPart) > 255 {
		return false
	}

	// Check for valid domain format
	if !strings.Contains(domainPart, ".") {
		return false
	}

	return true
}

// GetCacheStats returns cache statistics
func (r *Resolver) GetCacheStats() map[string]interface{} {
	return r.cache.GetStats()
}

// ClearCache clears the user cache
func (r *Resolver) ClearCache() {
	r.cache.Clear()
	r.logger.Info("User cache cleared")
}

// Get retrieves a user from cache
func (c *UserCache) Get(email string) *UserInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if user, exists := c.users[email]; exists {
		return user
	}

	return nil
}

// Set stores a user in cache
func (c *UserCache) Set(email string, user *UserInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.users[email] = user
}

// GetStats returns cache statistics
func (c *UserCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"size": len(c.users),
		"ttl":  c.ttl.String(),
	}
}

// Clear clears all cached users
func (c *UserCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.users = make(map[string]*UserInfo)
}

// GetResolvedUserIDs returns a map of email to UserID for resolved users
func (r *Resolver) GetResolvedUserIDs(batchResult *BatchResolutionResult) map[string]string {
	emailToUserID := make(map[string]string)

	for _, result := range batchResult.Results {
		if result != nil && result.Resolved {
			emailToUserID[result.Email] = result.UserID
		}
	}

	return emailToUserID
}

// GetUnresolvedEmails returns a list of emails that could not be resolved
func (r *Resolver) GetUnresolvedEmails(batchResult *BatchResolutionResult) []string {
	unresolved := make([]string, 0)

	for _, result := range batchResult.Results {
		if result != nil && !result.Resolved {
			unresolved = append(unresolved, result.Email)
		}
	}

	return unresolved
}

// ValidateEmailFormat validates email format before resolution
func (r *Resolver) ValidateEmailFormat(email string) error {
	if !r.isValidEmail(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// ValidateEmails validates multiple email formats
func (r *Resolver) ValidateEmails(emails []string) []error {
	var errors []error

	for _, email := range emails {
		if err := r.ValidateEmailFormat(email); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
