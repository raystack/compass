package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/record"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Logger                  logrus.FieldLogger
	TypeRepository          record.TypeRepository
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

	typeHandler := handlers.NewTypeHandler(
		config.Logger.WithField("reporter", "type-handler"),
		config.TypeRepository,
	)

	recordHandler := handlers.NewRecordHandler(
		config.Logger.WithField("reporter", "record-handler"),
		config.TypeRepository,
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
	setupV1TypeRoutes(router, typeHandler, recordHandler)

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

func setupV1TypeRoutes(router *mux.Router, th *handlers.TypeHandler, rh *handlers.RecordHandler) {
	typeURL := "/v1/types"

	router.Path(typeURL).
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(th.Get)

	router.Path(typeURL).
		Methods(http.MethodPut, http.MethodHead).
		HandlerFunc(th.Upsert)

	router.Path(typeURL+"/{name}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(th.Find)

	router.Path(typeURL+"/{name}").
		Methods(http.MethodDelete, http.MethodHead).
		HandlerFunc(th.Delete)

	recordURL := "/v1/types/{name}/records"
	router.Path(recordURL).
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(rh.GetByType)

	router.Path(recordURL+"/{id}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(rh.GetOneByType)

	router.Path(recordURL+"/{id}").
		Methods(http.MethodDelete, http.MethodHead).
		HandlerFunc(rh.Delete)

	router.Path(recordURL).
		Methods(http.MethodPut, http.MethodHead).
		HandlerFunc(rh.UpsertBulk)
}
