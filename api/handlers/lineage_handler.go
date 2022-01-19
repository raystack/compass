package handlers

import (
	"fmt"
	"github.com/odpf/salt/log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/record"
)

// interface to lineage.Service
// named provider to avoid ambiguitity with the service implementation
type LineageProvider interface {
	Graph() (lineage.Graph, error)
}

type LineageHandler struct {
	log             log.Logger
	lineageProvider LineageProvider
}

func NewLineageHandler(log log.Logger, provider LineageProvider) *LineageHandler {
	handler := &LineageHandler{
		log:             log,
		lineageProvider: provider,
	}

	return handler
}

func (handler *LineageHandler) ListLineage(w http.ResponseWriter, r *http.Request) {
	graph, err := handler.lineageProvider.Graph()
	if err != nil {
		handler.log.Error("failed to request graph", "error", err)

		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}

	opts := handler.parseOpts(r.URL.Query())
	res, err := graph.Query(opts)
	if err != nil {
		handler.log.Error("failed to query graph", "query", opts, "error", err)
		status := http.StatusBadRequest
		if _, ok := err.(record.ErrNoSuchType); ok {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (handler *LineageHandler) GetLineage(w http.ResponseWriter, r *http.Request) {
	graph, err := handler.lineageProvider.Graph()
	if err != nil {
		handler.log.Error("failed to request graph", "error", err)
		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	requestParams := mux.Vars(r)

	opts := handler.parseOpts(r.URL.Query())
	opts.Root = fmt.Sprintf("%s/%s", requestParams["type"], requestParams["id"])

	res, err := graph.Query(opts)
	if err != nil {
		handler.log.Error("failed to query graph", "query", opts, "error", err)
		status := http.StatusBadRequest
		if _, ok := err.(record.ErrNoSuchType); ok {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (handler *LineageHandler) parseOpts(u url.Values) lineage.QueryCfg {
	collapse, _ := strconv.ParseBool(u.Get("collapse"))
	return lineage.QueryCfg{
		TypeWhitelist: u["filter.type"],
		Collapse:      collapse,
	}
}
