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
	tenantservice "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/service"
	"go.opentelemetry.io/otel/attribute"
)

type createTenant struct {
	repo      repositorytenant.RepositoryTenant
	service   tenantservice.ServiceTenant
	cache     cache.RedisClientInterface
	logger    loggers.LoggerInterface
	telemetry telemetry.OtelObservability
}

type CreateTenantOutput struct {
	ID          uuid.UUID
	Name        string
	Description string
	Slug        string
	OwnerID     uuid.UUID
	Status      tenantmodel.StatusType
}

type CreateTetant interface {
	Execute(ctx context.Context, input tenantmodel.TenantModel) (*CreateTenantOutput, error)
}

func NewCreateTenant(
	repo repositorytenant.RepositoryTenant,
	service tenantservice.ServiceTenant,
	cache cache.RedisClientInterface,
	logger loggers.LoggerInterface,
	telemetry telemetry.OtelObservability,
) CreateTetant {
	return &createTenant{
		repo:      repo,
		service:   service,
		cache:     cache,
		logger:    logger,
		telemetry: telemetry,
	}
}

func (u *createTenant) Execute(ctx context.Context, input tenantmodel.TenantModel) (*CreateTenantOutput, error) {
	ctx, span := u.telemetry.Tracer(tracerName).Start(ctx, "CreateTenant.Execute")
	defer span.End()

	// 1. Pre-validation and Slug generation (Status PENDING is set here)
	if err := input.CreateValidate(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	span.SetAttributes(
		attribute.String("tenant_id", input.ID.String()),
		attribute.String("tenant_name", input.Name),
		attribute.String("tenant_description", description),
		attribute.String("tenant_slug", input.Slug),
	)

	// 2. Business Logic: Handle Slug and Name checks
	if input.Slug == "" {
		// Validar se existe algum com o mesmo nome
		filter := map[string]interface{}{"name": input.Name}
		existingByName, err := u.repo.List(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
		}
		if len(existingByName) > 0 {
			err := fmt.Errorf("tenant with name '%s' already exists", input.Name)
			span.RecordError(err)
			return nil, err
		}

		// Gere slug a partir do nome
		input.Slug = tenantmodel.NameToSlug(input.Name)

		// Valide se existe o slug depois
		existingBySlug, _ := u.repo.GetBySlug(ctx, input.Slug)
		if existingBySlug.ID != uuid.Nil {
			err := fmt.Errorf("generated slug '%s' already exists", input.Slug)
			span.RecordError(err)
			return nil, err
		}
	} else {
		// Se slug foi fornecido, valide unicidade
		existing, _ := u.repo.GetBySlug(ctx, input.Slug)
		if existing.ID != uuid.Nil {
			err := fmt.Errorf("tenant with slug '%s' already exists", input.Slug)
			span.RecordError(err)
			return nil, err
		}
	}

	// 3. Persist Initial Tenant (Status PENDING)
	if err := u.repo.Create(ctx, input); err != nil {
		span.RecordError(err)
		u.logger.Error("failed to create tenant in database", "error", err, "tenant_id", input.ID)
		return nil, fmt.Errorf("persistence failed: %w", err)
	}

	// 4. Publish 'tenant.created' to RabbitMQ for async provisioning
	if err := u.service.PublishTenantCreated(ctx, input); err != nil {
		u.logger.Warn("failed to publish tenant.created event", "error", err, "tenant_id", input.ID)
		// We don't fail the request here, but we should log it.
		// In a real scenario, we might want to use Outbox Pattern.
	}

	u.logger.Info("tenant created successfully (pending provisioning)", "id", input.ID, "slug", input.Slug)
	return &CreateTenantOutput{
		ID:          input.ID,
		Name:        input.Name,
		Description: description,
		Slug:        input.Slug,
		OwnerID:     input.OwnerID,
		Status:      input.Status,
	}, nil
}
