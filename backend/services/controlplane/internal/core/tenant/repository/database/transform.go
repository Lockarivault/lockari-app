package tenantdatabase

import (
	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
)

func toDomain(e tenantEntity) tenantmodel.TenantModel {
	id, _ := uuid.Parse(e.ID)
	ownerID, _ := uuid.Parse(e.OwnerID)

	// Transform SecurityMetadata Entity to Domain Interior Struct via public methods
	secMeta := encryption.NewEncryptMetadata(
		e.SecurityMetadata.KeyID,
		e.SecurityMetadata.KeyType,
		e.SecurityMetadata.Algorithm,
		e.SecurityMetadata.Provider,
	)
	secMeta.WithParentKeyID(e.SecurityMetadata.ParentKeyID).
		WithStatus(e.SecurityMetadata.Status).
		WithVersion(e.SecurityMetadata.Version).
		WithExpiresAt(e.SecurityMetadata.ExpiresAt).
		WithFingerprint(e.SecurityMetadata.Fingerprint)

	return tenantmodel.TenantModel{
		ID:               id,
		Name:             e.Name,
		Description:      e.Description,
		DisplayName:      e.DisplayName,
		Slug:             e.Slug,
		OwnerID:          ownerID,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
		DeletedAt:        e.DeletedAt,
		Status:           tenantmodel.StatusType(e.Status),
		FailureReason:    e.FailureReason,
		SecurityMetadata: *secMeta,
		Properties:       tenantmodel.NewProprieties(e.Properties.Items),
	}
}

func fromDomain(m tenantmodel.TenantModel) tenantEntity {
	return tenantEntity{
		ID:            m.ID.String(),
		Name:          m.Name,
		Description:   m.Description,
		DisplayName:   m.DisplayName,
		Slug:          m.Slug,
		OwnerID:       m.OwnerID.String(),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		DeletedAt:     m.DeletedAt,
		Status:        string(m.Status),
		FailureReason: m.FailureReason,
		SecurityMetadata: securityMetadataEntity{
			KeyID:       m.SecurityMetadata.GetKeyID(),
			ParentKeyID: m.SecurityMetadata.GetParentKeyID(),
			KeyType:     m.SecurityMetadata.GetKeyType(),
			Status:      m.SecurityMetadata.GetStatus(),
			Version:     m.SecurityMetadata.GetVersion(),
			Algorithm:   m.SecurityMetadata.GetAlgorithm(),
			Provider:    m.SecurityMetadata.GetProvider(),
			Fingerprint: m.SecurityMetadata.GetFingerprint(),
			CreatedAt:   m.SecurityMetadata.GetCreatedAt(),
			UpdatedAt:   m.SecurityMetadata.GetUpdatedAt(),
			ExpiresAt:   m.SecurityMetadata.GetExpiresAt(),
			RotatedAt:   m.SecurityMetadata.GetRotatedAt(),
		},
		Properties: propertiesEntity{
			Items: m.Properties.GetItems(),
		},
	}
}
func fromDomainMetadata(m encryption.EncryptMetadata) securityMetadataEntity {
	return securityMetadataEntity{
		KeyID:       m.GetKeyID(),
		ParentKeyID: m.GetParentKeyID(),
		KeyType:     m.GetKeyType(),
		Status:      m.GetStatus(),
		Version:     m.GetVersion(),
		Algorithm:   m.GetAlgorithm(),
		Provider:    m.GetProvider(),
		Fingerprint: m.GetFingerprint(),
		CreatedAt:   m.GetCreatedAt(),
		UpdatedAt:   m.GetUpdatedAt(),
		ExpiresAt:   m.GetExpiresAt(),
		RotatedAt:   m.GetRotatedAt(),
	}
}
