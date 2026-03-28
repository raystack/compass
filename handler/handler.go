package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/user"
)

type Handler struct {
	namespaceService   NamespaceService
	assetService       AssetService
	starService        StarService
	discussionService  DiscussionService
	tagService         TagService
	tagTemplateService TagTemplateService
	userService        UserService
}

var (
	errMissingUserInfo = errors.New("missing user information")
)

func New(
	namespaceService NamespaceService,
	assetService AssetService,
	starService StarService,
	discussionService DiscussionService,
	tagService TagService,
	tagTemplateService TagTemplateService,
	userService UserService,
) *Handler {
	return &Handler{
		namespaceService:   namespaceService,
		assetService:       assetService,
		starService:        starService,
		discussionService:  discussionService,
		tagService:         tagService,
		tagTemplateService: tagTemplateService,
		userService:        userService,
	}
}

func (server *Handler) validateUserInCtx(ctx context.Context, ns *namespace.Namespace) (string, error) {
	usr := user.FromContext(ctx)
	userID, err := server.userService.ValidateUser(ctx, ns, usr.UUID, usr.Email)
	if err != nil {
		if errors.Is(err, user.ErrNoUserInformation) {
			return "", connect.NewError(connect.CodeInvalidArgument, err)
		}
		if errors.As(err, &user.DuplicateRecordError{UUID: usr.UUID, Email: usr.Email}) {
			return "", connect.NewError(connect.CodeAlreadyExists, err)
		}
		return "", internalServerError(ctx, "error validating user", err)
	}
	if userID == "" {
		return "", connect.NewError(connect.CodeInvalidArgument, errMissingUserInfo)
	}
	return userID, nil
}

func internalServerError(ctx context.Context, msg string, err error) error {
	ref := time.Now().Unix()

	slog.ErrorContext(ctx, msg, "error", err, "ref", ref)
	return connect.NewError(connect.CodeInternal, fmt.Errorf(
		"%s - ref (%d)",
		http.StatusText(http.StatusInternalServerError),
		ref,
	))
}

func bodyParserErrorMsg(err error) string {
	return fmt.Sprintf("error parsing request body: %v", err)
}
