package cli

import (
	"fmt"
	"os"

	"github.com/odpf/compass/internal/server"
	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const configFlag = "config"

func cmdShowConfigs() *cobra.Command {
	return &cobra.Command{
		Use:   "configs",
		Short: "Display configurations currently loaded",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := loadConfig(cmd)
			if err != nil {
				fmt.Printf("failed to read configs: %v\n", err)
				os.Exit(1)
			}
			_ = yaml.NewEncoder(os.Stdout).Encode(cfg)
		},
	}
}

type Config struct {
	// Log
	LogLevel string `mapstructure:"LOG_LEVEL" default:"info"`

	// StatsD
	StatsdAddress string `mapstructure:"STATSD_ADDRESS" default:"127.0.0.1:8125"`
	StatsdPrefix  string `mapstructure:"STATSD_PREFIX" default:"compassApi"`
	StatsdEnabled bool   `mapstructure:"STATSD_ENABLED" default:"false"`

	// NewRelic
	NewRelicEnabled    bool   `mapstructure:"NEW_RELIC_ENABLED" default:"false"`
	NewRelicAppName    string `mapstructure:"NEW_RELIC_APP_NAME" default:"compass"`
	NewRelicLicenseKey string `mapstructure:"NEW_RELIC_LICENSE_KEY" default:""`

	// Elasticsearch
	ElasticSearchBrokers string `mapstructure:"ELASTICSEARCH_BROKERS" default:"http://localhost:9200"`

	// Database
	DBHost     string `mapstructure:"DB_HOST" default:"localhost"`
	DBPort     int    `mapstructure:"DB_PORT" default:"5432"`
	DBName     string `mapstructure:"DB_NAME" default:"postgres"`
	DBUser     string `mapstructure:"DB_USER" default:"root"`
	DBPassword string `mapstructure:"DB_PASSWORD" default:""`
	DBSSLMode  string `mapstructure:"DB_SSL_MODE" default:"disable"`

	// // User Identity
	Service server.Config `mapstructure:"service"`
}

func loadConfig(cmd *cobra.Command) (Config, error) {
	var opts []config.LoaderOption

	cfgFile, _ := cmd.Flags().GetString(configFlag)
	if cfgFile != "" {
		opts = append(opts, config.WithFile(cfgFile))
	} else {
		opts = append(opts,
			config.WithPath("./"),
			config.WithName("compass"),
		)
	}

	var cfg Config
	err := config.NewLoader(opts...).Load(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
