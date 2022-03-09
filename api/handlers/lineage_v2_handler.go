package handlers

import (
	"net/http"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/lineage/v2"
)

type LineageV2Handler struct {
	logger      log.Logger
	lineageRepo lineage.Repository
}

func NewLineageV2Handler(logger log.Logger, lineageRepo lineage.Repository) *LineageV2Handler {
	handler := &LineageV2Handler{
		logger:      logger,
		lineageRepo: lineageRepo,
	}

	return handler
}

func (handler *LineageV2Handler) GetGraph(w http.ResponseWriter, r *http.Request) {
	urn := mux.Vars(r)["urn"]

	graph, err := handler.lineageRepo.GetGraph(r.Context(), lineage.Node{URN: urn})
	if err != nil {
		internalServerError(w, handler.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, graph)
}
