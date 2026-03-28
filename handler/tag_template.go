package handler

//go:generate mockery --name=TagTemplateService -r --case underscore --with-expecter --structname TagTemplateService --filename tag_template_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/tag"
	"github.com/raystack/compass/internal/middleware"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
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
func (server *APIServer) GetAllTagTemplates(ctx context.Context, req *connect.Request[compassv1beta1.GetAllTagTemplatesRequest]) (*connect.Response[compassv1beta1.GetAllTagTemplatesResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	listOfDomainTemplate, err := server.tagTemplateService.GetTemplates(ctx, req.Msg.GetUrn())
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding templates: %s", err.Error()))
	}

	var templatesPB []*compassv1beta1.TagTemplate
	for _, template := range listOfDomainTemplate {
		templatesPB = append(templatesPB, tagTemplateToProto(template))
	}

	return connect.NewResponse(&compassv1beta1.GetAllTagTemplatesResponse{
		Data: templatesPB,
	}), nil
}

// CreateTagTemplate handles template creation requests
func (server *APIServer) CreateTagTemplate(ctx context.Context, req *connect.Request[compassv1beta1.CreateTagTemplateRequest]) (*connect.Response[compassv1beta1.CreateTagTemplateResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if req.Msg.GetUrn() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty urn"))
	}
	if req.Msg.GetDisplayName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty display name"))
	}
	if req.Msg.GetDescription() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty description"))
	}
	if req.Msg.GetFields() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty fields"))
	}

	var templateFields []tag.Field
	for _, fPB := range req.Msg.GetFields() {
		templateFields = append(templateFields, tagTemplateFieldFromProto(fPB))
	}

	template := tag.Template{
		URN:         req.Msg.GetUrn(),
		DisplayName: req.Msg.GetDisplayName(),
		Description: req.Msg.GetDescription(),
		Fields:      templateFields,
	}
	err := server.tagTemplateService.CreateTemplate(ctx, ns, &template)
	if errors.As(err, new(tag.DuplicateTemplateError)) {
		return nil, connect.NewError(connect.CodeAlreadyExists, err)
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error creating tag template: %s", err.Error()))
	}

	return connect.NewResponse(&compassv1beta1.CreateTagTemplateResponse{
		Data: tagTemplateToProto(template),
	}), nil
}

// GetTagTemplate handles template read requests based on URN
func (server *APIServer) GetTagTemplate(ctx context.Context, req *connect.Request[compassv1beta1.GetTagTemplateRequest]) (*connect.Response[compassv1beta1.GetTagTemplateResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	domainTemplate, err := server.tagTemplateService.GetTemplate(ctx, req.Msg.GetTemplateUrn())
	if err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding a template: %s", err.Error()))
	}

	return connect.NewResponse(&compassv1beta1.GetTagTemplateResponse{
		Data: tagTemplateToProto(domainTemplate),
	}), nil
}

func (server *APIServer) UpdateTagTemplate(ctx context.Context, req *connect.Request[compassv1beta1.UpdateTagTemplateRequest]) (*connect.Response[compassv1beta1.UpdateTagTemplateResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if req.Msg.GetDisplayName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty display name"))
	}
	if req.Msg.GetDescription() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty description"))
	}
	if req.Msg.GetFields() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty fields"))
	}

	var templateFields []tag.Field
	for _, fPB := range req.Msg.GetFields() {
		templateFields = append(templateFields, tagTemplateFieldFromProto(fPB))
	}

	template := tag.Template{
		URN:         req.Msg.GetTemplateUrn(),
		DisplayName: req.Msg.GetDisplayName(),
		Description: req.Msg.GetDescription(),
		Fields:      templateFields,
	}
	if err := server.tagTemplateService.UpdateTemplate(ctx, ns, req.Msg.TemplateUrn, &template); err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.As(err, new(tag.ValidationError)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error updating template: %s", err.Error()))
	}

	return connect.NewResponse(&compassv1beta1.UpdateTagTemplateResponse{
		Data: tagTemplateToProto(template),
	}), nil
}

// DeleteTagTemplate handles template delete request based on URN
func (server *APIServer) DeleteTagTemplate(ctx context.Context, req *connect.Request[compassv1beta1.DeleteTagTemplateRequest]) (*connect.Response[compassv1beta1.DeleteTagTemplateResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	err := server.tagTemplateService.DeleteTemplate(ctx, req.Msg.GetTemplateUrn())
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error deleting a template: %s", err.Error()))
	}
	return connect.NewResponse(&compassv1beta1.DeleteTagTemplateResponse{}), nil
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
