package identity

import (
	repositoryidentitydb "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/identity/repository/database"
	"go.uber.org/fx"
)

var Module = fx.Options(
	repositoryidentitydb.Module,
)
