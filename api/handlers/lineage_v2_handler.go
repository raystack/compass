package handlers

import (
	"net/http"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/lineage"
)

type LineageHandler struct {
	logger      log.Logger
	lineageRepo lineage.Repository
}

func NewLineageHandler(logger log.Logger, lineageRepo lineage.Repository) *LineageHandler {
	handler := &LineageHandler{
		logger:      logger,
		lineageRepo: lineageRepo,
	}

	return handler
}

func (handler *LineageHandler) GetGraph(w http.ResponseWriter, r *http.Request) {
	urn := mux.Vars(r)["urn"]

	graph, err := handler.lineageRepo.GetGraph(r.Context(), lineage.Node{URN: urn})
	if err != nil {
		internalServerError(w, handler.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, graph)
}
