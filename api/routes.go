package api

import (
	"net/http"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/api/middleware"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/comment"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/columbus/user"
)

type Config struct {
	Logger               log.Logger
	AssetRepository      asset.Repository
	DiscoveryRepository  discovery.Repository
	TagService           *tag.Service
	TagTemplateService   *tag.TemplateService
	UserService          *user.Service
	MiddlewareConfig     middleware.Config
	StarRepository       star.Repository
	LineageRepository    lineage.Repository
	DiscussionRepository discussion.Repository
	CommentRepository    comment.Repository

	// Deprecated
	DiscoveryService        *discovery.Service
	TypeRepository          discovery.TypeRepository
	RecordRepositoryFactory discovery.RecordRepositoryFactory
}

type Handlers struct {
	Asset       *handlers.AssetHandler
	Type        *handlers.TypeHandler
	Record      *handlers.RecordHandler
	Search      *handlers.SearchHandler
	Lineage     *handlers.LineageHandler
	Tag         *handlers.TagHandler
	TagTemplate *handlers.TagTemplateHandler
	User        *handlers.UserHandler
	Discussion  *handlers.DiscussionHandler
	Comment     *handlers.CommentHandler
}

func initHandlers(config Config) *Handlers {
	typeHandler := handlers.NewTypeHandler(
		config.Logger,
		config.TypeRepository,
	)

	assetHandler := handlers.NewAssetHandler(
		config.Logger,
		config.AssetRepository,
		config.DiscoveryRepository,
		config.StarRepository,
		config.LineageRepository,
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
		config.LineageRepository,
	)
	tagHandler := handlers.NewTagHandler(
		config.Logger,
		config.TagService,
	)
	tagTemplateHandler := handlers.NewTagTemplateHandler(
		config.Logger,
		config.TagTemplateService,
	)
	userHandler := handlers.NewUserHandler(
		config.Logger,
		config.StarRepository,
		config.DiscussionRepository,
	)

	discussionHandler := handlers.NewDiscussionHandler(
		config.Logger,
		config.DiscussionRepository,
	)

	return &Handlers{
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

func RegisterRoutes(router *mux.Router, config Config) {
	// By default mux will decode url and then match the decoded url against the route
	// we reverse the steps by telling mux to use encoded path to match the url
	// then we manually decode via custom middleware (decodeURLMiddleware).
	//
	// This is to allow urn that has "/" to be matched correctly to the route
	router.UseEncodedPath()
	router.Use(middleware.DecodeURL(config.MiddlewareConfig))

	handlerCollection := initHandlers(config)

	router.PathPrefix("/ping").Handler(handlers.NewHeartbeatHandler())

	v1Beta1SubRouter := router.PathPrefix("/v1beta1").Subrouter()
	v1Beta1SubRouter.Use(middleware.ValidateUser(config.MiddlewareConfig, config.UserService))
	setupV1Beta1Router(v1Beta1SubRouter, handlerCollection)

	router.NotFoundHandler = http.HandlerFunc(handlers.NotFound)
	router.MethodNotAllowedHandler = http.HandlerFunc(handlers.MethodNotAllowed)
}
