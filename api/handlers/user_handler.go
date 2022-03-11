package handlers

import (
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
)

// UserHandler exposes a REST interface to user
type UserHandler struct {
	starRepo       star.Repository
	discussionRepo discussion.Repository
	logger         log.Logger
}

func (h *UserHandler) GetStarredAssetsWithHeader(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	starCfg := buildStarConfig(h.logger, r.URL.Query())

	starredAssets, err := h.starRepo.GetAllAssetsByUserID(r.Context(), starCfg, userID)
	if errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.As(err, new(star.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, starredAssets)
}

func (h *UserHandler) GetStarredAssetsWithPath(w http.ResponseWriter, r *http.Request) {
	targetUserID := mux.Vars(r)["user_id"]
	if targetUserID == "" {
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	starCfg := buildStarConfig(h.logger, r.URL.Query())

	var starredAssets []asset.Asset

	//TODO: might want to remove get by email flow in the future
	// we can use user id or user email
	// get by email is a temporary flow and might be deleted in the future version
	// once we already introduce better solution (e.g. get by user name)
	_, err := mail.ParseAddress(targetUserID)
	if err == nil {
		starredAssets, err = h.starRepo.GetAllAssetsByUserEmail(r.Context(), starCfg, targetUserID)
	} else {
		starredAssets, err = h.starRepo.GetAllAssetsByUserID(r.Context(), starCfg, targetUserID)
	}
	if errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.As(err, new(star.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, starredAssets)
}

func (h *UserHandler) StarAsset(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	assetID := pathParams["asset_id"]

	starID, err := h.starRepo.Create(r.Context(), userID, assetID)
	if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.As(err, new(star.UserNotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.As(err, new(star.DuplicateRecordError)) {
		// idempotent
		writeJSON(w, http.StatusNoContent, starID)
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, starID)
}

func (h *UserHandler) GetStarredAsset(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	assetID := pathParams["asset_id"]

	ast, err := h.starRepo.GetAssetByUserID(r.Context(), userID, assetID)
	if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.As(err, new(star.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ast)
}

func (h *UserHandler) UnstarAsset(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	assetID := pathParams["asset_id"]

	err := h.starRepo.Delete(r.Context(), userID, assetID)
	if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.As(err, new(star.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, "success")
}

func (h *UserHandler) GetDiscussions(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	flt, err := h.buildGetDiscussionsFilter(r.URL.Query(), userID)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	dscs, err := h.discussionRepo.GetAll(r.Context(), flt)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dscs)
}

func (h *UserHandler) buildGetDiscussionsFilter(query url.Values, userID string) (discussion.Filter, error) {
	fl := discussion.Filter{
		Type:          query.Get("type"),
		State:         query.Get("state"),
		SortBy:        query.Get("sort"),
		SortDirection: query.Get("direction"),
	}

	filterQuery := "assigned"
	if len(strings.TrimSpace(query.Get("filter"))) > 0 {
		tempFilterQuery := query.Get("filter")
		if tempFilterQuery == "created" || tempFilterQuery == "all" {
			filterQuery = tempFilterQuery
		}
	}

	switch filterQuery {
	case "all":
		fl.Owner = userID
		fl.Assignees = []string{userID}
	case "created":
		fl.Owner = userID
	default:
		fl.Assignees = []string{userID}
	}

	assets := query.Get("asset")
	if assets != "" {
		fl.Assets = strings.Split(assets, ",")
	}

	labels := query.Get("labels")
	if labels != "" {
		fl.Labels = strings.Split(labels, ",")
	}

	sizeString := query.Get("size")
	if sizeString != "" {
		size, err := strconv.Atoi(sizeString)
		if err == nil {
			fl.Size = size
		}
	}

	offsetString := query.Get("offset")
	if offsetString != "" {
		offset, err := strconv.Atoi(offsetString)
		if err == nil {
			fl.Offset = offset
		}
	}

	if err := fl.Validate(); err != nil {
		return discussion.Filter{}, err
	}

	fl.AssignDefault()
	fl.DisjointAssigneeOwner = true

	return fl, nil
}

func NewUserHandler(
	logger log.Logger,
	starRepo star.Repository,
	discussionRepo discussion.Repository) *UserHandler {
	h := &UserHandler{
		logger:         logger,
		starRepo:       starRepo,
		discussionRepo: discussionRepo,
	}
	return h
}
