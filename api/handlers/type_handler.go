package handlers

import (
	"fmt"
	"net/http"
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
		clsList[idx] = cls.String()
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
		count, _ := counts[typ.Name.String()]

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

func (h *TypeHandler) responseStatusForError(err error) (int, string) {
	switch err.(type) {
	case record.ErrNoSuchType, record.ErrNoSuchRecord:
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}
