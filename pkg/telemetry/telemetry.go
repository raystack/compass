package telemetry

import (
	"context"
	"fmt"
	"time"

	log "github.com/raystack/salt/observability/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Init initializes OpenTelemetry with the given configuration.
// Returns a cleanup function that should be called on shutdown.
func Init(ctx context.Context, cfg Config, logger log.Logger) (func(), error) {
	if !cfg.OpenTelemetry.Enabled {
		logger.Info("OpenTelemetry is disabled")
		return func() {}, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Set up trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OpenTelemetry.CollectorAddr),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	sampler := trace.ParentBased(
		trace.TraceIDRatioBased(cfg.OpenTelemetry.TraceSampleProbability),
	)

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
		trace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set up metric exporter
	readInterval := cfg.OpenTelemetry.PeriodicReadInterval
	if readInterval == 0 {
		readInterval = 15 * time.Second
	}

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OpenTelemetry.CollectorAddr),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(readInterval),
		)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	logger.Info("OpenTelemetry initialized", "collector", cfg.OpenTelemetry.CollectorAddr)

	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			logger.Error("error shutting down tracer provider", "err", err)
		}
		if err := meterProvider.Shutdown(shutdownCtx); err != nil {
			logger.Error("error shutting down meter provider", "err", err)
		}
	}

	return cleanup, nil
}
