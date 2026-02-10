package providers

import (
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/cmd/api/hooks"
	"go.uber.org/fx"
)

// Module wires all API providers and hooks into a single Fx module.
var Module = fx.Module("api-providers",
	fx.Provide(
		ProvideConfig,
		ProvideAppSettings,
		ProvideLogger,
		ProvideServer,
		ProvideCache,
		ProvideNoSQL,
		ProvideObservability,
		ProvideQueue,
	),
	fx.Invoke(
		ProvideTelemetry,
		hooks.SetupHandlers,
		hooks.StartWebServer,
	),
)
