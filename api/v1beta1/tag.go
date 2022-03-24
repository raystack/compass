package v1beta1

import (
	"context"
	"errors"
	"fmt"

	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/tag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errEmptyRecordURN   = errors.New("record urn is empty")
	errEmptyRecordType  = errors.New("type is empty")
	errNilTagService    = errors.New("tag service is nil")
	errEmptyTemplateURN = errors.New("template urn is empty")
)

// GetTagsByRecordAndTemplate handles get tag by record requests
func (h *Handler) GetTagsByRecordAndTemplate(ctx context.Context, req *compassv1beta1.GetTagsByRecordAndTemplateRequest) (*compassv1beta1.GetTagsByRecordAndTemplateResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordType.Error())
	}
	if req.GetRecordUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordURN.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	tg, err := h.TagService.FindByRecordAndTemplate(ctx, req.GetType(), req.GetRecordUrn(), req.GetTemplateUrn())
	if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error finding a tag with record and template: %s", err.Error()))
	}

	tagPB, err := tg.ToProto()
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.GetTagsByRecordAndTemplateResponse{
		Data: tagPB,
	}, nil
}

// CreateTag handles tag creation requests
func (h *Handler) CreateTag(ctx context.Context, req *compassv1beta1.CreateTagRequest) (*compassv1beta1.CreateTagResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetRecordType() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordType.Error())
	}
	if req.GetRecordUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordURN.Error())
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
		RecordType:          req.GetRecordType(),
		RecordURN:           req.GetRecordUrn(),
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

	return &compassv1beta1.CreateTagResponse{
		Data: tagPB,
	}, nil
}

// UpdateTag handles tag update requests
func (h *Handler) UpdateTag(ctx context.Context, req *compassv1beta1.UpdateTagRequest) (*compassv1beta1.UpdateTagResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordType.Error())
	}
	if req.GetRecordUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordURN.Error())
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
		RecordType:          req.GetType(),
		RecordURN:           req.GetRecordUrn(),
		TemplateURN:         req.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.GetTemplateDisplayName(),
		TemplateDescription: req.GetTemplateDescription(),
	}

	err := h.TagService.Update(ctx, &tagDomain)
	if errors.As(err, new(tag.NotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if errors.As(err, new(tag.ValidationError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error updating a template: %s", err.Error()))
	}

	tagPB, err := tagDomain.ToProto()
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.UpdateTagResponse{
		Data: tagPB,
	}, nil
}

// DeleteTag handles delete tag by record and template requests
func (h *Handler) DeleteTag(ctx context.Context, req *compassv1beta1.DeleteTagRequest) (*compassv1beta1.DeleteTagResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordType.Error())
	}
	if req.GetRecordUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordURN.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	err := h.TagService.Delete(ctx, req.GetType(), req.GetRecordUrn(), req.GetTemplateUrn())
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error deleting a template: %s", err.Error()))
	}

	return &compassv1beta1.DeleteTagResponse{}, nil
}

// GetByRecord handles get tag by record requests
func (h *Handler) GetTagsByRecord(ctx context.Context, req *compassv1beta1.GetTagsByRecordRequest) (*compassv1beta1.GetTagsByRecordResponse, error) {
	if h.TagService == nil {
		return nil, internalServerError(h.Logger, errNilTagService.Error())
	}

	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordType.Error())
	}
	if req.GetRecordUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyRecordURN.Error())
	}

	tags, err := h.TagService.GetByRecord(ctx, req.GetType(), req.GetRecordUrn())
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error getting record tags: %s", err.Error()))
	}

	var tagsPB []*compassv1beta1.Tag
	for _, tg := range tags {
		tgPB, err := tg.ToProto()
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
		}
		tagsPB = append(tagsPB, tgPB)
	}

	return &compassv1beta1.GetTagsByRecordResponse{
		Data: tagsPB,
	}, nil
}
