package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/odpf/salt/log"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
)

// RecordHandler exposes a REST interface to types
type RecordHandler struct {
	typeRepository          discovery.TypeRepository
	recordRepositoryFactory discovery.RecordRepositoryFactory
	discoveryService        *discovery.Service
	logger                  log.Logger
}

func NewRecordHandler(
	logger log.Logger,
	typeRepository discovery.TypeRepository,
	discoveryService *discovery.Service,
	rrf discovery.RecordRepositoryFactory) *RecordHandler {
	handler := &RecordHandler{
		recordRepositoryFactory: rrf,
		discoveryService:        discoveryService,
		typeRepository:          typeRepository,
		logger:                  logger,
	}

	return handler
}

func (h *RecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		typeName   = vars["name"]
		recordID   = vars["id"]
		statusCode = http.StatusInternalServerError
		errMessage = fmt.Sprintf("error deleting record \"%s\" with type \"%s\"", recordID, typeName)
	)

	typName := asset.Type(typeName)
	if !typName.IsValid() {
		writeJSONError(w, http.StatusNotFound, "type is invalid")
		return
	}

	err := h.discoveryService.DeleteRecord(r.Context(), typName.String(), recordID)
	if err != nil {
		h.logger.Error("error deleting record", "type", typName, "error", err)

		if _, ok := err.(asset.NotFoundError); ok {
			statusCode = http.StatusNotFound
			errMessage = err.Error()
		}

		writeJSONError(w, statusCode, errMessage)
		return
	}

	h.logger.Info("deleted record", "record id", recordID, "type", typName)
	writeJSON(w, http.StatusNoContent, "success")
}

func (h *RecordHandler) UpsertBulk(w http.ResponseWriter, r *http.Request) {
	typeName := mux.Vars(r)["name"]

	var assets []asset.Asset
	err := json.NewDecoder(r.Body).Decode(&assets)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	typName := asset.Type(typeName)
	if !typName.IsValid() {
		writeJSONError(w, http.StatusNotFound, "type is invalid")
		return
	}

	var failedAssets = make(map[int]string)
	for idx, record := range assets {
		if err := h.validateRecord(record); err != nil {
			h.logger.Error("error validating record", "type", typName, "record", record, "error", err)
			failedAssets[idx] = err.Error()
		}
	}
	if len(failedAssets) > 0 {
		writeJSON(w, http.StatusBadRequest, NewValidationErrorResponse(failedAssets))
		return
	}

	if err := h.discoveryService.Upsert(r.Context(), typName.String(), assets); err != nil {
		h.logger.Error("error creating/updating assets", "type", typName, "error", err)
		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	h.logger.Info("created/updated assets", "record count", len(assets), "type", typName)
	writeJSON(w, http.StatusOK, StatusResponse{Status: "success"})
}

func (h *RecordHandler) GetByType(w http.ResponseWriter, r *http.Request) {
	typeName := mux.Vars(r)["name"]

	typName := asset.Type(typeName)
	if !typName.IsValid() {
		writeJSONError(w, http.StatusNotFound, "type is invalid")
		return
	}

	recordRepo, err := h.recordRepositoryFactory.For(typName.String())
	if err != nil {
		h.logger.Error("error constructing record repository", "type", typName, "error", err)
		status, message := h.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}
	getCfg, err := h.buildGetConfig(r.URL.Query())
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	recordList, err := recordRepo.GetAll(r.Context(), getCfg)
	if err != nil {
		h.logger.Error("error fetching assets: GetAll", "type", typName, "error", err)
		status, message := h.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}

	fieldsToSelect := h.parseSelectQuery(r.URL.Query().Get("select"))
	if len(fieldsToSelect) > 0 {
		recordList.Data = h.selectRecordFields(fieldsToSelect, recordList.Data)
	}
	writeJSON(w, http.StatusOK, recordList)
}

func (h *RecordHandler) GetOneByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		typeName = vars["name"]
		recordID = vars["id"]
	)

	typName := asset.Type(typeName)
	if !typName.IsValid() {
		writeJSONError(w, http.StatusNotFound, "type is invalid")
		return
	}

	recordRepo, err := h.recordRepositoryFactory.For(typName.String())
	if err != nil {
		h.logger.Error("internal: error construing record repository", "type", typName, "error", err)
		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}

	record, err := recordRepo.GetByID(r.Context(), recordID)
	if err != nil {
		h.logger.Error("error fetching record", "type", typName, "record id", recordID, "error", err)
		status, message := h.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (h *RecordHandler) buildGetConfig(params url.Values) (cfg discovery.GetConfig, err error) {
	if size := params.Get("size"); size != "" {
		cfg.Size, err = strconv.Atoi(size)
		if err != nil {
			return
		}
	}
	if from := params.Get("from"); from != "" {
		cfg.From, err = strconv.Atoi(from)
		if err != nil {
			return
		}
	}

	cfg.Filters = filterConfigFromValues(params)

	return
}

func (h *RecordHandler) parseSelectQuery(raw string) (fields []string) {
	tokens := strings.Split(raw, ",")
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		fields = append(fields, token)
	}
	return
}

func (h *RecordHandler) selectRecordFields(fields []string, assets []asset.Asset) (processedAssets []asset.Asset) {
	for _, ast := range assets {
		newData := map[string]interface{}{}
		for _, field := range fields {
			v, ok := ast.Data[field]
			if !ok {
				continue
			}
			newData[field] = v
		}
		ast.Data = newData
		processedAssets = append(processedAssets, ast)
	}
	return
}

func (h *RecordHandler) validateRecord(ast asset.Asset) error {
	if ast.URN == "" {
		return fmt.Errorf("urn is required")
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

func (h *RecordHandler) responseStatusForError(err error) (int, string) {
	switch err.(type) {
	case asset.NotFoundError:
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}

func bodyParserErrorMsg(err error) string {
	return fmt.Sprintf("error parsing request body: %v", err)
}
