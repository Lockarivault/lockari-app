package tenantmodel

import (
	"strings"
	"testing"

	"github.com/lockarivault/lockari-app/backend/libs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTenant(t *testing.T) {
	ownerID := uuid.New().String()
	name := "Test Tenant"
	slug := "test-tenant"
	description := "A test description"

	tests := []struct {
		name        string
		tenantName  string
		tenantSlug  string
		ownerID     string
		description *string
		wantErr     bool
		errMsgs     []string
	}{
		{
			name:        "valid tenant",
			tenantName:  name,
			tenantSlug:  slug,
			ownerID:     ownerID,
			description: &description,
			wantErr:     false,
		},
		{
			name:        "empty owner",
			tenantName:  name,
			tenantSlug:  slug,
			ownerID:     "",
			description: &description,
			wantErr:     true,
			errMsgs:     []string{"owner is required"},
		},
		{
			name:        "invalid owner uuid",
			tenantName:  name,
			tenantSlug:  slug,
			ownerID:     "invalid-uuid",
			description: &description,
			wantErr:     true,
		},
		{
			name:        "name too short",
			tenantName:  "ab",
			tenantSlug:  slug,
			ownerID:     ownerID,
			description: &description,
			wantErr:     true,
			errMsgs:     []string{"name is required"},
		},
		{
			name:        "name too long",
			tenantName:  strings.Repeat("a", 101),
			tenantSlug:  slug,
			ownerID:     ownerID,
			description: &description,
			wantErr:     true,
			errMsgs:     []string{"name is required"},
		},
		{
			name:        "empty slug (will fail validation because Validate requires it)",
			tenantName:  name,
			tenantSlug:  "",
			ownerID:     ownerID,
			description: &description,
			wantErr:     true,
			errMsgs:     []string{"tenant slug cannot be empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewTenant(tt.tenantName, tt.tenantSlug, tt.ownerID, tt.description)
			if tt.wantErr {
				assert.Error(t, err)
				if len(tt.errMsgs) > 0 {
					found := false
					for _, msg := range tt.errMsgs {
						if strings.Contains(err.Error(), msg) {
							found = true
							break
						}
					}
					assert.True(t, found, "error message '%s' not found in '%v'", tt.errMsgs, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, m)
				assert.Equal(t, tt.tenantName, m.Name)
				assert.Equal(t, tt.tenantSlug, m.Slug)
				assert.Equal(t, ownerID, m.OwnerID.String())
				assert.NotNil(t, m.Status)
				assert.NotNil(t, m.TenantSpec)
				assert.NotNil(t, m.Security)
				assert.Equal(t, StatusPending, m.Status.Status)
			}
		})
	}
}

func TestTenantModel_IsValidSlug(t *testing.T) {
	m := &TenantModel{}

	tests := []struct {
		slug  string
		valid bool
	}{
		{"valid-slug", true},
		{"valid_slug", true},
		{"valid123", true},
		{"-invalid-start", false},
		{"invalid-end-", false},
		{"_invalid-start", false},
		{"invalid-end_", false},
		{"Invalid-Case", false},
		{"invalid space", false},
		{"invalid$char", false},
		{"", false},
		{"a", true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.valid, m.isValidSlug(tt.slug), "slug: %s", tt.slug)
	}
}

func TestTenantModel_Validate(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name    string
		model   *TenantModel
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid model",
			model: &TenantModel{
				Name:    "Valid Name",
				Slug:    "valid-slug",
				OwnerID: ownerID,
				Status: &TenantStatus{
					Status: StatusActive,
				},
				TenantSpec: &TenantSpec{},
				Security:   &TenantSecurity{},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			model: &TenantModel{
				Name:    "  ",
				Slug:    "valid-slug",
				OwnerID: ownerID,
			},
			wantErr: true,
			errMsg:  "tenant name cannot be empty",
		},
		{
			name: "name too short",
			model: &TenantModel{
				Name:    "ab",
				Slug:    "valid-slug",
				OwnerID: ownerID,
			},
			wantErr: true,
			errMsg:  "tenant name must be between 3 and 100 characters",
		},
		{
			name: "empty slug",
			model: &TenantModel{
				Name:    "Valid Name",
				Slug:    "",
				OwnerID: ownerID,
			},
			wantErr: true,
			errMsg:  "tenant slug cannot be empty",
		},
		{
			name: "invalid slug",
			model: &TenantModel{
				Name:    "Valid Name",
				Slug:    "-invalid-",
				OwnerID: ownerID,
			},
			wantErr: true,
			errMsg:  "tenant slug must be lowercase, numbers, underscores and dashes only",
		},
		{
			name: "nil owner",
			model: &TenantModel{
				Name:    "Valid Name",
				Slug:    "valid-slug",
				OwnerID: uuid.Nil,
			},
			wantErr: true,
			errMsg:  "tenant owner ID is required",
		},
		{
			name: "invalid status",
			model: &TenantModel{
				Name:    "Valid Name",
				Slug:    "valid-slug",
				OwnerID: ownerID,
				Status: &TenantStatus{
					Status: StatusType("INVALID"),
				},
			},
			wantErr: true,
			errMsg:  "invalid tenant status",
		},
		{
			name: "nil security",
			model: &TenantModel{
				Name:         "Valid Name",
				Slug:         "valid-slug",
				OwnerID:      ownerID,
				Status: &TenantStatus{Status: StatusActive},
				TenantSpec:   &TenantSpec{},
				Security:     nil,
			},
			wantErr: true,
			errMsg:  "tenant security is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
