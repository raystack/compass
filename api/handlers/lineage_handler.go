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

	query := r.URL.Query()
	opts, err := handler.parseOpts(query)
	if err != nil {
		handler.log.
			WithField("query", query).
			Error(err)

		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	res, err := graph.Query(opts)
	if err != nil {
		handler.log.
			WithField("query", opts).
			Errorf("error querying graph: %v\ncfg: %v", err, opts)

		writeJSONError(w, http.StatusBadRequest, err.Error())
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

	query := r.URL.Query()
	opts, err := handler.parseOpts(query)
	if err != nil {
		handler.log.
			WithField("query", query).
			Error(err)

		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	opts.Root = fmt.Sprintf("%s/%s", requestParams["type"], requestParams["id"])

	res, err := graph.Query(opts)
	if err != nil {
		handler.log.
			WithField("query", opts).
			Errorf("error querying graph: %v\ncfg: %v", err, opts)

		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (handler *LineageHandler) parseOpts(u url.Values) (cfg lineage.QueryCfg, err error) {
	collapse, _ := strconv.ParseBool(u.Get("collapse"))
	filterTypes := u["filter.type"]
	var types []record.Type
	for _, ft := range filterTypes {
		var t record.Type
		t, err = validateType(ft)
		if err != nil {
			return
		}
		types = append(types, t)
	}

	cfg = lineage.QueryCfg{
		TypeWhitelist: types,
		Collapse:      collapse,
	}

	return
}
