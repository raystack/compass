package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/odpf/salt/log"

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
		logger:                  logger,
		discoveryService:        discoveryService,
		typeRepository:          typeRepository,
		recordRepositoryFactory: rrf,
	}

	return handler
}

func (h *RecordHandler) GetByType(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	typeName := pathParams["name"]

	typName := asset.Type(typeName)
	if !typName.IsValid() {
		WriteJSONError(w, http.StatusNotFound, "type is invalid")
		return
	}

	recordRepo, err := h.recordRepositoryFactory.For(typName.String())
	if err != nil {
		h.logger.Error("error constructing record repository", "type", typName, "error", err)
		status, message := h.responseStatusForError(err)
		WriteJSONError(w, status, message)
		return
	}
	getCfg, err := h.buildGetConfig(r.URL.Query())
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	recordList, err := recordRepo.GetAll(r.Context(), getCfg)
	if err != nil {
		h.logger.Error("error fetching assets: GetAll", "type", typName, "error", err)
		status, message := h.responseStatusForError(err)
		WriteJSONError(w, status, message)
		return
	}

	fieldsToSelect := h.parseSelectQuery(r.URL.Query().Get("select"))
	if len(fieldsToSelect) > 0 {
		recordList.Data = h.selectRecordFields(fieldsToSelect, recordList.Data)
	}
	writeJSON(w, http.StatusOK, recordList)
}

func (h *RecordHandler) GetOneByType(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var (
		typeName = pathParams["name"]
		recordID = pathParams["id"]
	)

	typName := asset.Type(typeName)
	if !typName.IsValid() {
		WriteJSONError(w, http.StatusNotFound, "type is invalid")
		return
	}

	recordRepo, err := h.recordRepositoryFactory.For(typName.String())
	if err != nil {
		h.logger.Error("internal: error construing record repository", "type", typName, "error", err)
		status := http.StatusInternalServerError
		WriteJSONError(w, status, http.StatusText(status))
		return
	}

	record, err := recordRepo.GetByID(r.Context(), recordID)
	if err != nil {
		h.logger.Error("error fetching record", "type", typName, "record id", recordID, "error", err)
		status, message := h.responseStatusForError(err)
		WriteJSONError(w, status, message)
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

func (h *RecordHandler) responseStatusForError(err error) (int, string) {
	switch err.(type) {
	case asset.NotFoundError:
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}
