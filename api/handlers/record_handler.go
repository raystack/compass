package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/odpf/salt/log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/record"
)

// RecordHandler exposes a REST interface to types
type RecordHandler struct {
	typeRepository          record.TypeRepository
	recordRepositoryFactory discovery.RecordRepositoryFactory
	discoveryService        *discovery.Service
	log                     log.Logger
}

func NewRecordHandler(
	log log.Logger,
	typeRepository record.TypeRepository,
	discoveryService *discovery.Service,
	rrf discovery.RecordRepositoryFactory) *RecordHandler {
	handler := &RecordHandler{
		recordRepositoryFactory: rrf,
		discoveryService:        discoveryService,
		typeRepository:          typeRepository,
		log:                     log,
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

	typName := record.TypeName(typeName)
	if err := typName.IsValid(); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	err := h.discoveryService.DeleteRecord(r.Context(), typName.String(), recordID)
	if err != nil {
		h.log.Error("error deleting record", "type", typName, "error", err)

		if _, ok := err.(record.ErrNoSuchRecord); ok {
			statusCode = http.StatusNotFound
			errMessage = err.Error()
		}

		writeJSONError(w, statusCode, errMessage)
		return
	}

	h.log.Info("deleted record", "record id", recordID, "type", typName)
	writeJSON(w, http.StatusNoContent, "success")
}

func (h *RecordHandler) UpsertBulk(w http.ResponseWriter, r *http.Request) {
	typeName := mux.Vars(r)["name"]

	var records []record.Record
	err := json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	typName := record.TypeName(typeName)
	if err := typName.IsValid(); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	var failedRecords = make(map[int]string)
	for idx, record := range records {
		if err := h.validateRecord(record); err != nil {
			h.log.Error("failed to validate record", "type", typName, "record", record, "error", err)
			failedRecords[idx] = err.Error()
		}
	}
	if len(failedRecords) > 0 {
		writeJSON(w, http.StatusBadRequest, NewValidationErrorResponse(failedRecords))
		return
	}

	if err := h.discoveryService.Upsert(r.Context(), typName.String(), records); err != nil {
		h.log.Error("error creating/updating records", "type", typName, "error", err)
		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	h.log.Info("created/updated records", "record count", len(records), "type", typName)
	writeJSON(w, http.StatusOK, StatusResponse{Status: "success"})
}

func (h *RecordHandler) GetByType(w http.ResponseWriter, r *http.Request) {
	typeName := mux.Vars(r)["name"]

	typName := record.TypeName(typeName)
	if err := typName.IsValid(); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	recordRepo, err := h.recordRepositoryFactory.For(typName.String())
	if err != nil {
		h.log.Error("failed to construct record repository", "type", typName, "error", err)
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
		h.log.Error("failed to fetch records: GetAll", "type", typName, "error", err)
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

	typName := record.TypeName(typeName)
	if err := typName.IsValid(); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	recordRepo, err := h.recordRepositoryFactory.For(typName.String())
	if err != nil {
		h.log.Error("internal: failed to construe record repository", "type", typName, "error", err)
		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}

	record, err := recordRepo.GetByID(r.Context(), recordID)
	if err != nil {
		h.log.Error("failed to fetch record", "type", typName, "record id", recordID, "error", err)
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

func (h *RecordHandler) selectRecordFields(fields []string, records []record.Record) (processedRecords []record.Record) {
	for _, record := range records {
		newData := map[string]interface{}{}
		for _, field := range fields {
			v, ok := record.Data[field]
			if !ok {
				continue
			}
			newData[field] = v
		}
		record.Data = newData
		processedRecords = append(processedRecords, record)
	}
	return
}

func (h *RecordHandler) validateRecord(record record.Record) error {
	if record.Urn == "" {
		return fmt.Errorf("urn is required")
	}
	if record.Name == "" {
		return fmt.Errorf("name is required")
	}
	if record.Data == nil {
		return fmt.Errorf("data is required")
	}
	if record.Service == "" {
		return fmt.Errorf("service is required")
	}

	return nil
}

func (h *RecordHandler) responseStatusForError(err error) (int, string) {
	switch err.(type) {
	case record.ErrNoSuchRecord:
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}

func bodyParserErrorMsg(err error) string {
	return fmt.Sprintf("error parsing request body: %v", err)
}
