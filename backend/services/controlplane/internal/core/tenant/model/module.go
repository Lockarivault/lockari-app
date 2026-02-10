package tenantmodel

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
)

type TenantModel struct {
	ID               uuid.UUID
	Name             string
	Description      *string
	DisplayName      *string
	Slug             string
	OwnerID          uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Status           StatusType
	FailureReason    *string
	SecurityMetadata encryption.EncryptMetadata
	Properties       ProprietiesTypes
}

// NewTenantModel creates a new tenant with default values.
func NewTenantModel(name, slug string, description *string, ownerID uuid.UUID) TenantModel {
	now := time.Now().UTC()

	return TenantModel{
		ID:          uuid.New(),
		Name:        name,
		Slug:        strings.ToLower(strings.TrimSpace(slug)),
		OwnerID:     ownerID,
		Description: description,
		Status:      StatusPending, // Inicia como PENDING para a esteira de provisionamento
		CreatedAt:   now,
		UpdatedAt:   now,
		Properties:  NewProprieties(nil),
	}
}

// CreateValidate prepares and validates the model for the creation flow.
func (m *TenantModel) CreateValidate() error {
	m.Name = strings.TrimSpace(m.Name)
	if m.Name == "" {
		return fmt.Errorf("tenant name cannot be empty")
	}

	if strings.TrimSpace(m.Slug) != "" {
		m.Slug = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(m.Slug)), " ", "-")
	}

	if m.OwnerID == uuid.Nil {
		return fmt.Errorf("tenant owner ID is required")
	}

	m.Status = StatusPending

	// If slug is already provided, we can run full validation.
	// If not, we'll validate again in the usecase after generation.
	if m.Slug != "" {
		return m.Validate()
	}

	return nil
}

// Validate checks if the tenant model adheres to business rules.
func (m *TenantModel) Validate() error {
	// Name validation
	name := strings.TrimSpace(m.Name)
	if name == "" {
		return fmt.Errorf("tenant name cannot be empty")
	}
	if len(name) < 3 || len(name) > 100 {
		return fmt.Errorf("tenant name must be between 3 and 100 characters")
	}

	// Slug validation
	if m.Slug == "" {
		return fmt.Errorf("tenant slug cannot be empty")
	}

	if m.Slug != "" {
		m.Slug = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(m.Slug)), " ", "-")
	}

	if !isValidSlug(m.Slug) {
		return fmt.Errorf("tenant slug must be lowercase, numbers, underscores and dashes only")
	}

	// Owner validation
	if m.OwnerID == uuid.Nil {
		return fmt.Errorf("tenant owner ID is required")
	}

	// Status validation
	if !m.Status.IsValid() {
		return fmt.Errorf("invalid tenant status: %s", m.Status)
	}

	// Nested components validation
	if err := m.Properties.Validate(); err != nil {
		return fmt.Errorf("invalid properties: %w", err)
	}

	// Security Metadata validation (only if populated)
	// Note: During creation it might be empty before the first encryption operation,
	// but once set, it must be valid.
	if m.SecurityMetadata.GetKeyID() != "" {
		if err := m.SecurityMetadata.Validate(); err != nil {
			return fmt.Errorf("invalid security metadata: %w", err)
		}
	}

	return nil
}

// NameToSlug converts a name to a slug using underscores.
func NameToSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "_")
	// Simple cleanup: remove non-alphanumeric (mostly)
	var sb strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sb.WriteRune(r)
		}
	}
	res := sb.String()
	// Remove multiple underscores
	for strings.Contains(res, "__") {
		res = strings.ReplaceAll(res, "__", "_")
	}
	return strings.Trim(res, "_")
}

// isValidSlug checks if a string is lowercase, numbers and underscores only.
func isValidSlug(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
		if (r == '_' || r == '-') && (i == 0 || i == len(s)-1) {
			return false
		}
	}
	return true
}
