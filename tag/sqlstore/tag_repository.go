package sqlstore

import (
	"fmt"
	"strings"

	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"

	"gorm.io/gorm"
)

var (
	errNilDomainTag    = errors.New("domain tag is nil")
	errEmptyRecordType = errors.New("record type should not be empty")
	errEmptyRecordURN  = errors.New("record urn should not be empty")
)

// TagRepository is a type that manages tag operation ot the primary database
type TagRepository struct {
	dbClient *gorm.DB
}

// Create inserts tag to database
func (r *TagRepository) Create(domainTag *tag.Tag) error {
	if r.dbClient == nil {
		return errNilDBClient
	}
	if domainTag == nil {
		return errNilDomainTag
	}
	recordModelTemplate, err := r.getTemplateWithFields(domainTag.TemplateURN)
	if err != nil {
		return err
	}
	domainTemplate := r.convertToDomainTemplate(recordModelTemplate)
	err = r.dbClient.Transaction(func(tx *gorm.DB) error {
		return r.insertDomainTagToDB(tx, domainTag)
	})
	if err == nil {
		listOfRecordModelTag, err := r.readListOfModelTagFromDB(r.dbClient, domainTag.RecordType, domainTag.RecordURN, domainTemplate)
		if err == nil {
			err = r.complementDomainTag(domainTag, domainTemplate, listOfRecordModelTag)
		}
	}
	return err
}

// Read reads tags grouped by its template
func (r *TagRepository) Read(filter tag.Tag) ([]tag.Tag, error) {
	if r.dbClient == nil {
		return nil, errNilDBClient
	}
	if filter.RecordType == "" {
		return nil, errEmptyRecordType
	}
	if filter.RecordURN == "" {
		return nil, errEmptyRecordURN
	}
	if filter.TemplateURN != "" {
		recordModelTemplate, err := r.getTemplateWithFields(filter.TemplateURN)
		if err != nil {
			return nil, err
		}
		domainTag, err := r.readByRecordAndTemplateFromDB(r.dbClient, filter.RecordType, filter.RecordURN, recordModelTemplate)
		if err != nil {
			return nil, err
		}
		return []tag.Tag{domainTag}, nil
	}
	return r.readByRecordFromDB(r.dbClient, filter.RecordType, filter.RecordURN)
}

// Update updates tags in the database
func (r *TagRepository) Update(domainTag *tag.Tag) error {
	if r.dbClient == nil {
		return errNilDBClient
	}
	if domainTag == nil {
		return errNilDomainTag
	}
	recordModelTemplate, err := r.getTemplateWithFields(domainTag.TemplateURN)
	if err != nil {
		return err
	}
	domainTemplate := r.convertToDomainTemplate(recordModelTemplate)
	err = r.dbClient.Transaction(func(tx *gorm.DB) error {
		return r.updateDomainTagToDB(tx, domainTag)
	})
	if err == nil {
		listOfRecordModelTag, err := r.readListOfModelTagFromDB(r.dbClient, domainTag.RecordType, domainTag.RecordURN, domainTemplate)
		if err == nil {
			err = r.complementDomainTag(domainTag, domainTemplate, listOfRecordModelTag)
		}
	}
	return err
}

// Delete deletes tags from database
func (r *TagRepository) Delete(domainTag tag.Tag) error {
	if r.dbClient == nil {
		return errNilDBClient
	}
	if domainTag.RecordURN == "" {
		return errEmptyRecordURN
	}
	if domainTag.TemplateURN != "" {
		recordModelTemplate, err := r.getTemplateWithFields(domainTag.TemplateURN)
		if err != nil {
			return err
		}
		return r.deleteByRecordAndTemplate(r.dbClient, domainTag.RecordType, domainTag.RecordURN, recordModelTemplate)
	}
	return r.dbClient.Where("record_type = ? and record_urn = ?", domainTag.RecordType, domainTag.RecordURN).Delete(&Tag{}).Error
}

func (r *TagRepository) updateDomainTagToDB(tx *gorm.DB, domainTag *tag.Tag) error {
	for _, value := range domainTag.TagValues {
		if value.FieldValue == nil || value.FieldValue == "" {
			continue
		}
		var recordModelTag Tag
		queryResult := tx.Where("record_type = ? and record_urn = ? and field_id = ?", domainTag.RecordType, domainTag.RecordURN, value.FieldID).First(&recordModelTag)
		recordModelTag.Value = fmt.Sprintf("%v", value.FieldValue)
		if queryResult.Error != nil {
			recordModelTag.RecordURN = domainTag.RecordURN
			recordModelTag.FieldID = value.FieldID
			createResult := tx.Create(&recordModelTag)
			if createResult.Error != nil {
				return createResult.Error
			}
			continue
		}
		updateResult := tx.Updates(&recordModelTag)
		if updateResult.Error != nil {
			return updateResult.Error
		}
	}
	return nil
}

func (r *TagRepository) convertToDomainTemplate(modelTemplate Template) tag.Template {
	listOfDomainField := []tag.Field{}
	for _, field := range modelTemplate.Fields {
		var options []string
		if field.Options != nil {
			options = strings.Split(*field.Options, ",")
		}
		listOfDomainField = append(listOfDomainField, tag.Field{
			ID:          field.ID,
			URN:         field.URN,
			DisplayName: field.DisplayName,
			Description: field.Description,
			DataType:    field.DataType,
			Required:    field.Required,
			Options:     options,
			CreatedAt:   field.CreatedAt,
			UpdatedAt:   field.UpdatedAt,
		})
	}
	return tag.Template{
		URN:         modelTemplate.URN,
		DisplayName: modelTemplate.DisplayName,
		Description: modelTemplate.Description,
		Fields:      listOfDomainField,
		CreatedAt:   modelTemplate.CreatedAt,
		UpdatedAt:   modelTemplate.UpdatedAt,
	}
}

func (r *TagRepository) deleteByRecordAndTemplate(tx *gorm.DB, recordType, recordURN string, template Template) error {
	var listOfFieldID []uint
	for _, field := range template.Fields {
		listOfFieldID = append(listOfFieldID, field.ID)
	}
	return tx.Where("record_type = ? and record_urn = ? and field_id in ?", recordType, recordURN, listOfFieldID).Delete(&Tag{}).Error
}

func (r *TagRepository) readByRecordFromDB(tx *gorm.DB, recordType, recordURN string) ([]tag.Tag, error) {
	var listOfRecordModelTag []Tag
	err := tx.Preload("Field").
		Where("record_type = ? AND record_urn = ?", recordType, recordURN).
		Find(&listOfRecordModelTag).Error
	if err != nil {
		return nil, err
	}
	listOfTemplateURN := r.getListOfTemplateURN(listOfRecordModelTag)
	listOfTemplate, err := r.getListOfTemplateByListOfURNFromDB(tx, listOfTemplateURN)
	if err != nil {
		return nil, err
	}
	templateByURN := r.mapTemplateByTemplateURN(listOfTemplate)
	tagsByTemplateURN := r.mapTagsByTemplateURN(listOfRecordModelTag)
	output := []tag.Tag{}
	for templateURN, tags := range tagsByTemplateURN {
		domainTag := r.convertToDomainTag(recordType, recordURN, templateByURN[templateURN], tags)
		output = append(output, domainTag)
	}
	return output, nil
}

func (r *TagRepository) mapTagsByTemplateURN(tags []Tag) map[string][]Tag {
	tagsByTemplateURN := make(map[string][]Tag)
	for _, t := range tags {
		key := t.Field.TemplateURN
		if tagsByTemplateURN[key] == nil {
			tagsByTemplateURN[key] = []Tag{}
		}
		tagsByTemplateURN[key] = append(tagsByTemplateURN[key], t)
	}
	return tagsByTemplateURN
}

func (r *TagRepository) mapTemplateByTemplateURN(templates []Template) map[string]Template {
	templateByURN := make(map[string]Template)
	for _, t := range templates {
		templateByURN[t.URN] = t
	}
	return templateByURN
}

func (r *TagRepository) getListOfTemplateURN(tags []Tag) []string {
	isURNRegistered := make(map[string]bool)
	for _, t := range tags {
		key := t.Field.TemplateURN
		if !isURNRegistered[key] {
			isURNRegistered[key] = true
		}
	}
	var output []string
	for urn := range isURNRegistered {
		output = append(output, urn)
	}
	return output
}

func (r *TagRepository) getListOfTemplateByListOfURNFromDB(tx *gorm.DB, listOfTemplateURN []string) ([]Template, error) {
	var listOfTemplate []Template
	err := tx.Where("urn in ?", listOfTemplateURN).Find(&listOfTemplate).Error
	if err != nil {
		return nil, err
	}
	return listOfTemplate, nil
}

func (r *TagRepository) readByRecordAndTemplateFromDB(tx *gorm.DB, recordType, recordURN string, template Template) (tag.Tag, error) {
	var listOfRecordModelTag []Tag
	for _, field := range template.Fields {
		var recordModelTag Tag
		err := tx.Preload("Field").
			Where("record_type = ? and record_urn = ? and field_id = ?", recordType, recordURN, field.ID).
			First(&recordModelTag).Error
		if err != nil {
			return tag.Tag{}, fmt.Errorf("no tag record is found for record urn [%s] and template urn [%s]",
				recordURN, template.URN)
		}
		listOfRecordModelTag = append(listOfRecordModelTag, recordModelTag)
	}
	return r.convertToDomainTag(recordType, recordURN, template, listOfRecordModelTag), nil
}

func (r *TagRepository) convertToDomainTag(recordType, recordURN string, template Template, tags []Tag) tag.Tag {
	var listOfDomainTagValue []tag.TagValue
	for _, t := range tags {
		var options []string
		if t.Field.Options != nil {
			options = strings.Split(*t.Field.Options, ",")
		}
		parsedValue, _ := tag.ParseTagValue(template.URN, t.FieldID, t.Field.DataType, t.Value, options)
		listOfDomainTagValue = append(listOfDomainTagValue, tag.TagValue{
			FieldID:          t.FieldID,
			FieldValue:       parsedValue,
			FieldURN:         t.Field.URN,
			FieldDisplayName: t.Field.DisplayName,
			FieldDescription: t.Field.Description,
			FieldDataType:    t.Field.DataType,
			FieldOptions:     options,
			FieldRequired:    t.Field.Required,
			CreatedAt:        t.CreatedAt,
			UpdatedAt:        t.UpdatedAt,
		})
	}
	return tag.Tag{
		RecordType:          recordType,
		RecordURN:           recordURN,
		TemplateURN:         template.URN,
		TagValues:           listOfDomainTagValue,
		TemplateDisplayName: template.DisplayName,
		TemplateDescription: template.Description,
	}
}

func (r *TagRepository) complementDomainTag(domainTag *tag.Tag, domainTemplate tag.Template, listOfRecordModelTag []Tag) error {
	tagByFieldID := make(map[uint]Tag)
	for _, t := range listOfRecordModelTag {
		tagByFieldID[t.FieldID] = t
	}
	var listOfDomainTagValue []tag.TagValue
	for _, field := range domainTemplate.Fields {
		t := tagByFieldID[field.ID]
		parsedValue, _ := tag.ParseTagValue(domainTag.TemplateURN, field.ID, field.DataType, t.Value, field.Options)
		listOfDomainTagValue = append(listOfDomainTagValue, tag.TagValue{
			FieldID:          field.ID,
			FieldValue:       parsedValue,
			FieldURN:         field.URN,
			FieldDisplayName: field.DisplayName,
			FieldDescription: field.Description,
			FieldDataType:    field.DataType,
			FieldOptions:     field.Options,
			FieldRequired:    field.Required,
			CreatedAt:        t.CreatedAt,
			UpdatedAt:        t.UpdatedAt,
		})
	}
	domainTag.TemplateURN = domainTemplate.URN
	domainTag.TemplateDescription = domainTemplate.Description
	domainTag.TemplateDisplayName = domainTemplate.DisplayName
	domainTag.TagValues = listOfDomainTagValue
	return nil
}

func (r *TagRepository) getTemplateWithFields(templateURN string) (Template, error) {
	var template Template
	queryResult := r.dbClient.Preload("Fields").First(&template, "urn = ?", templateURN)
	if queryResult.Error == gorm.ErrRecordNotFound {
		return template, tag.TemplateNotFoundError{URN: templateURN}
	}

	return template, queryResult.Error
}

func (r *TagRepository) readListOfModelTagFromDB(tx *gorm.DB, recordType, recordURN string, domainTemplate tag.Template) ([]Tag, error) {
	var tags []Tag
	for _, field := range domainTemplate.Fields {
		var t Tag
		err := tx.Preload("Field").Where("record_type = ? and record_urn = ? and field_id = ?", recordType, recordURN, field.ID).First(&t).Error
		if err != nil {
			continue
		}
		tags = append(tags, t)
	}
	return tags, nil
}
func (r *TagRepository) insertDomainTagToDB(tx *gorm.DB, domainTag *tag.Tag) error {
	for _, tv := range domainTag.TagValues {
		if tv.FieldValue == nil {
			continue
		}
		modelTag := Tag{
			RecordType: domainTag.RecordType,
			RecordURN:  domainTag.RecordURN,
			FieldID:    tv.FieldID,
			Value:      fmt.Sprintf("%v", tv.FieldValue),
		}
		createResult := tx.Create(&modelTag)
		if createResult.Error != nil {
			return createResult.Error
		}
	}
	return nil
}

// NewTagRepository initializes tag repository
func NewTagRepository(dbClient *gorm.DB) *TagRepository {
	if dbClient == nil {
		panic(errNilDBClient.Error())
	}
	return &TagRepository{
		dbClient: dbClient,
	}
}
