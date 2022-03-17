package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/user"
)

// CreateComment will create a new comment of a discussion
// field body is mandatory
func (h *DiscussionHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID}))
		return
	}

	var cmt discussion.Comment
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
	id, err := h.discussionRepository.CreateComment(r.Context(), &cmt)
	if errors.As(err, new(discussion.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id": id,
	})
}

// GetAllComments returns all comments of a discussion
func (h *DiscussionHandler) GetAllComments(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID}))
		return
	}

	flt, err := h.buildGetFilter(r.URL.Query())
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(err))
		return
	}

	cmts, err := h.discussionRepository.GetAllComments(r.Context(), discussionID, flt)
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cmts)
}

// GetComment returns a comment discussion by id from path
func (h *DiscussionHandler) GetComment(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID}))
		return
	}

	commentID := pathParams["id"]
	if err := h.validateID(commentID); err != nil {
		h.logger.Warn(err.Error(), "id", commentID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID, CommentID: commentID}))
		return
	}

	cmt, err := h.discussionRepository.GetComment(r.Context(), commentID, discussionID)
	if errors.As(err, new(discussion.NotFoundError)) {
		WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalServerError(w, h.logger, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cmt)
}

// UpdateComment is an api to update a comment by discussion id
func (h *DiscussionHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID}))
		return
	}

	commentID := pathParams["id"]
	if err := h.validateID(commentID); err != nil {
		h.logger.Warn(err.Error(), "id", commentID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID, CommentID: commentID}))
		return
	}

	var cmt discussion.Comment
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
	err := h.discussionRepository.UpdateComment(r.Context(), &cmt)
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

// DeleteComment is an api to delete a comment by discussion id
func (h *DiscussionHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID}))
		return
	}

	commentID := pathParams["id"]
	if err := h.validateID(commentID); err != nil {
		h.logger.Warn(err.Error(), "id", commentID)
		WriteJSONError(w, http.StatusBadRequest, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: discussionID, CommentID: commentID}))
		return
	}

	err := h.discussionRepository.DeleteComment(r.Context(), commentID, discussionID)
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
