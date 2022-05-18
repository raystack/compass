package cli

import (
	"fmt"
	"os"

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
	// Server Config
	ServerHost string `mapstructure:"SERVER_HOST" default:"0.0.0.0"`
	ServerPort string `mapstructure:"SERVER_PORT" default:"8080"`

	// Elasticsearch
	ElasticSearchBrokers string `mapstructure:"ELASTICSEARCH_BROKERS" default:"http://localhost:9200"`

	// StatsD
	StatsdAddress string `mapstructure:"STATSD_ADDRESS" default:"127.0.0.1:8125"`
	StatsdPrefix  string `mapstructure:"STATSD_PREFIX" default:"compassApi"`
	StatsdEnabled bool   `mapstructure:"STATSD_ENABLED" default:"false"`

	TypeWhiteListStr         string `mapstructure:"SEARCH_WHITELIST" default:""`
	SearchTypesCacheDuration int    `mapstructure:"SEARCH_TYPES_CACHE_DURATION" default:"300"`

	// Lineage
	LineageRefreshIntervalStr string `mapstructure:"LINEAGE_REFRESH_INTERVAL" default:"5m"`

	// NewRelic
	NewRelicEnabled    bool   `mapstructure:"NEW_RELIC_ENABLED" default:"false"`
	NewRelicAppName    string `mapstructure:"NEW_RELIC_APP_NAME" default:"compass"`
	NewRelicLicenseKey string `mapstructure:"NEW_RELIC_LICENSE_KEY" default:""`

	// Log
	LogLevel string `mapstructure:"LOG_LEVEL" default:"info"`

	// Database
	DBHost     string `mapstructure:"DB_HOST" default:"localhost"`
	DBPort     int    `mapstructure:"DB_PORT" default:"5432"`
	DBName     string `mapstructure:"DB_NAME" default:"postgres"`
	DBUser     string `mapstructure:"DB_USER" default:"root"`
	DBPassword string `mapstructure:"DB_PASSWORD" default:""`
	DBSSLMode  string `mapstructure:"DB_SSL_MODE" default:"disable"`

	// User Identity
	IdentityUUIDHeaderKey       string `mapstructure:"IDENTITY_UUID_HEADER" default:"Compass-User-UUID"`
	IdentityEmailHeaderKey      string `mapstructure:"IDENTITY_EMAIL_HEADER" default:"Compass-User-Email"`
	IdentityProviderDefaultName string `mapstructure:"IDENTITY_PROVIDER_DEFAULT_NAME" default:""`
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
