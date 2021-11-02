package lineage_test

import (
	"testing"

	"github.com/odpf/columbus/lineage"
)

func TestAdjacencyEntry(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		type testCase struct {
			Entry    lineage.AdjacencyEntry
			ExpectID string
		}

		var testCases = []testCase{
			{
				Entry: lineage.AdjacencyEntry{
					Type: "a",
					URN:  "b",
				},
				ExpectID: "a/b",
			},
			{
				Entry: lineage.AdjacencyEntry{
					Type: "  over",
					URN:  "under   ",
				},
				ExpectID: "over/under",
			},
			{
				Entry:    lineage.AdjacencyEntry{},
				ExpectID: "<unknown>/<unknown>",
			},
		}
		for _, tc := range testCases {
			if tc.Entry.ID() != tc.ExpectID {
				t.Errorf("expected %#v.ID() == %q, was %q", tc.Entry, tc.ExpectID, tc.Entry.ID())
			}
		}
	})
}
