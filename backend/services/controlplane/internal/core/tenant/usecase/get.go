package tenantusecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/database/cache"
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type getTenant struct {
	repo      repositorytenant.RepositoryTenant
	cache     cache.RedisClientInterface
	logger    loggers.LoggerInterface
	telemetry telemetry.OtelObservability
}

type GetTenant interface {
	ByID(ctx context.Context, id uuid.UUID) (tenantmodel.TenantModel, error)
	BySlug(ctx context.Context, slug string) (tenantmodel.TenantModel, error)
}

func NewGetTenant(
	repo repositorytenant.RepositoryTenant,
	cache cache.RedisClientInterface,
	logger loggers.LoggerInterface,
	telemetry telemetry.OtelObservability,
) GetTenant {
	return &getTenant{
		repo:      repo,
		cache:     cache,
		logger:    logger,
		telemetry: telemetry,
	}
}

func (u *getTenant) ByID(ctx context.Context, id uuid.UUID) (tenantmodel.TenantModel, error) {
	ctx, span := u.telemetry.Tracer(tracerName).Start(ctx, "GetTenant.ByID")
	defer span.End()

	span.SetAttributes(attribute.String("tenant_id", id.String()))

	cacheKey := fmt.Sprintf("tenant:%s", id.String())

	// 1. Try Cache Aside
	var model tenantmodel.TenantModel
	err := u.cache.GetJSON(ctx, cacheKey, &model)
	if err == nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return model, nil
	}
	span.SetAttributes(attribute.Bool("cache_hit", false))

	// 2. Fetch from Database
	model, err = u.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return tenantmodel.TenantModel{}, err
	}

	// 3. Update Cache (Non-blocking or ignore error)
	_ = u.cache.SetJSON(ctx, cacheKey, model, 1*time.Hour)

	return model, nil
}

func (u *getTenant) BySlug(ctx context.Context, slug string) (tenantmodel.TenantModel, error) {
	ctx, span := u.telemetry.Tracer(tracerName).Start(ctx, "GetTenant.BySlug")
	defer span.End()

	span.SetAttributes(attribute.String("tenant_slug", slug))

	cacheKey := fmt.Sprintf("tenant:slug:%s", slug)

	// 1. Try Cache Aside
	var model tenantmodel.TenantModel
	err := u.cache.GetJSON(ctx, cacheKey, &model)
	if err == nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return model, nil
	}
	span.SetAttributes(attribute.Bool("cache_hit", false))

	// 2. Fetch from Database
	model, err = u.repo.GetBySlug(ctx, slug)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return tenantmodel.TenantModel{}, err
	}

	// 3. Update Cache
	_ = u.cache.SetJSON(ctx, cacheKey, model, 1*time.Hour)

	return model, nil
}
