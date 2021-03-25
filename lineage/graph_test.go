package lineage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/models"
)

func TestInMemoryGraph(t *testing.T) {
	var sampleGraph = lineage.AdjacencyMap{
		"type_a/instance_a": lineage.AdjacencyEntry{
			Type:        "type_a",
			URN:         "instance_a",
			Downstreams: set.NewStringSet("type_b/instance_z"),
		},
		"type_a/instance_b": lineage.AdjacencyEntry{
			Type:      "type_a",
			URN:       "instance_b",
			Upstreams: set.NewStringSet("type_b/instance_z"),
		},
		"type_b/instance_z": lineage.AdjacencyEntry{
			Type:        "type_b",
			URN:         "instance_z",
			Upstreams:   set.NewStringSet("type_a/instance_a"),
			Downstreams: set.NewStringSet("type_a/instance_b"),
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
				Description: "filter by type",
				Supergraph:  sampleGraph,
				ExpectGraph: lineage.AdjacencyMap{
					"type_b/instance_z": sampleGraph["type_b/instance_z"],
				},
				Cfg: lineage.QueryCfg{
					TypeWhitelist: []string{"type_b"},
				},
			},
			{
				Description: "filter by type error when type not known",
				Supergraph:  sampleGraph,
				Cfg: lineage.QueryCfg{
					TypeWhitelist: []string{"type_c"},
				},
				Err: models.ErrNoSuchType{TypeName: "type_c"},
			},
			{
				Description: "collapse relationships",
				Supergraph:  sampleGraph,
				Cfg: lineage.QueryCfg{
					TypeWhitelist: []string{"type_a"},
					Collapse:      true,
				},
				ExpectGraph: lineage.AdjacencyMap{
					"type_a/instance_a": lineage.AdjacencyEntry{
						Type:        "type_a",
						URN:         "instance_a",
						Downstreams: set.NewStringSet("type_a/instance_b"),
						Upstreams:   set.NewStringSet(),
					},
					"type_a/instance_b": lineage.AdjacencyEntry{
						Type:        "type_a",
						URN:         "instance_b",
						Upstreams:   set.NewStringSet("type_a/instance_a"),
						Downstreams: set.NewStringSet(),
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
					enc.Encode(tc.ExpectGraph)
					fmt.Fprint(msg, "got: ")
					enc.Encode(result)
					t.Error(msg.String())
					return
				}
			})
		}
	})
}
