package discovery

//go:generate mockery --name Repository --outpkg mocks --output ../lib/mocks/ --with-expecter --structname DiscoveryRepository --filename discovery_repository.go
import (
	"context"

	"github.com/odpf/compass/asset"
)

type Repository interface {
	Upsert(context.Context, asset.Asset) error
	Delete(ctx context.Context, assetID string) error

	// GetTypes fetches types with assets count for all available types
	// and returns them as a map[typeName]count
	GetTypes(context.Context) (map[asset.Type]int, error)

	Search(ctx context.Context, cfg SearchConfig) (results []SearchResult, err error)
	Suggest(ctx context.Context, cfg SearchConfig) (suggestions []string, err error)
}
