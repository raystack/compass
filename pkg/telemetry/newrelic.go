package telemetry

import (
	"fmt"

	"github.com/goto/salt/log"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type NewRelicConfig struct {
	Enabled bool `mapstructure:"enabled" default:"false"`
	// Deprecated: Use Config.AppName instead
	AppName    string `mapstructure:"appname" default:"compass"`
	LicenseKey string `mapstructure:"licensekey" default:""`
}

func initNewRelicMonitor(appName string, cfg NewRelicConfig, logger log.Logger) (*newrelic.Application, error) {
	if !cfg.Enabled {
		logger.Info("New Relic monitoring is disabled.")
		return nil, nil
	}

	if appName == "" && cfg.AppName != "" {
		appName = cfg.AppName
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(cfg.LicenseKey),
	)
	if err != nil {
		return nil, fmt.Errorf("init new relic monitor: %w", err)
	}

	logger.Info("NewRelic monitoring is enabled", "app", appName)
	return app, nil
}
