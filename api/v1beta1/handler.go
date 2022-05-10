package v1beta1

import (
	"fmt"
	"net/http"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/discovery"
	"github.com/odpf/compass/discussion"
	"github.com/odpf/compass/lineage"
	"github.com/odpf/compass/star"
	"github.com/odpf/compass/tag"
	"github.com/odpf/compass/user"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Handler struct {
	compassv1beta1.UnimplementedCompassServiceServer
	Logger               log.Logger
	AssetRepository      asset.Repository
	TagService           *tag.Service
	TagTemplateService   *tag.TemplateService
	UserService          *user.Service
	StarRepository       star.Repository
	LineageRepository    lineage.Repository
	DiscussionRepository discussion.Repository
	DiscoveryRepository  discovery.Repository
	HealthServer         grpc_health_v1.HealthServer

	// deprecated
	TypeRepository   discovery.TypeRepository
	DiscoveryService *discovery.Service
}

func internalServerError(logger log.Logger, msg string) error {
	ref := time.Now().Unix()

	logger.Error(msg, "ref", ref)
	return status.Error(codes.Internal, fmt.Sprintf(
		"%s - ref (%d)",
		http.StatusText(http.StatusInternalServerError),
		ref,
	))
}
