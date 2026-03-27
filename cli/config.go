package cli

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/internal/client"
	"github.com/raystack/compass/internal/server"
	esStore "github.com/raystack/compass/internal/store/elasticsearch"
	"github.com/raystack/compass/internal/store/postgres"
	"github.com/raystack/compass/pkg/metrics"
	"github.com/raystack/compass/pkg/telemetry"
	"github.com/raystack/salt/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const configFlag = "config"

func configCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage server and client configurations",
		Example: heredoc.Doc(`
			$ compass config init
			$ compass config list`),
	}

	cmd.AddCommand(configInitCommand())
	cmd.AddCommand(configListCommand(cfg))

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
			loader := config.NewLoader(config.WithAppConfig("compass"))

			if err := loader.Init(&Config{}); err != nil {
				return err
			}

			fmt.Println("config created")
			return nil
		},
	}
}

func configListCommand(cfg *Config) *cobra.Command {
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
			return yaml.NewEncoder(os.Stdout).Encode(*cfg)
		},
	}
	return cmd
}

type Config struct {
	// Log
	LogLevel string `yaml:"log_level" mapstructure:"log_level" default:"info"`

	// NewRelic
	NewRelic metrics.NewRelicConfig `mapstructure:"newrelic"`

	// Telemetry
	Telemetry telemetry.Config `mapstructure:"telemetry"`

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
	var cfg Config

	// Try loading from current directory first
	err := LoadFromCurrentDir(&cfg)
	if err != nil {
		// Fall back to app config directory
		loader := config.NewLoader(config.WithAppConfig("compass"))
		if loadErr := loader.Load(&cfg); loadErr != nil {
			return &cfg, ErrConfigNotFound
		}
	}
	return &cfg, nil
}

func LoadFromCurrentDir(cfg *Config) error {
	return config.NewLoader(
		config.WithFile("./config.yaml"),
		config.WithEnvPrefix("COMPASS"),
	).Load(cfg)
}

func LoadConfigFromFlag(cfgFile string, cfg *Config) error {
	return config.NewLoader(config.WithFile(cfgFile)).Load(cfg)
}
