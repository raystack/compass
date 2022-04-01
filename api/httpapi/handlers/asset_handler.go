package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/odpf/salt/log"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
)

var (
	dataFilterPrefix = "data"
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
	cfg, err := h.buildAssetConfig(r.URL.Query())
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	assets, err := h.assetRepo.GetAll(r.Context(), cfg)
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
			Types:         cfg.Types,
			Services:      cfg.Services,
			Size:          cfg.Size,
			Offset:        cfg.Offset,
			SortBy:        cfg.SortBy,
			SortDirection: cfg.SortDirection,
			QueryFields:   cfg.QueryFields,
			Query:         cfg.Query,
			Data:          cfg.Data,
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

	var payload upsertAssetPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	ast := payload.Asset
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

	if err := h.saveLineage(r.Context(), ast, payload.Upstreams, payload.Downstreams); err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id": ast.ID,
	})
}

func (h *AssetHandler) UpsertPatch(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	var payload patchAssetPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	urn, typ, service, err := h.validatePatchPayload(payload.Asset)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	ast, err := h.assetRepo.Find(r.Context(), urn, asset.Type(typ), service)
	if err != nil && !errors.As(err, &asset.NotFoundError{}) {
		internalServerError(w, h.logger, err.Error())
		return
	}
	ast.Patch(payload.Asset)

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

	if err := h.saveLineage(r.Context(), ast, payload.Upstreams, payload.Downstreams); err != nil {
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
	cfg, err := h.buildAssetConfig(r.URL.Query())
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	pathParams := mux.Vars(r)
	assetID := pathParams["id"]

	assetVersions, err := h.assetRepo.GetVersionHistory(r.Context(), cfg, assetID)
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

func (h *AssetHandler) validatePatchPayload(assetPayload map[string]interface{}) (urn, typ, service string, err error) {
	urnVal, exists := assetPayload["urn"]
	if !exists {
		err = fmt.Errorf("urn is required")
		return
	}
	urn, valid := urnVal.(string)
	if !valid || urn == "" {
		err = fmt.Errorf("urn is invalid")
		return
	}

	typeVal, exists := assetPayload["type"]
	if !exists {
		err = fmt.Errorf("type is required")
		return
	}
	typ, valid = typeVal.(string)
	if !valid || typ == "" || !asset.Type(typ).IsValid() {
		err = fmt.Errorf("type is invalid")
		return
	}

	serviceVal, exists := assetPayload["service"]
	if !exists {
		err = fmt.Errorf("service is required")
		return
	}
	service, valid = serviceVal.(string)
	if !valid || service == "" {
		err = fmt.Errorf("service is invalid")
		return
	}

	return
}

func (h *AssetHandler) buildAssetConfig(query url.Values) (cfg asset.Config, err error) {
	cfg = asset.Config{
		SortBy:        query.Get("sort"),
		SortDirection: query.Get("direction"),
		Query:         query.Get("q"),
	}

	types := query.Get("types")
	if types != "" {
		typ := strings.Split(types, ",")
		for _, typeVal := range typ {
			cfg.Types = append(cfg.Types, asset.Type(typeVal))
		}
	}

	services := query.Get("services")
	if services != "" {
		cfg.Services = strings.Split(services, ",")
	}

	queriesFields := query.Get("q_fields")
	if queriesFields != "" {
		cfg.QueryFields = strings.Split(queriesFields, ",")
	}

	sizeString := query.Get("size")
	if sizeString != "" {
		size, err := strconv.Atoi(sizeString)
		if err == nil {
			cfg.Size = size
		}
	}

	offsetString := query.Get("offset")
	if offsetString != "" {
		offset, err := strconv.Atoi(offsetString)
		if err == nil {
			cfg.Offset = offset
		}
	}

	cfg.Data = dataAssetConfigValue(query)
	cfg.AssignDefault()
	if err = cfg.Validate(); err != nil {
		return asset.Config{}, err
	}

	return cfg, nil
}

func dataAssetConfigValue(queryString url.Values) map[string]string {
	dataFilter := make(map[string]string)
	preChar := "["
	postChar := "]"

	// Get substring between two strings.
	for key, values := range queryString {
		if !strings.HasPrefix(key, dataFilterPrefix) {
			continue
		}

		posFirst := strings.Index(key, preChar)
		if posFirst == -1 {
			return nil
		}
		posLast := strings.Index(key, postChar)
		if posLast == -1 {
			return nil
		}
		posFirstAdjusted := posFirst + len(preChar)
		if posFirstAdjusted >= posLast {
			return nil
		}

		filterKey := key[posFirstAdjusted:posLast]
		dataFilter[filterKey] = values[0] // cannot have duplicate query key, always get the first one
	}

	return dataFilter
}

func (h *AssetHandler) saveLineage(ctx context.Context, ast asset.Asset, upstreams, downstreams []lineage.Node) error {
	node := lineage.Node{
		URN:     ast.URN,
		Type:    ast.Type,
		Service: ast.Service,
	}

	return h.lineageRepo.Upsert(ctx, node, upstreams, downstreams)
}
