package handlersv1beta1

//go:generate mockery --name=TagService -r --case underscore --with-expecter --structname TagService --filename tag_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/tag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errEmptyAssetID     = errors.New("asset id is empty")
	errNilTagService    = errors.New("tag service is nil")
	errEmptyTemplateURN = errors.New("template urn is empty")
)

type TagService interface {
	Validate(tag *tag.Tag) error
	CreateTag(ctx context.Context, tag *tag.Tag) error
	GetTagByAssetID(ctx context.Context, assetID string) ([]tag.Tag, error)
	FindTagByAssetIDAndTemplateURN(ctx context.Context, assetID, templateURN string) (tag.Tag, error)
	DeleteTagByAssetIDAndTemplateURN(ctx context.Context, assetID, templateURN string) error
	UpdateTag(ctx context.Context, tag *tag.Tag) error
}

// GetTagsByAssetAndTemplate handles get tag by asset requests
func (server *APIServer) GetTagsByAssetAndTemplate(ctx context.Context, req *compassv1beta1.GetTagsByAssetAndTemplateRequest) (*compassv1beta1.GetTagsByAssetAndTemplateResponse, error) {
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
	if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding a tag with asset and template: %s", err.Error()))
	}

	tagPB, err := tg.ToProto()
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetTagsByAssetAndTemplateResponse{
		Data: tagPB,
	}, nil
}

// CreateTagAsset handles tag creation requests
func (server *APIServer) CreateTagAsset(ctx context.Context, req *compassv1beta1.CreateTagAssetRequest) (*compassv1beta1.CreateTagAssetResponse, error) {
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
		tagValues = append(tagValues, tag.NewTagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.GetAssetId(),
		TemplateURN:         req.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.GetTemplateDisplayName(),
		TemplateDescription: req.GetTemplateDescription(),
	}

	err := server.tagService.CreateTag(ctx, &tagDomain)
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

	tagPB, err := tagDomain.ToProto()
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.CreateTagAssetResponse{
		Data: tagPB,
	}, nil
}

// UpdateTagAsset handles tag update requests
func (server *APIServer) UpdateTagAsset(ctx context.Context, req *compassv1beta1.UpdateTagAssetRequest) (*compassv1beta1.UpdateTagAssetResponse, error) {
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
		tagValues = append(tagValues, tag.NewTagValueFromProto(tvPB))
	}

	tagDomain := tag.Tag{
		AssetID:             req.GetAssetId(),
		TemplateURN:         req.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: req.GetTemplateDisplayName(),
		TemplateDescription: req.GetTemplateDescription(),
	}

	err := server.tagService.UpdateTag(ctx, &tagDomain)
	if err != nil {
		if errors.As(err, new(tag.NotFoundError)) || errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.As(err, new(tag.ValidationError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error updating an asset's tag: %s", err.Error()))
	}

	tagPB, err := tagDomain.ToProto()
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.UpdateTagAssetResponse{
		Data: tagPB,
	}, nil
}

// DeleteTagAsset handles delete tag by asset and template requests
func (server *APIServer) DeleteTagAsset(ctx context.Context, req *compassv1beta1.DeleteTagAssetRequest) (*compassv1beta1.DeleteTagAssetResponse, error) {
	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}
	if req.GetTemplateUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyTemplateURN.Error())
	}

	err := server.tagService.DeleteTagByAssetIDAndTemplateURN(ctx, req.GetAssetId(), req.GetTemplateUrn())
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
	if server.tagService == nil {
		return nil, internalServerError(server.logger, errNilTagService.Error())
	}

	if req.GetAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errEmptyAssetID.Error())
	}

	tags, err := server.tagService.GetTagByAssetID(ctx, req.GetAssetId())
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error getting asset tags: %s", err.Error()))
	}

	var tagsPB []*compassv1beta1.Tag
	for _, tg := range tags {
		tgPB, err := tg.ToProto()
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		tagsPB = append(tagsPB, tgPB)
	}

	return &compassv1beta1.GetAllTagsByAssetResponse{
		Data: tagsPB,
	}, nil
}
