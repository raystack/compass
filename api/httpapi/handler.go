package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/httpapi/handlers"
)

type Handler struct {
	Asset       *handlers.AssetHandler
	Type        *handlers.TypeHandler
	Record      *handlers.RecordHandler
	Search      *handlers.SearchHandler
	Lineage     *handlers.LineageHandler
	Tag         *handlers.TagHandler
	TagTemplate *handlers.TagTemplateHandler
	User        *handlers.UserHandler
	Discussion  *handlers.DiscussionHandler
}

func RegisterRoutes(router *mux.Router, handlerCollection *Handler) {
	setupV1Beta1Router(router, handlerCollection)

	router.NotFoundHandler = http.HandlerFunc(handlers.NotFound)
	router.MethodNotAllowedHandler = http.HandlerFunc(handlers.MethodNotAllowed)
}
