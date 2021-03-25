package lineage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/models"
)

func TestAdjacencyEntry(t *testing.T) {
	t.Run("AdjacentEntriesInDir", func(t *testing.T) {

		entry := lineage.AdjacencyEntry{
			Type:        "entry",
			URN:         "one",
			Upstreams:   set.NewStringSet("upstream/one"),
			Downstreams: set.NewStringSet("downstream/two", "downstream/three"),
		}

		type testCase struct {
			Description string
			Dir         models.DataflowDir
			ExpectSet   set.StringSet
		}

		var testCases = []testCase{
			{
				Description: "upstream",
				Dir:         models.DataflowDirUpstream,
				ExpectSet:   set.NewStringSet("upstream/one"),
			},
			{
				Description: "downstream",
				Dir:         models.DataflowDirDownstream,
				ExpectSet:   set.NewStringSet("downstream/two", "downstream/three"),
			},
			{
				Description: "catch all",
				Dir:         models.DataflowDir("custom"),
				ExpectSet:   set.NewStringSet(),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				result := entry.AdjacentEntriesInDir(tc.Dir)
				var msg = new(bytes.Buffer)
				fmt.Fprintf(msg, "expected: ")
				json.NewEncoder(msg).Encode(tc.ExpectSet)
				fmt.Fprintf(msg, "got: ")
				json.NewEncoder(msg).Encode(result)
			})
		}
	})
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
