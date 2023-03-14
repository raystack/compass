package tag

import (
	"context"
	"errors"
	"fmt"

	"github.com/goto/compass/core/tag/validator"
)

// Service is a type of service that manages business process
type TemplateService struct {
	validator  validator.Validator
	repository TagTemplateRepository
}

// Validate validates domain template based on the business rule
func (s *TemplateService) Validate(template Template) error {
	err := s.validator.Validate(template)
	if err != nil {
		err = ValidationError{err}
	}

	return err
}

// Create handles create business operation for template
func (s *TemplateService) CreateTemplate(ctx context.Context, template *Template) error {
	if template == nil {
		return errors.New("template is nil")
	}
	err := s.Validate(*template)
	if err != nil {
		return err
	}

	templateAssets, err := s.repository.Read(ctx, template.URN)
	if err != nil {
		return fmt.Errorf("error checking template existence: %w", err)
	}
	if len(templateAssets) > 0 {
		return DuplicateTemplateError{URN: template.URN}
	}

	err = s.repository.Create(ctx, template)
	if err != nil {
		return fmt.Errorf("error creating template: %w", err)
	}

	return nil
}

// GetTemplates handles read business operation for template
func (s *TemplateService) GetTemplates(ctx context.Context, templateURN string) ([]Template, error) {
	if templateURN == "" {
		output, err := s.repository.ReadAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("error fetching templates: %w", err)
		}
		return output, nil
	}
	output, err := s.repository.Read(ctx, templateURN)
	if err != nil {
		return nil, fmt.Errorf("error fetching templates: %w", err)
	}
	return output, nil
}

// UpdateTemplate handles update business operation for template
func (s *TemplateService) UpdateTemplate(ctx context.Context, templateURN string, template *Template) error {
	if template == nil {
		return errors.New("template is nil")
	}
	err := s.Validate(*template)
	if err != nil {
		return err
	}
	templateAssets, err := s.repository.Read(ctx, templateURN)
	if err != nil {
		return fmt.Errorf("error checking template existence: %w", err)
	}
	if len(templateAssets) == 0 {
		return TemplateNotFoundError{URN: templateURN}
	}

	// check for duplication
	templateFromDB := templateAssets[0]
	isFieldIDPartOfTemplateMap := make(map[uint]bool)
	fieldURNToIDMap := make(map[string]uint)
	for _, f := range templateFromDB.Fields {
		isFieldIDPartOfTemplateMap[f.ID] = true
		fieldURNToIDMap[f.URN] = f.ID
	}
	for i, f := range template.Fields {
		if !isFieldIDPartOfTemplateMap[f.ID] {
			return buildFieldError(
				fmt.Sprintf("fields.[%d].id", i),
				fmt.Sprintf("[%d] is not part of the template", f.ID),
			)
		}
		if fieldURNToIDMap[f.URN] != f.ID {
			return buildFieldError(
				fmt.Sprintf("fields.[%d].urn", i),
				fmt.Sprintf("[%s] already exists within the template", f.URN),
			)
		}
	}

	err = s.repository.Update(ctx, templateURN, template)
	if err != nil {
		return fmt.Errorf("error updating template: %w", err)
	}
	return nil
}

// GetTemplate handles request to get template by urn
func (s *TemplateService) GetTemplate(ctx context.Context, urn string) (Template, error) {
	listOfDomainTemplate, err := s.repository.Read(ctx, urn)
	if err != nil {
		return Template{}, fmt.Errorf("error reading repository: %w", err)
	}
	if len(listOfDomainTemplate) == 0 {
		return Template{}, TemplateNotFoundError{URN: urn}
	}
	return listOfDomainTemplate[0], nil
}

// DeleteTemplate handles request to delete template by urn
func (s *TemplateService) DeleteTemplate(ctx context.Context, urn string) error {
	err := s.repository.Delete(ctx, urn)
	if err != nil {
		return fmt.Errorf("error deleting template: %w", err)
	}
	return nil
}

// NewTemplateService initializes service template service
func NewTemplateService(r TagTemplateRepository) *TemplateService {
	return &TemplateService{
		validator:  newTemplateValidator(),
		repository: r,
	}
}
