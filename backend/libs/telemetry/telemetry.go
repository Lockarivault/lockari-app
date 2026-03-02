package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	metricSDK "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Endpoint       string // e.g., localhost:4317
	Insecure       bool
	Headers        map[string]string
	EnableTraces   bool
	EnableMetrics  bool
	EnableLogs     bool
}

// OtelObservability defines the operations for observability.
type OtelObservability interface {
	Tracer(name string) trace.Tracer
	Meter(name string) metric.Meter
	TextMapPropagator() propagation.TextMapPropagator
}

type otelObs struct{}

func (o *otelObs) Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func (o *otelObs) Meter(name string) metric.Meter {
	return otel.GetMeterProvider().Meter(name)
}

func (o *otelObs) TextMapPropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// NewOtelObservability returns a new OtelObservability implementation.
func NewOtelObservability() OtelObservability {
	return &otelObs{}
}

func InitOTel(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			if e := fn(ctx); e != nil {
				err = e
			}
		}
		return err
	}

	// Propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// TRACING
	if cfg.EnableTraces {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithHeaders(cfg.Headers),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		traceExporter, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}

		tp := traceSDK.NewTracerProvider(
			traceSDK.WithBatcher(traceExporter),
			traceSDK.WithResource(res),
		)
		otel.SetTracerProvider(tp)
		shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	}

	// METRICS
	if cfg.EnableMetrics {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
			otlpmetricgrpc.WithHeaders(cfg.Headers),
		}
		if cfg.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}

		metricExporter, err := otlpmetricgrpc.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}

		mp := metricSDK.NewMeterProvider(
			metricSDK.WithReader(metricSDK.NewPeriodicReader(metricExporter, metricSDK.WithInterval(10*time.Second))),
			metricSDK.WithResource(res),
		)
		otel.SetMeterProvider(mp)
		shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	}

	// LOGS (Experimental)
	if cfg.EnableLogs {
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(cfg.Endpoint),
			otlploggrpc.WithHeaders(cfg.Headers),
		}
		if cfg.Insecure {
			opts = append(opts, otlploggrpc.WithInsecure())
		}

		logExporter, err := otlploggrpc.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create log exporter: %w", err)
		}

		lp := log.NewLoggerProvider(
			log.WithProcessor(log.NewBatchProcessor(logExporter)),
			log.WithResource(res),
		)
		// Set global logger provider if possible or provide a way to get it
		// otellog.SetLoggerProvider(lp) // Note: Global set might vary by version
		shutdownFuncs = append(shutdownFuncs, lp.Shutdown)
	}

	return shutdown, nil
}
