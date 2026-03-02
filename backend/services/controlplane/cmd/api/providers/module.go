package providers

import (
	"github.com/lockarivault/lockari-app/backend/services/controlplane/cmd/api/hooks"
	"go.uber.org/fx"
)

// Module wires all API providers and hooks into a single Fx module.
var Module = fx.Module("api-providers",
	fx.Provide(
		ProvideConfig,
		ProvideAppSettings,
		ProvideLogger,
		ProvideAuthService,
		ProvideServer,
		ProvideCache,
		ProvideNoSQL,
		ProvideObservability,
		ProvideQueue,
		ProvideAuditlog,
	),
	fx.Invoke(
		ProvideTelemetry,
		hooks.SetupHandlers,
		hooks.StartWebServer,
	),
)
