package tenantservice

import (
	"context"
	"fmt"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	auditmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/model"
	auditusecase "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/usecase"
	identitymodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/identity/model"
	repositoryidentity "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/identity/repository"
	vaultmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/vault/model"
	repositoryvault "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/vault/repository"
)

// SeedService defines the contract for provisioning initial tenant data.
type SeedService interface {
	ProvisionSeedData(ctx context.Context, tenantID uuid.UUID) error
}

type seedService struct {
	repoGroup    repositoryidentity.RepositoryIdentity
	repoVault    repositoryvault.RepositoryVault
	auditService auditusecase.AuditService
}

func NewSeedService(
	repoGroup repositoryidentity.RepositoryIdentity,
	repoVault repositoryvault.RepositoryVault,
	auditService auditusecase.AuditService,
) SeedService {
	return &seedService{
		repoGroup:    repoGroup,
		repoVault:    repoVault,
		auditService: auditService,
	}
}

func (s *seedService) ProvisionSeedData(ctx context.Context, tenantID uuid.UUID) error {
	// 1. Create "Administrators" Group (Step 6)
	adminGroup := identitymodel.NewGroup(tenantID, "Administrators", "Default group with full administrative access.")
	if err := s.repoGroup.SaveGroup(ctx, adminGroup); err != nil {
		return fmt.Errorf("failed to create administrators group: %w", err)
	}
	s.logEvent(ctx, tenantID, "GROUP_CREATE", adminGroup.ID.String(), "Administrators", auditmodel.TargetKey)

	// 2. Create "General" Vault (Step 7)
	generalVault := vaultmodel.NewVault(tenantID, "General", "Default vault for general secrets.", vaultmodel.VaultTypeGeneral)
	if err := s.repoVault.Save(ctx, generalVault); err != nil {
		return fmt.Errorf("failed to create general vault: %w", err)
	}
	s.logEvent(ctx, tenantID, "VAULT_CREATE", generalVault.ID.String(), "General", auditmodel.TargetSecret)

	return nil
}

func (s *seedService) logEvent(ctx context.Context, tenantID uuid.UUID, action string, targetID string, name string, tType auditmodel.TargetType) {
	log := auditmodel.NewAuditLog(
		auditmodel.ActorInfo{
			ID:   "system-seed-provisioner",
			Type: auditmodel.ActorSystem,
			IP:   "internal",
		},
		action,
		auditmodel.TargetInfo{
			ID:   targetID,
			Type: tType,
			Name: name,
		},
	)
	log.Outcome.Status = auditmodel.OutcomeSuccess
	log.Context.TenantID = tenantID

	_ = s.auditService.Log(ctx, log)
}
