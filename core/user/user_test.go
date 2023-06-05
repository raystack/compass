package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	type testCase struct {
		Title       string
		User        *User
		ExpectError error
	}

	testCases := []testCase{
		{
			Title:       "should return error no user information if user is nil",
			User:        nil,
			ExpectError: ErrNoUserInformation,
		},
		{
			Title:       "should return error invalid if uuid is empty",
			User:        &User{Provider: "provider"},
			ExpectError: InvalidError{},
		},
		{
			Title:       "should return nil if user is valid",
			User:        &User{UUID: "some-uuid", Provider: "provider"},
			ExpectError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			err := testCase.User.Validate()
			assert.Equal(t, testCase.ExpectError, err)
		})
	}
}
