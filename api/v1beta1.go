package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
)

func setupV1Beta1Router(router *mux.Router, handlers *Handlers) *mux.Router {
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

	// Deprecated: Use setupV1Beta1AssetRoutes instead
	setupV1Beta1TypeRoutes(router, handlers.Type, handlers.Record)

	return router
}

func setupV1Beta1AssetRoutes(router *mux.Router, ah *handlers.AssetHandler) {
	url := "/assets"

	router.Path(url).
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(ah.Get)

	router.Path(url).
		Methods(http.MethodPut, http.MethodHead).
		HandlerFunc(ah.Upsert)

	router.Path(url+"/{id}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(ah.GetByID)

	router.Path(url+"/{id}").
		Methods(http.MethodDelete, http.MethodHead).
		HandlerFunc(ah.Delete)
}

func setupV1Beta1TypeRoutes(router *mux.Router, th *handlers.TypeHandler, rh *handlers.RecordHandler) {
	typeURL := "/types"

	router.Path(typeURL).
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(th.Get)

	recordURL := "/types/{name}/records"
	router.Path(recordURL).
		Methods(http.MethodPut, http.MethodHead).
		HandlerFunc(rh.UpsertBulk)

	router.Path(recordURL).
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(rh.GetByType)

	router.Path(recordURL+"/{id}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(rh.GetOneByType)

	router.Path(recordURL+"/{id}").
		Methods(http.MethodDelete, http.MethodHead).
		HandlerFunc(rh.Delete)
}

func setupV1Beta1TagRoutes(router *mux.Router, baseURL string, th *handlers.TagHandler, tth *handlers.TagTemplateHandler) {
	router.Methods(http.MethodPost).Path(baseURL).HandlerFunc(th.Create)

	url := baseURL + "/types/{type}/records/{record_urn}/templates/{template_urn}"
	router.Methods(http.MethodGet).Path(url).HandlerFunc(th.FindByRecordAndTemplate)
	router.Methods(http.MethodPut).Path(url).HandlerFunc(th.Update)
	router.Methods(http.MethodDelete).Path(url).HandlerFunc(th.Delete)

	router.Methods(http.MethodGet).Path(baseURL + "/types/{type}/records/{record_urn}").HandlerFunc(th.GetByRecord)

	templateURL := baseURL + "/templates"
	router.Methods(http.MethodGet).Path(templateURL).HandlerFunc(tth.Index)
	router.Methods(http.MethodPost).Path(templateURL).HandlerFunc(tth.Create)
	router.Methods(http.MethodGet).Path(templateURL + "/{template_urn}").HandlerFunc(tth.Find)
	router.Methods(http.MethodPut).Path(templateURL + "/{template_urn}").HandlerFunc(tth.Update)
	router.Methods(http.MethodDelete).Path(templateURL + "/{template_urn}").HandlerFunc(tth.Delete)

}
