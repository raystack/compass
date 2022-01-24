package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/salt/log"

	"github.com/odpf/columbus/asset"
	esStore "github.com/odpf/columbus/store/elasticsearch"
	"github.com/odpf/columbus/store/postgres"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "problem with migration")

	}

	return nil
}

func migrateElasticsearch(logger log.Logger) (err error) {
	logger.Info("Initiating ES client...")
	esClient := initElasticsearch(config, logger)
	for _, supportedType := range asset.AllSupportedTypes {
		logger.Info("Migrating type\n", "type", supportedType)
		ctx, cancel := context.WithTimeout(context.Background(), esMigrationTimeout)
		defer cancel()
		err = esStore.Migrate(ctx, esClient, supportedType)
		if err != nil {
			err = errors.Wrapf(err, "error creating/replacing type: %q", supportedType)
			return
		}
		logger.Info("created/updated type\n", "type", supportedType)
	}
	return
}
