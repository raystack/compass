package handlersv1beta1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	statsDReporter     StatsDClient
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
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				jBytes, err := json.Marshal(md)
				if err != nil {
					server.logger.Debug("unable to marshal headers", "err", err)
				} else {
					server.logger.Debug("printing headers", "headers", string(jBytes))
				}
			} else {
				server.logger.Debug("could not get metadata")
			}

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
