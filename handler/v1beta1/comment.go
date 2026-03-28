package handlersv1beta1

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/discussion"
	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/internal/middleware"
	compassv1beta1 "github.com/raystack/compass/proto/gen/raystack/compass/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateComment will create a new comment of a discussion
// field body is mandatory
func (server *APIServer) CreateComment(ctx context.Context, req *connect.Request[compassv1beta1.CreateCommentRequest]) (*connect.Response[compassv1beta1.CreateCommentResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	userID, err := server.validateUserInCtx(ctx, ns)
	if err != nil {
		return nil, err
	}

	if err := server.validateIDInteger(req.Msg.DiscussionId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId})))
	}

	cmt := discussion.Comment{
		DiscussionID: req.Msg.DiscussionId,
		Body:         req.Msg.Body,
		Owner:        user.User{ID: userID},
		UpdatedBy:    user.User{ID: userID},
	}

	if err := cmt.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	id, err := server.discussionService.CreateComment(ctx, ns, &cmt)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return connect.NewResponse(&compassv1beta1.CreateCommentResponse{Id: id}), nil
}

// GetAllComments returns all comments of a discussion
func (server *APIServer) GetAllComments(ctx context.Context, req *connect.Request[compassv1beta1.GetAllCommentsRequest]) (*connect.Response[compassv1beta1.GetAllCommentsResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}
	if err := server.validateIDInteger(req.Msg.DiscussionId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId})))
	}

	flt, err := server.buildGetAllCommentsFilter(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(err)))
	}

	cmts, err := server.discussionService.GetComments(ctx, req.Msg.DiscussionId, flt)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	commentsProto := []*compassv1beta1.Comment{}
	for _, cmt := range cmts {
		commentsProto = append(commentsProto, commentToProto(cmt))
	}

	return connect.NewResponse(&compassv1beta1.GetAllCommentsResponse{Data: commentsProto}), nil
}

// GetComment returns a comment discussion by id from path
func (server *APIServer) GetComment(ctx context.Context, req *connect.Request[compassv1beta1.GetCommentRequest]) (*connect.Response[compassv1beta1.GetCommentResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}
	if err := server.validateIDInteger(req.Msg.DiscussionId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId})))
	}

	if err := server.validateIDInteger(req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId, CommentID: req.Msg.Id})))
	}

	cmt, err := server.discussionService.GetComment(ctx, req.Msg.Id, req.Msg.DiscussionId)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return connect.NewResponse(&compassv1beta1.GetCommentResponse{Data: commentToProto(cmt)}), nil
}

// UpdateComment is an api to update a comment by discussion id
func (server *APIServer) UpdateComment(ctx context.Context, req *connect.Request[compassv1beta1.UpdateCommentRequest]) (*connect.Response[compassv1beta1.UpdateCommentResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	userID, err := server.validateUserInCtx(ctx, ns)
	if err != nil {
		return nil, err
	}

	if err := server.validateIDInteger(req.Msg.DiscussionId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId})))
	}

	if err := server.validateIDInteger(req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId, CommentID: req.Msg.Id})))
	}

	cmt := discussion.Comment{
		ID:           req.Msg.Id,
		DiscussionID: req.Msg.DiscussionId,
		Body:         req.Msg.Body,
		UpdatedBy:    user.User{ID: userID},
	}

	if err := cmt.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	err = server.discussionService.UpdateComment(ctx, &cmt)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return connect.NewResponse(&compassv1beta1.UpdateCommentResponse{}), nil
}

// DeleteComment is an api to delete a comment by discussion id
func (server *APIServer) DeleteComment(ctx context.Context, req *connect.Request[compassv1beta1.DeleteCommentRequest]) (*connect.Response[compassv1beta1.DeleteCommentResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}
	if err := server.validateIDInteger(req.Msg.DiscussionId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId})))
	}

	if err := server.validateIDInteger(req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(bodyParserErrorMsg(discussion.InvalidError{DiscussionID: req.Msg.DiscussionId, CommentID: req.Msg.Id})))
	}

	err := server.discussionService.DeleteComment(ctx, req.Msg.Id, req.Msg.DiscussionId)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return connect.NewResponse(&compassv1beta1.DeleteCommentResponse{}), nil
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
