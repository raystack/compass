package lineage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
)

func TestInMemoryGraph(t *testing.T) {
	var sampleGraph = lineage.AdjacencyMap{
		"node-1": lineage.AdjacencyEntry{
			ID:          "node-1",
			Downstreams: set.NewStringSet("node-2"),
		},
		"node-3": lineage.AdjacencyEntry{
			ID:        "node-3",
			Upstreams: set.NewStringSet("node-2"),
		},
		"node-2": lineage.AdjacencyEntry{
			ID:          "node-2",
			Upstreams:   set.NewStringSet("node-1"),
			Downstreams: set.NewStringSet("node-3"),
		},
		"node-99": lineage.AdjacencyEntry{
			ID: "node-99",
		},
	}
	t.Run("Query", func(t *testing.T) {
		type testCase struct {
			Description string
			Supergraph  lineage.AdjacencyMap
			ExpectGraph lineage.AdjacencyMap
			Cfg         lineage.QueryCfg
			Err         error
		}

		var testCases = []testCase{
			{
				Description: "list the entire graph",
				Supergraph:  sampleGraph,
				ExpectGraph: sampleGraph,
			},
			{
				Description: "build lineage from Root",
				Supergraph:  sampleGraph,
				Cfg: lineage.QueryCfg{
					Root: "node-1",
				},
				ExpectGraph: lineage.AdjacencyMap{
					"node-1": lineage.AdjacencyEntry{
						ID:          "node-1",
						Downstreams: set.NewStringSet("node-2"),
					},
					"node-3": lineage.AdjacencyEntry{
						ID:        "node-3",
						Upstreams: set.NewStringSet("node-2"),
					},
					"node-2": lineage.AdjacencyEntry{
						ID:          "node-2",
						Upstreams:   set.NewStringSet("node-1"),
						Downstreams: set.NewStringSet("node-3"),
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				graph := lineage.NewInMemoryGraph(tc.Supergraph)
				result, err := graph.Query(tc.Cfg)
				if err != nil {
					if err != tc.Err {
						t.Errorf("expected graph.Query() to return %q error, but it returned %q instead", tc.Err, err)
					}
					return
				}

				if reflect.DeepEqual(tc.ExpectGraph, result) == false {
					var (
						msg = new(bytes.Buffer)
						enc = json.NewEncoder(msg)
					)
					enc.SetIndent("", "  ")
					fmt.Fprint(msg, "expected: ")
					if err := enc.Encode(tc.ExpectGraph); err != nil {
						t.Fatal(err)
					}
					fmt.Fprint(msg, "got: ")
					if err := enc.Encode(result); err != nil {
						t.Fatal(err)
					}
					t.Error(msg.String())
					return
				}
			})
		}
	})
}
