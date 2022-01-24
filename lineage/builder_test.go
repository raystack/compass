package lineage_test

import (
	"context"
	"testing"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/lineage/mocks"
	"github.com/stretchr/testify/assert"
)

func adjEntryWithTypeAndURN(id, urn, typ, service string) lineage.AdjacencyEntry {
	return lineage.AdjacencyEntry{
		ID:          id,
		Type:        typ,
		URN:         urn,
		Service:     service,
		Downstreams: set.NewStringSet(),
		Upstreams:   set.NewStringSet(),
	}
}

func TestDefaultBuilder(t *testing.T) {
	t.Run("graph construction algorithm", func(t *testing.T) {
		type testCase struct {
			Description string
			Edges       []lineage.Edge
			Result      lineage.AdjacencyMap
			QueryCfg    lineage.QueryCfg
			BuildErr    error
			QueryErr    error
		}
		var testCases = []testCase{
			{
				Description: "smoke test",
				Edges:       []lineage.Edge{},
				Result:      lineage.AdjacencyMap{},
			},
			{
				Description: "internal ref test (simple)",
				Edges: []lineage.Edge{
					{
						SourceID: "A",
						TargetID: "1",
					},
					{
						SourceID: "B",
						TargetID: "1",
					},
					{
						SourceID: "1",
						TargetID: "C",
					},
				},
				Result: lineage.AdjacencyMap{
					"1": lineage.AdjacencyEntry{
						ID:          "1",
						Downstreams: set.NewStringSet("C"),
						Upstreams:   set.NewStringSet("A", "B"),
					},
					"A": lineage.AdjacencyEntry{
						ID:          "A",
						Downstreams: set.NewStringSet("1"),
						Upstreams:   set.NewStringSet(),
					},
					"B": lineage.AdjacencyEntry{
						ID:          "B",
						Downstreams: set.NewStringSet("1"),
						Upstreams:   set.NewStringSet(),
					},
					"C": lineage.AdjacencyEntry{
						ID:          "C",
						Downstreams: set.NewStringSet(),
						Upstreams:   set.NewStringSet("1"),
					},
				},
			},
			{
				Description: "external ref test",
				Edges: []lineage.Edge{
					{
						SourceID: "data-booking",
						TargetID: "booking-aggregator",
					},
					{
						SourceID: "data-booking",
						TargetID: "booking-fraud-detector",
					},
				},
				Result: lineage.AdjacencyMap{
					"data-booking": lineage.AdjacencyEntry{
						ID:          "data-booking",
						Downstreams: set.NewStringSet("booking-aggregator", "booking-fraud-detector"),
						Upstreams:   set.NewStringSet(),
					},
					"booking-aggregator": lineage.AdjacencyEntry{
						ID:          "booking-aggregator",
						Downstreams: set.NewStringSet(),
						Upstreams:   set.NewStringSet("data-booking"),
					},
					"booking-fraud-detector": lineage.AdjacencyEntry{
						ID:          "booking-fraud-detector",
						Downstreams: set.NewStringSet(),
						Upstreams:   set.NewStringSet("data-booking"),
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				ctx := context.Background()
				repo := new(mocks.Repository)
				repo.On("GetEdges", ctx).Return(tc.Edges, nil)
				defer repo.AssertExpectations(t)

				graph, err := lineage.DefaultBuilder.Build(ctx, repo)
				if err != nil {
					if err != tc.BuildErr {
						t.Errorf("unexpected error when building graph: %v", err)
					}
					// the error was expected, terminate test case
					return
				}

				result, err := graph.Query(tc.QueryCfg)
				if err != nil {
					if err != tc.QueryErr {
						t.Errorf("unexpected error querying the graph: %v", err)
					}
					return
				}

				assert.Equal(t, tc.Result, result)
			})
		}
	})
}
