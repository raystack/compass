package cli

import (
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/compass/internal/client"
	"github.com/odpf/compass/internal/server"
	esStore "github.com/odpf/compass/internal/store/elasticsearch"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/compass/pkg/metrics"
	"github.com/odpf/compass/pkg/statsd"
	"github.com/odpf/salt/cmdx"
	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"
)

const configFlag = "config"

var cliConfig *Config

func configCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage server and client configurations",
		Example: heredoc.Doc(`
			$ compass config init
			$ compass config list`),
	}

	cmd.AddCommand(configInitCommand())
	cmd.AddCommand(configListCommand())

	return cmd
}

func configInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new sevrer and client configuration",
		Example: heredoc.Doc(`
			$ compass config init
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cmdx.SetConfig("compass")

			if err := cfg.Init(&Config{}); err != nil {
				return err
			}

			fmt.Printf("config created: %v\n", cfg.File())
			return nil
		},
	}
}

func configListCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List server and client configuration settings",
		Example: heredoc.Doc(`
			$ compass config list
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cmdx.SetConfig("compass")

			data, err := cfg.Read()
			if err != nil {
				return err
			}

			fmt.Println(data)
			return nil
		},
	}
	return cmd
}

type Config struct {
	// Log
	LogLevel string `yaml:"log_level" mapstructure:"log_level" default:"info"`

	// StatsD
	StatsD statsd.Config `mapstructure:"statsd"`

	// NewRelic
	NewRelic metrics.NewRelicConfig `mapstructure:"newrelic"`

	// Elasticsearch
	Elasticsearch esStore.Config `mapstructure:"elasticsearch"`

	// Database
	DB postgres.Config `mapstructure:"db"`

	// Service
	Service server.Config `mapstructure:"service"`

	// Client
	Client client.Config `mapstructure:"client"`
}

func LoadConfig() (*Config, error) {
	var config Config

	cfg := cmdx.SetConfig("compass")
	err := cfg.Load(&config)

	return &config, err
}

func initConfigFromFlag(configFile string) error {
	loader := config.NewLoader(config.WithFile(configFile))

	if err := loader.Load(cliConfig); err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			fmt.Println(err)
			return nil
		}
		return err
	}

	return nil
}
