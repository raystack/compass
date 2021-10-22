package models

import (
	"context"
	"fmt"
	"strings"
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

// TypeClassification specifies a class for an Type
type TypeClassification string

// list of valid classifications
var (
	TypeClassificationResource TypeClassification = "resource"
	TypeClassificationDataset  TypeClassification = "dataset"
	TypeClassificationSchema   TypeClassification = "schema"
	TypeClassificationMetric   TypeClassification = "metric"
)

// AllTypeClassifications holds a list of valid classifications
var AllTypeClassifications = []TypeClassification{
	TypeClassificationResource,
	TypeClassificationDataset,
	TypeClassificationSchema,
	TypeClassificationMetric,
}

// Type represents a named collection of records
// Entities are supposed to represent resources, datasets and schema.
// XXX(Aman): should Type names be case insensitive?
type Type struct {
	Name           string              `json:"name"`
	Classification TypeClassification  `json:"classification"`
	Lineage        []LineageDescriptor `json:"lineage"`
	Boost          map[string]float64  `json:"boost"`
}

func (e Type) Normalise() Type {
	normal := e
	normal.Name = strings.ToLower(e.Name)
	normal.Classification = TypeClassification(
		strings.ToLower(string(e.Classification)),
	)
	return normal
}

// TypeRepository is an interface to a storage
// system for types.
type TypeRepository interface {
	CreateOrReplace(context.Context, Type) error
	GetByName(context.Context, string) (Type, error)
	GetAll(context.Context) ([]Type, error)
	Delete(context.Context, string) error
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
