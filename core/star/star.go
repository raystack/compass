package star

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname StarRepository --filename star_repository.go --output=./mocks

import (
	"context"

	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/user"
)

type Repository interface {
	Create(ctx context.Context, ns *namespace.Namespace, userID string, entityID string) (string, error)
	GetStargazers(ctx context.Context, flt Filter, entityID string) ([]user.User, error)
	GetAllEntitiesByUserID(ctx context.Context, flt Filter, userID string) ([]entity.Entity, error)
	GetEntityByUserID(ctx context.Context, userID string, entityID string) (entity.Entity, error)
	Delete(ctx context.Context, userID string, entityID string) error
}
