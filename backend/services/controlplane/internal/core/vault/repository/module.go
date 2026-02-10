package repositoryvault

import (
	"context"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	vaultmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/vault/model"
)

// RepositoryVault defines the storage contract for Vaults.
type RepositoryVault interface {
	Save(ctx context.Context, vault vaultmodel.Vault) error
	GetByID(ctx context.Context, tenantID uuid.UUID, vaultID uuid.UUID) (vaultmodel.Vault, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]vaultmodel.Vault, error)
}
