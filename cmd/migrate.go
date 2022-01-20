package cmd

import (
	"context"
	"fmt"
	"github.com/odpf/salt/log"
	"time"

	"github.com/odpf/columbus/record"
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

func migratePostgres(log log.Logger) (err error) {
	log.Info("Initiating Postgres client...")

	pgConfig := postgres.Config{
		Host:     config.ServerHost,
		Port:     config.DBPort,
		Name:     config.DBName,
		User:     config.DBUser,
		Password: config.DBPassword,
		SSLMode:  config.DBSSLMode,
	}

	pgClient, err := postgres.NewClient(log, pgConfig)
	if err != nil {
		log.Errorf("failed to prepare migration: %s", err)
		return err
	}

	err = pgClient.Migrate(pgConfig)
	if err != nil {
		return errors.Wrap(err, "problem with migration")

	}

	return nil
}

func migrateElasticsearch(log log.Logger) (err error) {
	log.Info("Initiating ES client...")
	esClient := initElasticsearch(config, log)
	for _, supportedTypeName := range record.AllSupportedTypes {
		log.Info("Migrating type\n", "type", supportedTypeName)
		ctx, cancel := context.WithTimeout(context.Background(), esMigrationTimeout)
		defer cancel()
		err = esStore.Migrate(ctx, esClient, supportedTypeName)
		if err != nil {
			err = errors.Wrapf(err, "error creating/replacing type: %q", supportedTypeName)
			return
		}
		log.Info("created/updated type\n", "type", supportedTypeName)
	}
	return
}
