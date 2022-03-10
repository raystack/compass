package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupV1Router(router *mux.Router, handlers *Handlers) {
	setupV1Beta1AssetRoutes(router, handlers.Asset)
	setupV1Beta1TagRoutes(router, "/tags", handlers.Tag, handlers.TagTemplate)

	router.Path("/search").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Search.Search)

	router.Path("/search/suggest").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Search.Suggest)

	router.PathPrefix("/lineage/{id}").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Lineage.GetLineage)

	// Deprecated: This route will be removed in the future.
	// Use /lineage/{id} instead
	router.PathPrefix("/lineage/{type}/{id}").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Lineage.GetLineage)

	// Deprecated: This route will be removed in the future.
	router.PathPrefix("/lineage").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Lineage.ListLineage)

	// Deprecated: Use setupV1Beta1AssetRoutes instead
	setupV1Beta1TypeRoutes(router, handlers.Type, handlers.Record)

	userRouter := router.PathPrefix("/user").Subrouter()
	setupUserRoutes(userRouter, handlers.User)

	usersRouter := router.PathPrefix("/users").Subrouter()
	setupUsersRoutes(usersRouter, handlers.User)
}
