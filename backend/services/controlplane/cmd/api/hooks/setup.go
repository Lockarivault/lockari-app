package hooks

import (
	"context"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/services/controlplane/cmd/api/types"

	"go.uber.org/fx"
)

// Hook de inicialização para configurar handlers
func SetupHandlers(
	params types.AppParams,
	// registerUserRoutes handleruser.RegisterUserRoutesParams,
) error {
	params.Logger.Info("executing application setup handlers")

	// Validate critical dependencies
	if params.DB == nil {
		params.Logger.Error("critical dependency missing: NoSQL database service is nil")
		return fmt.Errorf("nosql database service is nil")
	}

	return nil

}

// Hook para iniciar o servidor web
func StartWebServer(lc fx.Lifecycle, params types.AppParams) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			errCh := make(chan error, 1)
			go func() {
				errCh <- params.Router.Start()
			}()

			select {
			case err := <-errCh:
				if err != nil {
					params.Logger.Error("failed to start web server", "error", err)
					return err
				}
				return nil
			case <-time.After(500 * time.Millisecond):
				// No immediate error observed; assume server started successfully.
				return nil
			}
		},
		OnStop: onStop(params),
	})
}

// Hook de shutdown
func onStop(params types.AppParams) func(context.Context) error {
	return func(ctx context.Context) error {

		// Shutdown do message queue
		// if params.MQ != nil {
		// 	if err := params.MQ.GracefulShutdown(ctx); err != nil {
		// 		params.Logger.Error("failed to shutdown message queue", "error", err)
		// 	} else {
		// 		params.Logger.Info("message queue shutdown completed")
		// 	}
		// }

		// Shutdown do servidor web
		if params.Router != nil {
			if err := params.Router.Shutdown(ctx); err != nil {
				params.Logger.Error("failed to shutdown web server", "error", err)
			} else {
				params.Logger.Info("web server shutdown completed")
			}
		}

		return nil
	}
}
