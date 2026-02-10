package tenantusecase

import (
	"context"
	"fmt"

	"github.com/Lockarivault/lockari-app/backend/libs/database/cache"
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type manageTenant struct {
	repo      repositorytenant.RepositoryTenant
	cache     cache.RedisClientInterface
	logger    loggers.LoggerInterface
	telemetry telemetry.OtelObservability
}

type ManageTenant interface {
	Update(ctx context.Context, tenant tenantmodel.TenantModel) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}) ([]tenantmodel.TenantModel, error)
}

func NewManageTenant(
	repo repositorytenant.RepositoryTenant,
	cache cache.RedisClientInterface,
	logger loggers.LoggerInterface,
	telemetry telemetry.OtelObservability,
) ManageTenant {
	return &manageTenant{
		repo:      repo,
		cache:     cache,
		logger:    logger,
		telemetry: telemetry,
	}
}

func (m *manageTenant) Update(ctx context.Context, tenant tenantmodel.TenantModel) error {
	ctx, span := m.telemetry.Tracer(tracerName).Start(ctx, "ManageTenant.Update")
	defer span.End()

	span.SetAttributes(attribute.String("tenant_id", tenant.ID.String()))

	if err := tenant.Validate(); err != nil {
		span.RecordError(err)
		return err
	}

	// 1. Update Database
	if err := m.repo.Update(ctx, tenant); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// 2. Invalidate Cache (ID and Slug)
	_ = m.invalidateCache(ctx, tenant.ID, tenant.Slug)

	return nil
}

func (m *manageTenant) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := m.telemetry.Tracer(tracerName).Start(ctx, "ManageTenant.Delete")
	defer span.End()

	span.SetAttributes(attribute.String("tenant_id", id.String()))

	// Get tenant first to know the slug for cache invalidation
	tenant, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 1. Delete from Database
	if err := m.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// 2. Invalidate Cache
	_ = m.invalidateCache(ctx, id, tenant.Slug)

	return nil
}

func (m *manageTenant) List(ctx context.Context, filter map[string]interface{}) ([]tenantmodel.TenantModel, error) {
	ctx, span := m.telemetry.Tracer(tracerName).Start(ctx, "ManageTenant.List")
	defer span.End()

	return m.repo.List(ctx, filter)
}

func (m *manageTenant) invalidateCache(ctx context.Context, id uuid.UUID, slug string) error {
	_ = m.cache.Delete(ctx, fmt.Sprintf("tenant:%s", id.String()))
	_ = m.cache.Delete(ctx, fmt.Sprintf("tenant:slug:%s", slug))
	return nil
}
