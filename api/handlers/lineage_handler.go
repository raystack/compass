package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/record"
	"github.com/sirupsen/logrus"
)

// interface to lineage.Service
// named provider to avoid ambiguitity with the service implementation
type LineageProvider interface {
	Graph() (lineage.Graph, error)
}

type LineageHandler struct {
	log             logrus.FieldLogger
	lineageProvider LineageProvider
}

func NewLineageHandler(log logrus.FieldLogger, provider LineageProvider) *LineageHandler {
	handler := &LineageHandler{
		log:             log,
		lineageProvider: provider,
	}

	return handler
}

func (handler *LineageHandler) ListLineage(w http.ResponseWriter, r *http.Request) {
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
		handler.log.Errorf("error requesting graph: %v", err)
		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	requestParams := mux.Vars(r)

	opts := handler.parseOpts(r.URL.Query())
	opts.Root = fmt.Sprintf("%s/%s", requestParams["type"], requestParams["id"])

	res, err := graph.Query(opts)
	if err != nil {
		handler.log.
			WithField("query", opts).
			Errorf("error querying graph: %v\ncfg: %v", err, opts)

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
