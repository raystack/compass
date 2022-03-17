package v1beta1

import (
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
)

type Handler struct {
	compassv1beta1.UnimplementedCompassServiceServer
	Logger               log.Logger
	AssetRepository      asset.Repository
	DiscoveryRepository  discovery.Repository
	TagService           *tag.Service
	TagTemplateService   *tag.TemplateService
	UserService          *user.Service
	StarRepository       star.Repository
	LineageRepository    lineage.Repository
	DiscussionRepository discussion.Repository
}
