package principal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	type testCase struct {
		Title       string
		Principal   *Principal
		ExpectError error
	}

	var testCases = []testCase{
		{
			Title:       "should return error no principal information if principal is nil",
			Principal:   nil,
			ExpectError: ErrNoPrincipalInformation,
		},
		{
			Title:       "should return error invalid if subject is empty",
			Principal:   &Principal{Type: "user"},
			ExpectError: InvalidError{},
		},
		{
			Title:       "should return nil if principal is valid",
			Principal:   &Principal{Subject: "some-subject", Type: "user"},
			ExpectError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			err := testCase.Principal.Validate()
			assert.Equal(t, testCase.ExpectError, err)
		})
	}
}
