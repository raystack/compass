package tag

import (
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
		err = ValidationError{err}
	}

	return err
}

// Create handles create business operation for template
func (s *TemplateService) Create(template *Template) error {
	if template == nil {
		return errors.New("template is nil")
	}
	err := s.Validate(*template)
	if err != nil {
		return err
	}

	filterForExistence := Template{
		URN: template.URN,
	}
	templateRecords, err := s.repository.Read(filterForExistence)
	if err != nil {
		return errors.Wrap(err, "error checking template existence")
	}
	if len(templateRecords) > 0 {
		return DuplicateTemplateError{URN: template.URN}
	}

	err = s.repository.Create(template)
	if err != nil {
		return errors.Wrap(err, "error creating template")
	}

	return nil
}

// Index handles read business operation for template
func (s *TemplateService) Index(template Template) ([]Template, error) {
	output, err := s.repository.Read(template)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching templates")
	}
	return output, nil
}

// Update handles update business operation for template
func (s *TemplateService) Update(template *Template) error {
	if template == nil {
		return errors.New("template is nil")
	}
	err := s.Validate(*template)
	if err != nil {
		return err
	}
	filterForExistence := Template{
		URN: template.URN,
	}
	templateRecords, err := s.repository.Read(filterForExistence)
	if err != nil {
		return errors.Wrap(err, "error checking template existence")
	}
	if len(templateRecords) == 0 {
		return TemplateNotFoundError{URN: template.URN}
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

	err = s.repository.Update(template)
	if err != nil {
		return errors.Wrap(err, "error updating template")
	}
	return nil
}

// Find handles request to get template by urn
func (s *TemplateService) Find(urn string) (Template, error) {
	queryDomainTemplate := Template{
		URN: urn,
	}
	listOfDomainTemplate, err := s.repository.Read(queryDomainTemplate)
	if err != nil {
		return Template{}, errors.Wrap(err, "error reading repository")
	}
	if len(listOfDomainTemplate) == 0 {
		return Template{}, TemplateNotFoundError{URN: urn}
	}
	return listOfDomainTemplate[0], nil
}

// Delete handles request to delete template by urn
func (s *TemplateService) Delete(urn string) error {
	template := Template{
		URN: urn,
	}
	err := s.repository.Delete(template)
	if err != nil {
		return errors.Wrap(err, "error deleting template")
	}
	return nil
}

// NewTemplateService initializes service template service
func NewTemplateService(r TemplateRepository) *TemplateService {
	if r == nil {
		panic("repository is nil")
	}
	return &TemplateService{
		validator:  newTemplateValidator(),
		repository: r,
	}
}
