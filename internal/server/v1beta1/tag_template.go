package handlersv1beta1

//go:generate mockery --name=TagTemplateService -r --case underscore --with-expecter --structname TagTemplateService --filename tag_template_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/pkg/grpc_interceptor"
	"time"

	"github.com/raystack/compass/core/tag"
	compassv1beta1 "github.com/raystack/compass/proto/raystack/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TagTemplateService interface {
	Validate(template tag.Template) error
	CreateTemplate(ctx context.Context, ns *namespace.Namespace, template *tag.Template) error
	GetTemplates(ctx context.Context, templateURN string) ([]tag.Template, error)
	UpdateTemplate(ctx context.Context, ns *namespace.Namespace, templateURN string, template *tag.Template) error
	GetTemplate(ctx context.Context, urn string) (tag.Template, error)
	DeleteTemplate(ctx context.Context, urn string) error
}

// GetAllTagTemplates handles template read requests
func (server *APIServer) GetAllTagTemplates(ctx context.Context, req *compassv1beta1.GetAllTagTemplatesRequest) (*compassv1beta1.GetAllTagTemplatesResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}
	ns := grpc_interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	listOfDomainTemplate, err := server.tagTemplateService.GetTemplates(ctx, req.GetUrn())
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding templates: %s", err.Error()))
	}

	var templatesPB []*compassv1beta1.TagTemplate
	for _, template := range listOfDomainTemplate {
		templatesPB = append(templatesPB, tagTemplateToProto(template))
	}

	return &compassv1beta1.GetAllTagTemplatesResponse{
		Data: templatesPB,
	}, nil
}

// CreateTagTemplate handles template creation requests
func (server *APIServer) CreateTagTemplate(ctx context.Context, req *compassv1beta1.CreateTagTemplateRequest) (*compassv1beta1.CreateTagTemplateResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}
	ns := grpc_interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if req.GetUrn() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty urn")
	}
	if req.GetDisplayName() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty display name")
	}
	if req.GetDescription() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty description")
	}
	if req.GetFields() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty fields")
	}

	var templateFields []tag.Field
	for _, fPB := range req.GetFields() {
		templateFields = append(templateFields, tagTemplateFieldFromProto(fPB))
	}

	template := tag.Template{
		URN:         req.GetUrn(),
		DisplayName: req.GetDisplayName(),
		Description: req.GetDescription(),
		Fields:      templateFields,
	}
	err := server.tagTemplateService.CreateTemplate(ctx, ns, &template)
	if errors.As(err, new(tag.DuplicateTemplateError)) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error creating tag template: %s", err.Error()))
	}

	return &compassv1beta1.CreateTagTemplateResponse{
		Data: tagTemplateToProto(template),
	}, nil
}

// GetTagTemplate handles template read requests based on URN
func (server *APIServer) GetTagTemplate(ctx context.Context, req *compassv1beta1.GetTagTemplateRequest) (*compassv1beta1.GetTagTemplateResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}
	ns := grpc_interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	domainTemplate, err := server.tagTemplateService.GetTemplate(ctx, req.GetTemplateUrn())
	if err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding a template: %s", err.Error()))
	}

	return &compassv1beta1.GetTagTemplateResponse{
		Data: tagTemplateToProto(domainTemplate),
	}, nil
}

func (server *APIServer) UpdateTagTemplate(ctx context.Context, req *compassv1beta1.UpdateTagTemplateRequest) (*compassv1beta1.UpdateTagTemplateResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}
	ns := grpc_interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if req.GetDisplayName() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty display name")
	}
	if req.GetDescription() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty description")
	}
	if req.GetFields() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty fields")
	}

	var templateFields []tag.Field
	for _, fPB := range req.GetFields() {
		templateFields = append(templateFields, tagTemplateFieldFromProto(fPB))
	}

	template := tag.Template{
		URN:         req.GetTemplateUrn(),
		DisplayName: req.GetDisplayName(),
		Description: req.GetDescription(),
		Fields:      templateFields,
	}
	if err := server.tagTemplateService.UpdateTemplate(ctx, ns, req.TemplateUrn, &template); err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.As(err, new(tag.ValidationError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error updating template: %s", err.Error()))
	}

	return &compassv1beta1.UpdateTagTemplateResponse{
		Data: tagTemplateToProto(template),
	}, nil
}

// DeleteTagTemplate handles template delete request based on URN
func (server *APIServer) DeleteTagTemplate(ctx context.Context, req *compassv1beta1.DeleteTagTemplateRequest) (*compassv1beta1.DeleteTagTemplateResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}
	ns := grpc_interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	err := server.tagTemplateService.DeleteTemplate(ctx, req.GetTemplateUrn())
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error deleting a template: %s", err.Error()))
	}
	return &compassv1beta1.DeleteTagTemplateResponse{}, nil
}

// tagTemplateToProto convert domain to protobuf
func tagTemplateToProto(t tag.Template) *compassv1beta1.TagTemplate {
	var templateFieldsPB []*compassv1beta1.TagTemplateField
	for _, tf := range t.Fields {
		templateFieldsPB = append(templateFieldsPB, tagTemplateFieldToProto(tf))
	}

	var createdAtPB *timestamppb.Timestamp
	if !t.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(t.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !t.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(t.UpdatedAt)
	}

	return &compassv1beta1.TagTemplate{
		Urn:         t.URN,
		DisplayName: t.DisplayName,
		Description: t.Description,
		Fields:      templateFieldsPB,
		CreatedAt:   createdAtPB,
		UpdatedAt:   updatedAtPB,
	}
}

// tagTemplateFromProto converts proto to tag.Template
func tagTemplateFromProto(pb *compassv1beta1.TagTemplate) tag.Template {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var fields []tag.Field
	if pb.GetFields() != nil {
		for _, tfPB := range pb.GetFields() {
			fields = append(fields, tagTemplateFieldFromProto(tfPB))
		}
	}

	return tag.Template{
		URN:         pb.GetUrn(),
		DisplayName: pb.GetDisplayName(),
		Description: pb.GetDescription(),
		Fields:      fields,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// tagTemplateFieldToProto convert domain to protobuf
func tagTemplateFieldToProto(f tag.Field) *compassv1beta1.TagTemplateField {
	var createdAtPB *timestamppb.Timestamp
	if !f.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(f.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !f.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(f.UpdatedAt)
	}

	return &compassv1beta1.TagTemplateField{
		Id:          uint32(f.ID),
		Urn:         f.URN,
		DisplayName: f.DisplayName,
		Description: f.Description,
		DataType:    f.DataType,
		Options:     f.Options,
		Required:    f.Required,
		CreatedAt:   createdAtPB,
		UpdatedAt:   updatedAtPB,
	}
}

// tagTemplateFieldFromProto converts proto to tag.Field
func tagTemplateFieldFromProto(pb *compassv1beta1.TagTemplateField) tag.Field {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	return tag.Field{
		ID:          uint(pb.GetId()),
		URN:         pb.GetUrn(),
		DisplayName: pb.GetDisplayName(),
		Description: pb.GetDescription(),
		DataType:    pb.GetDataType(),
		Options:     pb.GetOptions(),
		Required:    pb.GetRequired(),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
