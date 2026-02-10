package tenantusecase

import (
	"context"
	"errors"

	"github.com/Lockarivault/lockari-app/backend/libs/database/cache"
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	tenantservice "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/service"
)

type checkSlugTenant struct {
	repo      repositorytenant.RepositoryTenant
	service   tenantservice.ServiceTenant
	cache     cache.RedisClientInterface
	logger    loggers.LoggerInterface
	telemetry telemetry.OtelObservability
}

type CheckSlugTenant interface {
	CheckSlugAvailability(ctx context.Context, input string) (bool, error)
}

func NewCheckSlugTenant(
	repo repositorytenant.RepositoryTenant,
	service tenantservice.ServiceTenant,
	cache cache.RedisClientInterface,
	logger loggers.LoggerInterface,
	telemetry telemetry.OtelObservability,
) (CheckSlugTenant, error) {
	if repo == nil {
		return nil, errors.New("repository cannot be nil")
	}
	if service == nil {
		return nil, errors.New("service cannot be nil")
	}
	if cache == nil {
		return nil, errors.New("cache cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}
	if telemetry == nil {
		return nil, errors.New("telemetry cannot be nil")
	}
	return &checkSlugTenant{
		repo:      repo,
		service:   service,
		cache:     cache,
		logger:    logger,
		telemetry: telemetry,
	}, nil
}

// CheckSlugAvailability checks if a slug is available for a tenant.
// It returns true if the slug is available, false otherwise.
// It also returns an error if the input is empty or if the slug already exists.
func (c *checkSlugTenant) CheckSlugAvailability(ctx context.Context, input string) (bool, error) {

	if input == "" {
		return false, errors.New("input cannot be empty")
	}

	t, err := c.repo.GetBySlug(ctx, input)
	if err != nil {
		if t.ID == uuid.Nil {
			return true, nil
		}
		return false, err
	}

	if t.ID != uuid.Nil {
		if t.Slug == input {
			return false, nil
		}
	}

	return true, nil

}
