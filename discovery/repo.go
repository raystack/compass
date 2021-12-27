package discovery

import (
	"context"

	"github.com/odpf/columbus/record"
)

type RecordIterator interface {
	Scan() bool
	Next() []record.Record
	Close() error
}

// RecordRepository is an abstract storage for Records
type RecordRepository interface {
	CreateOrReplaceMany(context.Context, []record.Record) error

	// GetAll returns specific records from storage
	// GetConfig is used to configure fetching such as filters and offset
	GetAll(ctx context.Context, cfg GetConfig) (RecordList, error)

	// GetAllIterator returns RecordIterator to iterate records by batches
	GetAllIterator(context.Context) (RecordIterator, error)

	// GetByID returns a record by it's id.
	// The field that contains this ID is defined by the
	// type to which this record belongs
	GetByID(context.Context, string) (record.Record, error)

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
