package handler

import (
	"context"

	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/star"
	"github.com/raystack/compass/core/user"
)

// StarService defines star operations for the handler.
type StarService interface {
	GetStarredEntitiesByUserID(ctx context.Context, flt star.Filter, userID string) ([]entity.Entity, error)
	GetStarredEntityByUserID(ctx context.Context, userID string, entityID string) (entity.Entity, error)
	Stars(ctx context.Context, ns *namespace.Namespace, userID string, entityID string) (string, error)
	Unstars(ctx context.Context, userID string, entityID string) error
	GetStargazers(ctx context.Context, flt star.Filter, entityID string) ([]user.User, error)
}
