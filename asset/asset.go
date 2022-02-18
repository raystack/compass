package asset

//go:generate mockery --name Repository --outpkg mocks --output ../lib/mocks/ --structname AssetRepository --filename asset_repository.go
import (
	"context"
	"time"

	"github.com/odpf/columbus/user"

	"github.com/r3labs/diff/v2"
)

type Config struct {
	Text    string `json:"text"`
	Type    Type   `json:"type"`
	Service string `json:"service"`
	Size    int    `json:"size"`
	Offset  int    `json:"offset"`
}

type Repository interface {
	Get(context.Context, Config) ([]Asset, error)
	GetCount(context.Context, Config) (int, error)
	GetByID(ctx context.Context, id string) (Asset, error)
	GetLastVersions(ctx context.Context, cfg Config, id string) ([]AssetVersion, error)
	GetByVersion(ctx context.Context, id string, version string) (Asset, error)
	Upsert(ctx context.Context, userID string, ast *Asset) (string, error)
	Delete(ctx context.Context, id string) error
}

// Asset is a model that wraps arbitrary data with Columbus' context
type Asset struct {
	ID          string                 `json:"id" diff:"-"`
	URN         string                 `json:"urn" diff:"-"`
	Type        Type                   `json:"type" diff:"-"`
	Service     string                 `json:"service" diff:"-"`
	Name        string                 `json:"name" diff:"name"`
	Description string                 `json:"description" diff:"description"`
	Data        map[string]interface{} `json:"data" diff:"data"`
	Labels      map[string]string      `json:"labels" diff:"labels"`
	Owners      []user.User            `json:"owners,omitempty" diff:"owners"`
	CreatedAt   time.Time              `json:"created_at" diff:"-"`
	UpdatedAt   time.Time              `json:"updated_at" diff:"-"`
	Version     string                 `json:"version" diff:"-"`
	UpdatedBy   user.User              `json:"updated_by" diff:"-"`
	Changelog   diff.Changelog         `json:"changelog,omitempty" diff:"-"`
}

// Diff returns nil changelog with nil error if equal
// returns wrapped r3labs/diff Changelog struct with nil error if not equal
func (a *Asset) Diff(otherAsset *Asset) (diff.Changelog, error) {
	return diff.Diff(a, otherAsset, diff.DiscardComplexOrigin(), diff.AllowTypeMismatch(true))
}
