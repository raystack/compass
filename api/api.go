package api

import (
	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/httpapi"
	"github.com/odpf/columbus/api/httpapi/handlers"
	"github.com/odpf/columbus/api/httpapi/middleware"
	"github.com/odpf/columbus/api/v1beta1"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
)

type Dependencies struct {
	Logger               log.Logger
	AssetRepository      asset.Repository
	DiscoveryRepository  discovery.Repository
	TagService           *tag.Service
	TagTemplateService   *tag.TemplateService
	UserService          *user.Service
	StarRepository       star.Repository
	LineageRepository    lineage.Repository
	DiscussionRepository discussion.Repository

	// Deprecated
	DiscoveryService        *discovery.Service
	TypeRepository          discovery.TypeRepository
	RecordRepositoryFactory discovery.RecordRepositoryFactory
}

type Handlers struct {
	HTTPHandler *httpapi.Handler
	GRPCHandler *v1beta1.Handler
}

func NewHandlers(logger log.Logger, deps *Dependencies) *Handlers {
	return &Handlers{
		HTTPHandler: NewHTTPHandlers(deps),
		GRPCHandler: NewGRPCHandler(logger, deps),
	}
}

func NewHTTPHandlers(deps *Dependencies) *httpapi.Handler {
	typeHandler := handlers.NewTypeHandler(
		deps.Logger,
		deps.TypeRepository,
	)

	assetHandler := handlers.NewAssetHandler(
		deps.Logger,
		deps.AssetRepository,
		deps.DiscoveryRepository,
		deps.StarRepository,
		deps.LineageRepository,
	)

	recordHandler := handlers.NewRecordHandler(
		deps.Logger,
		deps.TypeRepository,
		deps.DiscoveryService,
		deps.RecordRepositoryFactory,
	)
	searchHandler := handlers.NewSearchHandler(
		deps.Logger,
		deps.DiscoveryService,
	)
	lineageHandler := handlers.NewLineageHandler(
		deps.Logger,
		deps.LineageRepository,
	)
	tagHandler := handlers.NewTagHandler(
		deps.Logger,
		deps.TagService,
	)
	tagTemplateHandler := handlers.NewTagTemplateHandler(
		deps.Logger,
		deps.TagTemplateService,
	)
	userHandler := handlers.NewUserHandler(
		deps.Logger,
		deps.StarRepository,
		deps.DiscussionRepository,
	)

	discussionHandler := handlers.NewDiscussionHandler(
		deps.Logger,
		deps.DiscussionRepository,
	)

	return &httpapi.Handler{
		Asset:       assetHandler,
		Type:        typeHandler,
		Record:      recordHandler,
		Search:      searchHandler,
		Lineage:     lineageHandler,
		Tag:         tagHandler,
		TagTemplate: tagTemplateHandler,
		User:        userHandler,
		Discussion:  discussionHandler,
	}
}

func NewGRPCHandler(l log.Logger, deps *Dependencies) *v1beta1.Handler {
	return &v1beta1.Handler{
		Logger:               l,
		DiscussionRepository: deps.DiscussionRepository,
	}
}

func RegisterHTTPRoutes(cfg Config, router *mux.Router, deps *Dependencies, handlerCollection *httpapi.Handler) {
	// By default mux will decode url and then match the decoded url against the route
	// we reverse the steps by telling mux to use encoded path to match the url
	// then we manually decode via custom middleware (decodeURLMiddleware).
	//
	// This is to allow urn that has "/" to be matched correctly to the route
	router.UseEncodedPath()
	router.Use(middleware.DecodeURL())

	router.PathPrefix("/ping").Handler(handlers.NewHeartbeatHandler())

	v1Beta1SubRouter := router.PathPrefix("/v1beta1").Subrouter()
	v1Beta1SubRouter.Use(middleware.ValidateUser(cfg.IdentityHeaderKey, deps.UserService))

	httpapi.RegisterRoutes(v1Beta1SubRouter, handlerCollection)

	// router.NotFoundHandler = http.HandlerFunc(handlers.NotFound)
	// router.MethodNotAllowedHandler = http.HandlerFunc(handlers.MethodNotAllowed)
}
