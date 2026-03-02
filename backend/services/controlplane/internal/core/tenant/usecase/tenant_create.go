package tenantusecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/lockarivault/lockari-app/backend/libs/auditlog"
	"github.com/lockarivault/lockari-app/backend/libs/loggers"
	tenantmodel "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	tenantrepository "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	tenanttools "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/tools"
	"github.com/lockarivault/lockari-app/backend/services/controlplane/internal/infrastructure/messaging"
)

type TenantUsecase interface{}

type tenant struct {
	database  tenantrepository.TenantRepository
	logger    loggers.LoggerInterface
	publisher messaging.Publisher
	auditLog  auditlog.Service
}

func NewTenantUsecase(
	db tenantrepository.TenantRepository,
	logger loggers.LoggerInterface,
	publisher messaging.Publisher,
	audit auditlog.Service,
) (TenantUsecase, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if db == nil {
		logger.Error("tenant repository is required")
		return nil, errors.New("tenant repository is required")
	}
	if publisher == nil {
		logger.Error("publisher is required")
		return nil, errors.New("publisher is required")
	}
	if audit == nil {
		logger.Error("audit log is required")
		return nil, errors.New("audit log is required")
	}
	return &tenant{
		database:  db,
		logger:    logger,
		publisher: publisher,
		auditLog:  audit,
	}, nil
}

func (t *tenant) CreateTenant(ctx context.Context, tenant *tenantmodel.TenantModel) error {
	if tenant == nil {
		return errors.New("tenant is required")
	}

	if err := tenant.Validate(); err != nil {
		return fmt.Errorf("tenant validation failed: %w", err)
	}

	// check if tenant already exists by id
	existsByID, err := t.database.GetByID(ctx, tenant.ID.String())
	if err != nil {
		return fmt.Errorf("failed to check existing tenant id: %w", err)
	}
	if existsByID != nil {
		return fmt.Errorf("tenant with id %s already exists", tenant.ID)
	}

	// check if tenant already exists by name
	existsByName, err := t.database.Filter(ctx, map[string]any{
		"name": tenant.Name,
	}, tenantrepository.Pagination{})
	if err != nil {
		return fmt.Errorf("failed to check existing tenant name: %w", err)
	}
	if len(existsByName) > 0 {
		return fmt.Errorf("tenant with name %s already exists", tenant.Name)
	}

	// check if tenant already exists by slug
	existsBySlug, err := t.database.Filter(ctx, map[string]any{
		"slug": tenant.Slug,
	}, tenantrepository.Pagination{})
	if err != nil {
		return fmt.Errorf("failed to check existing tenant slug: %w", err)
	}
	if len(existsBySlug) > 0 {
		return fmt.Errorf("tenant with slug %s already exists", tenant.Slug)
	}

	if err := t.database.Create(ctx, tenant); err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	err = t.auditLog.CreateAuditLog(ctx, auditlog.AuditEntry{
		TenantID:     tenant.ID.String(),
		UserID:       tenanttools.GetUserIDFromContext(ctx),
		ResourceType: tenanttools.ResourceType,
		ResourceID:   tenant.ID.String(),
		Action:       auditlog.ActionCreate,
		IPAddress:    tenanttools.GetIPAddressFromContext(ctx),
		UserAgent:    tenanttools.GetUserAgentFromContext(ctx),
		ActorType:    auditlog.UserType(tenanttools.GetActorTypeFromContext(ctx)),
		Metadata: map[string]any{
			"name":            tenant.Name,
			"slug":            tenant.Slug,
			"system_resource": tenanttools.SystemResource,
		},
	})
	if err != nil {
		return fmt.Errorf("tenant was created but failed to create audit log: %w", err)
	}

	if err := t.publish(ctx, tenant); err != nil {
		t.logger.Error("tenant created, but failed to publish event via messaging", "error", err.Error())
	}

	return nil
}

func (t *tenant) publish(ctx context.Context, tenant *tenantmodel.TenantModel) error {
	if tenant == nil {
		return errors.New("tenant is required")
	}

	if err := tenant.Validate(); err != nil {
		return fmt.Errorf("tenant validation failed: %w", err)
	}

	event := messaging.Event{
		Type:    "tenant.created",
		Payload: tenant,
		Headers: map[string]any{
			"tenant_id": tenant.ID.String(),
			"actor":     tenanttools.GetUserIDFromContext(ctx),
			"trace_id":  tenanttools.GetTraceIDFromContext(ctx),
			"span_id":   tenanttools.GetSpanIDFromContext(ctx),
		},
	}

	if err := t.publisher.Publish(ctx, "tenant_exchange", "tenant.created", event); err != nil {
		return fmt.Errorf("failed to publish tenant: %w", err)
	}

	return nil
}
