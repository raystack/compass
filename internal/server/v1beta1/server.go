package handlersv1beta1

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// type Dependencies struct {
// 	AssetRepository      asset.Repository
// 	TagService           *tag.Service
// 	TagTemplateService   *tag.TemplateService
// 	UserService          *user.Service
// 	StarRepository       star.Repository
// 	LineageRepository    lineage.Repository
// 	DiscussionRepository discussion.Repository
// 	DiscoveryRepository  discovery.Repository
// }

type APIServer struct {
	compassv1beta1.UnimplementedCompassServiceServer
	assetService       AssetService
	starService        StarService
	discussionService  DiscussionService
	tagService         TagService
	tagTemplateService TagTemplateService
	UserService        UserService
	// assetRepository      asset.Repository
	// tagService           *tag.Service
	// tagTemplateService   *tag.TemplateService
	// userService          *user.Service
	// starRepository       star.Repository
	// lineageRepository    lineage.Repository
	// discussionRepository discussion.Repository
	// discoveryRepository  discovery.Repository
	logger log.Logger
}

var (
	errMissingUserInfo = errors.New("missing user information")
)

func NewAPIServer(
	logger log.Logger,
	assetService AssetService,
	starService StarService,
	discussionService DiscussionService,
	tagService TagService,
	tagTemplateService TagTemplateService,
	userService UserService,
) *APIServer {
	return &APIServer{
		assetService:       assetService,
		starService:        starService,
		discussionService:  discussionService,
		tagService:         tagService,
		tagTemplateService: tagTemplateService,
		// assetRepository:      deps.AssetRepository,
		// tagService:           deps.TagService,
		// tagTemplateService:   deps.TagTemplateService,
		// userService:          deps.UserService,
		// starRepository:       deps.StarRepository,
		// lineageRepository:    deps.LineageRepository,
		// discussionRepository: deps.DiscussionRepository,
		// discoveryRepository:  deps.DiscoveryRepository,
		logger: logger,
	}
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

func bodyParserErrorMsg(err error) string {
	return fmt.Sprintf("error parsing request body: %v", err)
}
