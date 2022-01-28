package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/salt/log"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
)

// AssetHandler exposes a REST interface to types
type AssetHandler struct {
	assetRepository asset.Repository
	discoveryRepo   discovery.Repository
	logger          log.Logger
}

func NewAssetHandler(
	logger log.Logger,
	assetRepository asset.Repository,
	discoveryRepo discovery.Repository) *AssetHandler {
	handler := &AssetHandler{
		assetRepository: assetRepository,
		discoveryRepo:   discoveryRepo,
		logger:          logger,
	}

	return handler
}

func (h *AssetHandler) Get(w http.ResponseWriter, r *http.Request) {
	config := h.buildGetConfig(r.URL.Query())
	assets, err := h.assetRepository.Get(r.Context(), config)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	payload := map[string]interface{}{
		"data": assets,
	}

	withTotal, ok := r.URL.Query()["with_total"]
	if ok && len(withTotal) > 0 && withTotal[0] != "false" && withTotal[0] != "0" {
		total, err := h.assetRepository.GetCount(r.Context(), asset.Config{
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

	ast, err := h.assetRepository.GetByID(r.Context(), assetID)
	if err != nil {
		if _, ok := err.(asset.NotFoundError); ok {
			writeJSON(w, http.StatusNotFound, err.Error())
		} else {
			internalServerError(w, h.logger, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, ast)
}

func (h *AssetHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var ast asset.Asset
	err := json.NewDecoder(r.Body).Decode(&ast)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}
	if err := h.validateAsset(ast); err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.assetRepository.Upsert(r.Context(), &ast); err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}
	if err := h.discoveryRepo.Upsert(r.Context(), ast); err != nil {
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

	if err := h.assetRepository.Delete(r.Context(), assetID); err != nil {
		if _, ok := err.(asset.NotFoundError); ok {
			writeJSON(w, http.StatusNotFound, err.Error())
		} else {
			internalServerError(w, h.logger, err.Error())
		}
		return
	}

	if err := h.discoveryRepo.Delete(r.Context(), assetID); err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *AssetHandler) validateAssets(assets []asset.Asset) (faileds map[int]string) {
	faileds = make(map[int]string)
	for idx, ast := range assets {
		if err := h.validateAsset(ast); err != nil {
			h.logger.Error("error validating asset", "asset", ast, "error", err)
			faileds[idx] = err.Error()
		}
	}

	return
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

func (h *AssetHandler) buildGetConfig(query url.Values) asset.Config {
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
