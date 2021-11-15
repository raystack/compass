package lineage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/record"
	"github.com/stretchr/testify/require"
)

func TestInMemoryGraph(t *testing.T) {
	var sampleGraph = lineage.AdjacencyMap{
		"topic/instance_a": lineage.AdjacencyEntry{
			Type:        record.TypeTopic,
			URN:         "instance_a",
			Downstreams: set.NewStringSet("table/instance_z"),
		},
		"topic/instance_b": lineage.AdjacencyEntry{
			Type:      record.TypeTopic,
			URN:       "instance_b",
			Upstreams: set.NewStringSet("table/instance_z"),
		},
		"table/instance_z": lineage.AdjacencyEntry{
			Type:        record.TypeTable,
			URN:         "instance_z",
			Upstreams:   set.NewStringSet("topic/instance_a"),
			Downstreams: set.NewStringSet("topic/instance_b"),
		},
		"job/isolated_record": lineage.AdjacencyEntry{
			Type: record.TypeJob,
			URN:  "isolated_record",
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
					"table/instance_z": sampleGraph["table/instance_z"],
				},
				Cfg: lineage.QueryCfg{
					TypeWhitelist: []record.Type{record.TypeTable},
				},
			},
			{
				Description: "collapse relationships",
				Supergraph:  sampleGraph,
				Cfg: lineage.QueryCfg{
					TypeWhitelist: []record.Type{record.TypeTopic},
					Collapse:      true,
				},
				ExpectGraph: lineage.AdjacencyMap{
					"topic/instance_a": lineage.AdjacencyEntry{
						Type:        record.TypeTopic,
						URN:         "instance_a",
						Downstreams: set.NewStringSet("topic/instance_b"),
						Upstreams:   set.NewStringSet(),
					},
					"topic/instance_b": lineage.AdjacencyEntry{
						Type:        record.TypeTopic,
						URN:         "instance_b",
						Upstreams:   set.NewStringSet("topic/instance_a"),
						Downstreams: set.NewStringSet(),
					},
				},
			},
			{
				Description: "build lineage from Root",
				Supergraph:  sampleGraph,
				Cfg: lineage.QueryCfg{
					Root: "topic/instance_a",
				},
				ExpectGraph: lineage.AdjacencyMap{
					"topic/instance_a": lineage.AdjacencyEntry{
						Type:        record.TypeTopic,
						URN:         "instance_a",
						Downstreams: set.NewStringSet("table/instance_z"),
					},
					"topic/instance_b": lineage.AdjacencyEntry{
						Type:      record.TypeTopic,
						URN:       "instance_b",
						Upstreams: set.NewStringSet("table/instance_z"),
					},
					"table/instance_z": lineage.AdjacencyEntry{
						Type:        record.TypeTable,
						URN:         "instance_z",
						Upstreams:   set.NewStringSet("topic/instance_a"),
						Downstreams: set.NewStringSet("topic/instance_b"),
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
					err = enc.Encode(tc.ExpectGraph)
					require.NoError(t, err)
					fmt.Fprint(msg, "got: ")
					err = enc.Encode(result)
					require.NoError(t, err)
					t.Error(msg.String())
					return
				}
			})
		}
	})
}
