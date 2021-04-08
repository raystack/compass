package api

import (
	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/models"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Logger                  logrus.FieldLogger
	TypeRepository          models.TypeRepository
	RecordRepositoryFactory models.RecordRepositoryFactory
	RecordSearcher          models.RecordSearcher
	LineageProvider         handlers.LineageProvider
}

func RegisterRoutes(router *mux.Router, config Config) {
	typeHandler := handlers.NewTypeHandler(
		config.Logger.WithField("reporter", "type-handler"),
		config.TypeRepository,
		config.RecordRepositoryFactory,
	)
	searchHandler := handlers.NewSearchHandler(
		config.Logger.WithField("reporter", "search-handler"),
		config.RecordSearcher,
		config.TypeRepository,
	)
	lineageHandler := handlers.NewLineageHandler(
		config.Logger.WithField("reporter", "lineage-handler"),
		config.LineageProvider,
	)

	router.PathPrefix("/ping").Handler(handlers.NewHeartbeatHandler())
	router.PathPrefix("/v1/types").Handler(typeHandler)
	router.PathPrefix("/v1/entities").Handler(typeHandler) // For backward compatibility
	router.PathPrefix("/v1/search").Handler(searchHandler)
	router.PathPrefix("/v1/lineage").Handler(lineageHandler)
}
