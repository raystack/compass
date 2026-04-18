package principal_test

import (
	"testing"

	"github.com/raystack/compass/core/principal"
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
			Err:            principal.NotFoundError{Subject: "sub-123"},
			ExpectedString: "could not find principal with subject \"sub-123\"",
		},
		{
			Description:    "duplicate error return correct error string",
			Err:            principal.DuplicateRecordError{Subject: "sub-123"},
			ExpectedString: "duplicate principal with subject \"sub-123\"",
		},
		{
			Description:    "invalid error return correct error string",
			Err:            principal.InvalidError{Subject: ""},
			ExpectedString: "invalid principal: empty subject \"\"",
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
