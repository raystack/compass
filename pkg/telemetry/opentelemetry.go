package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/goto/salt/log"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/samplers/probability/consistent"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"google.golang.org/grpc/encoding/gzip"
)

type OpenTelemetryConfig struct {
	Enabled                bool          `yaml:"enabled" mapstructure:"enabled" default:"false"`
	CollectorAddr          string        `yaml:"collector_addr" mapstructure:"collector_addr" default:"localhost:4317"`
	PeriodicReadInterval   time.Duration `yaml:"periodic_read_interval" mapstructure:"periodic_read_interval" default:"1s"`
	TraceSampleProbability float64       `yaml:"trace_sample_probability" mapstructure:"trace_sample_probability" default:"1"`
}

func initOTLP(ctx context.Context, cfg Config, logger log.Logger) (func(), error) {
	if !cfg.OpenTelemetry.Enabled {
		logger.Info("OpenTelemetry monitoring is disabled.")
		return noOp, nil
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.AppName),
			semconv.ServiceVersion(cfg.AppVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	shutdownMetric, err := initGlobalMetrics(ctx, res, cfg.OpenTelemetry, logger)
	if err != nil {
		return nil, err
	}

	shutdownTracer, err := initGlobalTracer(ctx, res, cfg.OpenTelemetry, logger)
	if err != nil {
		shutdownMetric()
		return nil, err
	}

	shutdownProviders := func() {
		shutdownTracer()
		shutdownMetric()
	}

	if err := host.Start(); err != nil {
		shutdownProviders()
		return nil, err
	}

	if err := runtime.Start(); err != nil {
		shutdownProviders()
		return nil, err
	}

	return shutdownProviders, nil
}

func initGlobalMetrics(ctx context.Context, res *resource.Resource, cfg OpenTelemetryConfig, logger log.Logger) (func(), error) {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.CollectorAddr),
		otlpmetricgrpc.WithCompressor(gzip.Name),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(cfg.PeriodicReadInterval))
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
	otel.SetMeterProvider(provider)

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.Error("otlp metric-provider failed to shutdown", "err", err)
		}
	}, nil
}

func initGlobalTracer(ctx context.Context, res *resource.Resource, cfg OpenTelemetryConfig, logger log.Logger) (func(), error) {
	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(cfg.CollectorAddr),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithCompressor(gzip.Name),
	))
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(consistent.ProbabilityBased(cfg.TraceSampleProbability)),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()
		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			logger.Error("otlp trace-provider failed to shutdown", "err", err)
		}
	}, nil
}

func noOp() {}
