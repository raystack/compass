package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/record"
	"github.com/sirupsen/logrus"
)

var (
	validClassifications     map[record.TypeClassification]int
	validClassificationsList string
)

func init() {
	validClassifications = make(map[record.TypeClassification]int)
	clsList := make([]string, len(record.AllTypeClassifications))
	for idx, cls := range record.AllTypeClassifications {
		validClassifications[cls] = 0
		clsList[idx] = string(cls)
	}
	validClassificationsList = strings.Join(clsList, ",")
}

// TypeHandler exposes a REST interface to types
type TypeHandler struct {
	typeRepo record.TypeRepository
	log      logrus.FieldLogger
}

func NewTypeHandler(log logrus.FieldLogger, er record.TypeRepository) *TypeHandler {
	handler := &TypeHandler{
		typeRepo: er,
		log:      log,
	}

	return handler
}

func (handler *TypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	types, err := handler.typeRepo.GetAll(r.Context())
	if err != nil {
		handler.log.
			Errorf("error fetching types: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "error fetching types")
		return
	}

	writeJSON(w, http.StatusOK, types)
}

func (handler *TypeHandler) Find(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	recordType, err := handler.typeRepo.GetByName(r.Context(), name)
	if err != nil {
		handler.log.
			Errorf("error fetching type \"%s\": %v", name, err)

		var status int
		var msg string
		if _, ok := err.(record.ErrNoSuchType); ok {
			status = http.StatusNotFound
			msg = err.Error()
		} else {
			status = http.StatusInternalServerError
			msg = fmt.Sprintf("error fetching type \"%s\"", name)
		}

		writeJSONError(w, status, msg)
		return
	}

	writeJSON(w, http.StatusOK, recordType)
}

func (handler *TypeHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var payload record.Type
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	payload = payload.Normalise()
	if err := handler.validateType(payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = handler.typeRepo.CreateOrReplace(r.Context(), payload)
	if err != nil {
		handler.log.
			WithField("type", payload.Name).
			Errorf("error creating/replacing type: %v", err)

		var status int
		var msg string
		if _, ok := err.(record.ErrReservedTypeName); ok {
			status = http.StatusUnprocessableEntity
			msg = err.Error()
		} else {
			status = http.StatusInternalServerError
			msg = fmt.Sprintf("error creating type: %v", err)
		}

		writeJSONError(w, status, msg)
		return
	}
	handler.log.Infof("created/updated %q type", payload.Name)
	writeJSON(w, http.StatusCreated, payload)
}

func (handler *TypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	err := handler.typeRepo.Delete(r.Context(), name)
	if err != nil {
		handler.log.
			Errorf("error deleting type \"%s\": %v", name, err)

		var status int
		var msg string
		if _, ok := err.(record.ErrReservedTypeName); ok {
			status = http.StatusUnprocessableEntity
			msg = err.Error()
		} else {
			status = http.StatusInternalServerError
			msg = fmt.Sprintf("error deleting type \"%s\"", name)
		}

		writeJSONError(w, status, msg)
		return
	}

	handler.log.Infof("deleted type \"%s\"", name)
	writeJSON(w, http.StatusNoContent, "success")
}

func (handler *TypeHandler) parseSelectQuery(raw string) (fields []string) {
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

func (handler *TypeHandler) selectRecordFields(fields []string, records []record.Record) (processedRecords []record.Record) {
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

func (handler *TypeHandler) validateRecord(record record.Record) error {
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

func (handler *TypeHandler) validateType(e record.Type) error {
	// TODO(Aman): write a generic zero-value validator that uses reflection
	// TODO(Aman): how about moving this validation to the repository?
	// TODO(Aman): use reflection to compute the key namespace for recordType.Fields
	// TODO(Aman): add validation for recordType.Lineage
	trim := strings.TrimSpace
	switch {
	case trim(e.Name) == "":
		return fmt.Errorf("'name' is required")
	case trim(string(e.Classification)) == "":
		return fmt.Errorf("'classification' is required")
	case isClassificationValid(e.Classification) == false:
		return fmt.Errorf("'classification' must be one of [%s]", validClassificationsList)
	}
	return nil
}

func (handler *TypeHandler) responseStatusForError(err error) (int, string) {
	switch err.(type) {
	case record.ErrNoSuchType, record.ErrNoSuchRecord:
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}

func isClassificationValid(cls record.TypeClassification) bool {
	_, valid := validClassifications[cls]
	return valid
}

func getJSONKeyNameForField(structure interface{}, field string) string {
	structType := reflect.TypeOf(structure)
	structField, exists := structType.FieldByName(field)
	if !exists {
		msg := fmt.Sprintf("no such Field %q in %q", field, structType.Name())
		panic(msg)
	}
	tag := structField.Tag.Get("json")
	items := strings.Split(tag, ",")
	return strings.TrimSpace(items[0])
}
