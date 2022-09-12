package handlersv1beta1

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/odpf/compass/core/discussion"
	"github.com/odpf/compass/core/user"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateComment will create a new comment of a discussion
// field body is mandatory
func (server *APIServer) CreateComment(ctx context.Context, req *compassv1beta1.CreateCommentRequest) (*compassv1beta1.CreateCommentResponse, error) {
	userID, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := server.validateIDInteger(req.DiscussionId); err != nil {
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

	id, err := server.discussionService.CreateComment(ctx, &cmt)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.CreateCommentResponse{Id: id}, nil
}

// GetAllComments returns all comments of a discussion
func (server *APIServer) GetAllComments(ctx context.Context, req *compassv1beta1.GetAllCommentsRequest) (*compassv1beta1.GetAllCommentsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := server.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	flt, err := server.buildGetAllCommentsFilter(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	cmts, err := server.discussionService.GetComments(ctx, req.DiscussionId, flt)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	commentsProto := []*compassv1beta1.Comment{}
	for _, cmt := range cmts {
		commentsProto = append(commentsProto, commentToProto(cmt))
	}

	return &compassv1beta1.GetAllCommentsResponse{Data: commentsProto}, nil
}

// GetComment returns a comment discussion by id from path
func (server *APIServer) GetComment(ctx context.Context, req *compassv1beta1.GetCommentRequest) (*compassv1beta1.GetCommentResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := server.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	if err := server.validateIDInteger(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId, CommentID: req.Id}))
	}

	cmt, err := server.discussionService.GetComment(ctx, req.Id, req.DiscussionId)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, errMissingUserInfo.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetCommentResponse{Data: commentToProto(cmt)}, nil
}

// UpdateComment is an api to update a comment by discussion id
func (server *APIServer) UpdateComment(ctx context.Context, req *compassv1beta1.UpdateCommentRequest) (*compassv1beta1.UpdateCommentResponse, error) {
	userID, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := server.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	if err := server.validateIDInteger(req.Id); err != nil {
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

	err = server.discussionService.UpdateComment(ctx, &cmt)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.UpdateCommentResponse{}, nil
}

// DeleteComment is an api to delete a comment by discussion id
func (server *APIServer) DeleteComment(ctx context.Context, req *compassv1beta1.DeleteCommentRequest) (*compassv1beta1.DeleteCommentResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := server.validateIDInteger(req.DiscussionId); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId}))
	}

	if err := server.validateIDInteger(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.DiscussionId, CommentID: req.Id}))
	}

	err = server.discussionService.DeleteComment(ctx, req.Id, req.DiscussionId)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.DeleteCommentResponse{}, nil
}

func (server *APIServer) buildGetAllDiscussionsFilter(req *compassv1beta1.GetAllDiscussionsRequest) (discussion.Filter, error) {

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

func (server *APIServer) buildGetAllCommentsFilter(req *compassv1beta1.GetAllCommentsRequest) (discussion.Filter, error) {

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

// commentToProto transforms struct to proto
func commentToProto(c discussion.Comment) *compassv1beta1.Comment {

	var createdAtPB *timestamppb.Timestamp
	if !c.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(c.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !c.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(c.UpdatedAt)
	}

	return &compassv1beta1.Comment{
		Id:           c.ID,
		DiscussionId: c.DiscussionID,
		Body:         c.Body,
		Owner:        userToProto(c.Owner),
		UpdatedBy:    userToProto(c.UpdatedBy),
		CreatedAt:    createdAtPB,
		UpdatedAt:    updatedAtPB,
	}
}

// commentFromProto transforms proto to struct
func commentFromProto(pb *compassv1beta1.Comment) discussion.Comment {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var owner user.User
	if pb.GetOwner() != nil {
		owner = userFromProto(pb.GetOwner())
	}

	var updatedBy user.User
	if pb.GetUpdatedBy() != nil {
		updatedBy = userFromProto(pb.GetUpdatedBy())
	}

	return discussion.Comment{
		ID:           pb.GetId(),
		DiscussionID: pb.GetDiscussionId(),
		Body:         pb.GetBody(),
		Owner:        owner,
		UpdatedBy:    updatedBy,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
