package v1beta1

import (
	"context"
	"errors"
	"strings"

	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
	"github.com/odpf/columbus/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) GetUserStarredAssets(ctx context.Context, req *compassv1beta1.GetUserStarredAssetsRequest) (*compassv1beta1.GetUserStarredAssetsResponse, error) {

	starFilter := star.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}

	starredAssets, err := h.StarRepository.GetAllAssetsByUserID(ctx, starFilter, req.GetUserId())

	if errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(star.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	var starredAssetsPB []*compassv1beta1.Asset
	for _, ast := range starredAssets {
		astPB, err := ast.ToProto(false)
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
		}
		starredAssetsPB = append(starredAssetsPB, astPB)
	}

	return &compassv1beta1.GetUserStarredAssetsResponse{
		Data: starredAssetsPB,
	}, nil
}

func (h *Handler) GetMyStarredAssets(ctx context.Context, req *compassv1beta1.GetMyStarredAssetsRequest) (*compassv1beta1.GetMyStarredAssetsResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	starFilter := star.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}

	starredAssets, err := h.StarRepository.GetAllAssetsByUserID(ctx, starFilter, userID)

	if errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(star.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	var starredAssetsPB []*compassv1beta1.Asset
	for _, ast := range starredAssets {
		astPB, err := ast.ToProto(false)
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
		}
		starredAssetsPB = append(starredAssetsPB, astPB)
	}

	return &compassv1beta1.GetMyStarredAssetsResponse{
		Data: starredAssetsPB,
	}, nil
}

func (h *Handler) GetMyStarredAsset(ctx context.Context, req *compassv1beta1.GetMyStarredAssetRequest) (*compassv1beta1.GetMyStarredAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	ast, err := h.StarRepository.GetAssetByUserID(ctx, userID, req.GetAssetId())
	if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(star.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	astPB, err := ast.ToProto(false)
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.GetMyStarredAssetResponse{
		Data: astPB,
	}, nil
}

func (h *Handler) StarAsset(ctx context.Context, req *compassv1beta1.StarAssetRequest) (*compassv1beta1.StarAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	starID, err := h.StarRepository.Create(ctx, userID, req.GetAssetId())
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
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.StarAssetResponse{
		Id: starID,
	}, nil
}

func (h *Handler) UnstarAsset(ctx context.Context, req *compassv1beta1.UnstarAssetRequest) (*compassv1beta1.UnstarAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	err := h.StarRepository.Delete(ctx, userID, req.GetAssetId())
	if errors.Is(err, star.ErrEmptyAssetID) || errors.Is(err, star.ErrEmptyUserID) || errors.As(err, new(star.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.As(err, new(star.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.UnstarAssetResponse{}, nil
}

func (h *Handler) GetMyDiscussions(ctx context.Context, req *compassv1beta1.GetMyDiscussionsRequest) (*compassv1beta1.GetMyDiscussionsResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	flt, err := h.buildGetDiscussionsFilter(req, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dscs, err := h.DiscussionRepository.GetAll(ctx, flt)
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	var dscsPB []*compassv1beta1.Discussion
	for _, dsc := range dscs {
		dscsPB = append(dscsPB, dsc.ToProto())
	}

	return &compassv1beta1.GetMyDiscussionsResponse{
		Data: dscsPB,
	}, nil
}

func (h *Handler) buildGetDiscussionsFilter(req *compassv1beta1.GetMyDiscussionsRequest, userID string) (discussion.Filter, error) {
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
