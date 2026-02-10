package types

import (
	"github.com/Lockarivault/lockari-app/backend/libs/database/nosql"
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/libs/webserver"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/config"
	"go.uber.org/fx"
)

type ConfigParams struct {
	fx.In
	Config config.AppConfig
}

type AppParams struct {
	fx.In
	Config config.AppConfig
	Router webserver.ServerInterface
	Logger loggers.LoggerInterface
	DB     nosql.DatabaseService
}
