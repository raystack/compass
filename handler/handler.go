package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
)

type Handler struct {
	namespaceService NamespaceService
	entityService    EntityServiceV2
	edgeService      EdgeServiceV2
}

func New(
	namespaceService NamespaceService,
	entityService EntityServiceV2,
	edgeService EdgeServiceV2,
) *Handler {
	return &Handler{
		namespaceService: namespaceService,
		entityService:    entityService,
		edgeService:      edgeService,
	}
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
