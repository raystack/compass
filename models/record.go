package models

import (
	"context"
	"fmt"
)

// Record represents an arbitrary JSON message
// TODO(Aman): add validation for mandatory fields? (landscape for instance)
type Record = map[string]interface{}

// RecordFilter is a filter intended to be used as a search
// criteria for operations involving record search
type RecordFilter = map[string][]string

// RecordRepository is an abstract storage for Records
type RecordRepository interface {
	CreateOrReplaceMany(context.Context, []Record) error

	// GetAll returns specific records from storage
	// RecordFilter is an optional data structure that is
	// used for return documents matching the search criteria.
	GetAll(context.Context, RecordFilter) ([]Record, error)

	// GetByID returns a record by it's id.
	// The field that contains this ID is defined by the
	// type to which this record belongs
	GetByID(context.Context, string) (Record, error)

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
	For(Type) (RecordRepository, error)
}

type ErrNoSuchRecord struct {
	RecordID string
}

func (err ErrNoSuchRecord) Error() string {
	return fmt.Sprintf("no such record: %q", err.RecordID)
}
