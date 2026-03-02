package tenantrepository

import (
	"time"

	"github.com/lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
)

// TenantPersistence represents the database schema for a tenant.
// This struct exists to separate the database representation from the domain model,
// allowing the database schema to evolve independently of the business logic.
// It uses bson tags for MongoDB compatibility and json tags for general serialization.
type TenantPersistence struct {
	ID          uuid.UUID           `bson:"_id" json:"id"`
	Name        string              `bson:"name" json:"name"`
	Description *string             `bson:"description,omitempty" json:"description,omitempty"`
	DisplayName *string             `bson:"display_name,omitempty" json:"display_name,omitempty"`
	Slug        string              `bson:"slug" json:"slug"`
	OwnerID     uuid.UUID           `bson:"owner_id" json:"owner_id"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
	DeletedAt   *time.Time          `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	Properties  map[string]any      `bson:"properties" json:"properties"`
	Status      StatusPersistence   `bson:"status" json:"status"`
	Spec        SpecPersistence     `bson:"spec" json:"spec"`
	Security    SecurityPersistence `bson:"security" json:"security"`
}

type StatusPersistence struct {
	Status        string    `bson:"status" json:"status"`
	FailureReason *string   `bson:"failure_reason,omitempty" json:"failure_reason,omitempty"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

type SpecPersistence struct {
	// Add fields as they are defined in TenantSpec
}

type SecurityPersistence struct {
	KekID     string     `bson:"kek_id" json:"kek_id"`
	Algorithm string     `bson:"algorithm" json:"algorithm"`
	Provider  string     `bson:"provider" json:"provider"`
	RotatedAt *time.Time `bson:"rotated_at,omitempty" json:"rotated_at,omitempty"`
	Version   int        `bson:"version" json:"version"`
}

// ConvertToTenantModel transforms a persistence struct into a domain model.
// This ensures that the rest of the application only deals with rich domain objects
// and is shielded from database-specific types or structures.
func ConvertToTenantModel(p *TenantPersistence) *tenantmodel.TenantModel {
	if p == nil {
		return nil
	}

	return &tenantmodel.TenantModel{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		DisplayName: p.DisplayName,
		Slug:        p.Slug,
		OwnerID:     p.OwnerID,
		CreatedAt:   p.CreatedAt,
		DeletedAt:   p.DeletedAt,
		Properties:  tenantmodel.ProprietiesKeys(p.Properties),
		Status: &tenantmodel.TenantStatus{
			Status:        tenantmodel.StatusType(p.Status.Status),
			FailureReason: p.Status.FailureReason,
			UpdatedAt:     p.Status.UpdatedAt,
		},
		TenantSpec: &tenantmodel.TenantSpec{},
		Security: &tenantmodel.TenantSecurity{
			KekID:     p.Security.KekID,
			Algorithm: p.Security.Algorithm,
			Provider:  tenantmodel.SecurityProvider(p.Security.Provider),
			RotatedAt: p.Security.RotatedAt,
			Version:   p.Security.Version,
		},
	}
}

// ConvertToTenantStorage transforms a domain model into a persistence struct.
// This isolates the logic for preparing data for storage, such as mapping
// domain-specific types to database-friendly formats.
func ConvertToTenantStorage(m *tenantmodel.TenantModel) *TenantPersistence {
	if m == nil {
		return nil
	}

	p := &TenantPersistence{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		DisplayName: m.DisplayName,
		Slug:        m.Slug,
		OwnerID:     m.OwnerID,
		CreatedAt:   m.CreatedAt,
		DeletedAt:   m.DeletedAt,
		Properties:  map[string]any(m.Properties),
	}

	if m.Status != nil {
		p.Status = StatusPersistence{
			Status:        string(m.Status.Status),
			FailureReason: m.Status.FailureReason,
			UpdatedAt:     m.Status.UpdatedAt,
		}
	}

	if m.Security != nil {
		p.Security = SecurityPersistence{
			KekID:     m.Security.KekID,
			Algorithm: m.Security.Algorithm,
			Provider:  string(m.Security.Provider),
			RotatedAt: m.Security.RotatedAt,
			Version:   m.Security.Version,
		}
	}

	return p
}
