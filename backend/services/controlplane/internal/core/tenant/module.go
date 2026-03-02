package tenant

import (
	"log/slog"

	tenantrepositorycache "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository/cache"
	tenantrepositorydatabase "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository/database"
	"go.uber.org/fx"
)

type ModuleTenantParams struct {
	fx.In

	Observability *slog.Logger
}

// RepositoryCacheModule is the uber fx module for the tenant cache repository.
// It provides the TenantCache interface to the application.
var RepositoryCacheModule = fx.Module("tenant-cache-repository",
	fx.Provide(tenantrepositorycache.NewTenantCacheRepository),
)

// RepositoryDatabaseModule is the uber fx module for the tenant database repository.
// It provides the TenantDatabase interface to the application.
var RepositoryDatabaseModule = fx.Module("tenant-database-repository",
	fx.Provide(tenantrepositorydatabase.NewTenantDatabaseRepository),
)

var Module = fx.Module("tenant",
	RepositoryCacheModule,
	RepositoryDatabaseModule,
	fx.Invoke(
	// handleruser.RegisterUserRoutes,
	// invokeUserCreateConsumer,
	),
)
