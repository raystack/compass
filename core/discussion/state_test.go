package discussion_test

import (
	"testing"

	"github.com/goto/compass/core/discussion"
	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	t.Run("enum to string conversion", func(t *testing.T) {
		type TestCase struct {
			Description string
			State       discussion.State
			Result      string
		}

		var testCases = []TestCase{
			{
				Description: "state open converts to \"open\"",
				State:       discussion.StateOpen,
				Result:      "open",
			},
			{
				Description: "state closed converts to \"closed\"",
				State:       discussion.StateClosed,
				Result:      "closed",
			},
			{
				Description: "unknown state converts to \"open\"",
				State:       discussion.StateOpen,
				Result:      "open",
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				assert.Equal(t, tc.Result, tc.State.String())
			})
		}
	})

	t.Run("string to enum conversion", func(t *testing.T) {
		type TestCase struct {
			Description string
			StateString string
			Result      discussion.State
		}

		var testCases = []TestCase{
			{
				Description: "\"open\" converts to state open",
				StateString: "open",
				Result:      discussion.StateOpen,
			},
			{
				Description: "\"closed\" converts to state closed",
				StateString: "closed",
				Result:      discussion.StateClosed,
			},
			{
				Description: "other words converts to state open",
				StateString: "random",
				Result:      discussion.StateOpen,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				assert.Equal(t, tc.Result, discussion.GetStateEnum(tc.StateString))
			})
		}
	})

	t.Run("validating string", func(t *testing.T) {
		type TestCase struct {
			Description string
			StateString string
			IsValid     bool
		}

		var testCases = []TestCase{
			{
				Description: "supported state will return valid true",
				StateString: "open",
				IsValid:     true,
			},
			{
				Description: "unsupported state will return valid true",
				StateString: "random",
				IsValid:     false,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				assert.Equal(t, tc.IsValid, discussion.IsStateStringValid(tc.StateString))
			})
		}
	})
}
