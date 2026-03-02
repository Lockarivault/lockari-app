package tenantrepository

import (
	"context"
	"errors"

	tenantmodel "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	"go.uber.org/fx"
)

// Pagination defines the structure for paginated requests.
type Pagination struct {
	Limit int64
	Skip  int64
}

// Module is the uber fx module for the tenant repository.
// It provides the TenantRepository interface to the application,
// coordinating between Cache and Database implementations.
var Module = fx.Module("tenant-repository",
	fx.Provide(NewTenantRepository),
)

// TenantRepository defines the standard interface for tenant data access.
type TenantRepository interface {
	GetByID(ctx context.Context, id string) (*tenantmodel.TenantModel, error)
	Create(ctx context.Context, tenant *tenantmodel.TenantModel) error
	Update(ctx context.Context, tenant *tenantmodel.TenantModel) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, p Pagination) ([]*tenantmodel.TenantModel, error)
	Filter(ctx context.Context, filter map[string]any, p Pagination) ([]*tenantmodel.TenantModel, error)
}

// TenantCacheRepository is a specialized interface for cache operations.
// It focuses on simple key-value operations with TTL support.
type TenantCacheRepository interface {
	Get(ctx context.Context, id string) (*tenantmodel.TenantModel, error)
	Set(ctx context.Context, tenant *tenantmodel.TenantModel) error
	Delete(ctx context.Context, id string) error
}

// TenantDBRepository is a specialized interface for database operations.
// It provides full persistence and complex query capabilities.
type TenantDBRepository interface {
	GetByID(ctx context.Context, id string) (*tenantmodel.TenantModel, error)
	Create(ctx context.Context, tenant *tenantmodel.TenantModel) error
	Update(ctx context.Context, tenant *tenantmodel.TenantModel) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, p Pagination) ([]*tenantmodel.TenantModel, error)
	Filter(ctx context.Context, filter map[string]any, p Pagination) ([]*tenantmodel.TenantModel, error)
}

// tenant implements TenantRepository by coordinating cache and database.
// This implementation uses a cache-aside pattern to ensure performance
// while maintaining data consistency.
type tenant struct {
	cache TenantCacheRepository
	db    TenantDBRepository
}

// NewTenantRepository creates a new instance of TenantRepository.
// It requires both a cache and a database implementation to function.
func NewTenantRepository(cache TenantCacheRepository, db TenantDBRepository) (TenantRepository, error) {
	if cache == nil {
		return nil, errors.New("cache is required")
	}
	if db == nil {
		return nil, errors.New("db is required")
	}
	m := &tenant{
		cache: cache,
		db:    db,
	}
	return m, nil
}

// GetByID retrieves a tenant by its ID.
// It first attempts to find the tenant in the cache. If not found,
// it queries the database and populates the cache for future requests.
func (t *tenant) GetByID(ctx context.Context, id string) (*tenantmodel.TenantModel, error) {
	if id == "" {
		return nil, errors.New("tenant id is required")
	}

	// Try cache first
	m, err := t.cache.Get(ctx, id)
	if err == nil && m != nil {
		return m, nil
	}

	// Fallback to database
	m, err = t.db.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if m != nil {
		// Populate cache asynchronously to not block the response
		_ = t.cache.Set(ctx, m)
	}

	return m, nil
}

// Create persists a new tenant in both the database and the cache.
func (t *tenant) Create(ctx context.Context, model *tenantmodel.TenantModel) error {
	if model == nil {
		return errors.New("tenant is required")
	}

	// Persist to database first
	err := t.db.Create(ctx, model)
	if err != nil {
		return err
	}

	// Ensure cache is updated or invalidated
	return t.cache.Set(ctx, model)
}

// Update modifies an existing tenant in the database and updates the cache.
func (t *tenant) Update(ctx context.Context, model *tenantmodel.TenantModel) error {
	if model == nil {
		return errors.New("tenant is required")
	}

	err := t.db.Update(ctx, model)
	if err != nil {
		return err
	}

	return t.cache.Set(ctx, model)
}

// Delete performs a soft delete of a tenant in the database and removes it from the cache.
func (t *tenant) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	err := t.db.Delete(ctx, id)
	if err != nil {
		return err
	}

	_ = t.cache.Delete(ctx, id)
	return nil
}

// List retrieves all non-deleted tenants from the database with pagination.
func (t *tenant) List(ctx context.Context, p Pagination) ([]*tenantmodel.TenantModel, error) {
	results, err := t.db.List(ctx, p)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		// Populate cache in background to not block the response.
		// Use context.WithoutCancel to ensure background updates complete
		// even if the request context is cancelled.
		bgCtx := context.WithoutCancel(ctx)
		go func(tenants []*tenantmodel.TenantModel) {
			for _, res := range tenants {
				_ = t.cache.Set(bgCtx, res)
			}
		}(results)
	}

	return results, nil
}

// Filter retrieves tenants based on a filter map with pagination.
func (t *tenant) Filter(ctx context.Context, filter map[string]any, p Pagination) ([]*tenantmodel.TenantModel, error) {
	results, err := t.db.Filter(ctx, filter, p)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		// Populate cache in background to not block the response.
		bgCtx := context.WithoutCancel(ctx)
		go func(tenants []*tenantmodel.TenantModel) {
			for _, res := range tenants {
				_ = t.cache.Set(bgCtx, res)
			}
		}(results)
	}

	return results, nil
}
