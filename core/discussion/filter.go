package discussion

import (
	"strings"

	validator "github.com/raystack/compass/core/validator"
)

type Filter struct {
	Type                  string `json:"type" validate:"omitempty,oneof=openended issues qanda all"`
	State                 string `json:"state" validate:"omitempty,oneof=open closed all"`
	Assets                []string
	Owner                 string
	Assignees             []string
	Labels                []string
	SortBy                string `json:"sort" validate:"omitempty,oneof=created_at updated_at"`
	SortDirection         string `json:"direction" validate:"omitempty,oneof=asc desc"`
	Size                  int    `json:"size" validate:"omitempty,gte=0"`
	Offset                int    `json:"offset" validate:"omitempty,gte=0"`
	DisjointAssigneeOwner bool
}

// Validate will check whether fields in the filter fulfills the constraint
func (f *Filter) Validate() error {
	return validator.ValidateStruct(f)
}

// AssignDefault will populate default value to filter
func (f *Filter) AssignDefault() {
	if len(strings.TrimSpace(f.Type)) == 0 {
		f.Type = "all"
	}

	if len(strings.TrimSpace(f.State)) == 0 {
		f.State = StateOpen.String()
	}

	if len(strings.TrimSpace(f.SortBy)) == 0 {
		f.SortBy = "created_at"
	}

	if len(strings.TrimSpace(f.SortDirection)) == 0 {
		f.SortDirection = "desc"
	}
}
