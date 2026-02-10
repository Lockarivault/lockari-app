package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lockarivault/lockari-app/backend/services/controlplane/cmd/api/providers"
	_ "github.com/Lockarivault/lockari-app/backend/services/controlplane/docs/swagger"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/identity"
	entitytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/vault"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/pkg/notifications"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/pkg/webhooks"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// @title Lockari Vault Control Plane API
// @version 1.0
// @description API for managing tenants and core infrastructure.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8443
// @BasePath /
// @schemes https
func main() {
	app := fx.New(
		// Custom Fx logger - using NopLogger to keep console clean,
		// but can be swapped to &fxevent.ConsoleLogger{W: os.Stderr} for debugging.
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.NopLogger
		}),

		// Main module with all providers and hooks
		providers.Module,
		entitytenant.Module,
		audit.Module,
		identity.Module,
		vault.Module,
		notifications.Module,
		webhooks.Module,
	)

	// Initialization timeout
	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start application
	if err := app.Start(startCtx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start application: %v\n", err)

		// Attempt cleanup even after start failure
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer stopCancel()
		if stopErr := app.Stop(stopCtx); stopErr != nil {
			fmt.Fprintf(os.Stderr, "failed to stop application after start error: %v\n", stopErr)
		}

		os.Exit(1)
	}

	// Capture system signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wait for stop signal or internal termination
	select {
	case sig := <-sigs:
		fmt.Fprintf(os.Stdout, "received signal '%v', initiating shutdown\n", sig)
	case <-app.Wait():
		fmt.Fprintln(os.Stdout, "application stop requested via Fx lifecycle")
	}

	// Graceful shutdown
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop application: %v\n", err)
		os.Exit(1)
	}
}
