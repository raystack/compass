package api

import (
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

type Handlers struct {
	Type        *handlers.TypeHandler
	Record      *handlers.RecordHandler
	Search      *handlers.SearchHandler
	Lineage     *handlers.LineageHandler
	Tag         *handlers.TagHandler
	TagTemplate *handlers.TagTemplateHandler
}

func initHandlers(config Config) *Handlers {
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

	return &Handlers{
		Type:        typeHandler,
		Record:      recordHandler,
		Search:      searchHandler,
		Lineage:     lineageHandler,
		Tag:         tagHandler,
		TagTemplate: tagTemplateHandler,
	}
}

func RegisterRoutes(router *mux.Router, config Config) {
	// By default mux will decode url and then match the decoded url against the route
	// we reverse the steps by telling mux to use encoded path to match the url
	// then we manually decode via custom middleware (decodeURLMiddleware).
	//
	// This is to allow urn that has "/" to be matched correctly to the route
	router.UseEncodedPath()
	router.Use(decodeURLMiddleware(config.Logger))

	handlerCollection := initHandlers(config)

	router.PathPrefix("/ping").Handler(handlers.NewHeartbeatHandler())

	v1Beta1SubRouter := router.PathPrefix("/v1beta1").Subrouter()
	setupV1Beta1Router(v1Beta1SubRouter, handlerCollection)

	v1SubRouter := router.PathPrefix("/v1").Subrouter()
	setupV1Beta1Router(v1SubRouter, handlerCollection)
}
