package auditusecase

import (
	"context"

	auditmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/model"
	repositoryaudit "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/repository"
	"go.uber.org/fx"
)

type AuditService interface {
	// Log records an audit event with enrichment.
	Log(ctx context.Context, log auditmodel.AuditLog) error
}

type auditService struct {
	repo repositoryaudit.RepositoryAudit
}

func NewAuditService(repo repositoryaudit.RepositoryAudit) AuditService {
	return &auditService{
		repo: repo,
	}
}

func (s *auditService) Log(ctx context.Context, log auditmodel.AuditLog) error {
	// 1. Enrichment logic (can be expanded to pull from ctx)
	// For now, we assume the caller provides basic AATO info.

	return s.repo.Save(ctx, log)
}

var Module = fx.Options(
	fx.Provide(NewAuditService),
)
