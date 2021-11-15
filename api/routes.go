package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/discovery"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Logger                  logrus.FieldLogger
	RecordRepositoryFactory discovery.RecordRepositoryFactory
	DiscoveryService        *discovery.Service
	LineageProvider         handlers.LineageProvider
}

func RegisterRoutes(router *mux.Router, config Config) {
	// By default mux will decode url and then match the decoded url against the route
	// we reverse the steps by telling mux to use encoded path to match the url
	// then we manually decode via custom middleware (decodeURLMiddleware).
	//
	// This is to allow urn that has "/" to be matched correctly to the route
	router.UseEncodedPath()
	router.Use(decodeURLMiddleware(config.Logger))

	typeHandler := handlers.NewRecordHandler(
		config.Logger.WithField("reporter", "type-handler"),
		config.DiscoveryService,
		config.RecordRepositoryFactory,
	)
	searchHandler := handlers.NewSearchHandler(
		config.Logger.WithField("reporter", "search-handler"),
		config.DiscoveryService,
	)

	lineageHandler := handlers.NewLineageHandler(
		config.Logger.WithField("reporter", "lineage-handler"),
		config.LineageProvider,
	)

	router.PathPrefix("/ping").Handler(handlers.NewHeartbeatHandler())
	setupV1TypeRoutes(router, typeHandler)

	router.Path("/v1/search").
		Methods(http.MethodGet).
		HandlerFunc(searchHandler.Search)

	router.PathPrefix("/v1/lineage/{type}/{id}").
		Methods(http.MethodGet).
		HandlerFunc(lineageHandler.GetLineage)

	router.PathPrefix("/v1/lineage").
		Methods(http.MethodGet).
		HandlerFunc(lineageHandler.ListLineage)
}

func setupV1TypeRoutes(router *mux.Router, typeHandler *handlers.RecordHandler) {
	baseURL := "/v1/types"
	router.Path(baseURL+"/{name}/records").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(typeHandler.GetByType)

	router.Path(baseURL + "/{name}/records/{id}").
		Methods(http.MethodDelete).
		HandlerFunc(typeHandler.Delete)

	router.Path(baseURL + "/{name}/records").
		Methods(http.MethodPut).
		HandlerFunc(typeHandler.UpsertBulk)

	router.Path(baseURL+"/{name}/records/{id}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(typeHandler.GetOneByType)
}
