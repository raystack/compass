package cli

import (
	"errors"
	"fmt"
	"os"

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
		// SilencePersistentFlag: true,
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
			_ = yaml.NewEncoder(os.Stdout).Encode(*cfg)
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
	var cfg Config
	err := cmdx.SetConfig("compass").Load(&cfg)
	if err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			return LoadFromCurrentDir()
		}
		return &cfg, err
	}
	return &cfg, nil
}

func LoadFromCurrentDir() (*Config, error) {
	var cfg Config
	var opts []config.LoaderOption

	opts = append(opts,
		config.WithPath("./"),
		config.WithName("compass.yaml"),
		config.WithEnvKeyReplacer(".", "_"),
		config.WithEnvPrefix("COMPASS"),
	)

	if err := config.NewLoader(opts...).Load(&cfg); err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			return &cfg, ErrConfigNotFound
		}
		return &cfg, err
	}
	return &cfg, nil
}

func LoadConfigFromFlag(cfgFile string, cfg *Config) error {
	var opts []config.LoaderOption
	opts = append(opts, config.WithFile(cfgFile))

	return config.NewLoader(opts...).Load(cfg)
}
