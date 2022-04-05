package v1beta1

import (
	"fmt"
	"net/http"
	"time"

	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
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
