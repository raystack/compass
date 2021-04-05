package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/models"
	"github.com/sirupsen/logrus"
)

var (
	validClassifications     map[models.TypeClassification]int
	validClassificationsList string

	// typeFieldNamespace refers to the keyname of the models.TypeFields
	// inside models.Type. This is used to avoid hard coding the JSON keyname
	// for type fields
	typeFieldNamespace = getJSONKeyNameForField(models.Type{}, "Fields")
)

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

func init() {
	validClassifications = make(map[models.TypeClassification]int)
	clsList := make([]string, len(models.AllTypeClassifications))
	for idx, cls := range models.AllTypeClassifications {
		validClassifications[cls] = 0
		clsList[idx] = string(cls)
	}
	validClassificationsList = strings.Join(clsList, ",")
}

func isClassificationValid(cls models.TypeClassification) bool {
	_, valid := validClassifications[cls]
	return valid
}

func bodyParserErrorMsg(err error) string {
	return fmt.Sprintf("error parsing request body: %v", err)
}

// TypeHandler exposes a REST interface to types
type TypeHandler struct {
	mux                     *mux.Router
	typeRepo                models.TypeRepository
	recordRepositoryFactory models.RecordRepositoryFactory
	log                     logrus.FieldLogger
}

func (handler *TypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mux.ServeHTTP(w, r)
}

func (handler *TypeHandler) getAll(w http.ResponseWriter, r *http.Request) {
	types, err := handler.typeRepo.GetAll()
	if err != nil {
		handler.log.
			Errorf("error fetching types: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "error fetching types")
		return
	}

	writeJSON(w, http.StatusOK, types)
}

func (handler *TypeHandler) getType(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	recordType, err := handler.typeRepo.GetByName(name)
	if err != nil {
		handler.log.
			Errorf("error fetching type \"%s\": %v", name, err)

		var status int
		var msg string
		if _, ok := err.(models.ErrNoSuchType); ok {
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

func (handler *TypeHandler) createOrReplaceType(w http.ResponseWriter, r *http.Request) {
	var payload models.Type
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

	err = handler.typeRepo.CreateOrReplace(payload)
	if err != nil {
		handler.log.
			WithField("type", payload.Name).
			Errorf("error creating/replacing type: %v", err)

		var status int
		var msg string
		if _, ok := err.(models.ErrReservedTypeName); ok {
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

func (handler *TypeHandler) deleteType(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	err := handler.typeRepo.Delete(name)
	if err != nil {
		handler.log.
			Errorf("error deleting type \"%s\": %v", name, err)

		var status int
		var msg string
		if _, ok := err.(models.ErrReservedTypeName); ok {
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

func (handler *TypeHandler) deleteRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		typeName = vars["name"]
		recordID = vars["id"]
	)

	statusCode := http.StatusInternalServerError
	errMessage := fmt.Sprintf("error deleting record \"%s\" with type \"%s\"", recordID, typeName)

	recordType, err := handler.typeRepo.GetByName(typeName)
	if err != nil {
		handler.log.
			Errorf("error getting type \"%s\": %v", typeName, err)

		if _, ok := err.(models.ErrNoSuchType); ok {
			statusCode = http.StatusNotFound
			errMessage = err.Error()
		}

		writeJSONError(w, statusCode, errMessage)
		return
	}

	recordRepoFactory, _ := handler.recordRepositoryFactory.For(recordType)
	if err != nil {
		handler.log.
			Errorf("error creating record repository for \"%s\": %v", typeName, err)
		writeJSONError(w, statusCode, errMessage)
		return
	}

	err = recordRepoFactory.Delete(recordID)
	if err != nil {
		handler.log.
			Errorf("error deleting record \"%s\": %v", typeName, err)

		if _, ok := err.(models.ErrNoSuchRecord); ok {
			statusCode = http.StatusNotFound
			errMessage = err.Error()
		}

		writeJSONError(w, statusCode, errMessage)
		return
	}

	handler.log.Infof("deleted record \"%s\" with type \"%s\"", recordID, typeName)
	writeJSON(w, http.StatusNoContent, "success")
}

func (handler *TypeHandler) ingestRecord(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	recordType, err := handler.typeRepo.GetByName(name)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(models.ErrNoSuchType); ok {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, err.Error())
		return
	}

	var records []models.Record
	err = json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	var failedRecords = make(map[int]string)
	for idx, record := range records {
		if err := handler.validateRecord(recordType, record); err != nil {
			handler.log.WithField("type", recordType).
				WithField("record", record).
				Errorf("error validating record: %v", err)
			failedRecords[idx] = err.Error()
		}
	}
	if len(failedRecords) > 0 {
		writeJSON(w, http.StatusBadRequest, NewValidationErrorResponse(failedRecords))
		return
	}

	recordRepo, err := handler.recordRepositoryFactory.For(recordType)
	if err != nil {
		handler.log.WithField("type", recordType.Name).
			Errorf("error creating record repository: %v", err)

		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	if err = recordRepo.CreateOrReplaceMany(records); err != nil {
		handler.log.WithField("type", recordType.Name).
			Errorf("error creating/updating records: %v", err)

		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}
	handler.log.Infof("created/updated %d records for %q type", len(records), recordType.Name)
	writeJSON(w, http.StatusOK, StatusResponse{Status: "success"})
}

func (handler *TypeHandler) listTypeRecords(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	recordType, err := handler.typeRepo.GetByName(name)
	if err != nil {
		status, message := handler.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}

	recordRepo, err := handler.recordRepositoryFactory.For(recordType)
	if err != nil {
		handler.log.WithField("type", recordType).
			Errorf("error constructing record repository: %v", err)
		status, message := handler.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}
	filterCfg := filterConfigFromValues(r.URL.Query())

	records, err := recordRepo.GetAll(filterCfg)
	if err != nil {
		handler.log.WithField("type", recordType).
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

func (handler *TypeHandler) getTypeRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		typeName = vars["name"]
		recordID = vars["id"]
	)
	recordType, err := handler.typeRepo.GetByName(typeName)

	// TODO(Aman): make error handling a bit more DRY
	if err != nil {
		handler.log.WithField("type", typeName).
			Errorf("error fetching type: %v", err)

		status, message := handler.responseStatusForError(err)
		writeJSONError(w, status, message)
		return
	}
	recordRepo, err := handler.recordRepositoryFactory.For(recordType)
	if err != nil {
		handler.log.WithField("type", typeName).
			Errorf("internal: error construing record repository: %v", err)

		status := http.StatusInternalServerError
		writeJSONError(w, status, http.StatusText(status))
		return
	}

	record, err := recordRepo.GetByID(recordID)
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

func (handler *TypeHandler) selectRecordFields(fields []string, records []models.Record) (processedRecords []models.Record) {
	for _, record := range records {
		currentRecord := make(models.Record)
		for _, field := range fields {
			v, ok := record[field]
			if !ok {
				continue
			}
			currentRecord[field] = v
		}
		processedRecords = append(processedRecords, currentRecord)
	}
	return
}

func (handler *TypeHandler) validateRecord(recordType models.Type, record map[string]interface{}) error {
	var shouldHave = []string{recordType.Fields.Title, recordType.Fields.ID}
	for _, key := range shouldHave {
		val, ok := record[key]
		if !ok {
			return fmt.Errorf("%q is required", key)
		}
		stringVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("%q must be string", key)
		}
		if strings.TrimSpace(stringVal) == "" {
			return fmt.Errorf("%q cannot be empty", key)
		}
	}
	return nil
}

func (handler *TypeHandler) validateType(e models.Type) error {
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
	case trim(e.Fields.Title) == "":
		return fmt.Errorf("'%s.title' is required", typeFieldNamespace)
	case trim(e.Fields.ID) == "":
		return fmt.Errorf("'%s.id' is required", typeFieldNamespace)
	}
	for idx, desc := range e.Lineage {
		if desc.Dir.Valid() == false {
			return fmt.Errorf("lineage[%d].dir: invalid direction %q", idx, desc.Dir)
		}
		if strings.TrimSpace(desc.Query) == "" {
			return fmt.Errorf("lineage[%d].query: query cannot be empty", idx)
		}
		if strings.TrimSpace(desc.Type) == "" {
			return fmt.Errorf("lineage[%d].query: type cannot be empty", idx)
		}
	}
	return nil
}

func (handler *TypeHandler) responseStatusForError(err error) (int, string) {
	switch err.(type) {
	case models.ErrNoSuchType, models.ErrNoSuchRecord:
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}

func NewTypeHandler(log logrus.FieldLogger, er models.TypeRepository, rrf models.RecordRepositoryFactory) *TypeHandler {
	handler := &TypeHandler{
		mux:                     mux.NewRouter(),
		typeRepo:                er,
		recordRepositoryFactory: rrf,
		log:                     log,
	}

	mapHandlers(handler, "/v1/types")
	// For backward compatibility
	mapHandlers(handler, "/v1/entities")

	return handler
}

func mapHandlers(handler *TypeHandler, baseURL string) {
	handler.mux.Path(baseURL).
		Methods(http.MethodGet).
		HandlerFunc(handler.getAll)

	// TODO: remove this route when
	// getting type details already handled on GET baseUrl/{name}
	handler.mux.Path(baseURL+"/{name}/details").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(handler.getType)

	// TODO: switch this route to return type details
	handler.mux.Path(baseURL+"/{name}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(handler.listTypeRecords)

	handler.mux.Path(baseURL+"/{name}/records").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(handler.listTypeRecords)

	handler.mux.Path(baseURL).
		Methods(http.MethodPut).
		HandlerFunc(handler.createOrReplaceType)

	handler.mux.Path(baseURL + "/{name}").
		Methods(http.MethodDelete).
		HandlerFunc(handler.deleteType)

	handler.mux.Path(baseURL + "/{name}/records/{id}").
		Methods(http.MethodDelete).
		HandlerFunc(handler.deleteRecord)

	handler.mux.Path(baseURL + "/{name}").
		Methods(http.MethodPut).
		HandlerFunc(handler.ingestRecord)

	handler.mux.Path(baseURL+"/{name}/records/{id}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(handler.getTypeRecord)

	// TODO: remove this once no more request is coming
	handler.mux.Path(baseURL+"/{name}/{id}").
		Methods(http.MethodGet, http.MethodHead).
		HandlerFunc(handler.getTypeRecord)
}
