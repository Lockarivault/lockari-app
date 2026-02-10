package audit

import (
	repositoryauditdb "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/repository/database"
	auditusecase "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/usecase"
	"go.uber.org/fx"
)

var Module = fx.Options(
	repositoryauditdb.Module,
	auditusecase.Module,
)
