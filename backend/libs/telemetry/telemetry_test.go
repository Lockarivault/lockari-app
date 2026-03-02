package telemetry

import (
	"context"
	"testing"
)

func TestInitOTel(t *testing.T) {
	ctx := context.Background()

	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Endpoint:       "localhost:4317",
		Insecure:       true,
		EnableTraces:   true,
		EnableMetrics:  true,
		EnableLogs:     true,
	}

	shutdown, err := InitOTel(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to init OTel: %v", err)
	}

	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}

	// Clean up - we log instead of erroring because connection refused is expected without a local collector
	if err := shutdown(ctx); err != nil {
		t.Logf("OTel shutdown finished with expected dial error: %v", err)
	}
}
