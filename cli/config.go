package cli

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/internal/config"
	saltconfig "github.com/raystack/salt/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const configFlag = "config"

func configCommand(cfg *config.Config) *cobra.Command {
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
		Short: "Initialize a new server and client configuration",
		Example: heredoc.Doc(`
			$ compass config init
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := saltconfig.NewLoader(saltconfig.WithAppConfig("compass"))

			if err := loader.Init(&config.Config{}); err != nil {
				return err
			}

			fmt.Println("config created")
			return nil
		},
	}
}

func configListCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
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
}
