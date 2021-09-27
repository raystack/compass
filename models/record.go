package models

import (
	"context"
	"fmt"
	"time"
)

// RecordV1 represents an arbitrary JSON message
// TODO(Aman): add validation for mandatory fields? (landscape for instance)
type RecordV1 = map[string]interface{}

// RecordV2 is a model that wraps arbitrary data with Columbus' context
type RecordV2 struct {
	Urn         string                 `json:"urn" mapstructure:"urn"`
	Name        string                 `json:"name" mapstructure:"name"`
	Description string                 `json:"description" mapstructure:"description"`
	Data        map[string]interface{} `json:"data" mapstructure:"data"`
	Labels      map[string]string      `json:"labels" mapstructure:"labels"`
	CreatedAt   time.Time              `json:"created_at" mapstructure:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" mapstructure:"updated_at"`
}

// RecordFilter is a filter intended to be used as a search
// criteria for operations involving record search
type RecordFilter = map[string][]string

// RecordRepository is an abstract storage for RecordV1s
type RecordRepository interface {
	CreateOrReplaceMany(context.Context, []RecordV1) error
	CreateOrReplaceManyV2(context.Context, []RecordV2) error

	// GetAll returns specific records from storage
	// RecordFilter is an optional data structure that is
	// used for return documents matching the search criteria.
	GetAll(context.Context, RecordFilter) ([]RecordV1, error)

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
