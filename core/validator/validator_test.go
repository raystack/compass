package validator_test

import (
	"testing"

	"github.com/odpf/compass/core/validator"
	"gotest.tools/assert"
)

func TestValidateStruct(t *testing.T) {
	type DummyStruct struct {
		VarOneOf string `json:"varoneof" validate:"omitempty,oneof=type1 type2 type3"`
		VarInt   int    `json:"varint" validate:"omitempty,gte=0"`
	}

	type TestCase struct {
		Description string
		Struct      interface{}
		ErrString   string
	}

	testCases := []TestCase{
		{
			Description: "return error with supported values in oneof type validation",
			Struct: DummyStruct{
				VarOneOf: "random",
			},
			ErrString: "error value \"random\" for key \"varoneof\" not recognized, only support \"type1 type2 type3\"",
		},
		{
			Description: "return error should greater than 0 in integer type validation",
			Struct: DummyStruct{
				VarInt: -1,
			},
			ErrString: "varint cannot be less than 0",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			err := validator.ValidateStruct(tc.Struct)
			assert.Equal(t, tc.ErrString, err.Error())
		})
	}
}

func TestValidateOneOf(t *testing.T) {
	type TestCase struct {
		Description string
		Value       string
		Enums       []string
		ErrString   string
	}

	testCases := []TestCase{
		{
			Description: "return error with supported values",
			Value:       "random",
			Enums:       []string{"type1", "type2", "type3"},
			ErrString:   "error value \"random\" not recognized, only support \"type1 type2 type3\"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			err := validator.ValidateOneOf(tc.Value, tc.Enums...)
			assert.Equal(t, tc.ErrString, err.Error())
		})
	}
}
