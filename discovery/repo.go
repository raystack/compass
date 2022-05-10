package discovery

//go:generate mockery --name Repository --outpkg mocks --output ../lib/mocks/ --with-expecter --structname DiscoveryRepository --filename discovery_repository.go
//go:generate mockery --name AssetSearcher --outpkg mocks --output ../lib/mocks/ --with-expecter --structname DiscoveryAssetSearcher --filename discovery_asset_searcher.go
//go:generate mockery --name AssetIterator --outpkg mocks --output ../lib/mocks/ --with-expecter --structname DiscoveryAssetIterator --filename discovery_asset_iterator.go
//go:generate mockery --name AssetRepository --outpkg mocks --output ../lib/mocks/ --with-expecter --structname DiscoveryAssetRepository --filename discovery_asset_repository.go
//go:generate mockery --name AssetRepositoryFactory --outpkg mocks --output ../lib/mocks/ --with-expecter --structname DiscoveryAssetRepositoryFactory --filename discovery_asset_repository_factory.go
//go:generate mockery --name TypeRepository --outpkg mocks --output ../lib/mocks/ --with-expecter --structname TypeRepository --filename type_repository.go
import (
	"context"

	"github.com/odpf/compass/asset"
)

type AssetIterator interface {
	Scan() bool
	Next() []asset.Asset
	Close() error
}

type Repository interface {
	Upsert(context.Context, asset.Asset) error
	Delete(ctx context.Context, assetID string) error
}

// AssetRepository is an abstract storage for Assets
type AssetRepository interface {
	CreateOrReplaceMany(context.Context, []asset.Asset) error

	// GetAll returns specific assets from storage
	// GetConfig is used to configure fetching such as filters and offset
	GetAll(ctx context.Context, cfg GetConfig) (AssetList, error)

	// GetAllIterator returns AssetIterator to iterate assets by batches
	GetAllIterator(context.Context) (AssetIterator, error)

	// GetByID returns a asset by it's id.
	// The field that contains this ID is defined by the
	// type to which this asset belongs
	GetByID(context.Context, string) (asset.Asset, error)

	// Delete deletes a asset by it's id.
	// The field that contains this ID is defined by the
	// type to which this asset belongs
	Delete(context.Context, string) error
}

// AssetRepositoryFactory represents a type capable
// of constructing a AssetRepository for a certain type
type AssetRepositoryFactory interface {
	For(Type string) (AssetRepository, error)
}

type AssetSearcher interface {
	Search(ctx context.Context, cfg SearchConfig) (results []SearchResult, err error)
	Suggest(ctx context.Context, cfg SearchConfig) (suggestions []string, err error)
}

// TypeRepository is an interface to a storage
// system for types.
type TypeRepository interface {
	// GetAll fetches types with assets count for all available types
	// and returns them as a map[typeName]count
	GetAll(context.Context) (map[asset.Type]int, error)
}
