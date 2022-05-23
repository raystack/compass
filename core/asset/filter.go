package asset

import (
	"github.com/odpf/compass/core/validator"
)

type Filter struct {
	Types         []Type
	Services      []string
	Size          int
	Offset        int
	SortBy        string `validate:"omitempty,oneof=name type service created_at updated_at"`
	SortDirection string `validate:"omitempty,oneof=asc desc"`
	QueryFields   []string
	Query         string
	Data          map[string]string
}

func (f *Filter) Validate() error {
	return validator.ValidateStruct(f)
}
