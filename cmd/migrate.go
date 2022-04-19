package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/salt/log"

	"github.com/odpf/compass/asset"
	esStore "github.com/odpf/compass/store/elasticsearch"
	"github.com/odpf/compass/store/postgres"
)

const (
	esMigrationTimeout = 5 * time.Second
)

func RunMigrate() {
	fmt.Println("Preparing migration...")
	if err := loadConfig(); err != nil {
		panic(err)
	}

	logger := initLogger(config.LogLevel)
	logger.Info("compass is migrating", "version", Version)

	logger.Info("Migrating Postgres...")
	if err := migratePostgres(logger); err != nil {
		panic(err)
	}
	logger.Info("Migration Postgres done.")

	logger.Info("Migrating ES...")
	if err := migrateElasticsearch(logger); err != nil {
		panic(err)
	}
	logger.Info("Migration ES done.")
}

func migratePostgres(logger log.Logger) (err error) {
	logger.Info("Initiating Postgres client...")

	pgConfig := postgres.Config{
		Host:     config.DBHost,
		Port:     config.DBPort,
		Name:     config.DBName,
		User:     config.DBUser,
		Password: config.DBPassword,
		SSLMode:  config.DBSSLMode,
	}

	pgClient, err := postgres.NewClient(pgConfig)
	if err != nil {
		logger.Error("failed to prepare migration", "error", err)
		return err
	}

	err = pgClient.Migrate(pgConfig)
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}

	return nil
}

func migrateElasticsearch(logger log.Logger) error {
	logger.Info("Initiating ES client...")
	esClient := initElasticsearch(config, logger)
	for _, supportedType := range asset.AllSupportedTypes {
		logger.Info("Migrating type\n", "type", supportedType)
		ctx, cancel := context.WithTimeout(context.Background(), esMigrationTimeout)
		defer cancel()
		err := esStore.Migrate(ctx, esClient, supportedType)
		if err != nil {
			return fmt.Errorf("error creating/replacing type %q: %w", supportedType, err)
		}
		logger.Info("created/updated type\n", "type", supportedType)
	}
	return nil
}
