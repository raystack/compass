package v1beta1

import (
	"context"
	"errors"
	"strings"

	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateComment will create a new comment of a discussion
// field body is mandatory
func (h *Handler) CreateComment(ctx context.Context, req *compassv1beta1.CreateCommentRequest) (*compassv1beta1.CreateCommentResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := h.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	cmt := discussion.Comment{
		DiscussionID: req.DiscussionId,
		Body:         req.Body,
		Owner:        user.User{ID: userID},
		UpdatedBy:    user.User{ID: userID},
	}

	if err := cmt.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	id, err := h.DiscussionRepository.CreateComment(ctx, &cmt)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.CreateCommentResponse{Id: id}, nil
}

// GetAllComments returns all comments of a discussion
func (h *Handler) GetAllComments(ctx context.Context, req *compassv1beta1.GetAllCommentsRequest) (*compassv1beta1.GetAllCommentsResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := h.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	flt, err := h.buildGetAllCommentsFilter(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	cmts, err := h.DiscussionRepository.GetAllComments(ctx, req.DiscussionId, flt)
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	commentsProto := []*compassv1beta1.Comment{}
	for _, cmt := range cmts {
		commentsProto = append(commentsProto, cmt.ToProto())
	}

	return &compassv1beta1.GetAllCommentsResponse{Data: commentsProto}, nil
}

// GetComment returns a comment discussion by id from path
func (h *Handler) GetComment(ctx context.Context, req *compassv1beta1.GetCommentRequest) (*compassv1beta1.GetCommentResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := h.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	if err := h.validateIDInteger(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId, CommentID: req.Id}))
	}

	cmt, err := h.DiscussionRepository.GetComment(ctx, req.Id, req.DiscussionId)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, errMissingUserInfo.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.GetCommentResponse{Data: cmt.ToProto()}, nil
}

// UpdateComment is an api to update a comment by discussion id
func (h *Handler) UpdateComment(ctx context.Context, req *compassv1beta1.UpdateCommentRequest) (*compassv1beta1.UpdateCommentResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := h.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	if err := h.validateIDInteger(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId, CommentID: req.Id}))
	}

	cmt := discussion.Comment{
		ID:           req.Id,
		DiscussionID: req.DiscussionId,
		Body:         req.Body,
		UpdatedBy:    user.User{ID: userID},
	}

	if err := cmt.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := h.DiscussionRepository.UpdateComment(ctx, &cmt)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.UpdateCommentResponse{}, nil
}

// DeleteComment is an api to delete a comment by discussion id
func (h *Handler) DeleteComment(ctx context.Context, req *compassv1beta1.DeleteCommentRequest) (*compassv1beta1.DeleteCommentResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := h.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	if err := h.validateIDInteger(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId, CommentID: req.Id}))
	}

	err := h.DiscussionRepository.DeleteComment(ctx, req.Id, req.DiscussionId)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.DeleteCommentResponse{}, nil
}

func (h *Handler) buildGetAllDiscussionsFilter(req *compassv1beta1.GetAllDiscussionsRequest) (discussion.Filter, error) {

	fl := discussion.Filter{
		Type:          req.GetType(),
		State:         req.GetState(),
		Owner:         req.GetOwner(),
		SortBy:        req.GetSort(),
		SortDirection: req.GetDirection(),
	}

	assignees := req.GetAssignee()
	if assignees != "" {
		fl.Assignees = strings.Split(assignees, ",")
	}

	assets := req.GetAsset()
	if assets != "" {
		fl.Assets = strings.Split(assets, ",")
	}

	labels := req.GetLabels()
	if labels != "" {
		fl.Labels = strings.Split(labels, ",")
	}

	fl.Size = int(req.GetSize())
	fl.Offset = int(req.GetOffset())

	if err := fl.Validate(); err != nil {
		return discussion.Filter{}, err
	}

	fl.AssignDefault()

	return fl, nil
}

func (h *Handler) buildGetAllCommentsFilter(req *compassv1beta1.GetAllCommentsRequest) (discussion.Filter, error) {

	fl := discussion.Filter{
		SortBy:        req.GetSort(),
		SortDirection: req.GetDirection(),
	}

	fl.Size = int(req.GetSize())
	fl.Offset = int(req.GetOffset())

	if err := fl.Validate(); err != nil {
		return discussion.Filter{}, err
	}

	fl.AssignDefault()

	return fl, nil
}
