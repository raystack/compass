package handlersv1beta1

//go:generate mockery --name=DiscussionService -r --case underscore --with-expecter --structname DiscussionService --filename discussion_service.go --output=./mocks
import (
	"context"
	"errors"
	"strconv"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/discussion"
	"github.com/odpf/compass/core/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiscussionService interface {
	GetDiscussions(ctx context.Context, filter discussion.Filter) ([]discussion.Discussion, error)
	CreateDiscussion(ctx context.Context, discussion *discussion.Discussion) (string, error)
	GetDiscussion(ctx context.Context, did string) (discussion.Discussion, error)
	PatchDiscussion(ctx context.Context, discussion *discussion.Discussion) error
	GetComments(ctx context.Context, discussionID string, filter discussion.Filter) ([]discussion.Comment, error)
	CreateComment(ctx context.Context, cmt *discussion.Comment) (string, error)
	GetComment(ctx context.Context, commentID string, discussionID string) (discussion.Comment, error)
	UpdateComment(ctx context.Context, cmt *discussion.Comment) error
	DeleteComment(ctx context.Context, commentID string, discussionID string) error
}

// GetAll returns all discussion based on filter in query params
// supported query params are type,state,owner,assignee,asset,labels (supporterd array separated by comma)
// query params sort,direction to sort asc or desc
// query params size,offset for pagination
func (server *APIServer) GetAllDiscussions(ctx context.Context, req *compassv1beta1.GetAllDiscussionsRequest) (*compassv1beta1.GetAllDiscussionsResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	flt, err := server.buildGetAllDiscussionsFilter(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	dscs, err := server.discussionService.GetDiscussions(ctx, flt)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	discussionsProto := []*compassv1beta1.Discussion{}
	for _, dsc := range dscs {
		discussionsProto = append(discussionsProto, dsc.ToProto())
	}

	return &compassv1beta1.GetAllDiscussionsResponse{Data: discussionsProto}, nil
}

// Create will create a new discussion
// field title, body, and type are mandatory
func (server *APIServer) CreateDiscussion(ctx context.Context, req *compassv1beta1.CreateDiscussionRequest) (*compassv1beta1.CreateDiscussionResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	dsc := discussion.Discussion{
		Title:     req.Title,
		Body:      req.Body,
		Type:      discussion.Type(req.Type),
		State:     discussion.GetStateEnum(req.State),
		Labels:    req.GetLabels(),
		Assets:    req.Assets,
		Assignees: req.Assignees,
		Owner:     user.User{ID: userID},
	}

	if err := dsc.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	id, err := server.discussionService.CreateDiscussion(ctx, &dsc)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.CreateDiscussionResponse{Id: id}, nil
}

func (server *APIServer) GetDiscussion(ctx context.Context, req *compassv1beta1.GetDiscussionRequest) (*compassv1beta1.GetDiscussionResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := server.validateIDInteger(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.ErrInvalidID))
	}

	dsc, err := server.discussionService.GetDiscussion(ctx, req.Id)
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetDiscussionResponse{Data: dsc.ToProto()}, nil
}

// Patch updates a specific field in discussion
// empty array in assets,labels,assignees will be considered
// and clear all assets,labels,assignees from the discussion
func (server *APIServer) PatchDiscussion(ctx context.Context, req *compassv1beta1.PatchDiscussionRequest) (*compassv1beta1.PatchDiscussionResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := server.validateIDInteger(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(discussion.ErrInvalidID))
	}

	dsc := discussion.Discussion{
		ID:        req.GetId(),
		Title:     req.GetTitle(),
		Body:      req.GetBody(),
		Type:      discussion.Type(req.GetType()),
		State:     discussion.State(req.GetState()),
		Labels:    req.GetLabels(),
		Assets:    req.GetAssets(),
		Assignees: req.GetAssignees(),
	}

	if isEmpty := dsc.IsEmpty(); isEmpty {
		err := errors.New("empty discussion body")
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	if err := dsc.ValidateConstraint(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := server.discussionService.PatchDiscussion(ctx, &dsc)
	if errors.Is(err, discussion.ErrInvalidID) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(discussion.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.PatchDiscussionResponse{}, nil
}

func (server *APIServer) validateIDInteger(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return err
	}

	if idInt < 1 {
		return errors.New("id cannot be < 1")
	}

	return nil
}
