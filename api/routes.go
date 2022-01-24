package api

import (
	"net/http"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/tag"
)

type Config struct {
	Logger                  log.Logger
	TagService              *tag.Service
	TagTemplateService      *tag.TemplateService
	TypeRepository          discovery.TypeRepository
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
		config.Logger,
		config.TypeRepository,
	)

	recordHandler := handlers.NewRecordHandler(
		config.Logger,
		config.TypeRepository,
		config.DiscoveryService,
		config.RecordRepositoryFactory,
	)
	searchHandler := handlers.NewSearchHandler(
		config.Logger,
		config.DiscoveryService,
	)
	lineageHandler := handlers.NewLineageHandler(
		config.Logger,
		config.LineageProvider,
	)
	tagHandler := handlers.NewTagHandler(
		config.Logger,
		config.TagService,
	)
	tagTemplateHandler := handlers.NewTagTemplateHandler(
		config.Logger,
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

	router.NotFoundHandler = http.HandlerFunc(handlers.NotFound)
	router.MethodNotAllowedHandler = http.HandlerFunc(handlers.MethodNotAllowed)
}
