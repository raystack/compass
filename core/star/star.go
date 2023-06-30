package star

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname StarRepository --filename star_repository.go --output=./mocks

import (
	"context"
	"github.com/raystack/compass/core/namespace"

	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/user"
)

type Repository interface {
	Create(ctx context.Context, ns *namespace.Namespace, userID string, assetID string) (string, error)
	GetStargazers(ctx context.Context, flt Filter, assetID string) ([]user.User, error)
	GetAllAssetsByUserID(ctx context.Context, flt Filter, userID string) ([]asset.Asset, error)
	GetAssetByUserID(ctx context.Context, userID string, assetID string) (asset.Asset, error)
	Delete(ctx context.Context, userID string, assetID string) error
}
