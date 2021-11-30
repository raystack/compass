package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/models"
)

func TestLineageHandler(t *testing.T) {
	t.Run("ListLineage", func(t *testing.T) {
		t.Run("should return 404 if a non-existent type is requested", func(t *testing.T) {
			graph := new(mock.Graph)
			graph.On(
				"Query",
				lineage.QueryCfg{TypeWhitelist: []string{"bqtable"}}).
				Return(lineage.AdjacencyMap{}, models.ErrNoSuchType{TypeName: "bqtable"})
			lp := new(mock.LineageProvider)
			lp.On("Graph").Return(graph, nil)

			handler := handlers.NewLineageHandler(new(mock.Logger), lp)

			rr := httptest.NewRequest("GET", "/?filter.type=bqtable", nil)
			rw := httptest.NewRecorder()

			handler.ListLineage(rw, rr)

			if rw.Code != http.StatusNotFound {
				t.Errorf("expected handler to respond with status %d, was %d instead", http.StatusNotFound, rw.Code)
				return
			}

			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			if err != nil {
				t.Errorf("error decoding handler response: %v", err)
				return
			}

			expectedReason := `no such type: "bqtable"`
			if response.Reason != expectedReason {
				t.Errorf("expected handler to report reason %q, was %q instead", expectedReason, response.Reason)
			}
		})
		t.Run("should return graph filtered by type", func(t *testing.T) {
			var filteredGraph = lineage.AdjacencyMap{
				"topic/a": lineage.AdjacencyEntry{
					Type:        "topic",
					URN:         "a",
					Upstreams:   set.NewStringSet(),
					Downstreams: set.NewStringSet("dagger/ab"),
				},
				"topic/b": lineage.AdjacencyEntry{
					Type:        "topic",
					URN:         "a",
					Upstreams:   set.NewStringSet("dagger/ab"),
					Downstreams: set.NewStringSet(),
				},
				"dagger/ab": lineage.AdjacencyEntry{
					Type:        "dagger",
					URN:         "ab",
					Upstreams:   set.NewStringSet("topic/a"),
					Downstreams: set.NewStringSet("topic/b"),
				},
			}
			graph := new(mock.Graph)
			graph.On("Query", lineage.QueryCfg{TypeWhitelist: []string{"topic", "dagger"}}).Return(filteredGraph, nil)

			lp := new(mock.LineageProvider)
			lp.On("Graph").Return(graph, nil)

			handler := handlers.NewLineageHandler(new(mock.Logger), lp)

			rr := httptest.NewRequest("GET", "/?filter.type=topic&filter.type=dagger", nil)
			rw := httptest.NewRecorder()

			handler.ListLineage(rw, rr)

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

			if reflect.DeepEqual(response, filteredGraph) == false {
				t.Errorf("expected handler response to be %#v, was %#v instead", response, filteredGraph)
			}
		})
		t.Run("should return http 500 error if requesting the graph fails", func(t *testing.T) {
			errNoGraph := errors.New("no graph available")
			graph := new(mock.Graph)

			lp := new(mock.LineageProvider)
			lp.On("Graph").Return(graph, errNoGraph)

			handler := handlers.NewLineageHandler(new(mock.Logger), lp)

			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()

			handler.ListLineage(rw, rr)
			if rw.Code != http.StatusInternalServerError {
				t.Errorf("expected handler to respond with status %d, was %d instead", http.StatusInternalServerError, rw.Code)
				return
			}
		})
		t.Run("should return the entire graph if no options are specified", func(t *testing.T) {

			// using the same graph
			var fullGraph = lineage.AdjacencyMap{
				"bqtable/raw": lineage.AdjacencyEntry{
					Type:        "bqtable",
					URN:         "raw",
					Upstreams:   set.NewStringSet(),
					Downstreams: set.NewStringSet("s3/std"),
				},
				"s3/std": lineage.AdjacencyEntry{
					Type:        "s3",
					URN:         "std",
					Upstreams:   set.NewStringSet("bqtable/raw"),
					Downstreams: set.NewStringSet(),
				},
			}
			graph := new(mock.Graph)
			graph.On("Query", lineage.QueryCfg{}).Return(fullGraph, nil)

			lp := new(mock.LineageProvider)
			lp.On("Graph").Return(graph, nil)

			handler := handlers.NewLineageHandler(new(mock.Logger), lp)

			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()

			handler.ListLineage(rw, rr)

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

			if reflect.DeepEqual(response, fullGraph) == false {
				t.Errorf("expected handler response to be %#v, was %#v instead", response, fullGraph)
			}
		})
	})
	t.Run("GetLineage", func(t *testing.T) {
		t.Run("should return a graph containing the requested resource, along with it's related resources", func(t *testing.T) {
			var subGraph = lineage.AdjacencyMap{
				"bqtable/raw": lineage.AdjacencyEntry{
					Type:        "bqtable",
					URN:         "raw",
					Upstreams:   set.NewStringSet(),
					Downstreams: set.NewStringSet("bqtable/std"),
				},
				"bqtable/std": lineage.AdjacencyEntry{
					Type:        "bqtable",
					URN:         "std",
					Upstreams:   set.NewStringSet("bqtable/raw", "bqtable/to-be-removed"),
					Downstreams: set.NewStringSet(),
				},
				"bqtable/presentation": lineage.AdjacencyEntry{
					Type:        "bqtable",
					URN:         "presentation",
					Upstreams:   set.NewStringSet(),
					Downstreams: set.NewStringSet(),
				},
			}

			graph := new(mock.Graph)
			graph.On("Query", lineage.QueryCfg{Root: "bqtable/raw"}).Return(subGraph, nil)

			lp := new(mock.LineageProvider)
			lp.On("Graph").Return(graph, nil)

			handler := handlers.NewLineageHandler(new(mock.Logger), lp)

			rr := httptest.NewRequest("GET", "/v1/lineage/bqtable/raw", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"type": "bqtable",
				"id":   "raw",
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
