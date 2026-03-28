package cli

import (
	"os/signal"
	"syscall"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/compass/internal/config"
	compassserver "github.com/raystack/compass/internal/server"
	"github.com/spf13/cobra"
)

// Version of the current build. overridden by the build system.
// see "Makefile" for more information
var (
	Version string
)

func serverCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server <command>",
		Aliases: []string{"s"},
		Short:   "Run compass server",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ compass server start
			$ compass server start -c ./config.yaml
			$ compass server migrate
			$ compass server migrate -c ./config.yaml
		`),
	}

	cmd.AddCommand(
		serverStartCommand(cfg),
		serverMigrateCommand(cfg),
	)

	return cmd
}

func serverStartCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "start",
		Short:   "Start server on default port 8080",
		Example: "compass server start",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			return compassserver.Start(ctx, cfg, Version)
		},
	}
}

func serverMigrateCommand(cfg *config.Config) *cobra.Command {
	var down bool
	c := &cobra.Command{
		Use:   "migrate",
		Short: "Run storage migration",
		Example: heredoc.Doc(`
			$ compass server migrate
			$ compass server migrate --down
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			if down {
				return compassserver.MigrateDown(ctx, cfg, Version)
			}
			return compassserver.Migrate(ctx, cfg, Version)
		},
	}
	c.Flags().BoolVar(&down, "down", false, "rollback migration one step")
	return c
}
