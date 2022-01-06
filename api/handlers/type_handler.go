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
	h := &TypeHandler{
		typeRepo: er,
		log:      log,
	}

	return h
}

func (h *TypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	types, err := h.typeRepo.GetAll(r.Context())
	if err != nil {
		internalServerError(w, h.log, "error fetching types")
		return
	}

	counts, err := h.typeRepo.GetRecordsCount(r.Context())
	if err != nil {
		internalServerError(w, h.log, "error fetching records counts")
		return
	}

	type TypeWithCount struct {
		record.Type
		Count int `json:"count"`
	}

	results := []TypeWithCount{}
	for _, typ := range types {
		count, _ := counts[typ.Name]

		results = append(results, TypeWithCount{
			Type:  typ,
			Count: count,
		})
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *TypeHandler) Find(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	recordType, err := h.typeRepo.GetByName(r.Context(), name)
	if err != nil {
		h.log.
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

func (h *TypeHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var payload record.Type
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	payload = payload.Normalise()
	if err := h.validateType(payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.typeRepo.CreateOrReplace(r.Context(), payload)
	if err != nil {
		h.log.
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
	h.log.Infof("created/updated %q type", payload.Name)
	writeJSON(w, http.StatusCreated, payload)
}

func (h *TypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	err := h.typeRepo.Delete(r.Context(), name)
	if err != nil {
		h.log.
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

	h.log.Infof("deleted type \"%s\"", name)
	writeJSON(w, http.StatusNoContent, "success")
}

func (h *TypeHandler) parseSelectQuery(raw string) (fields []string) {
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

func (h *TypeHandler) selectRecordFields(fields []string, records []record.Record) (processedRecords []record.Record) {
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

func (h *TypeHandler) validateRecord(record record.Record) error {
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

func (h *TypeHandler) validateType(e record.Type) error {
	// TODO(Aman): write a generic zero-value validator that uses reflection
	// TODO(Aman): how about moving this validation to the repository?
	// TODO(Aman): use reflection to compute the key namespace for recordType.Fields
	// TODO(Aman): add validation for recordType.Lineage
	trim := strings.TrimSpace
	switch {
	case trim(string(e.Name)) == "":
		return fmt.Errorf("'name' is required")
	case trim(string(e.Classification)) == "":
		return fmt.Errorf("'classification' is required")
	case isClassificationValid(e.Classification) == false:
		return fmt.Errorf("'classification' must be one of [%s]", validClassificationsList)
	}
	return nil
}

func (h *TypeHandler) responseStatusForError(err error) (int, string) {
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
