package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/goto/compass/internal/client"
	"github.com/goto/compass/internal/server"
	esStore "github.com/goto/compass/internal/store/elasticsearch"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/pkg/statsd"
	"github.com/goto/compass/pkg/telemetry"
	"github.com/goto/salt/cmdx"
	"github.com/goto/salt/config"
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
	cmd := &cobra.Command{
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

	// OpenTelemetry and Newrelic
	Telemetry telemetry.Config `mapstructure:"telemetry"`

	// StatsD
	StatsD statsd.Config `mapstructure:"statsd"`

	// Deprecated: Use Config.Telemetry instead
	NewRelic telemetry.NewRelicConfig `mapstructure:"newrelic"`

	// Elasticsearch
	Elasticsearch esStore.Config `mapstructure:"elasticsearch"`

	// Database
	DB postgres.Config `mapstructure:"db"`

	// Service
	Service server.Config `mapstructure:"service"`

	// Async worker
	Worker workermanager.Config `mapstructure:"worker"`

	// Client
	Client client.Config `mapstructure:"client"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	defer func() {
		if cfg.NewRelic != (telemetry.NewRelicConfig{}) && cfg.Telemetry.NewRelic == (telemetry.NewRelicConfig{}) {
			cfg.Telemetry.NewRelic = cfg.NewRelic
		}
	}()

	err := LoadFromCurrentDir(&cfg)

	if errors.As(err, &config.ConfigFileNotFoundError{}) {
		err := cmdx.SetConfig("compass").
			Load(&cfg, cmdx.WithLoaderOptions(
				config.WithEnvKeyReplacer(".", "_"),
				config.WithEnvPrefix("COMPASS"),
			))
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
