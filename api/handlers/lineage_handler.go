package handlers

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/lineage"
)

// interface to lineage.Service
// named provider to avoid ambiguitity with the service implementation
type LineageProvider interface {
	Graph() (lineage.Graph, error)
}

type LineageHandler struct {
	logger          log.Logger
	lineageProvider LineageProvider
}

func NewLineageHandler(logger log.Logger, provider LineageProvider) *LineageHandler {
	handler := &LineageHandler{
		logger:          logger,
		lineageProvider: provider,
	}

	return handler
}

func (handler *LineageHandler) GetLineage(w http.ResponseWriter, r *http.Request) {
	graph, err := handler.lineageProvider.Graph()
	if err != nil {
		handler.logger.Error("error requesting graph", "error", err)
		status := http.StatusInternalServerError
		WriteJSONError(w, status, http.StatusText(status))
		return
	}
	requestParams := mux.Vars(r)

	opts := handler.parseOpts(r.URL.Query())
	opts.Root = requestParams["id"]

	res, err := graph.Query(opts)
	if err != nil {
		handler.logger.Error("error querying graph", "query", opts, "error", err)
		status := http.StatusBadRequest
		WriteJSONError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (handler *LineageHandler) parseOpts(u url.Values) lineage.QueryCfg {
	collapse, _ := strconv.ParseBool(u.Get("collapse"))
	return lineage.QueryCfg{
		Collapse: collapse,
	}
}
