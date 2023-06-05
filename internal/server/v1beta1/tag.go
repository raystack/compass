package handlersv1beta1

//go:generate mockery --name=TagService -r --case underscore --with-expecter --structname TagService --filename tag_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goto/compass/core/tag"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	errEmptyAssetID     = errors.New("asset id is empty")
	errNilTagService    = errors.New("tag service is nil")
	errEmptyTemplateURN = errors.New("template urn is empty")
)

type TagService interface {
	Validate(tag *tag.Tag) error
	CreateTag(ctx context.Context, tag *tag.Tag) error
	GetTagsByAssetID(ctx context.Context, assetID string) ([]tag.Tag, error)
	FindTagByAssetIDAndTemplateURN(ctx context.Context, assetID, templateURN string) (tag.Tag, error)
	DeleteTag(ctx context.Context, assetID, templateURN string) error
	UpdateTag(ctx context.Context, tag *tag.Tag) error
}

// GetTagByAssetAndTemplate handles get tag by asset requests
func (server *APIServer) GetTagByAssetAndTemplate(ctx context.Context, req *compassv1beta1.GetTagByAssetAndTemplateRequest) (*compassv1beta1.GetTagByAssetAndTemplateResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	tg, err := server.tagService.FindTagByAssetIDAndTemplateURN(ctx, req.GetAssetId(), req.GetTemplateUrn())
	if err != nil {
		if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding a tag with asset and template: %s", err.Error()))
	}

	tagPB, err := tagToProto(tg)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetTagByAssetAndTemplateResponse{
		Data: tagPB,
	}, nil
}

// CreateTagAsset handles tag creation requests
func (server *APIServer) CreateTagAsset(ctx context.Context, req *compassv1beta1.CreateTagAssetRequest) (*compassv1beta1.CreateTagAssetResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
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
		tagValues = append(tagValues, tagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.GetAssetId(),
		TemplateURN:         req.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.GetTemplateDisplayName(),
		TemplateDescription: req.GetTemplateDescription(),
	}

	err = server.tagService.CreateTag(ctx, &tagDomain)
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
		return nil, internalServerError(server.logger, fmt.Sprintf("error creating tag: %s", err.Error()))
	}

	tagPB, err := tagToProto(tagDomain)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.CreateTagAssetResponse{
		Data: tagPB,
	}, nil
}

// UpdateTagAsset handles tag update requests
func (server *APIServer) UpdateTagAsset(ctx context.Context, req *compassv1beta1.UpdateTagAssetRequest) (*compassv1beta1.UpdateTagAssetResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
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
		tagValues = append(tagValues, tagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.GetAssetId(),
		TemplateURN:         req.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.GetTemplateDisplayName(),
		TemplateDescription: req.GetTemplateDescription(),
	}

	err = server.tagService.UpdateTag(ctx, &tagDomain)
	if err != nil {
		if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.As(err, new(tag.ValidationError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error updating an asset's tag: %s", err.Error()))
	}

	tagPB, err := tagToProto(tagDomain)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.UpdateTagAssetResponse{
		Data: tagPB,
	}, nil
}

// DeleteTagAsset handles delete tag by asset and template requests
func (server *APIServer) DeleteTagAsset(ctx context.Context, req *compassv1beta1.DeleteTagAssetRequest) (*compassv1beta1.DeleteTagAssetResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	err = server.tagService.DeleteTag(ctx, req.GetAssetId(), req.GetTemplateUrn())
	if err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) || errors.As(err, new(tag.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error deleting a tag: %s", err.Error()))
	}

	return &compassv1beta1.DeleteTagAssetResponse{}, nil
}

// GetAllTagsByAsset handles get all tags by asset requests
func (server *APIServer) GetAllTagsByAsset(ctx context.Context, req *compassv1beta1.GetAllTagsByAssetRequest) (*compassv1beta1.GetAllTagsByAssetResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}

	tags, err := server.tagService.GetTagsByAssetID(ctx, req.GetAssetId())
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error getting asset tags: %s", err.Error()))
	}

	var tagsPB []*compassv1beta1.Tag
	for _, tg := range tags {
		tgPB, err := tagToProto(tg)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		tagsPB = append(tagsPB, tgPB)
	}

	return &compassv1beta1.GetAllTagsByAssetResponse{
		Data: tagsPB,
	}, nil
}

// tagToProto convert domain to protobuf
func tagToProto(t tag.Tag) (*compassv1beta1.Tag, error) {
	var tagValuesPB []*compassv1beta1.TagValue
	for _, tv := range t.TagValues {
		tvPB, err := tagValueToProto(tv)
		if err != nil {
			return nil, err
		}
		tagValuesPB = append(tagValuesPB, tvPB)
	}

	return &compassv1beta1.Tag{
		AssetId:             t.AssetID,
		TemplateUrn:         t.TemplateURN,
		TagValues:           tagValuesPB,
		TemplateDisplayName: t.TemplateDisplayName,
		TemplateDescription: t.TemplateDescription,
	}, nil
}

// tagFromProto converts proto to tag.Tag
func tagFromProto(pb *compassv1beta1.Tag) tag.Tag {
	var tagValues []tag.TagValue
	if pb.GetTagValues() != nil {
		for _, tvPB := range pb.GetTagValues() {
			tv := tagValueFromProto(tvPB)
			tagValues = append(tagValues, tv)
		}
	}

	return tag.Tag{
		AssetID:             pb.GetAssetId(),
		TemplateURN:         pb.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: pb.GetTemplateDisplayName(),
		TemplateDescription: pb.GetTemplateDescription(),
	}
}

// tagValueToProto convert domain to protobuf
func tagValueToProto(tv tag.TagValue) (*compassv1beta1.TagValue, error) {
	var err error
	var createdAtPB *timestamppb.Timestamp
	if !tv.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(tv.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !tv.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(tv.UpdatedAt)
	}

	var fieldValuePB *structpb.Value
	if tv.FieldValue != nil {
		fieldValuePB, err = structpb.NewValue(tv.FieldValue)
		if err != nil {
			return nil, err
		}
	}

	return &compassv1beta1.TagValue{
		FieldId:          uint32(tv.FieldID),
		FieldValue:       fieldValuePB,
		FieldUrn:         tv.FieldURN,
		FieldDisplayName: tv.FieldDisplayName,
		FieldDescription: tv.FieldDescription,
		FieldDataType:    tv.FieldDataType,
		FieldOptions:     tv.FieldOptions,
		FieldRequired:    tv.FieldRequired,
		CreatedAt:        createdAtPB,
		UpdatedAt:        updatedAtPB,
	}, nil
}

// tagValueFromProto converts proto to tag.TagValue
func tagValueFromProto(pb *compassv1beta1.TagValue) tag.TagValue {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var fieldValue interface{}
	if pb.GetFieldValue() != nil {
		fieldValue = pb.GetFieldValue().AsInterface()
	}

	return tag.TagValue{
		FieldID:          uint(pb.GetFieldId()),
		FieldValue:       fieldValue,
		FieldURN:         pb.GetFieldUrn(),
		FieldDisplayName: pb.GetFieldDisplayName(),
		FieldDescription: pb.GetFieldDescription(),
		FieldDataType:    pb.GetFieldDataType(),
		FieldOptions:     pb.GetFieldOptions(),
		FieldRequired:    pb.GetFieldRequired(),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}
