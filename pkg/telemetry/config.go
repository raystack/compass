package telemetry

import "time"

type Config struct {
	// ServiceName is the name of the service for telemetry identification
	ServiceName string `mapstructure:"service_name" default:"compass"`

	// OpenTelemetry configuration
	OpenTelemetry OpenTelemetryConfig `mapstructure:"open_telemetry"`
}

type OpenTelemetryConfig struct {
	// Enabled enables OpenTelemetry instrumentation
	Enabled bool `mapstructure:"enabled" default:"false"`

	// CollectorAddr is the address of the OTLP collector
	CollectorAddr string `mapstructure:"collector_addr" default:"localhost:4317"`

	// PeriodicReadInterval is the interval for periodic metric reads
	PeriodicReadInterval time.Duration `mapstructure:"periodic_read_interval" default:"15s"`

	// TraceSampleProbability is the probability of sampling traces (0.0 to 1.0)
	TraceSampleProbability float64 `mapstructure:"trace_sample_probability" default:"1"`
}
