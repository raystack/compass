package discussion_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/compass/discussion"
	"github.com/stretchr/testify/assert"
)

func TestValidateFilter(t *testing.T) {
	type testCase struct {
		Description string
		Filter      *discussion.Filter
		errString   string
	}

	var testCases = []testCase{
		{
			Description: "empty filter will be valid",
			Filter:      &discussion.Filter{},
		},
		{
			Description: "invalid type will return error",
			Filter:      &discussion.Filter{Type: "random"},
			errString:   "error value \"random\" for key \"type\" not recognized, only support \"openended issues qanda all\"",
		},
		{
			Description: "invalid state will return error",
			Filter:      &discussion.Filter{State: "random"},
			errString:   "error value \"random\" for key \"state\" not recognized, only support \"open closed all\"",
		},
		{
			Description: "invalid sort and direction will return error",
			Filter:      &discussion.Filter{SortBy: "random", SortDirection: "random"},
			errString:   "error value \"random\" for key \"sort\" not recognized, only support \"created_at updated_at\" and error value \"random\" for key \"direction\" not recognized, only support \"asc desc\"",
		},
		{
			Description: "invalid size and offset will return error",
			Filter:      &discussion.Filter{Size: -12, Offset: -1},
			errString:   "size cannot be less than 0 and offset cannot be less than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			err := tc.Filter.Validate()
			if err != nil {
				assert.Equal(t, tc.errString, err.Error())
			}
		})
	}
}

func TestAssignDefault(t *testing.T) {
	type testCase struct {
		Description    string
		Filter         *discussion.Filter
		ExpectedFilter *discussion.Filter
	}

	var testCases = []testCase{
		{
			Description: "non empty fields in filter won't be changed",
			Filter: &discussion.Filter{
				Type:          "a-type",
				State:         discussion.StateClosed.String(),
				SortBy:        "sort-by",
				SortDirection: "sort-direction",
			},
			ExpectedFilter: &discussion.Filter{
				Type:          "a-type",
				State:         discussion.StateClosed.String(),
				SortBy:        "sort-by",
				SortDirection: "sort-direction",
			},
		},
		{
			Description: "empty default fields will set to defaults",
			Filter:      &discussion.Filter{},
			ExpectedFilter: &discussion.Filter{
				Type:          "all",
				State:         discussion.StateOpen.String(),
				SortBy:        "created_at",
				SortDirection: "desc",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			tc.Filter.AssignDefault()
			if diff := cmp.Diff(tc.Filter, tc.ExpectedFilter); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectedFilter, tc.Filter)
			}
		})
	}
}
