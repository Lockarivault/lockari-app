package tenantcache

import (
	"context"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/database/cache"
	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	tenanttools "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/tools"
)

type tenantCacheRepository struct {
	cache cache.RedisClientInterface
	obs   telemetry.OtelObservability
}

func NewTenantCacheRepository(r cache.RedisClientInterface, obs telemetry.OtelObservability) (repositorytenant.RepositoryTenant, error) {
	if r == nil {
		return nil, tenanttools.ErrNilRepository
	}
	if obs == nil {
		return nil, tenanttools.ErrNilTelemetry
	}
	return &tenantCacheRepository{
		cache: r,
		obs:   obs,
	}, nil
}

func (r *tenantCacheRepository) Create(ctx context.Context, tenant tenantmodel.TenantModel) error {
	dto := fromDomain(tenant)
	key := fmt.Sprintf("tenant:%s", dto.ID)
	return r.cache.SetJSON(ctx, key, dto, 24*time.Hour)
}

func (r *tenantCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (tenantmodel.TenantModel, error) {
	var dto tenantCacheDTO
	key := fmt.Sprintf("tenant:%s", id.String())
	if err := r.cache.GetJSON(ctx, key, &dto); err != nil {
		return tenantmodel.TenantModel{}, err
	}
	return toDomain(dto), nil
}

func (r *tenantCacheRepository) GetBySlug(ctx context.Context, slug string) (tenantmodel.TenantModel, error) {
	// Cache listing or searching by slug usually requires secondary indexes or separate keys.
	// For now, we'll implement a simple key if slug is provided.
	var dto tenantCacheDTO
	key := fmt.Sprintf("tenant:slug:%s", slug)
	if err := r.cache.GetJSON(ctx, key, &dto); err != nil {
		return tenantmodel.TenantModel{}, err
	}
	return toDomain(dto), nil
}

func (r *tenantCacheRepository) Update(ctx context.Context, tenant tenantmodel.TenantModel) error {
	return r.Create(ctx, tenant)
}

func (r *tenantCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("tenant:%s", id.String())
	return r.cache.Delete(ctx, key)
}

func (r *tenantCacheRepository) List(ctx context.Context, filter map[string]interface{}) ([]tenantmodel.TenantModel, error) {
	return nil, fmt.Errorf("list operation not supported in cache")
}

func (r *tenantCacheRepository) UpdateSecurityMetadata(ctx context.Context, id uuid.UUID, metadata encryption.EncryptMetadata) error {
	return r.Delete(ctx, id)
}

func (r *tenantCacheRepository) ActivateTenant(ctx context.Context, id uuid.UUID) error {
	return r.Delete(ctx, id)
}

func (r *tenantCacheRepository) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	return r.Delete(ctx, id)
}

func (r *tenantCacheRepository) FailTenant(ctx context.Context, id uuid.UUID, reason string) error {
	return r.Delete(ctx, id)
}

func (r *tenantCacheRepository) UpdateProprietiesTypes(ctx context.Context, id uuid.UUID, properties tenantmodel.ProprietiesTypes) error {
	return r.Delete(ctx, id)
}
