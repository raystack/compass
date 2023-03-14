package statsd

import (
	"time"

	std "github.com/DataDog/datadog-go/v5/statsd"
	"github.com/goto/salt/log"
)

// StatsD provides functions for reporting metrics.
type Reporter struct {
	client *std.Client
	logger log.Logger
	config Config
}

// New validates the config and initializes the statsD client.
func Init(logger log.Logger, cfg Config) (*Reporter, error) {
	reporter := &Reporter{}
	if !cfg.Enabled {
		logger.Warn("statsd is disabled")
		return reporter, nil
	}

	client, err := std.New(cfg.Address,
		std.WithNamespace(cfg.Prefix),
		std.WithoutTelemetry())
	if err != nil {
		return nil, err
	}

	reporter.client = client
	reporter.logger = logger
	reporter.config = cfg
	return reporter, nil
}

// Close closes statsd connection
func (sd *Reporter) Close() {
	if sd != nil && sd.client != nil {
		sd.Close()
	}
}

// Incr returns a increment counter metric.
func (sd *Reporter) Incr(name string) *Metric {
	return &Metric{
		rate:          sd.config.SamplingRate,
		logger:        sd.logger,
		name:          name,
		withInfluxTag: sd.config.WithInfluxTagFormat,
		publishFunc: func(name string, tags []string, rate float64) error {
			if sd == nil || sd.client == nil {
				return nil
			}

			return sd.client.Incr(name, tags, rate)
		},
	}
}

// Timing returns a timer metric.
func (sd *Reporter) Timing(name string, value time.Duration) *Metric {
	return &Metric{
		rate:          sd.config.SamplingRate,
		logger:        sd.logger,
		name:          name,
		withInfluxTag: sd.config.WithInfluxTagFormat,
		publishFunc: func(name string, tags []string, rate float64) error {
			if sd == nil || sd.client == nil {
				return nil
			}

			return sd.client.Timing(name, value, tags, rate)
		},
	}
}

// Gauge creates and returns a new gauge metric.
func (sd *Reporter) Gauge(name string, value float64) *Metric {
	return &Metric{
		rate:          sd.config.SamplingRate,
		logger:        sd.logger,
		name:          name,
		withInfluxTag: sd.config.WithInfluxTagFormat,
		publishFunc: func(name string, tags []string, rate float64) error {
			if sd == nil || sd.client == nil {
				return nil
			}

			return sd.client.Gauge(name, value, tags, rate)
		},
	}
}

// Histogram creates and returns a rate & gauge metric.
func (sd *Reporter) Histogram(name string, value float64) *Metric {
	return &Metric{
		rate:          sd.config.SamplingRate,
		logger:        sd.logger,
		name:          name,
		withInfluxTag: sd.config.WithInfluxTagFormat,
		publishFunc: func(name string, tags []string, rate float64) error {
			if sd == nil || sd.client == nil {
				return nil
			}

			return sd.client.Histogram(name, value, tags, rate)
		},
	}
}
