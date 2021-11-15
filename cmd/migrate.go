package cmd

import (
	"context"

	esStore "github.com/odpf/columbus/discovery/elasticsearch"
)

func Migrate() {
	if err := loadConfig(); err != nil {
		log.Fatal(err)
	}

	rootLogger := initLogger(config.LogLevel)
	log = rootLogger.WithField("reporter", "migrate")

	log.Info("starting elasticsearch migration")
	esClient := initElasticsearch(config)
	err := esStore.Migrate(context.Background(), esClient)
	if err != nil {
		log.Fatal("error when migrating elasticsearch", "err", err)
	}
	log.Info("elasticsearch migration succeeded")
}
