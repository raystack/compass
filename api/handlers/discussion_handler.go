package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
)

// DiscussionHandler exposes a REST interface to discussion
type DiscussionHandler struct {
	logger               log.Logger
	discussionRepository discussion.Repository
}

// GetAll returns all discussion based on filter in query params
// supported query params are type,state,owner,assignee,asset,labels (supporterd array separated by comma)
// query params sort,direction to sort asc or desc
// query params size,offset for pagination
func (h *DiscussionHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	flt, err := h.buildGetFilter(r.URL.Query())
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	dscs, err := h.discussionRepository.GetAll(r.Context(), flt)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dscs)
}

// Create will create a new discussion
// field title, body, and type are mandatory
func (h *DiscussionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	var dsc discussion.Discussion
	if err := json.NewDecoder(r.Body).Decode(&dsc); err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}
	err := dsc.Validate()
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	dsc.Owner = user.User{ID: userID}
	id, err := h.discussionRepository.Create(r.Context(), &dsc)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id": id,
	})
}

// Get returns a discussion by id from path
func (h *DiscussionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	discussionID := pathParams["id"]
	if err := h.validateID(discussionID); err != nil {
		h.logger.Warn(err.Error(), "id", discussionID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.ErrInvalidID))
		return
	}

	dsc, err := h.discussionRepository.Get(r.Context(), discussionID)
	if errors.As(err, new(discussion.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dsc)
}

// Patch updates a specific field in discussion
// empty array in assets,labels,assignees will be considered
// and clear all assets,labels,assignees from the discussion
func (h *DiscussionHandler) Patch(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	discussionID := pathParams["id"]
	if err := h.validateID(discussionID); err != nil {
		h.logger.Warn(err.Error(), "id", discussionID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.ErrInvalidID))
		return
	}

	var dsc discussion.Discussion
	if err := json.NewDecoder(r.Body).Decode(&dsc); err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}
	if isEmpty := dsc.IsEmpty(); isEmpty {
		writeJSON(w, http.StatusNoContent, nil)
		return
	}
	if err := dsc.ValidateConstraint(); err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	dsc.ID = discussionID
	err := h.discussionRepository.Patch(r.Context(), &dsc)
	if errors.Is(err, discussion.ErrInvalidID) {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.As(err, new(discussion.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *DiscussionHandler) buildGetFilter(query url.Values) (discussion.Filter, error) {

	fl := discussion.Filter{
		Type:          query.Get("type"),
		State:         query.Get("state"),
		Owner:         query.Get("owner"),
		SortBy:        query.Get("sort"),
		SortDirection: query.Get("direction"),
	}

	assignees := query.Get("assignee")
	if assignees != "" {
		fl.Assignees = strings.Split(assignees, ",")
	}

	assets := query.Get("asset")
	if assets != "" {
		fl.Assets = strings.Split(assets, ",")
	}

	labels := query.Get("labels")
	if labels != "" {
		fl.Labels = strings.Split(labels, ",")
	}

	sizeString := query.Get("size")
	if sizeString != "" {
		size, err := strconv.Atoi(sizeString)
		if err == nil {
			fl.Size = size
		}
	}

	offsetString := query.Get("offset")
	if offsetString != "" {
		offset, err := strconv.Atoi(offsetString)
		if err == nil {
			fl.Offset = offset
		}
	}

	if err := fl.Validate(); err != nil {
		return discussion.Filter{}, err
	}

	fl.AssignDefault()

	return fl, nil
}

func (h *DiscussionHandler) validateID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return err
	}

	if idInt < 1 {
		return errors.New("id cannot be < 1")
	}

	return nil
}

func NewDiscussionHandler(
	logger log.Logger,
	discussionRepository discussion.Repository) *DiscussionHandler {
	handler := &DiscussionHandler{
		logger:               logger,
		discussionRepository: discussionRepository,
	}
	return handler
}
