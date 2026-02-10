package providers

import (
	"context"
	"fmt"

	"github.com/Lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/config"
	"go.uber.org/fx"
)

// ProvideTelemetry initializes OpenTelemetry and provides a shutdown function.
func ProvideTelemetry(lc fx.Lifecycle, cfg config.AppConfig) error {
	enabled := true // Default to true if not specified
	endpoint := "localhost:4317"
	serviceName := "control-plane-api"
	serviceVersion := "1.0.0"
	insecure := true

	if val, ok := cfg["telemetry"].(map[string]interface{}); ok {
		if v, ok := val["enabled"].(bool); ok {
			enabled = v
		}
		if v, ok := val["endpoint"].(string); ok {
			endpoint = v
		}
		if v, ok := val["service_name"].(string); ok {
			serviceName = v
		}
		if v, ok := val["service_version"].(string); ok {
			serviceVersion = v
		}
		if v, ok := val["insecure"].(bool); ok {
			insecure = v
		}
	}

	if !enabled {
		return nil
	}

	shutdown, err := telemetry.InitOTel(context.Background(), telemetry.Config{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Endpoint:       endpoint,
		Insecure:       insecure,
		EnableTraces:   true,
		EnableMetrics:  true,
		EnableLogs:     true,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return shutdown(ctx)
		},
	})

	return nil
}

// ProvideObservability provides the OTel observability implementation.
func ProvideObservability() telemetry.OtelObservability {
	return telemetry.NewOtelObservability()
}
