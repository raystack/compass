package cli

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/spf13/cobra"
)

func cmdMigrate() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run storage migration",
		Example: heredoc.Doc(`
			$ compass migrate
		`),
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			return runMigrations(ctx, cfg)
		},
	}
}

func runMigrations(ctx context.Context, config Config) error {
	fmt.Println("Preparing migration...")

	logger := initLogger(config.LogLevel)
	logger.Info("compass is migrating", "version", Version)

	logger.Info("Migrating Postgres...")
	if err := migratePostgres(logger, config); err != nil {
		return err
	}
	logger.Info("Migration Postgres done.")

	return nil
}

func migratePostgres(logger log.Logger, config Config) (err error) {
	logger.Info("Initiating Postgres client...")

	pgClient, err := postgres.NewClient(config.DB)
	if err != nil {
		logger.Error("failed to prepare migration", "error", err)
		return err
	}

	err = pgClient.Migrate(config.DB)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}

	return nil
}
