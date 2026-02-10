package vault

import (
	repositoryvaultdb "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/vault/repository/database"
	"go.uber.org/fx"
)

var Module = fx.Options(
	repositoryvaultdb.Module,
)
