package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/record"
	"github.com/sirupsen/logrus"
)

// RecordHandler exposes a REST interface to types
type RecordHandler struct {
	recordRepositoryFactory discovery.RecordRepositoryFactory
	discoveryService        *discovery.Service
	log                     logrus.FieldLogger
}

func NewRecordHandler(log logrus.FieldLogger, discoveryService *discovery.Service, rrf discovery.RecordRepositoryFactory) *RecordHandler {
	handler := &RecordHandler{
		recordRepositoryFactory: rrf,
		discoveryService:        discoveryService,
		log:                     log,
	}

	return handler
}

func (handler *RecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		typeName   = vars["name"]
		recordID   = vars["id"]
		statusCode = http.StatusInternalServerError
		errMessage = fmt.Sprintf("error deleting record \"%s\" with type \"%s\"", recordID, typeName)
	)

	typ, err := validateType(typeName)
	if err != nil {
		handler.log.WithField("type", typeName).
			Error(err)
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	err = handler.discoveryService.DeleteRecord(r.Context(), typ, recordID)
	if err != nil {
		handler.log.
			Errorf("error deleting record \"%s\": %v", typ, err)

		if _, ok := err.(record.ErrNoSuchRecord); ok {
			statusCode = http.StatusNotFound
			errMessage = err.Error()
		}

		writeJSONError(w, statusCode, errMessage)
		return
	}

	handler.log.Infof("deleted record \"%s\" with type \"%s\"", recordID, typeName)
	writeJSON(w, http.StatusNoContent, "success")
}

func (handler *RecordHandler) UpsertBulk(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	typ, err := validateType(name)
	if err != nil {
		handler.log.WithField("type", name).
			Error(err)
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	var records []record.Record
	err = json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	var failedRecords = make(map[int]string)
	for idx, record := range records {
		if err := handler.validateRecord(record); err != nil {
			handler.log.WithField("type", typ).
				WithField("record", record).
				Errorf("error validating record: %v", err)
			failedRecords[idx] = err.Error()
		}
	}
	if len(failedRecords) > 0 {
		writeJSON(w, http.StatusBadRequest, NewValidationErrorResponse(failedRecords))
		return
	}

	if err := handler.discoveryService.Upsert(r.Context(), typ, records); err != nil {
		handler.log.WithField("type", typ).
			Errorf("error creating/updating records: %v", err)

		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	handler.log.Infof("created/updated %d records for %q type", len(records), typ)
	writeJSON(w, http.StatusOK, StatusResponse{Status: "success"})
}

func (handler *RecordHandler) GetByType(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	typ, err := validateType(name)
	if err != nil {
		handler.log.WithField("type", name).
			Error(err)
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	recordRepo, err := handler.recordRepositoryFactory.For(typ)
	if err != nil {
		handler.log.WithField("type", typ).
			Errorf("error constructing record repository: %v", err)
		status, message := handler.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}
	filterCfg := filterConfigFromValues(r.URL.Query())

	records, err := recordRepo.GetAll(r.Context(), filterCfg)
	if err != nil {
		handler.log.WithField("type", typ).
			Errorf("error fetching records: GetAll: %v", err)
		status, message := handler.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}

	fieldsToSelect := handler.parseSelectQuery(r.URL.Query().Get("select"))
	if len(fieldsToSelect) > 0 {
		records = handler.selectRecordFields(fieldsToSelect, records)
	}
	writeJSON(w, http.StatusOK, records)
}

func (handler *RecordHandler) GetOneByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		typeName = vars["name"]
		recordID = vars["id"]
	)

	typ, err := validateType(typeName)
	if err != nil {
		handler.log.WithField("type", typeName).
			Error(err)
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	recordRepo, err := handler.recordRepositoryFactory.For(typ)
	if err != nil {
		handler.log.WithField("type", typeName).
			Errorf("internal: error construing record repository: %v", err)

		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}

	record, err := recordRepo.GetByID(r.Context(), recordID)
	if err != nil {
		handler.log.WithField("type", typeName).
			WithField("record", recordID).
			Errorf("error fetching record: %v", err)

		status, message := handler.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (handler *RecordHandler) parseSelectQuery(raw string) (fields []string) {
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

func (handler *RecordHandler) selectRecordFields(fields []string, records []record.Record) (processedRecords []record.Record) {
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

func (handler *RecordHandler) validateRecord(record record.Record) error {
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

func (handler *RecordHandler) responseStatusForError(err error) (int, string) {
	switch err.(type) {
	case record.ErrNoSuchRecord:
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}

func bodyParserErrorMsg(err error) string {
	return fmt.Sprintf("error parsing request body: %v", err)
}
