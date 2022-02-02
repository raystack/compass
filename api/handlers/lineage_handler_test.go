package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
)

func TestLineageHandler(t *testing.T) {
	logger := log.NewNoop()
	t.Run("GetLineage", func(t *testing.T) {
		t.Run("should return a graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			var subGraph = lineage.AdjacencyMap{
				"asset-A": lineage.AdjacencyEntry{
					ID:          "asset-A",
					Upstreams:   set.NewStringSet(),
					Downstreams: set.NewStringSet("asset-B"),
				},
				"asset-B": lineage.AdjacencyEntry{
					ID:          "asset-B",
					Upstreams:   set.NewStringSet("asset-A", "asset-Z"),
					Downstreams: set.NewStringSet(),
				},
				"asset-C": lineage.AdjacencyEntry{
					ID:          "asset-C",
					Upstreams:   set.NewStringSet(),
					Downstreams: set.NewStringSet(),
				},
			}

			graph := new(mock.Graph)
			graph.On("Query", lineage.QueryCfg{Root: "asset-A"}).Return(subGraph, nil)

			lp := new(mock.LineageProvider)
			lp.On("Graph").Return(graph, nil)

			handler := handlers.NewLineageHandler(logger, lp)

			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"id": "asset-A",
			})

			handler.GetLineage(rw, rr)

			if rw.Code != http.StatusOK {
				t.Errorf("expected handler to respond with status %d, was %d instead", http.StatusOK, rw.Code)
				return
			}

			var response lineage.AdjacencyMap
			err := json.NewDecoder(rw.Body).Decode(&response)
			if err != nil {
				t.Errorf("error decoding handler response: %v", err)
				return
			}

			if reflect.DeepEqual(response, subGraph) == false {
				t.Errorf("expected handler response to be: %#v\n was %#v instead", subGraph, response)
			}
		})
	})
}
