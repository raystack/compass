package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/comment"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
)

// CommentHandler exposes a REST interface to discussion
type CommentHandler struct {
	logger      log.Logger
	commentRepo comment.Repository
}

// Create will create a new comment of a discussion
// field body is mandatory
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	discussionID := pathParams["discussion_id"]
	if err := h.validateID(discussionID); err != nil {
		h.logger.Warn(err.Error(), "discussion_id", discussionID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID}))
		return
	}

	var cmt comment.Comment
	if err := json.NewDecoder(r.Body).Decode(&cmt); err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	cmt.DiscussionID = discussionID
	if err := cmt.Validate(); err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	cmt.DiscussionID = discussionID
	cmt.Owner = user.User{ID: userID}
	cmt.UpdatedBy = user.User{ID: userID}
	id, err := h.commentRepo.Create(r.Context(), &cmt)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id": id,
	})
}

// GetAll returns all comments of a discussion
func (h *CommentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	discussionID := pathParams["discussion_id"]
	if err := h.validateID(discussionID); err != nil {
		h.logger.Warn(err.Error(), "discussion_id", discussionID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID}))
		return
	}

	flt, err := h.buildGetFilter(r.URL.Query())
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	cmts, err := h.commentRepo.GetAll(r.Context(), discussionID, flt)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cmts)
}

// Get returns a comment discussion by id from path
func (h *CommentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	discussionID := pathParams["discussion_id"]
	if err := h.validateID(discussionID); err != nil {
		h.logger.Warn(err.Error(), "discussion_id", discussionID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID}))
		return
	}

	commentID := pathParams["id"]
	if err := h.validateID(commentID); err != nil {
		h.logger.Warn(err.Error(), "id", commentID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID, CommentID: commentID}))
		return
	}

	cmt, err := h.commentRepo.Get(r.Context(), commentID, discussionID)
	if errors.As(err, new(comment.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cmt)
}

// Update is an api to update a comment by discussion id
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	discussionID := pathParams["discussion_id"]
	if err := h.validateID(discussionID); err != nil {
		h.logger.Warn(err.Error(), "discussion_id", discussionID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID}))
		return
	}

	commentID := pathParams["id"]
	if err := h.validateID(commentID); err != nil {
		h.logger.Warn(err.Error(), "id", commentID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID, CommentID: commentID}))
		return
	}

	var cmt comment.Comment
	if err := json.NewDecoder(r.Body).Decode(&cmt); err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	cmt.ID = commentID
	cmt.DiscussionID = discussionID
	if err := cmt.Validate(); err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	cmt.ID = commentID
	cmt.DiscussionID = discussionID
	cmt.UpdatedBy = user.User{ID: userID}
	err := h.commentRepo.Update(r.Context(), &cmt)
	if errors.As(err, new(comment.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

// Delete is an api to delete a comment by discussion id
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := user.FromContext(r.Context())
	if userID == "" {
		h.logger.Warn(errMissingUserInfo.Error())
		WriteJSONError(w, http.StatusBadRequest, errMissingUserInfo.Error())
		return
	}

	pathParams := mux.Vars(r)
	discussionID := pathParams["discussion_id"]
	if err := h.validateID(discussionID); err != nil {
		h.logger.Warn(err.Error(), "discussion_id", discussionID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID}))
		return
	}

	commentID := pathParams["id"]
	if err := h.validateID(commentID); err != nil {
		h.logger.Warn(err.Error(), "id", commentID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(comment.InvalidError{DiscussionID: discussionID, CommentID: commentID}))
		return
	}

	err := h.commentRepo.Delete(r.Context(), commentID, discussionID)
	if errors.As(err, new(comment.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *CommentHandler) buildGetFilter(query url.Values) (comment.Filter, error) {

	fl := comment.Filter{
		SortBy:        query.Get("sort"),
		SortDirection: query.Get("direction"),
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
		return comment.Filter{}, err
	}

	fl.AssignDefault()

	return fl, nil
}

func (h *CommentHandler) validateID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return err
	}

	if idInt < 1 {
		return errors.New("id cannot be < 1")
	}

	return nil
}

func NewCommentHandler(
	logger log.Logger,
	commentRepo comment.Repository) *CommentHandler {
	handler := &CommentHandler{
		logger:      logger,
		commentRepo: commentRepo,
	}
	return handler
}
