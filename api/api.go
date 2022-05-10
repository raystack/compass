package api

import (
	"github.com/newrelic/go-agent/v3/newrelic"
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
	DiscoveryService                *discovery.Service
	TypeRepository                  discovery.TypeRepository
	DiscoveryAssetRepositoryFactory discovery.AssetRepositoryFactory
}

type Handlers struct {
	HealthHandler *HealthHandler
	GRPCHandler   *v1beta1.Handler
}

func NewHandlers(logger log.Logger, deps *Dependencies) *Handlers {
	return &Handlers{
		HealthHandler: &HealthHandler{},
		GRPCHandler:   NewGRPCHandler(logger, deps),
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

		// deprecated
		TypeRepository:   deps.TypeRepository,
		DiscoveryService: deps.DiscoveryService,
	}
}
