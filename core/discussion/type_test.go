package discussion_test

import (
	"testing"

	"github.com/goto/compass/core/discussion"
	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	t.Run("enum to string conversion", func(t *testing.T) {
		type TestCase struct {
			Description string
			Type        discussion.Type
			Result      string
		}

		testCases := []TestCase{
			{
				Description: "type openended converts to \"openended\"",
				Type:        discussion.TypeOpenEnded,
				Result:      "openended",
			},
			{
				Description: "type issues converts to \"issues\"",
				Type:        discussion.TypeIssues,
				Result:      "issues",
			},
			{
				Description: "type qanda converts to \"qanda\"",
				Type:        discussion.TypeQAndA,
				Result:      "qanda",
			},
			{
				Description: "unknown state converts to \"openended\"",
				Type:        discussion.TypeOpenEnded,
				Result:      "openended",
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				assert.Equal(t, tc.Result, tc.Type.String())
			})
		}
	})

	t.Run("string to enum conversion", func(t *testing.T) {
		type TestCase struct {
			Description string
			TypeString  string
			Result      discussion.Type
		}

		testCases := []TestCase{
			{
				Description: "\"openended\" converts to type openended",
				TypeString:  "openended",
				Result:      discussion.TypeOpenEnded,
			},
			{
				Description: "\"issues\" converts to type issues",
				TypeString:  "issues",
				Result:      discussion.TypeIssues,
			},
			{
				Description: "\"qanda\" converts to type qanda",
				TypeString:  "qanda",
				Result:      discussion.TypeQAndA,
			},
			{
				Description: "other words fallback to type openended",
				TypeString:  "random",
				Result:      discussion.TypeOpenEnded,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				assert.Equal(t, tc.Result, discussion.GetTypeEnum(tc.TypeString))
			})
		}
	})

	t.Run("validating string", func(t *testing.T) {
		type TestCase struct {
			Description string
			TypeString  string
			IsValid     bool
		}

		testCases := []TestCase{
			{
				Description: "supported type will return valid true",
				TypeString:  "openended",
				IsValid:     true,
			},
			{
				Description: "unsupported type will return valid true",
				TypeString:  "random",
				IsValid:     false,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				assert.Equal(t, tc.IsValid, discussion.IsTypeStringValid(tc.TypeString))
			})
		}
	})
}
