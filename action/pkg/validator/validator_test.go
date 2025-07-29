package validator

import (
	"context"
	"testing"
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-org/nobl9-action/pkg/logger"
	"github.com/your-org/nobl9-action/pkg/nobl9"
	"github.com/your-org/nobl9-action/pkg/resolver"
)

// MockClient is a mock implementation of the Nobl9 client
type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetProject(ctx context.Context, name string) (*rolebinding.RoleBinding, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*rolebinding.RoleBinding), args.Error(1)
}

func (m *MockClient) GetRoleBinding(ctx context.Context, projectName, name string) (*rolebinding.RoleBinding, error) {
	args := m.Called(ctx, projectName, name)
	return args.Get(0).(*rolebinding.RoleBinding), args.Error(1)
}

func (m *MockClient) ListRoleBindings(ctx context.Context, projectName string) ([]rolebinding.RoleBinding, error) {
	args := m.Called(ctx, projectName)
	return args.Get(0).([]rolebinding.RoleBinding), args.Error(1)
}

func (m *MockClient) GetUser(ctx context.Context, email string) (interface{}, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

// MockResolver is a mock implementation of the resolver
type MockResolver struct {
	mock.Mock
}

func (m *MockResolver) ValidateEmailFormat(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockResolver) ResolveEmail(ctx context.Context, email string) (*resolver.ResolutionResult, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*resolver.ResolutionResult), args.Error(1)
}

func TestNew(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &nobl9.Client{}
	resolver := &resolver.Resolver{}

	validator := New(client, resolver, log)

	assert.NotNil(t, validator)
	assert.Equal(t, client, validator.client)
	assert.Equal(t, resolver, validator.resolver)
	assert.Equal(t, log, validator.logger)
}

func TestValidateRoleBindingName(t *testing.T) {
	t.Skip("Skipping test that requires real Nobl9 client connection")
}

func TestValidateProjectName(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &MockClient{}
	resolver := &MockResolver{}
	validator := New(client, resolver, log)

	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "valid name",
			input:    "valid-project",
			expected: nil,
		},
		{
			name:     "empty name",
			input:    "",
			expected: assert.AnError,
		},
		{
			name:     "name too long",
			input:    "a-very-long-project-name-that-exceeds-the-maximum-length-of-sixty-three-characters",
			expected: assert.AnError,
		},
		{
			name:     "invalid characters",
			input:    "invalid_project",
			expected: assert.AnError,
		},
		{
			name:     "uppercase letters",
			input:    "Invalid-Project",
			expected: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateProjectName(tt.input)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGetRoleBindingRequirements(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &MockClient{}
	resolver := &MockResolver{}
	validator := New(client, resolver, log)

	tests := []struct {
		name     string
		role     string
		expected *RoleBindingRequirements
	}{
		{
			name: "project-owner role",
			role: "project-owner",
			expected: &RoleBindingRequirements{
				MinUsers:        1,
				MaxUsers:        10,
				RequiredRoles:   []string{"project-owner"},
				AllowedRoles:    []string{"project-owner"},
				ProjectRequired: true,
			},
		},
		{
			name: "project-editor role",
			role: "project-editor",
			expected: &RoleBindingRequirements{
				MinUsers:        0,
				MaxUsers:        50,
				RequiredRoles:   []string{"project-editor"},
				AllowedRoles:    []string{"project-editor"},
				ProjectRequired: true,
			},
		},
		{
			name: "project-viewer role",
			role: "project-viewer",
			expected: &RoleBindingRequirements{
				MinUsers:        0,
				MaxUsers:        100,
				RequiredRoles:   []string{"project-viewer"},
				AllowedRoles:    []string{"project-viewer"},
				ProjectRequired: true,
			},
		},
		{
			name: "custom role",
			role: "custom-role",
			expected: &RoleBindingRequirements{
				MinUsers:        0,
				MaxUsers:        50,
				RequiredRoles:   []string{"custom-role"},
				AllowedRoles:    []string{"custom-role"},
				ProjectRequired: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.getRoleBindingRequirements(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateRoleBindingRequirements(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &MockClient{}
	resolver := &MockResolver{}
	validator := New(client, resolver, log)

	tests := []struct {
		name        string
		validation  *RoleBindingValidation
		expectError bool
	}{
		{
			name: "valid requirements",
			validation: &RoleBindingValidation{
				Users: []*UserValidation{
					{CanBeAssigned: true},
					{CanBeAssigned: true},
				},
				Requirements: &RoleBindingRequirements{
					MinUsers: 1,
					MaxUsers: 10,
				},
			},
			expectError: false,
		},
		{
			name: "insufficient users",
			validation: &RoleBindingValidation{
				Users: []*UserValidation{
					{CanBeAssigned: true},
				},
				Requirements: &RoleBindingRequirements{
					MinUsers: 2,
					MaxUsers: 10,
				},
			},
			expectError: true,
		},
		{
			name: "too many users",
			validation: &RoleBindingValidation{
				Users: []*UserValidation{
					{CanBeAssigned: true},
					{CanBeAssigned: true},
					{CanBeAssigned: true},
				},
				Requirements: &RoleBindingRequirements{
					MinUsers: 1,
					MaxUsers: 2,
				},
			},
			expectError: true,
		},
		{
			name: "insufficient valid users",
			validation: &RoleBindingValidation{
				Users: []*UserValidation{
					{CanBeAssigned: false},
					{CanBeAssigned: true},
				},
				Requirements: &RoleBindingRequirements{
					MinUsers: 2,
					MaxUsers: 10,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRoleBindingRequirements(tt.validation)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCountValidUsers(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &MockClient{}
	resolver := &MockResolver{}
	validator := New(client, resolver, log)

	users := []*UserValidation{
		{CanBeAssigned: true},
		{CanBeAssigned: false},
		{CanBeAssigned: true},
		{CanBeAssigned: false},
		{CanBeAssigned: true},
	}

	count := validator.countValidUsers(users)
	assert.Equal(t, 3, count)
}

func TestCountInvalidUsers(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &MockClient{}
	resolver := &MockResolver{}
	validator := New(client, resolver, log)

	users := []*UserValidation{
		{CanBeAssigned: true},
		{CanBeAssigned: false},
		{CanBeAssigned: true},
		{CanBeAssigned: false},
		{CanBeAssigned: true},
	}

	count := validator.countInvalidUsers(users)
	assert.Equal(t, 2, count)
}

func TestGetValidationSummary(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &MockClient{}
	resolver := &MockResolver{}
	validator := New(client, resolver, log)

	validation := &RoleBindingValidation{
		Name:        "test-role-binding",
		ProjectName: "test-project",
		Role:        "project-owner",
		Users: []*UserValidation{
			{CanBeAssigned: true},
			{CanBeAssigned: false},
			{CanBeAssigned: true},
		},
		IsValid:  true,
		Errors:   []error{},
		Warnings: []string{"warning1"},
		Duration: 100 * time.Millisecond,
	}

	summary := validator.GetValidationSummary(validation)

	assert.Equal(t, "test-role-binding", summary["role_binding_name"])
	assert.Equal(t, "test-project", summary["project_name"])
	assert.Equal(t, "project-owner", summary["role"])
	assert.Equal(t, true, summary["is_valid"])
	assert.Equal(t, 3, summary["total_users"])
	assert.Equal(t, 2, summary["valid_users"])
	assert.Equal(t, 1, summary["invalid_users"])
	assert.Equal(t, 0, summary["error_count"])
	assert.Equal(t, 1, summary["warning_count"])
	assert.Equal(t, "100ms", summary["duration"])
}

func TestValidateUsers(t *testing.T) {
	log := logger.New(logger.LevelInfo, logger.FormatJSON)
	client := &MockClient{}
	resolver := &MockResolver{}
	validator := New(client, resolver, log)

	ctx := context.Background()
	emails := []string{"user1@example.com", "user2@example.com"}
	emailToUserID := map[string]string{
		"user1@example.com": "user-123",
		"user2@example.com": "user-456",
	}

	// Mock resolver behavior
	resolver.On("ValidateEmailFormat", "user1@example.com").Return(nil)
	resolver.On("ValidateEmailFormat", "user2@example.com").Return(nil)
	resolver.On("ResolveEmail", ctx, "user1@example.com").Return(&resolver.ResolutionResult{
		Email:    "user1@example.com",
		UserID:   "user-123",
		Resolved: true,
	}, nil)
	resolver.On("ResolveEmail", ctx, "user2@example.com").Return(&resolver.ResolutionResult{
		Email:    "user2@example.com",
		UserID:   "user-456",
		Resolved: true,
	}, nil)

	// Mock client behavior
	client.On("GetUser", ctx, "user1@example.com").Return(&struct{}{}, nil)
	client.On("GetUser", ctx, "user2@example.com").Return(&struct{}{}, nil)

	validations, err := validator.ValidateUsers(ctx, emails, emailToUserID)

	assert.NoError(t, err)
	assert.Len(t, validations, 2)
	assert.True(t, validations[0].CanBeAssigned)
	assert.True(t, validations[1].CanBeAssigned)

	resolver.AssertExpectations(t)
	client.AssertExpectations(t)
}
