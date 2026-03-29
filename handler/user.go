package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/star"
	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/internal/middleware"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
)

type UserService interface {
	ValidateUser(ctx context.Context, ns *namespace.Namespace, uuid, email string) (string, error)
}

func (server *Handler) GetUserStarredEntities(ctx context.Context, req *connect.Request[compassv1beta1.GetUserStarredEntitiesRequest]) (*connect.Response[compassv1beta1.GetUserStarredEntitiesResponse], error) {
	entities, err := server.starService.GetStarredEntitiesByUserID(ctx, star.Filter{
		Size:   int(req.Msg.GetSize()),
		Offset: int(req.Msg.GetOffset()),
	}, req.Msg.GetUserId())
	if err != nil {
		return nil, internalServerError(ctx, "error getting starred entities", err)
	}

	data := make([]*compassv1beta1.Entity, len(entities))
	for i, e := range entities {
		data[i] = entityToProto(e)
	}
	return connect.NewResponse(&compassv1beta1.GetUserStarredEntitiesResponse{Data: data}), nil
}

func (server *Handler) GetMyStarredEntities(ctx context.Context, req *connect.Request[compassv1beta1.GetMyStarredEntitiesRequest]) (*connect.Response[compassv1beta1.GetMyStarredEntitiesResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	userID, err := server.validateUserInCtx(ctx, ns)
	if err != nil {
		return nil, err
	}

	entities, err := server.starService.GetStarredEntitiesByUserID(ctx, star.Filter{
		Size:   int(req.Msg.GetSize()),
		Offset: int(req.Msg.GetOffset()),
	}, userID)
	if err != nil {
		return nil, internalServerError(ctx, "error getting my starred entities", err)
	}

	data := make([]*compassv1beta1.Entity, len(entities))
	for i, e := range entities {
		data[i] = entityToProto(e)
	}
	return connect.NewResponse(&compassv1beta1.GetMyStarredEntitiesResponse{Data: data}), nil
}

func (server *Handler) GetMyStarredEntity(ctx context.Context, req *connect.Request[compassv1beta1.GetMyStarredEntityRequest]) (*connect.Response[compassv1beta1.GetMyStarredEntityResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	userID, err := server.validateUserInCtx(ctx, ns)
	if err != nil {
		return nil, err
	}

	ent, err := server.starService.GetStarredEntityByUserID(ctx, userID, req.Msg.GetEntityId())
	if err != nil {
		return nil, internalServerError(ctx, "error getting starred entity", err)
	}

	return connect.NewResponse(&compassv1beta1.GetMyStarredEntityResponse{Data: entityToProto(ent)}), nil
}

func (server *Handler) StarEntity(ctx context.Context, req *connect.Request[compassv1beta1.StarEntityRequest]) (*connect.Response[compassv1beta1.StarEntityResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	userID, err := server.validateUserInCtx(ctx, ns)
	if err != nil {
		return nil, err
	}

	id, err := server.starService.Stars(ctx, ns, userID, req.Msg.GetEntityId())
	if err != nil {
		if errors.Is(err, star.ErrEmptyUserID) || errors.Is(err, star.ErrEmptyEntityID) || errors.As(err, new(star.InvalidError)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		if errors.As(err, new(star.UserNotFoundError)) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.As(err, new(star.DuplicateRecordError)) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return nil, internalServerError(ctx, "error starring entity", err)
	}

	return connect.NewResponse(&compassv1beta1.StarEntityResponse{Id: id}), nil
}

func (server *Handler) UnstarEntity(ctx context.Context, req *connect.Request[compassv1beta1.UnstarEntityRequest]) (*connect.Response[compassv1beta1.UnstarEntityResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	userID, err := server.validateUserInCtx(ctx, ns)
	if err != nil {
		return nil, err
	}

	if err := server.starService.Unstars(ctx, userID, req.Msg.GetEntityId()); err != nil {
		if errors.Is(err, star.ErrEmptyUserID) || errors.Is(err, star.ErrEmptyEntityID) || errors.As(err, new(star.InvalidError)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		if errors.As(err, new(star.NotFoundError)) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, internalServerError(ctx, "error unstarring entity", err)
	}

	return connect.NewResponse(&compassv1beta1.UnstarEntityResponse{}), nil
}

func (server *Handler) GetEntityStargazers(ctx context.Context, req *connect.Request[compassv1beta1.GetEntityStargazersRequest]) (*connect.Response[compassv1beta1.GetEntityStargazersResponse], error) {
	users, err := server.starService.GetStargazers(ctx, star.Filter{
		Size:   int(req.Msg.GetSize()),
		Offset: int(req.Msg.GetOffset()),
	}, req.Msg.GetId())
	if err != nil {
		return nil, internalServerError(ctx, "error getting stargazers", err)
	}

	data := make([]*compassv1beta1.User, len(users))
	for i, u := range users {
		data[i] = &compassv1beta1.User{Id: u.ID, Uuid: u.UUID, Email: u.Email}
	}
	return connect.NewResponse(&compassv1beta1.GetEntityStargazersResponse{Data: data}), nil
}

// suppress unused import warnings
var (
	_ = user.FromContext
)
