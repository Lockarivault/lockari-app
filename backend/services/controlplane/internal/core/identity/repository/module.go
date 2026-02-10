package repositoryidentity

import (
	"context"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	identitymodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/identity/model"
)

// RepositoryIdentity defines the storage contract for Identity-related resources (Groups, etc).
type RepositoryIdentity interface {
	SaveGroup(ctx context.Context, group identitymodel.Group) error
	GetGroupByID(ctx context.Context, tenantID uuid.UUID, groupID uuid.UUID) (identitymodel.Group, error)
	ListGroups(ctx context.Context, tenantID uuid.UUID) ([]identitymodel.Group, error)
}
