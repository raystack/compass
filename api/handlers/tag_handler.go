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

var (
	errEmptyRecordURN   = errors.New("record urn is empty")
	errEmptyRecordType  = errors.New("type is empty")
	errNilTagService    = errors.New("tag service is nil")
	errEmptyTemplateURN = errors.New("template urn is empty")
)

// TagHandler is handler to manage tag related requests
type TagHandler struct {
	service *tag.Service
	logger  logrus.FieldLogger
}

// NewTagHandler initializes tag handler based on the service
func NewTagHandler(logger logrus.FieldLogger, service *tag.Service) *TagHandler {
	if service == nil {
		panic("cannot create TagHandler with nil tag.Service")
	}

	return &TagHandler{
		service: service,
		logger:  logger,
	}
}

// Create handles tag creation requests
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSONError(w, http.StatusInternalServerError, errNilTagService.Error())
		return
	}

	var requestBody tag.Tag
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.Create(&requestBody)
	if errors.As(err, new(tag.DuplicateError)) {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.As(err, new(tag.ValidationError)) {
		writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error creating tag: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusCreated, requestBody)
}

// GetByRecord handles get tag by record requests
func (h *TagHandler) GetByRecord(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSONError(w, http.StatusInternalServerError, errNilTagService.Error())
		return
	}

	muxVar := mux.Vars(r)
	recordType, exists := muxVar["type"]
	if !exists || recordType == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordType.Error())
		return
	}
	recordURN, exists := muxVar["record_urn"]
	if !exists || recordURN == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordURN.Error())
		return
	}
	tags, err := h.service.GetByRecord(recordType, recordURN)
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error getting record tags: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, tags)
}

// FindByRecordAndTemplate handles get tag by record requests
func (h *TagHandler) FindByRecordAndTemplate(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSONError(w, http.StatusInternalServerError, errNilTagService.Error())
		return
	}

	muxVar := mux.Vars(r)
	recordType, exists := muxVar["type"]
	if !exists || recordType == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordType.Error())
		return
	}
	recordURN, exists := muxVar["record_urn"]
	if !exists || recordURN == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordURN.Error())
		return
	}
	templateURN, exists := muxVar["template_urn"]
	if !exists || templateURN == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}
	tags, err := h.service.FindByRecordAndTemplate(recordType, recordURN, templateURN)
	if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error finding a tag with record and template: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, tags)
}

// Update handles tag update requests
func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSONError(w, http.StatusInternalServerError, errNilTagService.Error())
		return
	}

	muxVar := mux.Vars(r)
	recordURN, exists := muxVar["record_urn"]
	if !exists || recordURN == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordURN.Error())
		return
	}
	recordType, exists := muxVar["type"]
	if !exists || recordType == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordType.Error())
		return
	}
	templateURN, exists := muxVar["template_urn"]
	if !exists || templateURN == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	var requestBody tag.Tag
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	requestBody.RecordURN = recordURN
	requestBody.RecordType = recordType
	requestBody.TemplateURN = templateURN
	err := h.service.Update(&requestBody)
	if errors.As(err, new(tag.NotFoundError)) {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error updating a template: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, requestBody)
}

// Delete handles delete tag by record and template requests
func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSONError(w, http.StatusInternalServerError, errNilTagService.Error())
		return
	}
	muxVar := mux.Vars(r)
	recordType, exists := muxVar["type"]
	if !exists || recordType == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordType.Error())
		return
	}
	recordURN, exists := muxVar["record_urn"]
	if !exists || recordURN == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyRecordURN.Error())
		return
	}
	templateURN, exists := muxVar["template_urn"]
	if !exists || templateURN == "" {
		writeJSONError(w, http.StatusBadRequest, errEmptyTemplateURN.Error())
		return
	}

	err := h.service.Delete(recordType, recordURN, templateURN)
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, fmt.Sprintf("error deleting a template: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
