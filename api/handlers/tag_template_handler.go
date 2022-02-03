package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/odpf/salt/log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/tag"
)

// TagTemplateHandler is handler to manage template related requests
type TagTemplateHandler struct {
	service *tag.TemplateService
	logger  log.Logger
}

// NewTagTemplateHandler initializes template handler based on the service
func NewTagTemplateHandler(logger log.Logger, service *tag.TemplateService) *TagTemplateHandler {
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
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := h.service.Create(r.Context(), &requestBody)
	if errors.As(err, new(tag.DuplicateTemplateError)) {
		WriteJSONError(w, http.StatusConflict, err.Error())
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
	listOfDomainTemplate, err := h.service.Index(r.Context(), urn)
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error finding templates: %s", err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, listOfDomainTemplate)
}

// Update handles template update requests
func (h *TagTemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	targetTemplateURN, ok := params["template_urn"]
	if !ok || targetTemplateURN == "" {
		WriteJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	var requestBody tag.Template
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	requestBody.URN = targetTemplateURN
	err := h.service.Update(r.Context(), targetTemplateURN, &requestBody)
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.As(err, new(tag.ValidationError)) {
		WriteJSONError(w, http.StatusUnprocessableEntity, err.Error())
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
		WriteJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	domainTemplate, err := h.service.Find(r.Context(), urn)
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
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
		WriteJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	err := h.service.Delete(r.Context(), urn)
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error deleting a template: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
