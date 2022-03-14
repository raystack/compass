package comment

import (
	"strings"

	"github.com/odpf/columbus/filter"
)

type Filter struct {
	SortBy        string `json:"sort" validate:"omitempty,oneof=created_at updated_at"`
	SortDirection string `json:"direction" validate:"omitempty,oneof=asc desc"`
	Size          int
	Offset        int
}

// Validate will check whether fields in the filter fulfills the constraint
func (f *Filter) Validate() error {
	return filter.ValidateStruct(f)
}

// AssignDefault will populate default value to filter
func (f *Filter) AssignDefault() {
	if len(strings.TrimSpace(f.SortBy)) == 0 {
		f.SortBy = "created_at"
	}

	if len(strings.TrimSpace(f.SortDirection)) == 0 {
		f.SortDirection = "desc"
	}
}
