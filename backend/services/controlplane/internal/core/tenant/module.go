package entitytenant

import (
	handlertenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/handler"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	tenantdatabase "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository/database"
	tenantservice "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/service"
	tenantusecase "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/usecase"

	"go.uber.org/fx"
)

// Module wires the tenant components into the Fx container.
var Module = fx.Module("tenant",
	fx.Provide(
		fx.Annotate(
			tenantdatabase.NewTenantMongoRepository,
			fx.As(new(repositorytenant.RepositoryTenant)),
		),
		fx.Annotate(
			tenantdatabase.NewKeyMongoRepository,
			fx.As(new(repositorytenant.RepositoryKey)),
		),
		tenantservice.InnicializeServiceTenant,
		tenantservice.NewSeedService,
		tenantusecase.NewCreateTenant,
		tenantusecase.NewGetTenant,
		tenantusecase.NewManageTenant,
		tenantusecase.NewWorkerTenant,
		tenantusecase.InnicializeUsecaseTenant,
	),
	fx.Invoke(
		handlertenant.NewHandler,
		func(w tenantusecase.WorkerTenant) {}, // Force instantiation of the worker
	),
)
