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

// GetAllTagTemplates handles template read requests
func (h *Handler) GetAllTagTemplates(ctx context.Context, req *compassv1beta1.GetAllTagTemplatesRequest) (*compassv1beta1.GetAllTagTemplatesResponse, error) {
	listOfDomainTemplate, err := h.TagTemplateService.Index(ctx, req.GetUrn())
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error finding templates: %s", err.Error()))
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
func (h *Handler) CreateTagTemplate(ctx context.Context, req *compassv1beta1.CreateTagTemplateRequest) (*compassv1beta1.CreateTagTemplateResponse, error) {

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
	err := h.TagTemplateService.Create(ctx, &template)
	if errors.As(err, new(tag.DuplicateTemplateError)) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error creating tag template: %s", err.Error()))
	}

	return &compassv1beta1.CreateTagTemplateResponse{
		Data: template.ToProto(),
	}, nil
}

// GetTagTemplate handles template read requests based on URN
func (h *Handler) GetTagTemplate(ctx context.Context, req *compassv1beta1.GetTagTemplateRequest) (*compassv1beta1.GetTagTemplateResponse, error) {
	domainTemplate, err := h.TagTemplateService.Find(ctx, req.GetTemplateUrn())
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error finding a template: %s", err.Error()))
	}

	return &compassv1beta1.GetTagTemplateResponse{
		Data: domainTemplate.ToProto(),
	}, nil
}

func (h *Handler) UpdateTagTemplate(ctx context.Context, req *compassv1beta1.UpdateTagTemplateRequest) (*compassv1beta1.UpdateTagTemplateResponse, error) {

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
	err := h.TagTemplateService.Update(ctx, req.TemplateUrn, &template)
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if errors.As(err, new(tag.ValidationError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error updating template: %s", err.Error()))
	}

	return &compassv1beta1.UpdateTagTemplateResponse{
		Data: template.ToProto(),
	}, nil
}

// DeleteTagTemplate handles template delete request based on URN
func (h *Handler) DeleteTagTemplate(ctx context.Context, req *compassv1beta1.DeleteTagTemplateRequest) (*compassv1beta1.DeleteTagTemplateResponse, error) {
	err := h.TagTemplateService.Delete(ctx, req.GetTemplateUrn())
	if errors.As(err, new(tag.TemplateNotFoundError)) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error deleting a template: %s", err.Error()))
	}
	return &compassv1beta1.DeleteTagTemplateResponse{}, nil
}
