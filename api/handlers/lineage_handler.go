package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/lineage/v1"
)

// interface to lineage.Service
// named provider to avoid ambiguitity with the service implementation
//go:generate mockery --name LineageProvider --outpkg mocks --output ../../lib/mocks/ --structname LineageProvider --filename lineage_provider.go
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

func (handler *LineageHandler) ListLineage(w http.ResponseWriter, r *http.Request) {
	graph, err := handler.lineageProvider.Graph()
	if err != nil {
		errMessage := fmt.Sprintf("error requesting graph: %s", err.Error())
		internalServerError(w, handler.logger, errMessage)
		return
	}

	opts := handler.parseOpts(r.URL.Query())
	res, err := graph.Query(opts)
	if err != nil {
		handler.logger.Error("error querying graph", "query", opts, "error", err)
		status := http.StatusBadRequest
		if errors.Is(err, asset.ErrUnknownType) {
			status = http.StatusNotFound
		}
		WriteJSONError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (handler *LineageHandler) GetLineage(w http.ResponseWriter, r *http.Request) {
	graph, err := handler.lineageProvider.Graph()
	if err != nil {
		errMessage := fmt.Sprintf("error requesting graph: %s", err.Error())
		internalServerError(w, handler.logger, errMessage)
		return
	}
	requestParams := mux.Vars(r)

	opts := handler.parseOpts(r.URL.Query())
	opts.Root = fmt.Sprintf("%s/%s", requestParams["type"], requestParams["id"])

	res, err := graph.Query(opts)
	if err != nil {
		handler.logger.Error("error querying graph", "query", opts, "error", err)
		status := http.StatusBadRequest
		if errors.Is(err, asset.ErrUnknownType) {
			status = http.StatusNotFound
		}
		WriteJSONError(w, status, err.Error())
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
