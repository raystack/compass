package tag

import (
	"fmt"

	"github.com/odpf/columbus/tag/validator"
	"github.com/pkg/errors"
)

// Service is a type that manages business process
type Service struct {
	validator       validator.Validator
	repository      TagRepository
	templateService *TemplateService
}

// Validate validates domain tag based on business requirement
func (s *Service) Validate(tag *Tag) error {
	if tag == nil {
		return errors.New("tag is nil")
	}

	err := s.validator.Validate(*tag)
	if err != nil {
		err = ValidationError{err}
	}

	return err
}

// Create handles business process for create
func (s *Service) Create(tag *Tag) error {
	if err := s.Validate(tag); err != nil {
		return err
	}
	template, err := s.templateService.Find(tag.TemplateURN)
	if err != nil {
		return errors.Wrap(err, "error finding template")
	}
	if err := s.validateFieldIsMemberOfTemplate(*tag, template); err != nil {
		return err
	}
	if err := s.validateRequiredFieldIsPassed(*tag, template); err != nil {
		return err
	}
	if err := s.validateFieldValueIsValid(*tag, template); err != nil {
		return err
	}
	if err := s.repository.Create(tag); err != nil {
		return err
	}

	return nil
}

// GetByRecord handles business process to get tags by its resource urn
func (s *Service) GetByRecord(recordType, recordURN string) ([]Tag, error) {
	tag := Tag{RecordType: recordType, RecordURN: recordURN}
	return s.repository.Read(tag)
}

// FindByRecordAndTemplate handles business process to get tags by its resource id and template id
func (s *Service) FindByRecordAndTemplate(recordType, recordURN, templateURN string) (Tag, error) {
	_, err := s.templateService.Find(templateURN)
	if err != nil {
		return Tag{}, err
	}
	listOfTag, err := s.repository.Read(Tag{RecordType: recordType, RecordURN: recordURN, TemplateURN: templateURN})
	if err != nil {
		return Tag{}, err
	}
	var output Tag
	if len(listOfTag) == 0 {
		return Tag{}, NotFoundError{Type: recordType, URN: recordURN, Template: templateURN}
	}

	output = listOfTag[0]
	return output, err
}

// Delete handles business process to delete a tag
func (s *Service) Delete(recordType, recordURN, templateURN string) error {
	_, err := s.templateService.Find(templateURN)
	if err != nil {
		return errors.Wrap(err, "error finding template")
	}
	if err := s.repository.Delete(Tag{
		RecordType:  recordType,
		RecordURN:   recordURN,
		TemplateURN: templateURN,
	}); err != nil {
		return errors.Wrap(err, "error deleting tag")
	}
	return nil
}

// Update handles business process for update
func (s *Service) Update(tag *Tag) error {
	if err := s.Validate(tag); err != nil {
		return err
	}
	template, err := s.templateService.Find(tag.TemplateURN)
	if err != nil {
		return errors.Wrap(err, "error finding template")
	}
	existingTags, err := s.repository.Read(Tag{
		RecordType:  tag.RecordType,
		RecordURN:   tag.RecordURN,
		TemplateURN: tag.TemplateURN,
	})
	if err != nil {
		return errors.Wrap(err, "error finding existing tag")
	}
	if len(existingTags) == 0 {
		return NotFoundError{URN: tag.RecordURN, Type: tag.RecordType, Template: tag.TemplateURN}
	}

	if err = s.validateFieldIsMemberOfTemplate(*tag, template); err != nil {
		return err
	}
	if err := s.validateFieldValueIsValid(*tag, template); err != nil {
		return err
	}
	if err := s.repository.Update(tag); err != nil {
		return errors.Wrap(err, "error updating tag")
	}
	return nil
}

func (s *Service) validateFieldIsMemberOfTemplate(tag Tag, template Template) error {
	fieldIsMemberOfTemplate := make(map[uint]bool)
	for _, field := range template.Fields {
		fieldIsMemberOfTemplate[field.ID] = true
	}
	for i, value := range tag.TagValues {
		if !fieldIsMemberOfTemplate[value.FieldID] {
			return buildFieldError(
				fmt.Sprintf("fields[%d].id", i),
				fmt.Sprintf("not part of template [%s]",
					template.URN),
			)
		}
	}
	return nil
}

func (s *Service) validateRequiredFieldIsPassed(tag Tag, template Template) error {
	passedFieldMap := make(map[uint]bool)
	for _, value := range tag.TagValues {
		if value.FieldValue != nil && value.FieldValue != "" {
			passedFieldMap[value.FieldID] = true
		}
	}
	for i, field := range template.Fields {
		if field.Required && !passedFieldMap[field.ID] {
			return buildFieldError(
				fmt.Sprintf("fields[%d].id", i), fmt.Sprintf("required by template [%s]", template.URN),
			)
		}
	}
	return nil
}

func (s *Service) validateFieldValueIsValid(tag Tag, template Template) error {
	domainFieldByID := make(map[uint]Field)
	for _, field := range template.Fields {
		domainFieldByID[field.ID] = field
	}
	for i, value := range tag.TagValues {
		if value.FieldValue != nil && value.FieldValue != "" {
			fieldValue := fmt.Sprintf("%v", value.FieldValue)
			fieldID := value.FieldID
			domainField := domainFieldByID[fieldID]
			_, err := ParseTagValue(template.URN, domainField.ID,
				domainField.DataType, fieldValue, domainField.Options,
			)
			if err != nil {
				return buildFieldError(fmt.Sprintf("fields[%d].value", i), err.Error())
			}
		}
	}
	return nil
}

// NewService initializes service tag
func NewService(repository TagRepository, templateService *TemplateService) *Service {
	return &Service{
		validator:       newValidator(),
		repository:      repository,
		templateService: templateService,
	}
}
