package api

import (
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/compass/api/httpapi"
	"github.com/odpf/compass/api/httpapi/handlers"
	"github.com/odpf/compass/api/httpapi/middleware"
	"github.com/odpf/compass/api/v1beta1"
	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/discovery"
	"github.com/odpf/compass/discussion"
	"github.com/odpf/compass/lineage"
	"github.com/odpf/compass/metrics"
	"github.com/odpf/compass/star"
	"github.com/odpf/compass/tag"
	"github.com/odpf/compass/user"
	"github.com/odpf/salt/log"
)

type Dependencies struct {
	Logger               log.Logger
	NRApp                *newrelic.Application
	StatsdMonitor        *metrics.StatsdMonitor
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

	recordHandler := handlers.NewRecordHandler(
		deps.Logger,
		deps.TypeRepository,
		deps.DiscoveryService,
		deps.RecordRepositoryFactory,
	)

	return &httpapi.Handler{
		Record: recordHandler,
	}
}

func NewGRPCHandler(l log.Logger, deps *Dependencies) *v1beta1.Handler {
	return &v1beta1.Handler{
		Logger:               l,
		DiscussionRepository: deps.DiscussionRepository,
		AssetRepository:      deps.AssetRepository,
		LineageRepository:    deps.LineageRepository,
		StarRepository:       deps.StarRepository,
		UserService:          deps.UserService,
		TagService:           deps.TagService,
		TagTemplateService:   deps.TagTemplateService,
		DiscoveryRepository:  deps.DiscoveryRepository,

		//deprecated
		TypeRepository:   deps.TypeRepository,
		DiscoveryService: deps.DiscoveryService,
	}
}

func RegisterHTTPRoutes(cfg Config, mux *runtime.ServeMux, deps *Dependencies, handlerCollection *httpapi.Handler) error {
	if err := mux.HandlePath(http.MethodGet, "/ping", runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		fmt.Fprintf(w, "pong")
	})); err != nil {
		return err
	}

	if err := mux.HandlePath(http.MethodGet, "/v1beta1/types/{name}/records",
		middleware.NewRelic(deps.NRApp, http.MethodGet, "/v1beta1/types/{name}/records",
			middleware.StatsD(deps.StatsdMonitor,
				middleware.ValidateUser(cfg.IdentityUUIDHeaderKey, cfg.IdentityEmailHeaderKey, deps.UserService, handlerCollection.Record.GetByType)))); err != nil {
		return err
	}

	if err := mux.HandlePath(http.MethodGet, "/v1beta1/types/{name}/records/{id}",
		middleware.NewRelic(deps.NRApp, http.MethodGet, "/v1beta1/types/{name}/records/{id}",
			middleware.StatsD(deps.StatsdMonitor,
				middleware.ValidateUser(cfg.IdentityUUIDHeaderKey, cfg.IdentityEmailHeaderKey, deps.UserService, handlerCollection.Record.GetOneByType)))); err != nil {
		return err
	}

	return nil
}
