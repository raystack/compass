package cli

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/spf13/cobra"
)

const (
	esMigrationTimeout = 5 * time.Second
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

	logger.Info("Migrating ES...")
	if err := migrateElasticsearch(logger, config); err != nil {
		return err
	}
	logger.Info("Migration ES done.")
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

func migrateElasticsearch(logger log.Logger, config Config) error {
	logger.Info("Initiating ES client...")
	esClient, err := initElasticsearch(logger, config.Elasticsearch)
	if err != nil {
		return err
	}

	for _, supportedType := range asset.AllSupportedTypes {
		logger.Info("Migrating type\n", "type", supportedType)
		ctx, cancel := context.WithTimeout(context.Background(), esMigrationTimeout)
		defer cancel()
		err := esClient.Migrate(ctx, supportedType)
		if err != nil {
			return fmt.Errorf("error creating/replacing type %q: %w", supportedType, err)
		}
		logger.Info("created/updated type\n", "type", supportedType)
	}
	return nil
}
