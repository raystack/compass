package lineage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lineage"
)

func TestCachedGraph(t *testing.T) {
	type testCase struct {
		Description  string
		Cfg          lineage.QueryCfg
		Graph        func(testCase) *mock.Graph
		Setup        func(testCase, *lineage.CachedGraph)
		ExpectCalls  int
		ExpectResult lineage.AdjacencyMap
		ExpectError  error
	}

	var graphFromTestCase = func(tc testCase) *mock.Graph {
		g := new(mock.Graph)
		g.On("Query", tc.Cfg).Return(tc.ExpectResult, tc.ExpectError)
		return g
	}

	var testCases = []testCase{
		{
			Description:  "pass through test",
			Cfg:          lineage.QueryCfg{},
			Graph:        graphFromTestCase,
			ExpectResult: lineage.AdjacencyMap{},
			ExpectCalls:  1,
		},
		{
			Description: "simple cache test",
			Cfg:         lineage.QueryCfg{},
			Setup: func(tc testCase, g *lineage.CachedGraph) {
				g.Query(tc.Cfg) // cache the request
			},
			Graph:        graphFromTestCase,
			ExpectResult: lineage.AdjacencyMap{},
			ExpectCalls:  1,
		},
		{
			Description: "test error handling",
			Cfg: lineage.QueryCfg{
				Collapse: true,
			},
			Graph: func(tc testCase) *mock.Graph {
				g := new(mock.Graph)
				g.On("Query", tc.Cfg).Return(lineage.AdjacencyMap{}, fmt.Errorf("bad implementation"))
				return g
			},
			ExpectError: fmt.Errorf("CachedGraph: error calling Query() on source graph: bad implementation"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			src := tc.Graph(tc)
			graph := lineage.NewCachedGraph(src)
			if tc.Setup != nil {
				tc.Setup(tc, graph)
			}
			result, err := graph.Query(tc.Cfg)
			if err != nil {
				if err.Error() != tc.ExpectError.Error() {
					t.Errorf("unexpected error while querying graph: %v", err)
				}
				return
			}

			if reflect.DeepEqual(tc.ExpectResult, result) == false {
				var (
					msg = new(bytes.Buffer)
					enc = json.NewEncoder(msg)
				)
				enc.SetIndent("", "  ")
				fmt.Fprint(msg, "expected: ")
				enc.Encode(tc.ExpectResult)
				fmt.Fprint(msg, "got: ")
				enc.Encode(result)
				t.Error(msg.String())
				return
			}

			src.AssertNumberOfCalls(t, "Query", tc.ExpectCalls)
		})
	}
}
