package vaultmodel

import (
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
)

// VaultType defines the usage pattern of the vault.
type VaultType string

const (
	VaultTypeGeneral VaultType = "GENERAL"
	VaultTypeShared  VaultType = "SHARED"
	VaultTypePrivate VaultType = "PRIVATE"
)

// Vault represents a container for secrets within a tenant.
type Vault struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Type        VaultType `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewVault creates a new vault for a specific tenant.
func NewVault(tenantID uuid.UUID, name string, description string, vType VaultType) Vault {
	now := time.Now().UTC()
	return Vault{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		Type:        vType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
