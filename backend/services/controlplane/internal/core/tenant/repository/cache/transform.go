package tenantcache

import (
	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
)

func toDomain(d tenantCacheDTO) tenantmodel.TenantModel {
	id, _ := uuid.Parse(d.ID)
	ownerID, _ := uuid.Parse(d.OwnerID)

	secMeta := encryption.NewEncryptMetadata(
		d.SecurityMetadata.KeyID,
		d.SecurityMetadata.KeyType,
		d.SecurityMetadata.Algorithm,
		d.SecurityMetadata.Provider,
	)
	secMeta.WithParentKeyID(d.SecurityMetadata.ParentKeyID).
		WithStatus(d.SecurityMetadata.Status).
		WithVersion(d.SecurityMetadata.Version).
		WithExpiresAt(d.SecurityMetadata.ExpiresAt).
		WithFingerprint(d.SecurityMetadata.Fingerprint)

	return tenantmodel.TenantModel{
		ID:               id,
		Name:             d.Name,
		Description:      d.Description,
		DisplayName:      d.DisplayName,
		Slug:             d.Slug,
		OwnerID:          ownerID,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
		DeletedAt:        d.DeletedAt,
		Status:           tenantmodel.StatusType(d.Status),
		SecurityMetadata: *secMeta,
		Properties:       tenantmodel.NewProprieties(d.Properties.Items),
	}
}

func fromDomain(m tenantmodel.TenantModel) tenantCacheDTO {
	return tenantCacheDTO{
		ID:          m.ID.String(),
		Name:        m.Name,
		Description: m.Description,
		DisplayName: m.DisplayName,
		Slug:        m.Slug,
		OwnerID:     m.OwnerID.String(),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
		Status:      string(m.Status),
		SecurityMetadata: securityMetadataDTO{
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
		Properties: propertiesDTO{
			Items: m.Properties.GetItems(),
		},
	}
}
