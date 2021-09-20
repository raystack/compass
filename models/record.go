package models

import (
	"context"
	"fmt"
)

// RecordV1 represents an arbitrary JSON message
// TODO(Aman): add validation for mandatory fields? (landscape for instance)
type RecordV1 = map[string]interface{}

// RecordV1Filter is a filter intended to be used as a search
// criteria for operations involving record search
type RecordV1Filter = map[string][]string

// RecordV1Repository is an abstract storage for RecordV1s
type RecordV1Repository interface {
	CreateOrReplaceMany(context.Context, []RecordV1) error

	// GetAll returns specific records from storage
	// RecordV1Filter is an optional data structure that is
	// used for return documents matching the search criteria.
	GetAll(context.Context, RecordV1Filter) ([]RecordV1, error)

	// GetByID returns a record by it's id.
	// The field that contains this ID is defined by the
	// type to which this record belongs
	GetByID(context.Context, string) (RecordV1, error)

	// Delete deletes a record by it's id.
	// The field that contains this ID is defined by the
	// type to which this record belongs
	Delete(context.Context, string) error

	// TODO: we should probably switch to iterator types for returning
	// records, or we could add options for pagination
}

// RecordV1RepositoryFactory represents a type capable
// of constructing a RecordV1Repository for a certain type
type RecordV1RepositoryFactory interface {
	For(Type) (RecordV1Repository, error)
}

type ErrNoSuchRecordV1 struct {
	RecordV1ID string
}

func (err ErrNoSuchRecordV1) Error() string {
	return fmt.Sprintf("no such record: %q", err.RecordV1ID)
}
