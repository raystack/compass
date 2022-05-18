package handlersv1beta1

//go:generate mockery --name=UserService -r --case underscore --with-expecter --structname UserService --filename user_service.go --output=./mocks
import (
	"context"
	"errors"
	"strings"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/discussion"
	"github.com/odpf/compass/core/star"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/core/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService interface {
	ValidateUser(ctx context.Context, uuid, email string) (string, error)
}

func (server *APIServer) GetUserStarredAssets(ctx context.Context, req *compassv1beta1.GetUserStarredAssetsRequest) (*compassv1beta1.GetUserStarredAssetsResponse, error) {

	starFilter := star.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}

	starredAssets, err := server.starService.GetStarredAssetsByUserID(ctx, starFilter, req.GetUserId())

	if errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(star.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	var starredAssetsPB []*compassv1beta1.Asset
	for _, ast := range starredAssets {
		astPB, err := ast.ToProto(false)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		starredAssetsPB = append(starredAssetsPB, astPB)
	}

	return &compassv1beta1.GetUserStarredAssetsResponse{
		Data: starredAssetsPB,
	}, nil
}

func (server *APIServer) GetMyStarredAssets(ctx context.Context, req *compassv1beta1.GetMyStarredAssetsRequest) (*compassv1beta1.GetMyStarredAssetsResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	starFilter := star.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}

	starredAssets, err := server.starService.GetStarredAssetsByUserID(ctx, starFilter, userID)

	if errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(star.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	var starredAssetsPB []*compassv1beta1.Asset
	for _, ast := range starredAssets {
		astPB, err := ast.ToProto(false)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		starredAssetsPB = append(starredAssetsPB, astPB)
	}

	return &compassv1beta1.GetMyStarredAssetsResponse{
		Data: starredAssetsPB,
	}, nil
}

func (server *APIServer) GetMyStarredAsset(ctx context.Context, req *compassv1beta1.GetMyStarredAssetRequest) (*compassv1beta1.GetMyStarredAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	ast, err := server.starService.GetStarredAssetByUserID(ctx, userID, req.GetAssetId())
	if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(star.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	astPB, err := ast.ToProto(false)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetMyStarredAssetResponse{
		Data: astPB,
	}, nil
}

func (server *APIServer) StarAsset(ctx context.Context, req *compassv1beta1.StarAssetRequest) (*compassv1beta1.StarAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	starID, err := server.starService.Stars(ctx, userID, req.GetAssetId())
	if err != nil {
		if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(star.UserNotFoundError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(star.DuplicateRecordError)) {
			// idempotent
			return &compassv1beta1.StarAssetResponse{
				Id: starID,
			}, nil
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.StarAssetResponse{
		Id: starID,
	}, nil
}

func (server *APIServer) UnstarAsset(ctx context.Context, req *compassv1beta1.UnstarAssetRequest) (*compassv1beta1.UnstarAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	err := server.starService.Unstars(ctx, userID, req.GetAssetId())
	if err != nil {
		if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(star.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.UnstarAssetResponse{}, nil
}

func (server *APIServer) GetMyDiscussions(ctx context.Context, req *compassv1beta1.GetMyDiscussionsRequest) (*compassv1beta1.GetMyDiscussionsResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	flt, err := server.buildGetDiscussionsFilter(req, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dscs, err := server.discussionService.GetDiscussions(ctx, flt)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	var dscsPB []*compassv1beta1.Discussion
	for _, dsc := range dscs {
		dscsPB = append(dscsPB, dsc.ToProto())
	}

	return &compassv1beta1.GetMyDiscussionsResponse{
		Data: dscsPB,
	}, nil
}

func (server *APIServer) buildGetDiscussionsFilter(req *compassv1beta1.GetMyDiscussionsRequest, userID string) (discussion.Filter, error) {
	fl := discussion.Filter{
		Type:          req.GetType(),
		State:         req.GetState(),
		SortBy:        req.GetSort(),
		SortDirection: req.GetDirection(),
		Size:          int(req.GetSize()),
		Offset:        int(req.GetOffset()),
	}

	filterQuery := req.GetFilter()
	if err := validator.ValidateOneOf(filterQuery, "assigned", "created", "all"); err != nil {
		return discussion.Filter{}, err
	}

	if len(strings.TrimSpace(filterQuery)) > 0 {
		if !(filterQuery == "created" || filterQuery == "all") {
			filterQuery = "assigned" // default value
		}
	}

	switch filterQuery {
	case "all":
		fl.Owner = userID
		fl.Assignees = []string{userID}
		fl.DisjointAssigneeOwner = true
	case "created":
		fl.Owner = userID
	default:
		fl.Assignees = []string{userID}
	}

	assets := req.GetAsset()
	if assets != "" {
		fl.Assets = strings.Split(assets, ",")
	}

	labels := req.GetLabels()
	if labels != "" {
		fl.Labels = strings.Split(labels, ",")
	}

	if err := fl.Validate(); err != nil {
		return discussion.Filter{}, err
	}

	fl.AssignDefault()
	return fl, nil
}
