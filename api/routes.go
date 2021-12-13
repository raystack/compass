package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/record"
	"github.com/odpf/columbus/tag"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Logger                  logrus.FieldLogger
	TagService              *tag.Service
	TagTemplateService      *tag.TemplateService
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
	tagHandler := handlers.NewTagHandler(
		config.Logger.WithField("reporter", "tag-handler"),
		config.TagService,
	)
	tagTemplateHandler := handlers.NewTagTemplateHandler(
		config.Logger.WithField("reporter", "tag-template-handler"),
		config.TagTemplateService,
	)

	router.PathPrefix("/ping").Handler(handlers.NewHeartbeatHandler())
	setupV1TypeRoutes(router, typeHandler, recordHandler)
	setupV1TagRoutes(router, "/v1/tags", tagHandler, tagTemplateHandler)

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
		Methods(http.MethodPut, http.MethodHead).
		HandlerFunc(rh.UpsertBulk)

	router.Path(recordURL).
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(rh.GetByType)

	router.Path(recordURL+"/{id}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(rh.GetOneByType)

	router.Path(recordURL+"/{id}").
		Methods(http.MethodDelete, http.MethodHead).
		HandlerFunc(rh.Delete)

}

func setupV1TagRoutes(router *mux.Router, baseURL string, th *handlers.TagHandler, tth *handlers.TagTemplateHandler) {
	router.Methods(http.MethodPost).Path(baseURL).HandlerFunc(th.Create)

	url := baseURL + "/types/{type}/records/{record_urn}/templates/{template_urn}"
	router.Methods(http.MethodGet).Path(url).HandlerFunc(th.FindByRecordAndTemplate)
	router.Methods(http.MethodPut).Path(url).HandlerFunc(th.Update)
	router.Methods(http.MethodDelete).Path(url).HandlerFunc(th.Delete)

	router.Methods(http.MethodGet).Path(baseURL + "/types/{type}/records/{record_urn}").HandlerFunc(th.GetByRecord)

	templateURL := baseURL + "/templates"
	router.Methods(http.MethodGet).Path(templateURL).HandlerFunc(tth.Index)
	router.Methods(http.MethodPost).Path(templateURL).HandlerFunc(tth.Create)
	router.Methods(http.MethodGet).Path(templateURL + "/{template_urn}").HandlerFunc(tth.Find)
	router.Methods(http.MethodPut).Path(templateURL + "/{template_urn}").HandlerFunc(tth.Update)
	router.Methods(http.MethodDelete).Path(templateURL + "/{template_urn}").HandlerFunc(tth.Delete)

}
