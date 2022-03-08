package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/salt/log"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lineage/v2"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
)

// AssetHandler exposes a REST interface to types
type AssetHandler struct {
	logger        log.Logger
	assetRepo     asset.Repository
	discoveryRepo discovery.Repository
	starRepo      star.Repository
	lineageRepo   lineage.Repository
}

func NewAssetHandler(
	logger log.Logger,
	assetRepo asset.Repository,
	discoveryRepo discovery.Repository,
	starRepo star.Repository,
	lineageRepo lineage.Repository,
) *AssetHandler {
	handler := &AssetHandler{
		logger:        logger,
		assetRepo:     assetRepo,
		discoveryRepo: discoveryRepo,
		starRepo:      starRepo,
		lineageRepo:   lineageRepo,
	}

	return handler
}

func (h *AssetHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	config := h.buildAssetConfig(r.URL.Query())
	assets, err := h.assetRepo.GetAll(r.Context(), config)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	payload := map[string]interface{}{
		"data": assets,
	}

	withTotal, ok := r.URL.Query()["with_total"]
	if ok && len(withTotal) > 0 && withTotal[0] != "false" && withTotal[0] != "0" {
		total, err := h.assetRepo.GetCount(r.Context(), asset.Config{
			Type:    config.Type,
			Service: config.Service,
			Text:    config.Text,
		})
		if err != nil {
			internalServerError(w, h.logger, err.Error())
			return
		}
		payload["total"] = total
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *AssetHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID := vars["id"]

	ast, err := h.assetRepo.GetByID(r.Context(), assetID)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(asset.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ast)
}

func (h *AssetHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	var ast asset.Asset
	err := json.NewDecoder(r.Body).Decode(&ast)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}
	if err := h.validateAsset(ast); err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	ast.UpdatedBy.ID = userID
	assetID, err := h.assetRepo.Upsert(r.Context(), &ast)
	if errors.As(err, new(asset.InvalidError)) {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	ast.ID = assetID
	if err := h.discoveryRepo.Upsert(r.Context(), ast); err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	if err := h.saveLineage(r.Context(), ast); err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id": ast.ID,
	})
}

func (h *AssetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID := vars["id"]

	if err := h.assetRepo.Delete(r.Context(), assetID); err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(asset.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	if err := h.discoveryRepo.Delete(r.Context(), assetID); err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *AssetHandler) GetStargazers(w http.ResponseWriter, r *http.Request) {
	starCfg := buildStarConfig(h.logger, r.URL.Query())

	pathParams := mux.Vars(r)
	assetID := pathParams["id"]

	users, err := h.starRepo.GetStargazers(r.Context(), starCfg, assetID)
	if err != nil {
		if errors.Is(err, star.ErrEmptyUserID) || errors.Is(err, star.ErrEmptyAssetID) || errors.As(err, new(star.InvalidError)) {
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

	writeJSON(w, http.StatusOK, users)
}

func (h *AssetHandler) GetVersionHistory(w http.ResponseWriter, r *http.Request) {
	config := h.buildAssetConfig(r.URL.Query())

	pathParams := mux.Vars(r)
	assetID := pathParams["id"]

	assetVersions, err := h.assetRepo.GetVersionHistory(r.Context(), config, assetID)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(asset.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, assetVersions)
}

func (h *AssetHandler) GetByVersion(w http.ResponseWriter, r *http.Request) {

	pathParams := mux.Vars(r)
	assetID := pathParams["id"]
	version := pathParams["version"]

	if _, err := asset.ParseVersion(version); err != nil {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	ast, err := h.assetRepo.GetByVersion(r.Context(), assetID, version)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.As(err, new(asset.NotFoundError)) {
			WriteJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ast)
}

func (h *AssetHandler) validateAsset(ast asset.Asset) error {
	if ast.URN == "" {
		return fmt.Errorf("urn is required")
	}
	if ast.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !ast.Type.IsValid() {
		return fmt.Errorf("type is invalid")
	}
	if ast.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ast.Data == nil {
		return fmt.Errorf("data is required")
	}
	if ast.Service == "" {
		return fmt.Errorf("service is required")
	}

	return nil
}

func (h *AssetHandler) buildAssetConfig(query url.Values) asset.Config {
	config := asset.Config{
		Text:    query.Get("text"),
		Type:    asset.Type(query.Get("type")),
		Service: query.Get("service"),
	}

	sizeString := query.Get("size")
	if sizeString != "" {
		size, err := strconv.Atoi(sizeString)
		if err == nil {
			config.Size = size
		}
	}
	offsetString := query.Get("offset")
	if offsetString != "" {
		offset, err := strconv.Atoi(offsetString)
		if err == nil {
			config.Offset = offset
		}
	}

	return config
}

func (h *AssetHandler) saveLineage(ctx context.Context, ast asset.Asset) error {
	node := lineage.Node{
		URN:     ast.URN,
		Type:    ast.Type,
		Service: ast.Service,
	}

	upstreams := []lineage.Node{}
	for _, n := range ast.Upstreams {
		upstreams = append(upstreams, lineage.Node{
			URN:     n.URN,
			Type:    n.Type,
			Service: n.Service,
		})
	}

	downstreams := []lineage.Node{}
	for _, n := range ast.Downstreams {
		downstreams = append(downstreams, lineage.Node{
			URN:     n.URN,
			Type:    n.Type,
			Service: n.Service,
		})
	}

	return h.lineageRepo.Upsert(ctx, node, upstreams, downstreams)
}
