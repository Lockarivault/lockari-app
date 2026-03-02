package tenantrepositorycache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lockarivault/lockari-app/backend/libs/database/cache"
	tenantmodel "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	tenantrepository "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
)

// prefix is the prefix used for tenant keys in the cache.
const prefix = "tenant:"

// TenantCache defines the interface for tenant cache operations.
type TenantCache interface {
	tenantrepository.TenantCacheRepository
}

// tenantCache provides an implementation of TenantCache using Redis.
// This exists to provide fast lookup of tenant information, reducing
// the load on the primary database for frequent read operations.
type tenantCache struct {
	cache cache.RedisClientInterface
}

// NewTenantCacheRepository creates a new instance of TenantCache.
// It requires a RedisClientInterface to interact with the cache layer.
func NewTenantCacheRepository(cache cache.RedisClientInterface) (TenantCache, error) {
	if cache == nil {
		return nil, errors.New("cache is required")
	}
	return &tenantCache{
		cache: cache,
	}, nil
}

// Get retrieves a tenant from the cache by its ID.
// If the tenant is not found in the cache, it returns nil, nil.
func (t *tenantCache) Get(ctx context.Context, id string) (*tenantmodel.TenantModel, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	key := fmt.Sprintf("%s%s", prefix, id)
	data, err := t.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if data == "" {
		return nil, nil
	}

	var p tenantrepository.TenantPersistence
	err = json.Unmarshal([]byte(data), &p)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling tenant: %w", err)
	}

	return tenantrepository.ConvertToTenantModel(&p), nil
}

// Set stores a tenant in the cache.
// This is typically called after a tenant is created in the database
// or during a cache-aside population.
func (t *tenantCache) Set(ctx context.Context, tenant *tenantmodel.TenantModel) error {
	if tenant == nil {
		return errors.New("tenant is required")
	}

	p := tenantrepository.ConvertToTenantStorage(tenant)
	key := fmt.Sprintf("%s%s", prefix, p.ID.String())

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("error marshalling tenant: %w", err)
	}

	return t.cache.Set(ctx, key, data, 24*time.Hour)
}

// Delete removes a tenant from the cache.
func (t *tenantCache) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	key := fmt.Sprintf("%s%s", prefix, id)
	return t.cache.Delete(ctx, key)
}
