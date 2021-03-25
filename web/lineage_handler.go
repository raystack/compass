package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/models"
	"github.com/sirupsen/logrus"
)

// interface to lineage.Service
// named provider to avoid ambiguitity with the service implementation
type LineageProvider interface {
	Graph() (lineage.Graph, error)
}

type LineageHandler struct {
	mux             *mux.Router
	log             logrus.FieldLogger
	lineageProvider LineageProvider
}

func (handler *LineageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mux.ServeHTTP(w, r)
}

func (handler *LineageHandler) listLineage(w http.ResponseWriter, r *http.Request) {
	graph, err := handler.lineageProvider.Graph()
	if err != nil {
		handler.log.
			Errorf("error requesting graph: %v", err)

		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}

	opts := handler.parseOpts(r.URL.Query())
	res, err := graph.Query(opts)
	if err != nil {
		handler.log.
			WithField("query", opts).
			Errorf("error querying graph: %v\ncfg: %v", err, opts)

		status := http.StatusBadRequest
		if _, ok := err.(models.ErrNoSuchType); ok {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (handler *LineageHandler) getLineage(w http.ResponseWriter, r *http.Request) {
	graph, err := handler.lineageProvider.Graph()
	if err != nil {
		handler.log.Errorf("error requesting graph: %v", err)
		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	requestParams := mux.Vars(r)
	typeName := requestParams["type"]

	opts := handler.parseOpts(r.URL.Query())
	opts.TypeWhitelist = []string{typeName}

	res, err := graph.Query(opts)
	if err != nil {
		handler.log.
			WithField("query", opts).
			Errorf("error querying graph: %v\ncfg: %v", err, opts)

		status := http.StatusBadRequest
		if _, ok := err.(models.ErrNoSuchType); ok {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, err.Error())
		return
	}

	root := fmt.Sprintf("%s/%s", typeName, requestParams["id"])
	if _, found := res[root]; !found {
		msg := fmt.Sprintf("no such resource %q", root)
		writeJSONError(w, http.StatusNotFound, msg)
		return
	}

	// TODO: move the business logic for querying sub-graphs
	// to lineage.Graph
	res = handler.subgraph(res, root)
	writeJSON(w, http.StatusOK, res)
}

func (handler *LineageHandler) subgraph(superGraph lineage.AdjacencyMap, root string) lineage.AdjacencyMap {
	var (
		result  = make(lineage.AdjacencyMap)
		rootElm = superGraph[root]
	)
	result[rootElm.ID()] = rootElm
	for _, dir := range models.AllDataflowDir {
		handler.addAdjecentsInDir(result, superGraph, rootElm, dir)
	}
	return handler.pruneRefs(result)
}

func (handler *LineageHandler) addAdjecentsInDir(graph lineage.AdjacencyMap, superGraph lineage.AdjacencyMap, root lineage.AdjacencyEntry, dir models.DataflowDir) {
	queue := []lineage.AdjacencyEntry{root}
	for len(queue) > 0 {
		n := len(queue)
		el := queue[n-1]
		queue = queue[:n-1]
		for adjacent := range el.AdjacentEntriesInDir(dir) {
			adjacentEl, exists := superGraph[adjacent]
			if !exists {
				continue
			}
			graph[adjacentEl.ID()] = adjacentEl
			queue = append(queue, adjacentEl)
		}
	}
	return
}

// remove refs for each entry that don't exist in the built graph
// during handler.addAdjacentsInDir we follow the declarations to add
// upstreams/downstreams to the graph. Once that is done, we've constructed the lineage
// graph for the requested resource. During this step, we remove outgoing refs to any
// resource that doesn't belong in the graph
func (handler *LineageHandler) pruneRefs(graph lineage.AdjacencyMap) lineage.AdjacencyMap {
	pruned := make(lineage.AdjacencyMap)
	for _, entry := range graph {
		entry.Upstreams = handler.filterRefs(entry.Upstreams, graph)
		entry.Downstreams = handler.filterRefs(entry.Downstreams, graph)
		pruned[entry.ID()] = entry
	}
	return pruned
}

func (handler *LineageHandler) filterRefs(refs set.StringSet, graph lineage.AdjacencyMap) set.StringSet {
	rv := set.NewStringSet()
	for ref := range refs {
		if _, exists := graph[ref]; exists {
			rv.Add(ref)
		}
	}
	return rv
}

func (handler *LineageHandler) parseOpts(u url.Values) lineage.QueryCfg {
	collapse, _ := strconv.ParseBool(u.Get("collapse"))
	return lineage.QueryCfg{
		TypeWhitelist: u["filter.type"],
		Collapse:      collapse,
	}
}

func NewLineageHandler(log logrus.FieldLogger, provider LineageProvider) *LineageHandler {
	handler := &LineageHandler{
		log:             log,
		mux:             mux.NewRouter(),
		lineageProvider: provider,
	}

	handler.mux.PathPrefix("/v1/lineage/{type}/{id}").
		Methods(http.MethodGet).
		HandlerFunc(handler.getLineage)

	handler.mux.PathPrefix("/v1/lineage").
		Methods(http.MethodGet).
		HandlerFunc(handler.listLineage)

	return handler
}
