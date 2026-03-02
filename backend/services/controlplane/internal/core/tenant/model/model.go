package tenantmodel

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lockarivault/lockari-app/backend/libs/uuid"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`)

type TenantModel struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	DisplayName  *string
	Slug         string
	OwnerID      uuid.UUID
	CreatedAt    time.Time
	DeletedAt    *time.Time
	Properties   ProprietiesKeys
	Status *TenantStatus
	TenantSpec   *TenantSpec
	Security     *TenantSecurity
}

func NewTenant(name, slug, ownerID string, description *string) (*TenantModel, error) {

	now := time.Now().UTC()
	m := &TenantModel{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: description,
		CreatedAt:   now,
	}

	m.Status = NewTenantStatus(StatusPending)
	m.TenantSpec = NewTenantSpec()
	m.Security = newTenantSecurity()

	err := m.createValidate(ownerID)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// createValidate prepares and validates the model for the creation flow.
func (m *TenantModel) createValidate(owner string) error {

	if owner == "" {
		return errors.New("owner is required")
	}

	u, err := uuid.Parse(owner)
	if err != nil {
		return err
	}

	m.OwnerID = u

	m.Name = strings.TrimSpace(m.Name)
	if m.Name == "" || len(m.Name) < 3 || len(m.Name) > 100 {
		return errors.New("name is required")
	}

	if strings.TrimSpace(m.Slug) != "" {
		m.Slug = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(m.Slug)), " ", "-")
	}

	err = m.Validate()
	if err != nil {
		return err
	}
	return nil
}

// isValidSlug checks if a string is lowercase, numbers and underscores only.
func (m *TenantModel) isValidSlug(s string) bool {
	return slugRegex.MatchString(s)
}

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

	if !m.isValidSlug(m.Slug) {
		return fmt.Errorf("tenant slug must be lowercase, numbers, underscores and dashes only")
	}

	// Owner validation
	if m.OwnerID == uuid.Nil {
		return fmt.Errorf("tenant owner ID is required")
	}

	// Status validation
	if m.Status == nil {
		return fmt.Errorf("tenant status is required")
	}
	if !m.Status.Status.IsValid() {
		return fmt.Errorf("invalid tenant status: %s", m.Status.Status)
	}

	// Spec validation
	if m.TenantSpec == nil {
		return fmt.Errorf("tenant spec is required")
	}

	// Security validation
	if m.Security == nil {
		return fmt.Errorf("tenant security is required")
	}

	return nil
}
