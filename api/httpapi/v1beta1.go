package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/httpapi/handlers"
)

func setupV1Beta1Router(router *mux.Router, handlers *Handler) {
	setupV1Beta1AssetRoutes("/assets", router, handlers.Asset)
	setupV1Beta1TagRoutes("/tags", router, handlers.Tag, handlers.TagTemplate)

	router.Path("/search").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Search.Search)

	router.Path("/search/suggest").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Search.Suggest)

	router.PathPrefix("/lineage/{urn}").
		Methods(http.MethodGet).
		HandlerFunc(handlers.Lineage.GetGraph)

	// Deprecated: Use setupV1Beta1AssetRoutes instead
	setupV1Beta1TypeRoutes("/types", router, handlers.Type, handlers.Record)

	setupUserRoutes("/user", router, handlers.User)

	setupUsersRoutes("/users", router, handlers.User)

	setupDiscussionsRoutes("/discussions", router, handlers.Discussion)
}

func setupV1Beta1AssetRoutes(baseURL string, router *mux.Router, ah *handlers.AssetHandler) {
	router.Path(baseURL).
		Methods(http.MethodGet).
		HandlerFunc(ah.GetAll)

	router.Path(baseURL).
		Methods(http.MethodPut).
		HandlerFunc(ah.Upsert)

	router.Path(baseURL).
		Methods(http.MethodPatch).
		HandlerFunc(ah.UpsertPatch)

	router.Path(baseURL + "/{id}").
		Methods(http.MethodGet).
		HandlerFunc(ah.GetByID)

	router.Path(baseURL + "/{id}").
		Methods(http.MethodDelete).
		HandlerFunc(ah.Delete)

	router.Path(baseURL + "/{id}/stargazers").
		Methods(http.MethodGet).
		HandlerFunc(ah.GetStargazers)

	router.Path(baseURL + "/{id}/versions").
		Methods(http.MethodGet).
		HandlerFunc(ah.GetVersionHistory)

	router.Path(baseURL + "/{id}/versions/{version}").
		Methods(http.MethodGet).
		HandlerFunc(ah.GetByVersion)
}

func setupV1Beta1TypeRoutes(baseURL string, router *mux.Router, th *handlers.TypeHandler, rh *handlers.RecordHandler) {
	router.Path(baseURL).
		Methods(http.MethodGet).
		HandlerFunc(th.Get)

	recordURL := baseURL + "/{name}/records"
	router.Path(recordURL).
		Methods(http.MethodPut).
		HandlerFunc(rh.UpsertBulk)

	router.Path(recordURL).
		Methods(http.MethodGet).
		HandlerFunc(rh.GetByType)

	router.Path(recordURL + "/{id}").
		Methods(http.MethodGet).
		HandlerFunc(rh.GetOneByType)

	router.Path(recordURL + "/{id}").
		Methods(http.MethodDelete).
		HandlerFunc(rh.Delete)
}

func setupV1Beta1TagRoutes(baseURL string, router *mux.Router, th *handlers.TagHandler, tth *handlers.TagTemplateHandler) {
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

func setupUserRoutes(baseURL string, router *mux.Router, ush *handlers.UserHandler) {

	router.Path(baseURL + "/starred").
		Methods(http.MethodGet).
		HandlerFunc(ush.GetStarredAssetsWithHeader)

	userAssetsURL := baseURL + "/starred/{asset_id}"
	router.Methods(http.MethodPut).Path(userAssetsURL).HandlerFunc(ush.StarAsset)
	router.Methods(http.MethodGet).Path(userAssetsURL).HandlerFunc(ush.GetStarredAsset)
	router.Methods(http.MethodDelete).Path(userAssetsURL).HandlerFunc(ush.UnstarAsset)

	router.Path(baseURL + "/discussions").
		Methods(http.MethodGet).
		HandlerFunc(ush.GetDiscussions)
}

func setupUsersRoutes(baseURL string, router *mux.Router, ush *handlers.UserHandler) {

	router.Path(baseURL + "/{user_id}/starred").
		Methods(http.MethodGet).
		HandlerFunc(ush.GetStarredAssetsWithPath)
}

func setupDiscussionsRoutes(baseURL string, router *mux.Router, dh *handlers.DiscussionHandler) {
	router.Path(baseURL).
		Methods(http.MethodPost).
		HandlerFunc(dh.Create)

	router.Path(baseURL).
		Methods(http.MethodGet).
		HandlerFunc(dh.GetAll)

	router.Path(baseURL + "/{id}").
		Methods(http.MethodGet).
		HandlerFunc(dh.Get)

	router.Path(baseURL + "/{id}").
		Methods(http.MethodPatch).
		HandlerFunc(dh.Patch)

	commentURL := baseURL + "/{discussion_id}/comments"
	router.Path(commentURL).
		Methods(http.MethodPost).
		HandlerFunc(dh.CreateComment)

	router.Path(commentURL).
		Methods(http.MethodGet).
		HandlerFunc(dh.GetAllComments)

	router.Path(commentURL + "/{id}").
		Methods(http.MethodGet).
		HandlerFunc(dh.GetComment)

	router.Path(commentURL + "/{id}").
		Methods(http.MethodPut).
		HandlerFunc(dh.UpdateComment)

	router.Path(commentURL + "/{id}").
		Methods(http.MethodDelete).
		HandlerFunc(dh.DeleteComment)
}
