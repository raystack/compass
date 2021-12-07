package cmd

import (
	"fmt"

	"github.com/jeremywohl/flatten"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	ServerHost                string `mapstructure:"SERVER_HOST" default:"0.0.0.0"`
	ServerPort                string `mapstructure:"SERVER_PORT" default:"8080"`
	ElasticSearchBrokers      string `mapstructure:"ELASTICSEARCH_BROKERS" default:"http://localhost:9200"`
	StatsdAddress             string `mapstructure:"STATSD_ADDRESS" default:"127.0.0.1:8125"`
	StatsdPrefix              string `mapstructure:"STATSD_PREFIX" default:"columbusApi"`
	StatsdEnabled             bool   `mapstructure:"STATSD_ENABLED" default:"false"`
	LineageRefreshIntervalStr string `mapstructure:"LINEAGE_REFRESH_INTERVAL" default:"5m"`
	NewRelicEnabled           bool   `mapstructure:"NEW_RELIC_ENABLED" default:"false"`
	NewRelicAppName           string `mapstructure:"NEW_RELIC_APP_NAME" default:"columbus"`
	NewRelicLicenseKey        string `mapstructure:"NEW_RELIC_LICENSE_KEY" default:""`
	LogLevel                  string `mapstructure:"LOG_LEVEL" default:"info"`
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
		return fmt.Errorf("unable to unmarshal config to struct: %v\n", err)
	}

	return nil
}

func bindEnvVars() {
	err, configKeys := getFlattenedStructKeys(Config{})
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

func getFlattenedStructKeys(config Config) (error, []string) {
	var structMap map[string]interface{}
	err := mapstructure.Decode(config, &structMap)
	if err != nil {
		return err, nil
	}

	flat, err := flatten.Flatten(structMap, "", flatten.DotStyle)
	if err != nil {
		return err, nil
	}

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}

	return nil, keys
}
