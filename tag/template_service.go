package tag

import (
	"context"
	"fmt"

	"github.com/odpf/columbus/tag/validator"
	"github.com/pkg/errors"
)

var validDomainFieldDataType = []string{
	"string",
	"double",
	"boolean",
	"enumerated",
	"datetime",
}

// Service is a type of service that manages business process
type TemplateService struct {
	validator  validator.Validator
	repository TemplateRepository
}

// Validate validates domain template based on the business rule
func (s *TemplateService) Validate(template Template) error {
	err := s.validator.Validate(template)
	if err != nil {
		err = ErrValidation{err}
	}

	return err
}

// Create handles create business operation for template
func (s *TemplateService) Create(ctx context.Context, template *Template) error {
	if template == nil {
		return errors.New("template is nil")
	}
	err := s.Validate(*template)
	if err != nil {
		return err
	}

	templateRecords, err := s.repository.Read(ctx, template.URN)
	if err != nil {
		return errors.Wrap(err, "error checking template existence")
	}
	if len(templateRecords) > 0 {
		return ErrDuplicateTemplate{URN: template.URN}
	}

	err = s.repository.Create(ctx, template)
	if err != nil {
		return errors.Wrap(err, "error creating template")
	}

	return nil
}

// Index handles read business operation for template
func (s *TemplateService) Index(ctx context.Context, templateURN string) ([]Template, error) {
	if templateURN == "" {
		output, err := s.repository.ReadAll(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching templates")
		}
		return output, nil
	}
	output, err := s.repository.Read(ctx, templateURN)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching templates")
	}
	return output, nil
}

// Update handles update business operation for template
func (s *TemplateService) Update(ctx context.Context, templateURN string, template *Template) error {
	if template == nil {
		return errors.New("template is nil")
	}
	err := s.Validate(*template)
	if err != nil {
		return err
	}
	templateRecords, err := s.repository.Read(ctx, templateURN)
	if err != nil {
		return errors.Wrap(err, "error checking template existence")
	}
	if len(templateRecords) == 0 {
		return ErrTemplateNotFound{URN: templateURN}
	}

	// check for duplication
	templateFromDB := templateRecords[0]
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
		return errors.Wrap(err, "error updating template")
	}
	return nil
}

// Find handles request to get template by urn
func (s *TemplateService) Find(ctx context.Context, urn string) (Template, error) {
	listOfDomainTemplate, err := s.repository.Read(ctx, urn)
	if err != nil {
		return Template{}, errors.Wrap(err, "error reading repository")
	}
	if len(listOfDomainTemplate) == 0 {
		return Template{}, ErrTemplateNotFound{URN: urn}
	}
	return listOfDomainTemplate[0], nil
}

// Delete handles request to delete template by urn
func (s *TemplateService) Delete(ctx context.Context, urn string) error {
	err := s.repository.Delete(ctx, urn)
	if err != nil {
		return errors.Wrap(err, "error deleting template")
	}
	return nil
}

// NewTemplateService initializes service template service
func NewTemplateService(r TemplateRepository) *TemplateService {
	return &TemplateService{
		validator:  newTemplateValidator(),
		repository: r,
	}
}
