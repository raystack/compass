package handlersv1beta1

//go:generate mockery --name=TagTemplateService -r --case underscore --with-expecter --structname TagTemplateService --filename tag_template_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/tag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TagTemplateService interface {
	Validate(template tag.Template) error
	CreateTemplate(ctx context.Context, template *tag.Template) error
	GetTemplates(ctx context.Context, templateURN string) ([]tag.Template, error)
	UpdateTemplate(ctx context.Context, templateURN string, template *tag.Template) error
	GetTemplate(ctx context.Context, urn string) (tag.Template, error)
	DeleteTemplate(ctx context.Context, urn string) error
}

// GetAllTagTemplates handles template read requests
func (server *APIServer) GetAllTagTemplates(ctx context.Context, req *compassv1beta1.GetAllTagTemplatesRequest) (*compassv1beta1.GetAllTagTemplatesResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	listOfDomainTemplate, err := server.tagTemplateService.GetTemplates(ctx, req.GetUrn())
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error finding templates: %s", err.Error()))
	}

	var templatesPB []*compassv1beta1.TagTemplate
	for _, template := range listOfDomainTemplate {
		templatesPB = append(templatesPB, template.ToProto())
	}

	return &compassv1beta1.GetAllTagTemplatesResponse{
		Data: templatesPB,
	}, nil
}

// CreateTagTemplate handles template creation requests
func (server *APIServer) CreateTagTemplate(ctx context.Context, req *compassv1beta1.CreateTagTemplateRequest) (*compassv1beta1.CreateTagTemplateResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
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
		templateFields = append(templateFields, tag.NewTemplateFieldFromProto(fPB))
	}

	template := tag.Template{
		URN:         req.GetUrn(),
		DisplayName: req.GetDisplayName(),
		Description: req.GetDescription(),
		Fields:      templateFields,
	}
	err = server.tagTemplateService.CreateTemplate(ctx, &template)
	if errors.As(err, new(tag.DuplicateTemplateError)) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error creating tag template: %s", err.Error()))
	}

	return &compassv1beta1.CreateTagTemplateResponse{
		Data: template.ToProto(),
	}, nil
}

// GetTagTemplate handles template read requests based on URN
func (server *APIServer) GetTagTemplate(ctx context.Context, req *compassv1beta1.GetTagTemplateRequest) (*compassv1beta1.GetTagTemplateResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
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
		Data: domainTemplate.ToProto(),
	}, nil
}

func (server *APIServer) UpdateTagTemplate(ctx context.Context, req *compassv1beta1.UpdateTagTemplateRequest) (*compassv1beta1.UpdateTagTemplateResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
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
		templateFields = append(templateFields, tag.NewTemplateFieldFromProto(fPB))
	}

	template := tag.Template{
		URN:         req.GetTemplateUrn(),
		DisplayName: req.GetDisplayName(),
		Description: req.GetDescription(),
		Fields:      templateFields,
	}
	if err = server.tagTemplateService.UpdateTemplate(ctx, req.TemplateUrn, &template); err != nil {
		if errors.As(err, new(tag.TemplateNotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.As(err, new(tag.ValidationError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, internalServerError(server.logger, fmt.Sprintf("error updating template: %s", err.Error()))
	}

	return &compassv1beta1.UpdateTagTemplateResponse{
		Data: template.ToProto(),
	}, nil
}

// DeleteTagTemplate handles template delete request based on URN
func (server *APIServer) DeleteTagTemplate(ctx context.Context, req *compassv1beta1.DeleteTagTemplateRequest) (*compassv1beta1.DeleteTagTemplateResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = server.tagTemplateService.DeleteTemplate(ctx, req.GetTemplateUrn())
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error deleting a template: %s", err.Error()))
	}
	return &compassv1beta1.DeleteTagTemplateResponse{}, nil
}
