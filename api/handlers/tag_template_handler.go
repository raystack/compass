package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/odpf/columbus/tag"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// TagTemplateHandler is handler to manage template related requests
type TagTemplateHandler struct {
	service *tag.TemplateService
	logger  logrus.FieldLogger
}

// NewTagTemplateHandler initializes template handler based on the service
func NewTagTemplateHandler(logger logrus.FieldLogger, service *tag.TemplateService) *TagTemplateHandler {
	if service == nil {
		panic("template service is nil")
	}
	return &TagTemplateHandler{
		service: service,
		logger:  logger,
	}
}

// Create handles template creation requests
func (h *TagTemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var requestBody tag.Template
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := h.service.Create(&requestBody)
	if errors.As(err, new(tag.DuplicateTemplateError)) {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error creating tag template: %s", err.Error()))

		return
	}

	writeJSON(w, http.StatusCreated, requestBody)
}

// Index handles template read requests
func (h *TagTemplateHandler) Index(w http.ResponseWriter, r *http.Request) {
	urn := r.URL.Query().Get("urn")
	queryDomainTemplate := tag.Template{
		URN: urn,
	}
	listOfDomainTemplate, err := h.service.Index(queryDomainTemplate)
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error finding templates: %s", err.Error()))

		return
	}
	writeJSON(w, http.StatusOK, listOfDomainTemplate)
}

// Update handles template update requests
func (h *TagTemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	urn, ok := params["template_urn"]
	if !ok || urn == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	var requestBody tag.Template
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	requestBody.URN = urn
	err := h.service.Update(&requestBody)
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.As(err, new(tag.ValidationError)) {
		writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error updating template: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, requestBody)
}

// Find handles template read requests based on URN
func (h *TagTemplateHandler) Find(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	urn, ok := params["template_urn"]
	if !ok || urn == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	domainTemplate, err := h.service.Find(urn)
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error finding a template: %s", err.Error()))

		return
	}

	writeJSON(w, http.StatusOK, domainTemplate)
}

// Delete handles template delete request based on URN
func (h *TagTemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	urn, ok := params["template_urn"]
	if !ok || urn == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	err := h.service.Delete(urn)
	if err != nil {
		e := new(tag.TemplateNotFoundError)
		if errors.As(err, e) {
			writeJSONError(w, http.StatusNotFound, e.Error())
			return
		}

		internalServerError(w, h.logger, fmt.Sprintf("error deleting a template: %s", err.Error()))

		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
