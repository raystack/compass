package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/columbus/record"
	esStore "github.com/odpf/columbus/store/elasticsearch"
	"github.com/odpf/columbus/store/postgres"
)

const esMigrationTimeout = 5 * time.Second

func Migrate() {
	fmt.Println("Preparing migration...")
	if err := loadConfig(); err != nil {
		panic(err)
	}

	rootLogger := initLogger(config.LogLevel)
	log = rootLogger.WithField("reporter", "main")
	log.Infof("columbus %s is migrating", Version)

	log.Info("Migrating Postgres...")
	if err := migratePostgres(); err != nil {
		panic(err)
	}
	log.Info("Migration Postgres done.")

	log.Info("Migrating ES...")
	if err := migrateElasticsearch(); err != nil {
		panic(err)
	}
	log.Info("Migration ES done.")
}

func migratePostgres() (err error) {
	log.Info("Initiating Postgres client...")
	pgClient, err := postgres.NewClient(postgres.Config{
		Port:     config.DBPort,
		Host:     config.DBHost,
		Name:     config.DBName,
		User:     config.DBUser,
		Password: config.DBPassword,
		SSLMode:  config.DBSSLMode,
	})
	if err != nil {
		return err
	}

	log.Info("Migrating DB...")
	if err := pgClient.AutoMigrate(
		&postgres.Template{},
		&postgres.Field{},
		&postgres.Tag{},
	); err != nil {
		return err
	}

	return nil
}

func migrateElasticsearch() (err error) {
	log.Info("Initiating ES client...")
	esClient := initElasticsearch(config)
	tr := esStore.NewTypeRepository(esClient)
	for _, supportedType := range record.AllSupportedTypes {
		log.Infof("Migrating %q type\n", supportedType.Name)
		ctx, cancel := context.WithTimeout(context.Background(), esMigrationTimeout)
		defer cancel()
		err = tr.CreateOrReplace(ctx, supportedType)
		if err != nil {
			err = fmt.Errorf("error creating/replacing type: %q, err: %w", supportedType.Name, err)
			return
		}
		log.Infof("created/updated %q type\n", supportedType.Name)
	}
	return
}
