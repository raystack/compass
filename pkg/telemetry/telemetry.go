package telemetry

import (
	"context"
	"time"

	"github.com/goto/salt/log"
	"github.com/newrelic/go-agent/v3/newrelic"
)

const gracePeriod = 5 * time.Second

type Config struct {
	AppVersion string

	AppName       string              `yaml:"app_name" mapstructure:"app_name" default:"compass"`
	NewRelic      NewRelicConfig      `yaml:"newrelic" mapstructure:"newrelic"`
	OpenTelemetry OpenTelemetryConfig `yaml:"open_telemetry" mapstructure:"open_telemetry"`
}

func Init(ctx context.Context, cfg Config, logger log.Logger) (nrApp *newrelic.Application, cleanUp func(), err error) {
	shutdown, err := initOTLP(ctx, cfg, logger)
	if err != nil {
		return nil, noOp, err
	}

	nrApp, err = initNewRelicMonitor(cfg.AppName, cfg.NewRelic, logger)
	if err != nil {
		shutdown()
		return nil, noOp, err
	}

	return nrApp, func() {
		nrApp.Shutdown(gracePeriod)
		shutdown()
	}, nil
}
