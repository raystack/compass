package handlersv1beta1

//go:generate mockery --name=TagService -r --case underscore --with-expecter --structname TagService --filename tag_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/tag"
	"github.com/raystack/compass/pkg/server/interceptor"
	compassv1beta1 "github.com/raystack/compass/proto/compassv1beta1"
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
	CreateTag(ctx context.Context, ns *namespace.Namespace, tag *tag.Tag) error
	GetTagsByAssetID(ctx context.Context, assetID string) ([]tag.Tag, error)
	FindTagByAssetIDAndTemplateURN(ctx context.Context, assetID, templateURN string) (tag.Tag, error)
	DeleteTag(ctx context.Context, assetID, templateURN string) error
	UpdateTag(ctx context.Context, tag *tag.Tag) error
}

// GetTagByAssetAndTemplate handles get tag by asset requests
func (server *APIServer) GetTagByAssetAndTemplate(ctx context.Context, req *connect.Request[compassv1beta1.GetTagByAssetAndTemplateRequest]) (*connect.Response[compassv1beta1.GetTagByAssetAndTemplateResponse], error) {
	if err := req.Msg.ValidateAll(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s", bodyParserErrorMsg(err)))
	}
	ns := interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.Msg.GetAssetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyAssetID)
	}
	if req.Msg.GetTemplateUrn() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyTemplateURN)
	}

	tg, err := server.tagService.FindTagByAssetIDAndTemplateURN(ctx, req.Msg.GetAssetId(), req.Msg.GetTemplateUrn())
	if err != nil {
		if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding a tag with asset and template: %s", err.Error()))
	}

	tagPB, err := tagToProto(tg)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return connect.NewResponse(&compassv1beta1.GetTagByAssetAndTemplateResponse{
		Data: tagPB,
	}), nil
}

// CreateTagAsset handles tag creation requests
func (server *APIServer) CreateTagAsset(ctx context.Context, req *connect.Request[compassv1beta1.CreateTagAssetRequest]) (*connect.Response[compassv1beta1.CreateTagAssetResponse], error) {
	if err := req.Msg.ValidateAll(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s", bodyParserErrorMsg(err)))
	}
	ns := interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.Msg.GetAssetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyAssetID)
	}
	if req.Msg.GetTemplateUrn() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyTemplateURN)
	}
	if len(req.Msg.GetTagValues()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty tag values"))
	}

	var tagValues []tag.TagValue
	for _, tvPB := range req.Msg.GetTagValues() {
		tagValues = append(tagValues, tagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.Msg.GetAssetId(),
		TemplateURN:         req.Msg.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.Msg.GetTemplateDisplayName(),
		TemplateDescription: req.Msg.GetTemplateDescription(),
	}

	err := server.tagService.CreateTag(ctx, ns, &tagDomain)
	if errors.As(err, new(tag.DuplicateError)) {
		return nil, connect.NewError(connect.CodeAlreadyExists, err)
	}
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if errors.As(err, new(tag.ValidationError)) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error creating tag: %s", err.Error()))
	}

	tagPB, err := tagToProto(tagDomain)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return connect.NewResponse(&compassv1beta1.CreateTagAssetResponse{
		Data: tagPB,
	}), nil
}

// UpdateTagAsset handles tag update requests
func (server *APIServer) UpdateTagAsset(ctx context.Context, req *connect.Request[compassv1beta1.UpdateTagAssetRequest]) (*connect.Response[compassv1beta1.UpdateTagAssetResponse], error) {
	if err := req.Msg.ValidateAll(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s", bodyParserErrorMsg(err)))
	}
	ns := interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.Msg.GetAssetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyAssetID)
	}
	if req.Msg.GetTemplateUrn() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyTemplateURN)
	}

	if len(req.Msg.GetTagValues()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty tag values"))
	}

	var tagValues []tag.TagValue
	for _, tvPB := range req.Msg.GetTagValues() {
		tagValues = append(tagValues, tagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.Msg.GetAssetId(),
		TemplateURN:         req.Msg.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.Msg.GetTemplateDisplayName(),
		TemplateDescription: req.Msg.GetTemplateDescription(),
	}

	err := server.tagService.UpdateTag(ctx, &tagDomain)
	if err != nil {
		if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.As(err, new(tag.ValidationError)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error updating an asset's tag: %s", err.Error()))
	}

	tagPB, err := tagToProto(tagDomain)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return connect.NewResponse(&compassv1beta1.UpdateTagAssetResponse{
		Data: tagPB,
	}), nil
}

// DeleteTagAsset handles delete tag by asset and template requests
func (server *APIServer) DeleteTagAsset(ctx context.Context, req *connect.Request[compassv1beta1.DeleteTagAssetRequest]) (*connect.Response[compassv1beta1.DeleteTagAssetResponse], error) {
	if err := req.Msg.ValidateAll(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s", bodyParserErrorMsg(err)))
	}
	ns := interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.Msg.GetAssetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyAssetID)
	}
	if req.Msg.GetTemplateUrn() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyTemplateURN)
	}

	err := server.tagService.DeleteTag(ctx, req.Msg.GetAssetId(), req.Msg.GetTemplateUrn())
	if err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) || errors.As(err, new(tag.NotFoundError)) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error deleting a tag: %s", err.Error()))
	}

	return connect.NewResponse(&compassv1beta1.DeleteTagAssetResponse{}), nil
}

// GetAllTagsByAsset handles get all tags by asset requests
func (server *APIServer) GetAllTagsByAsset(ctx context.Context, req *connect.Request[compassv1beta1.GetAllTagsByAssetRequest]) (*connect.Response[compassv1beta1.GetAllTagsByAssetResponse], error) {
	if err := req.Msg.ValidateAll(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s", bodyParserErrorMsg(err)))
	}
	ns := interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.Msg.GetAssetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errEmptyAssetID)
	}

	tags, err := server.tagService.GetTagsByAssetID(ctx, req.Msg.GetAssetId())
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

	return connect.NewResponse(&compassv1beta1.GetAllTagsByAssetResponse{
		Data: tagsPB,
	}), nil
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
