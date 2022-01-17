package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
)

var (
	errNilDomainTag    = errors.New("domain tag is nil")
	errEmptyRecordType = errors.New("record type should not be empty")
	errEmptyRecordURN  = errors.New("record urn should not be empty")
)

// TagRepository is a type that manages tag operation ot the primary database
type TagRepository struct {
	client *Client
}

// Create inserts tag to database
func (r *TagRepository) Create(ctx context.Context, domainTag *tag.Tag) error {
	if domainTag == nil {
		return errNilDomainTag
	}
	modelTemplatesFields, err := readTemplatesJoinFieldsFromDB(ctx, r.client.db, domainTag.TemplateURN)
	if err != nil {
		return err
	}
	domainTemplates := r.convertToDomainTemplates(modelTemplatesFields)
	if len(domainTemplates) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}

	var insertedModelTags []Tag
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		insertedModelTags, err = insertDomainTagToDBTx(ctx, tx, domainTag, time.Now().UTC())
		return err
	}); err != nil {
		return errors.Wrap(err, "failed to create domain tag")
	}

	return r.complementDomainTag(domainTag, domainTemplates[0], insertedModelTags)
}

// Read reads tags grouped by its template
func (r *TagRepository) Read(ctx context.Context, filter tag.Tag) ([]tag.Tag, error) {
	tmpltsMap := make(map[string]Template, 0) // template urn as key
	tagsMap := make(map[uint]Tag, 0)

	if filter.RecordType == "" {
		return nil, errEmptyRecordType
	}
	if filter.RecordURN == "" {
		return nil, errEmptyRecordURN
	}

	templateTagFields, err := readTemplatesJoinTagsJoinFieldsFromDB(ctx, r.client.db, filter)
	if err != nil {
		return nil, err
	}

	for _, ttf := range templateTagFields {
		// build template
		if _, ok := tmpltsMap[ttf.Template.URN]; !ok {
			tmpltsMap[ttf.Template.URN] = ttf.Template
		}

		templatePtr := tmpltsMap[ttf.Template.URN]
		templatePtr.Fields = append(templatePtr.Fields, ttf.Field)
		tmpltsMap[ttf.Template.URN] = templatePtr

		if _, ok := tagsMap[ttf.Tag.ID]; !ok {
			ttf.Tag.Field = ttf.Field
			tagsMap[ttf.Tag.ID] = ttf.Tag
		}
	}

	var templates []Template
	for _, tmp := range tmpltsMap {
		templates = append(templates, tmp)
	}
	var tags []Tag
	for _, tg := range tagsMap {
		tags = append(tags, tg)
	}

	templateByURN := r.mapTemplateByTemplateURN(templates)
	tagsByTemplateURN := r.mapTagsByTemplateURN(tags)

	output := []tag.Tag{}
	for templateURN, tags := range tagsByTemplateURN {
		domainTag := r.convertToDomainTag(filter.RecordType, filter.RecordURN, templateByURN[templateURN], tags)
		output = append(output, domainTag)
	}
	return output, nil
}

// Update updates tags in the database
func (r *TagRepository) Update(ctx context.Context, domainTag *tag.Tag) error {
	if domainTag == nil {
		return errNilDomainTag
	}
	modelTemplatesFields, err := readTemplatesJoinFieldsFromDB(ctx, r.client.db, domainTag.TemplateURN)
	if err != nil {
		return err
	}
	domainTemplates := r.convertToDomainTemplates(modelTemplatesFields)
	if len(domainTemplates) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}

	var updatedModelTags []Tag
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		updatedModelTags, err = updateDomainTagToDBTx(ctx, tx, domainTag, time.Now().UTC())
		return err
	}); err != nil {
		return errors.Wrap(err, "failed to update domain Tag")
	}

	return r.complementDomainTag(domainTag, domainTemplates[0], updatedModelTags)
}

// Delete deletes tags from database
func (r *TagRepository) Delete(ctx context.Context, domainTag tag.Tag) error {
	if domainTag.RecordURN == "" {
		return errEmptyRecordURN
	}
	deletedModelTags := []Tag{}
	fieldIDMap := map[uint]bool{}
	if domainTag.TemplateURN != "" {
		recordTemplatesFields, err := readTemplatesJoinFieldsFromDB(ctx, r.client.db, domainTag.TemplateURN)
		if err != nil {
			return err
		}
		for _, tf := range recordTemplatesFields {
			fieldIDMap[tf.Field.ID] = true
			deletedModelTags = append(deletedModelTags, Tag{
				RecordURN:  domainTag.RecordURN,
				RecordType: domainTag.RecordType,
				FieldID:    tf.Field.ID,
			})
		}
	} else {
		deletedModelTags = append(deletedModelTags, Tag{
			RecordURN:  domainTag.RecordURN,
			RecordType: domainTag.RecordType,
		})
	}
	return deleteTagsFromDBTx(ctx, r.client.db, deletedModelTags)
}

func (r *TagRepository) convertToDomainTemplates(modelTemplateFields []TemplateField) []tag.Template {
	templates := []tag.Template{}
	templatesMap := map[string]Template{}
	for _, tf := range modelTemplateFields {
		// build template
		if _, ok := templatesMap[tf.Template.URN]; !ok {
			templatesMap[tf.Template.URN] = tf.Template
		}

		templatePtr := templatesMap[tf.Template.URN]
		templatePtr.Fields = append(templatePtr.Fields, tf.Field)
		templatesMap[tf.Template.URN] = templatePtr
	}

	for _, template := range templatesMap {
		listOfDomainField := []tag.Field{}
		for _, field := range template.Fields {
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

		templates = append(templates, tag.Template{
			URN:         template.URN,
			DisplayName: template.DisplayName,
			Description: template.Description,
			Fields:      listOfDomainField,
			CreatedAt:   template.CreatedAt,
			UpdatedAt:   template.UpdatedAt,
		})
	}
	return templates
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

func (r *TagRepository) complementDomainTag(domainTag *tag.Tag, domainTemplate tag.Template, modelTags []Tag) error {
	tagByFieldID := make(map[uint]Tag)
	for _, t := range modelTags {
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

// NewTagRepository initializes tag repository
func NewTagRepository(client *Client) (*TagRepository, error) {
	if client == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &TagRepository{
		client: client,
	}, nil
}
