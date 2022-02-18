package star

//go:generate mockery --name Repository --outpkg mocks --output ../lib/mocks/ --structname StarRepository --filename star_repository.go

import (
	"context"
	"time"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/user"
)

type Star struct {
	ID        string      `json:"id"`
	Asset     asset.Asset `json:"asset"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, userID string, assetID string) (string, error)
	GetStargazers(ctx context.Context, cfg Config, assetID string) ([]user.User, error)
	GetAllAssetsByUserID(ctx context.Context, cfg Config, userID string) ([]asset.Asset, error)
	GetAllAssetsByUserEmail(ctx context.Context, cfg Config, userID string) ([]asset.Asset, error)
	GetAssetByUserID(ctx context.Context, userID string, assetID string) (*asset.Asset, error)
	Delete(ctx context.Context, userID string, assetID string) error
}
