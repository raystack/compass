package v1beta1

import (
	"context"
	"errors"
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/tag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errEmptyAssetID     = errors.New("asset id is empty")
	errNilTagService    = errors.New("tag service is nil")
	errEmptyTemplateURN = errors.New("template urn is empty")
)

// GetTagsByAssetAndTemplate handles get tag by asset requests
func (h *Handler) GetTagsByAssetAndTemplate(ctx context.Context, req *compassv1beta1.GetTagsByAssetAndTemplateRequest) (*compassv1beta1.GetTagsByAssetAndTemplateResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	tg, err := h.TagService.FindByAssetAndTemplate(ctx, req.GetAssetId(), req.GetTemplateUrn())
	if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error finding a tag with asset and template: %s", err.Error()))
	}

	tagPB, err := tg.ToProto()
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.GetTagsByAssetAndTemplateResponse{
		Data: tagPB,
	}, nil
}

// CreateTagAsset handles tag creation requests
func (h *Handler) CreateTagAsset(ctx context.Context, req *compassv1beta1.CreateTagAssetRequest) (*compassv1beta1.CreateTagAssetResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}
	if req.GetTagValues() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty tag values")
	}

	var tagValues []tag.TagValue
	for _, tvPB := range req.GetTagValues() {
		tagValues = append(tagValues, tag.NewTagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.GetAssetId(),
		TemplateURN:         req.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.GetTemplateDisplayName(),
		TemplateDescription: req.GetTemplateDescription(),
	}

	err := h.TagService.Create(ctx, &tagDomain)
	if errors.As(err, new(tag.DuplicateError)) {
		return nil, status.Error(codes.AlreadyExists, err.Error())

	}
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if errors.As(err, new(tag.ValidationError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error creating tag: %s", err.Error()))

	}

	tagPB, err := tagDomain.ToProto()
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.CreateTagAssetResponse{
		Data: tagPB,
	}, nil
}

// UpdateTagAsset handles tag update requests
func (h *Handler) UpdateTagAsset(ctx context.Context, req *compassv1beta1.UpdateTagAssetRequest) (*compassv1beta1.UpdateTagAssetResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	if req.GetTagValues() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty tag values")
	}

	var tagValues []tag.TagValue
	for _, tvPB := range req.GetTagValues() {
		tagValues = append(tagValues, tag.NewTagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.GetAssetId(),
		TemplateURN:         req.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.GetTemplateDisplayName(),
		TemplateDescription: req.GetTemplateDescription(),
	}

	err := h.TagService.Update(ctx, &tagDomain)
	if err != nil {
		if errors.As(err, new(tag.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.As(err, new(tag.ValidationError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, internalServerError(h.Logger, fmt.Sprintf("error updating an asset's tag: %s", err.Error()))
	}

	tagPB, err := tagDomain.ToProto()
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.UpdateTagAssetResponse{
		Data: tagPB,
	}, nil
}

// DeleteTagAsset handles delete tag by asset and template requests
func (h *Handler) DeleteTagAsset(ctx context.Context, req *compassv1beta1.DeleteTagAssetRequest) (*compassv1beta1.DeleteTagAssetResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	err := h.TagService.Delete(ctx, req.GetAssetId(), req.GetTemplateUrn())
	if err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) || errors.As(err, new(tag.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(h.Logger, fmt.Sprintf("error deleting a tag: %s", err.Error()))
	}

	return &compassv1beta1.DeleteTagAssetResponse{}, nil
}

// GetAllTagsByAsset handles get all tags by asset requests
func (h *Handler) GetAllTagsByAsset(ctx context.Context, req *compassv1beta1.GetAllTagsByAssetRequest) (*compassv1beta1.GetAllTagsByAssetResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}

	tags, err := h.TagService.GetByAsset(ctx, req.GetAssetId())
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error getting asset tags: %s", err.Error()))
	}

	var tagsPB []*compassv1beta1.Tag
	for _, tg := range tags {
		tgPB, err := tg.ToProto()
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
		}
		tagsPB = append(tagsPB, tgPB)
	}

	return &compassv1beta1.GetAllTagsByAssetResponse{
		Data: tagsPB,
	}, nil
}
