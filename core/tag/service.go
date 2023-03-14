package tag

import (
	"context"
	"errors"
	"fmt"

	"github.com/goto/compass/core/tag/validator"
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
func (s *Service) CreateTag(ctx context.Context, tag *Tag) error {
	if err := s.Validate(tag); err != nil {
		return err
	}
	template, err := s.templateService.GetTemplate(ctx, tag.TemplateURN)
	if err != nil {
		return fmt.Errorf("error finding template: %w", err)
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
	if err := s.repository.Create(ctx, tag); err != nil {
		return err
	}

	return nil
}

// GetTagsByAssetID handles business process to get tags by its asset id
func (s *Service) GetTagsByAssetID(ctx context.Context, assetID string) ([]Tag, error) {
	tag := Tag{AssetID: assetID}
	return s.repository.Read(ctx, tag)
}

// FindByAssetAndTemplate handles business process to get tags by its asset id and template id
func (s *Service) FindTagByAssetIDAndTemplateURN(ctx context.Context, assetID, templateURN string) (Tag, error) {
	_, err := s.templateService.GetTemplate(ctx, templateURN)
	if err != nil {
		return Tag{}, err
	}
	listOfTag, err := s.repository.Read(ctx, Tag{AssetID: assetID, TemplateURN: templateURN})
	if err != nil {
		return Tag{}, err
	}
	var output Tag
	if len(listOfTag) == 0 {
		return Tag{}, NotFoundError{AssetID: assetID, Template: templateURN}
	}

	output = listOfTag[0]
	return output, err
}

// DeleteTag handles business process to delete a tag
func (s *Service) DeleteTag(ctx context.Context, assetID, templateURN string) error {
	_, err := s.templateService.GetTemplate(ctx, templateURN)
	if err != nil {
		return fmt.Errorf("error finding template: %w", err)
	}
	if err := s.repository.Delete(ctx, Tag{
		AssetID:     assetID,
		TemplateURN: templateURN,
	}); err != nil {
		return err
	}
	return nil
}

// Update handles business process for update
func (s *Service) UpdateTag(ctx context.Context, tag *Tag) error {
	if err := s.Validate(tag); err != nil {
		return err
	}
	template, err := s.templateService.GetTemplate(ctx, tag.TemplateURN)
	if err != nil {
		return TemplateNotFoundError{URN: tag.TemplateURN}
	}
	existingTags, err := s.repository.Read(ctx, Tag{
		AssetID:     tag.AssetID,
		TemplateURN: tag.TemplateURN,
	})
	if err != nil {
		return NotFoundError{AssetID: tag.AssetID, Template: tag.TemplateURN}
	}
	if len(existingTags) == 0 {
		return NotFoundError{AssetID: tag.AssetID, Template: tag.TemplateURN}
	}

	if err = s.validateFieldIsMemberOfTemplate(*tag, template); err != nil {
		return err
	}
	if err := s.validateFieldValueIsValid(*tag, template); err != nil {
		return err
	}
	if err := s.repository.Update(ctx, tag); err != nil {
		return fmt.Errorf("error updating tag: %w", err)
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
