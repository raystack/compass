package discussion_test

import (
	"testing"

	"github.com/odpf/columbus/discussion"
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
			errString:   "error filter \"random\" for key \"type\" not recognized, only support \"openended issues qanda all\"",
		},
		{
			Description: "invalid state will return error",
			Filter:      &discussion.Filter{State: "random"},
			errString:   "error filter \"random\" for key \"state\" not recognized, only support \"open closed\"",
		},
		{
			Description: "invalid sort and direction will return error",
			Filter:      &discussion.Filter{SortBy: "random", SortDirection: "random"},
			errString:   "error filter \"random\" for key \"sort\" not recognized, only support \"created_at updated_at\" and error filter \"random\" for key \"direction\" not recognized, only support \"asc desc\"",
		},
		{
			Description: "invalid size and offset will return error",
			Filter:      &discussion.Filter{Size: -12, Offset: -1},
			errString:   "size cannot be less than 0",
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
