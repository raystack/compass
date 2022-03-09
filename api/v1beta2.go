package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupV1Beta2Router(router *mux.Router, handlers *Handlers) *mux.Router {
	router.Path("/lineage/{urn}").
		Methods(http.MethodGet).
		HandlerFunc(handlers.LineageV2.GetGraph)

	return router
}
