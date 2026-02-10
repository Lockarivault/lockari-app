package identitymodel

import (
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
)

// Group represents a logical grouping of users within a tenant.
type Group struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewGroup creates a new group for a specific tenant.
func NewGroup(tenantID uuid.UUID, name string, description string) Group {
	now := time.Now().UTC()
	return Group{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
