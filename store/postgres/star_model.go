package postgres

import (
	"time"

	"github.com/odpf/columbus/star"
)

type StarConfig struct {
	Limit            int
	Offset           int
	SortKey          string
	SortDirectionKey string
}

type StarModel struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	AssetType string    `db:"asset_type"`
	AssetURN  string    `db:"asset_urn"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *StarModel) toStar(assetModel *AssetModel) *star.Star {
	return &star.Star{
		ID:        s.ID,
		Asset:     assetModel.toAsset(),
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
