package user_test

import (
	"testing"

	"github.com/odpf/compass/user"
)

func TestErrors(t *testing.T) {
	type testCase struct {
		Description    string
		Err            error
		ExpectedString string
	}

	var testCases = []testCase{
		{
			Description:    "not found error return correct error string",
			Err:            user.NotFoundError{UUID: "uuid", Email: "email"},
			ExpectedString: "could not find user with uuid \"uuid\" with email \"email\"",
		},
		{
			Description:    "duplicate error return correct error string",
			Err:            user.DuplicateRecordError{UUID: "uuid", Email: "email"},
			ExpectedString: "duplicate user with uuid \"uuid\" with email \"email\"",
		},
		{
			Description:    "invalid error return correct error string",
			Err:            user.InvalidError{UUID: "uuid"},
			ExpectedString: "empty field with uuid \"uuid\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			if tc.ExpectedString != tc.Err.Error() {
				t.Fatalf("actual is \"%+v\" but expected was \"%+v\"", tc.Err.Error(), tc.ExpectedString)
			}
		})
	}
}
