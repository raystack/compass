package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
)

// UserHandler exposes a REST interface to user
type UserHandler struct {
	service *star.Service
	logger  log.Logger
}

func (h *UserHandler) GetStarredAssetsWithHeader(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserID.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserID.Error())
		return
	}

	starCfg := buildStarConfig(h.logger, r.URL.Query())

	starredAssets, err := h.service.GetAllAssetsByUserID(r.Context(), starCfg, userID)
	if err != nil {
		if errors.As(err, new(star.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(star.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, starredAssets)
}

func (h *UserHandler) GetStarredAssetsWithPath(w http.ResponseWriter, r *http.Request) {
	targetUserID := mux.Vars(r)["user_id"]
	if targetUserID == "" {
		WriteJSONError(w, http.StatusBadRequest, errMissingUserID.Error())
		return
	}

	starCfg := buildStarConfig(h.logger, r.URL.Query())

	starredAssets, err := h.service.GetAllAssetsByUserID(r.Context(), starCfg, targetUserID)
	if err != nil {
		if errors.As(err, new(star.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(star.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, starredAssets)
}

func (h *UserHandler) StarAsset(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserID.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserID.Error())
		return
	}

	starring := buildStarFromPath(mux.Vars(r))

	starID, err := h.service.Star(r.Context(), userID, starring)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) || errors.As(err, new(star.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(asset.NotFoundError)) || errors.As(err, new(star.UserNotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.As(err, new(star.DuplicateRecordError)) {
			// idempotent
			writeJSON(w, http.StatusNoContent, starID)
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, starID)
}

func (h *UserHandler) GetStarredAsset(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserID.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserID.Error())
		return
	}

	starring := buildStarFromPath(mux.Vars(r))

	starID, err := h.service.GetAssetByUserID(r.Context(), userID, starring)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) || errors.As(err, new(star.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(asset.NotFoundError)) || errors.As(err, new(star.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, starID)
}

func (h *UserHandler) UnstarAsset(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserID.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserID.Error())
		return
	}

	starring := buildStarFromPath(mux.Vars(r))

	err := h.service.Unstar(r.Context(), userID, starring)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) || errors.As(err, new(star.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(asset.NotFoundError)) || errors.As(err, new(star.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, "success")
}

func NewUserHandler(logger log.Logger, svc *star.Service) *UserHandler {
	h := &UserHandler{
		service: svc,
		logger:  logger,
	}
	return h
}
