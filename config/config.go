package config

import (
	"fmt"
	"strings"

	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
)

type Config struct {
	ServerHost                string `mapstructure:"SERVER_HOST" default:"0.0.0.0"`
	ServerPort                string `mapstructure:"SERVER_PORT" default:"8080"`
	ElasticSearchBrokers      string `mapstructure:"ELASTICSEARCH_BROKERS" default:"http://localhost:9200"`
	StatsdAddress             string `mapstructure:"STATSD_ADDRESS" default:"127.0.0.1:8125"`
	StatsdPrefix              string `mapstructure:"STATSD_PREFIX" default:"columbusApi"`
	StatsdEnabled             bool   `mapstructure:"STATSD_ENABLED" default:"false"`
	TypeWhiteListStr          string `mapstructure:"SEARCH_WHITELIST" default:""`
	LineageRefreshIntervalStr string `mapstructure:"LINEAGE_REFRESH_INTERVAL" default:"5m"`
	NewRelicEnabled           bool   `mapstructure:"NEW_RELIC_ENABLED" default:"false"`
	NewRelicAppName           string `mapstructure:"NEW_RELIC_APP_NAME" default:"columbus"`
	NewRelicLicenseKey        string `mapstructure:"NEW_RELIC_LICENSE_KEY" default:""`
	LogLevel                  string `mapstructure:"LOG_LEVEL" default:"info"`
}

var config Config

// LoadConfig returns application configuration
func LoadConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("config file was not found. Env vars and defaults will be used")
		} else {
			return config, err
		}
	}

	defaults.SetDefaults(&config)

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("unable to unmarshal config to struct: %v\n", err)
	}

	return config, nil
}
