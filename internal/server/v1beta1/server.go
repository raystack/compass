package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type APIServer struct {
	compassv1beta1.UnimplementedCompassServiceServer
	assetService       AssetService
	starService        StarService
	discussionService  DiscussionService
	tagService         TagService
	tagTemplateService TagTemplateService
	userService        UserService
	logger             log.Logger
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
		userService:        userService,
		logger:             logger,
	}
}

func (server *APIServer) validateUserInCtx(ctx context.Context) (string, error) {
	usr := user.FromContext(ctx)
	userID, err := server.userService.ValidateUser(ctx, usr.UUID, usr.Email)
	if err != nil {
		if errors.Is(err, user.ErrNoUserInformation) {
			return "", status.Errorf(codes.InvalidArgument, err.Error())
		}
		return "", status.Errorf(codes.Internal, codes.Internal.String())
	}
	if userID == "" {
		return "", status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}
	return userID, nil
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
