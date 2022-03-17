package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/httpapi/handlers"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/lineage"
)

func TestLineageHandler(t *testing.T) {
	logger := log.NewNoop()
	t.Run("GetLineage", func(t *testing.T) {
		t.Run("should return a graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			node := lineage.Node{
				URN: "job-1",
			}
			var graph = lineage.Graph{
				{Source: "job-1", Target: "table-2"},
				{Source: "table-2", Target: "table-31"},
				{Source: "table-31", Target: "dashboard-30"},
			}

			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"urn": node.URN,
			})

			lr := new(mocks.LineageRepository)
			lr.On("GetGraph", rr.Context(), node).Return(graph, nil)

			handler := handlers.NewLineageHandler(logger, lr)
			handler.GetGraph(rw, rr)

			if rw.Code != http.StatusOK {
				t.Errorf("expected handler to respond with status %d, was %d instead", http.StatusOK, rw.Code)
				return
			}

			var response lineage.Graph
			err := json.NewDecoder(rw.Body).Decode(&response)
			if err != nil {
				t.Errorf("error decoding handler response: %v", err)
				return
			}

			if reflect.DeepEqual(response, graph) == false {
				t.Errorf("expected handler response to be: %#v\n was %#v instead", graph, response)
			}
		})
	})
}
