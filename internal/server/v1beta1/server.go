package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/odpf/compass/core/user"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type APIServer struct {
	compassv1beta1.UnimplementedCompassServiceServer
	namespaceService   NamespaceService
	assetService       AssetService
	starService        StarService
	discussionService  DiscussionService
	tagService         TagService
	tagTemplateService TagTemplateService
	userService        UserService
	logger             log.Logger
	statsDReporter     StatsDClient
}

var (
	errMissingUserInfo = errors.New("missing user information")
)

func NewAPIServer(
	logger log.Logger,
	namespaceService NamespaceService,
	assetService AssetService,
	starService StarService,
	discussionService DiscussionService,
	tagService TagService,
	tagTemplateService TagTemplateService,
	userService UserService,
) *APIServer {
	return &APIServer{
		namespaceService:   namespaceService,
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
		if errors.As(err, &user.DuplicateRecordError{UUID: usr.UUID, Email: usr.Email}) {
			return "", status.Errorf(codes.AlreadyExists, err.Error())
		}
		return "", status.Errorf(codes.Internal, codes.Internal.String())
	}
	if userID == "" {
		return "", status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}
	return userID, nil
}

func (server *APIServer) sendStatsDCounterMetric(metricName string, kvTags map[string]string) {
	if server.statsDReporter != nil {
		metric := server.statsDReporter.Incr(metricName)
		for k, v := range kvTags {
			metric.Tag(k, v)
		}
		metric.Publish()
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
