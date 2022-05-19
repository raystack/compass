package cli

import (
	"fmt"
	"os"

	"github.com/odpf/compass/internal/server"
	esStore "github.com/odpf/compass/internal/store/elasticsearch"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/compass/pkg/metrics"
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
	StatsD metrics.StatsdConfig `mapstructure:"statsd"`

	// NewRelic
	NewRelic metrics.NewRelicConfig `mapstructure:"newrelic"`

	// Elasticsearch
	Elasticsearch esStore.Config `mapstructure:"elasticsearch"`

	// Database
	DB postgres.Config `mapstructure:"db"`

	// Service
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
