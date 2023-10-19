package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goto/compass/core/user"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
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

	assetUpdateCounter metric.Int64Counter
}

var errMissingUserInfo = errors.New("missing user information")

type APIServerDeps struct {
	AssetSvc       AssetService
	StarSvc        StarService
	DiscussionSvc  DiscussionService
	TagSvc         TagService
	TagTemplateSvc TagTemplateService
	UserSvc        UserService
	Logger         log.Logger
}

func NewAPIServer(d APIServerDeps) *APIServer {
	assetUpdateCounter, err := otel.Meter("github.com/goto/compass/internal/server/v1beta1").
		Int64Counter("compass.asset.update")
	if err != nil {
		otel.Handle(err)
	}

	return &APIServer{
		assetService:       d.AssetSvc,
		starService:        d.StarSvc,
		discussionService:  d.DiscussionSvc,
		tagService:         d.TagSvc,
		tagTemplateService: d.TagTemplateSvc,
		userService:        d.UserSvc,
		logger:             d.Logger,

		assetUpdateCounter: assetUpdateCounter,
	}
}

func (server *APIServer) ValidateUserInCtx(ctx context.Context) (string, error) {
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
