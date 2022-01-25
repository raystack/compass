package discovery

import (
	"context"

	"github.com/odpf/columbus/asset"
)

type Repository interface {
	Upsert(context.Context, asset.Asset) error
	Delete(ctx context.Context, assetID string) error
}

// RecordRepository is an abstract storage for Assets
type RecordRepository interface {
	CreateOrReplaceMany(context.Context, []asset.Asset) error

	// GetAll returns specific assets from storage
	// GetConfig is used to configure fetching such as filters and offset
	GetAll(ctx context.Context, cfg GetConfig) (RecordList, error)

	// GetByID returns a record by it's id.
	// The field that contains this ID is defined by the
	// type to which this record belongs
	GetByID(context.Context, string) (asset.Asset, error)

	// Delete deletes a record by it's id.
	// The field that contains this ID is defined by the
	// type to which this record belongs
	Delete(context.Context, string) error
}

// RecordRepositoryFactory represents a type capable
// of constructing a RecordRepository for a certain type
type RecordRepositoryFactory interface {
	For(Type string) (RecordRepository, error)
}

type RecordSearcher interface {
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
