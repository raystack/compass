package cmd

import (
	"context"
	"fmt"
	"github.com/odpf/salt/log"
	"time"

	"github.com/odpf/columbus/record"
	esStore "github.com/odpf/columbus/store/elasticsearch"
	"github.com/odpf/columbus/store/postgres"
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
	logger.Info("columbus is migrating", "version", Version)

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

func migrateElasticsearch(logger log.Logger) (err error) {
	logger.Info("Initiating ES client...")
	esClient := initElasticsearch(config, logger)
	for _, supportedTypeName := range record.AllSupportedTypes {
		logger.Info("Migrating type\n", "type", supportedTypeName)
		ctx, cancel := context.WithTimeout(context.Background(), esMigrationTimeout)
		defer cancel()
		err = esStore.Migrate(ctx, esClient, supportedTypeName)
		if err != nil {
			err = fmt.Errorf("error creating/replacing type %q: %w", supportedTypeName, err)
			return
		}
		logger.Info("created/updated type\n", "type", supportedTypeName)
	}
	return
}
