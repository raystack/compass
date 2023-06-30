package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/internal/client"
	"github.com/raystack/compass/internal/server"
	esStore "github.com/raystack/compass/internal/store/elasticsearch"
	"github.com/raystack/compass/internal/store/postgres"
	"github.com/raystack/compass/pkg/metrics"
	"github.com/raystack/compass/pkg/statsd"
	"github.com/raystack/salt/cmdx"
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
			cfg := cmdx.SetConfig("compass")

			if err := cfg.Init(&Config{}); err != nil {
				return err
			}

			fmt.Printf("config created: %v\n", cfg.File())
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
	var cfg Config
	err := LoadFromCurrentDir(&cfg)

	if errors.As(err, &config.ConfigFileNotFoundError{}) {
		err := cmdx.SetConfig("compass").Load(&cfg)
		if err != nil {
			if errors.As(err, &config.ConfigFileNotFoundError{}) {
				return &cfg, ErrConfigNotFound
			}
			return &cfg, err
		}
	}
	return &cfg, nil
}

func LoadFromCurrentDir(cfg *Config) error {
	var opts []config.LoaderOption
	opts = append(opts,
		config.WithPath("./"),
		config.WithFile("compass.yaml"),
		config.WithEnvKeyReplacer(".", "_"),
		config.WithEnvPrefix("COMPASS"),
	)

	return config.NewLoader(opts...).Load(cfg)
}

func LoadConfigFromFlag(cfgFile string, cfg *Config) error {
	var opts []config.LoaderOption
	opts = append(opts, config.WithFile(cfgFile))

	return config.NewLoader(opts...).Load(cfg)
}
