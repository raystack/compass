package comment_test

import (
	"testing"

	"github.com/odpf/columbus/comment"
	"github.com/stretchr/testify/assert"
)

func TestValidateFilter(t *testing.T) {
	type testCase struct {
		Description string
		Filter      *comment.Filter
		errString   string
	}

	var testCases = []testCase{
		{
			Description: "empty filter will be valid",
			Filter:      &comment.Filter{},
		},
		{
			Description: "invalid sort and direction will return error",
			Filter:      &comment.Filter{SortBy: "random", SortDirection: "random"},
			errString:   "error value \"random\" for key \"sort\" not recognized, only support \"created_at updated_at\" and error value \"random\" for key \"direction\" not recognized, only support \"asc desc\"",
		},
		{
			Description: "invalid size and offset will return error",
			Filter:      &comment.Filter{Size: -12, Offset: -1},
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
