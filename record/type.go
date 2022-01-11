package record

import (
	"context"
	"fmt"
)

// TypeFields describe what fields of an Type
// record designate what.
// For instance the Value of the Title field will be
// the 'key' on the record that represents the title.
type TypeFields struct {
	// ID designates the idType for a record.
	// At any time, len(records) == len(records.GroupBy(id))
	// This is used by repository implementations to make a create or replace
	// decision. Think of it as the primary key for records.
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
}

// TypeName specifies a supported type name
type TypeName string

var (
	TypeNameTable     TypeName = "table"
	TypeNameJob       TypeName = "job"
	TypeNameDashboard TypeName = "dashboard"
	TypeNameTopic     TypeName = "topic"
)

// String cast TypeName to string
func (tn TypeName) String() string {
	return string(tn)
}

// Type represents a typename wrapped in a JSON
type Type struct {
	Name TypeName `json:"name"`
}

// AllSupportedTypes holds a list of all supported types struct
var AllSupportedTypes = []TypeName{
	TypeNameTable,
	TypeNameJob,
	TypeNameDashboard,
	TypeNameTopic,
}

// TypeRepository is an interface to a storage
// system for types.
type TypeRepository interface {
	CreateOrReplace(context.Context, TypeName) error
	GetByName(context.Context, string) (TypeName, error)
	GetAll(context.Context) ([]TypeName, error)
	// GetRecordsCount fetches records count for all available types
	// and returns them as a map[type]count
	GetRecordsCount(context.Context) (map[string]int, error)
}

type ErrNoSuchType struct {
	TypeName string
}

func (err ErrNoSuchType) Error() string {
	return fmt.Sprintf("no such type: %q", err.TypeName)
}

type ErrReservedTypeName struct {
	TypeName string
}

func (err ErrReservedTypeName) Error() string {
	return fmt.Sprintf("type is reserved: %q", err.TypeName)
}
