package discussion_test

import (
	"errors"
	"testing"
	"time"

	"github.com/odpf/columbus/discussion"
	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	type TestCase struct {
		Description string
		Discussion  discussion.Discussion
		IsEmpty     bool
	}

	var testCases = []TestCase{
		{
			Description: "all necessary fields are empty and nil will be considered empty",
			Discussion:  discussion.Discussion{ID: "123", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			IsEmpty:     true,
		},
		{
			Description: "nil slice will be considered empty",
			Discussion:  discussion.Discussion{Labels: nil},
			IsEmpty:     true,
		},
		{
			Description: "empty slice won't be considered empty",
			Discussion:  discussion.Discussion{Labels: []string{}},
			IsEmpty:     false,
		},
		{
			Description: "title exist won't be considered empty",
			Discussion:  discussion.Discussion{Title: "title"},
			IsEmpty:     false,
		},
		{
			Description: "body exist won't be considered empty",
			Discussion:  discussion.Discussion{Body: "body"},
			IsEmpty:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.IsEmpty, tc.Discussion.IsEmpty())
		})
	}
}

func TestValidateConstraint(t *testing.T) {
	type TestCase struct {
		Description string
		Discussion  discussion.Discussion
		Err         error
	}

	var testCases = []TestCase{
		{
			Description: "type is not one of supported types will return error",
			Discussion:  discussion.Discussion{Type: "random"},
			Err:         discussion.ErrInvalidType,
		},
		{
			Description: "state is not one of supported states will return error",
			Discussion:  discussion.Discussion{State: "random"},
			Err:         discussion.ErrInvalidState,
		},
		{
			Description: "labels is more than MAX will return error",
			Discussion:  discussion.Discussion{Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
			Err:         errors.New("labels cannot be more than 10"),
		},
		{
			Description: "assets is more than MAX will return error",
			Discussion:  discussion.Discussion{Assets: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
			Err:         errors.New("assets cannot be more than 10"),
		},
		{
			Description: "assignees is more than MAX will return error",
			Discussion:  discussion.Discussion{Assignees: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
			Err:         errors.New("assignees cannot be more than 10"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.Err, tc.Discussion.ValidateConstraint())
		})
	}
}

func TestValidateDiscussion(t *testing.T) {
	type TestCase struct {
		Description string
		Discussion  discussion.Discussion
		Err         error
	}

	var testCases = []TestCase{
		{
			Description: "empty title will return error",
			Discussion:  discussion.Discussion{},
			Err:         errors.New("title cannot be empty"),
		},
		{
			Description: "empty body will return error",
			Discussion:  discussion.Discussion{Title: "title"},
			Err:         errors.New("body cannot be empty"),
		},
		{
			Description: "empty type will return error",
			Discussion:  discussion.Discussion{Title: "title", Body: "body"},
			Err:         errors.New("type must be specified"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.Err, tc.Discussion.Validate())
		})
	}
}
