package cmd

import (
	"fmt"

	"github.com/jeremywohl/flatten"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	// Server Config
	ServerHost string `mapstructure:"SERVER_HOST" default:"0.0.0.0"`
	ServerPort string `mapstructure:"SERVER_PORT" default:"8080"`

	// Elasticsearch
	ElasticSearchBrokers string `mapstructure:"ELASTICSEARCH_BROKERS" default:"http://localhost:9200"`

	// StatsD
	StatsdAddress string `mapstructure:"STATSD_ADDRESS" default:"127.0.0.1:8125"`
	StatsdPrefix  string `mapstructure:"STATSD_PREFIX" default:"columbusApi"`
	StatsdEnabled bool   `mapstructure:"STATSD_ENABLED" default:"false"`

	TypeWhiteListStr         string `mapstructure:"SEARCH_WHITELIST" default:""`
	SearchTypesCacheDuration int    `mapstructure:"SEARCH_TYPES_CACHE_DURATION" default:"300"`

	// Lineage
	LineageRefreshIntervalStr string `mapstructure:"LINEAGE_REFRESH_INTERVAL" default:"5m"`

	// NewRelic
	NewRelicEnabled    bool   `mapstructure:"NEW_RELIC_ENABLED" default:"false"`
	NewRelicAppName    string `mapstructure:"NEW_RELIC_APP_NAME" default:"columbus"`
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
	IdentityHeader              string `mapstructure:"IDENTITY_HEADER" default:"Columbus-User-Email"`
	IdentityProviderDefaultName string `mapstructure:"IDENTITY_PROVIDER_DEFAULT_NAME" default:""`
}

var config Config

// LoadConfig returns application configuration
func loadConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("config file was not found. Env vars and defaults will be used")
		} else {
			return err
		}
	}

	bindEnvVars()
	defaults.SetDefaults(&config)
	err = viper.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("unable to unmarshal config to struct: %w", err)
	}

	return nil
}

func bindEnvVars() {
	configKeys, err := getFlattenedStructKeys(Config{})
	if err != nil {
		panic(err)
	}

	// Bind each conf fields to environment vars
	for key := range configKeys {
		err := viper.BindEnv(configKeys[key])
		if err != nil {
			panic(err)
		}
	}
}

func getFlattenedStructKeys(config Config) ([]string, error) {
	var structMap map[string]interface{}
	err := mapstructure.Decode(config, &structMap)
	if err != nil {
		return nil, err
	}

	flat, err := flatten.Flatten(structMap, "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}

	return keys, nil
}
