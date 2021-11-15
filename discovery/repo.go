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
	// RecordFilter is an optional data structure that is
	// used for return documents matching the search criteria.
	GetAll(context.Context, RecordFilter) ([]record.Record, error)

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

	// TODO: we should probably switch to iterator types for returning
	// records, or we could add options for pagination
}

// RecordRepositoryFactory represents a type capable
// of constructing a RecordRepository for a certain type
type RecordRepositoryFactory interface {
	For(Type record.Type) (RecordRepository, error)
}

type RecordSearcher interface {
	Search(ctx context.Context, cfg SearchConfig) (results []record.Record, err error)
}
