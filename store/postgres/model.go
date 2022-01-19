package postgres

import (
	"strings"
	"time"

	"github.com/odpf/columbus/tag"
)

// Tag is a model for tag value in database table
type Tag struct {
	ID         uint      `db:"id"`
	Value      string    `db:"value"`
	RecordURN  string    `db:"record_urn"`
	RecordType string    `db:"record_type"`
	FieldID    uint      `db:"field_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	Field      Field     `db:"-"`
}

type Tags []Tag

func (ts Tags) buildMapByTemplateURN() map[string][]Tag {
	tagsByTemplateURN := make(map[string][]Tag)
	for _, t := range ts {
		key := t.Field.TemplateURN
		if tagsByTemplateURN[key] == nil {
			tagsByTemplateURN[key] = []Tag{}
		}
		tagsByTemplateURN[key] = append(tagsByTemplateURN[key], t)
	}
	return tagsByTemplateURN
}

func (ts Tags) toDomainTags(recordType, recordURN string, templates Templates) []tag.Tag {
	templateByURN := templates.buildMapByURN()
	tagsByTemplateURN := ts.buildMapByTemplateURN()

	output := []tag.Tag{}
	for templateURN, modelTags := range tagsByTemplateURN {
		var listOfDomainTagValue []tag.TagValue
		modelTemplate := templateByURN[templateURN]
		for _, t := range modelTags {
			var options []string
			if t.Field.Options != nil {
				options = strings.Split(*t.Field.Options, ",")
			}
			parsedValue, _ := tag.ParseTagValue(modelTemplate.URN, t.FieldID, t.Field.DataType, t.Value, options)
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
		output = append(output, tag.Tag{
			RecordType:          recordType,
			RecordURN:           recordURN,
			TemplateURN:         modelTemplate.URN,
			TagValues:           listOfDomainTagValue,
			TemplateDisplayName: modelTemplate.DisplayName,
			TemplateDescription: modelTemplate.Description,
		})
	}
	return output
}

// Template is a model for template database table
type Template struct {
	URN         string    `db:"urn"`
	DisplayName string    `db:"display_name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Fields      Fields    `db:"-"`
}

func (tmp *Template) toDomainTemplate() tag.Template {
	return tag.Template{
		URN:         tmp.URN,
		DisplayName: tmp.DisplayName,
		Description: tmp.Description,
		Fields:      tmp.Fields.toDomainFields(),
		CreatedAt:   tmp.CreatedAt,
		UpdatedAt:   tmp.UpdatedAt,
	}
}

func (tmp *Template) buildFromDomainTemplate(domainTemplate tag.Template) {
	modelFields := &Fields{}
	modelFields.buildFromDomainFields(domainTemplate.Fields)

	tmp.URN = domainTemplate.URN
	tmp.DisplayName = domainTemplate.DisplayName
	tmp.Description = domainTemplate.Description
	tmp.Fields = *modelFields
}

type Templates []Template

func (tmps Templates) buildMapByURN() map[string]Template {
	templateByURN := make(map[string]Template)
	for _, t := range tmps {
		templateByURN[t.URN] = t
	}
	return templateByURN
}

// Field is a model for field tag in database table
type Field struct {
	ID          uint      `db:"id"`
	URN         string    `db:"urn"`
	DisplayName string    `db:"display_name"`
	Description string    `db:"description"`
	DataType    string    `db:"data_type"`
	Options     *string   `db:"options"`
	Required    bool      `db:"required"`
	TemplateURN string    `db:"template_urn"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Template    Template  `db:"-"`
}

type Fields []Field

func (fs *Fields) isIDExist(id uint) bool {
	for _, field := range *fs {
		if field.ID == id {
			return true
		}
	}
	return false
}

func (fs *Fields) toDomainFields() []tag.Field {
	output := make([]tag.Field, len(*fs))
	for i, field := range *fs {
		var options []string
		if field.Options != nil {
			options = strings.Split(*field.Options, fieldOptionSeparator)
		}
		output[i] = tag.Field{
			ID:          field.ID,
			URN:         field.URN,
			DisplayName: field.DisplayName,
			Description: field.Description,
			DataType:    field.DataType,
			Options:     options,
			Required:    field.Required,
			CreatedAt:   field.CreatedAt,
			UpdatedAt:   field.UpdatedAt,
		}
	}
	return output
}

func (fs *Fields) buildFromDomainFields(listOfDomainField []tag.Field) {
	newFields := Fields{}
	for _, field := range listOfDomainField {
		var options *string
		if len(field.Options) > 0 {
			joinedOptions := strings.Join(field.Options, fieldOptionSeparator)
			options = &joinedOptions
		}
		newFields = append(newFields, Field{
			ID:          field.ID,
			URN:         field.URN,
			DisplayName: field.DisplayName,
			Description: field.Description,
			DataType:    field.DataType,
			Options:     options,
			Required:    field.Required,
		})
	}
	*fs = newFields
}

// TemplateFields is a slice of placeholder for joined template and field
type TemplateFields []TemplateField

func (tfs TemplateFields) toModelTemplates() (templates []Template) {
	templateMap := make(map[string]Template, 0)
	// fieldMap := make(map[uint]Field, 0)

	for _, tf := range tfs {
		if _, ok := templateMap[tf.Template.URN]; !ok {
			templateMap[tf.Template.URN] = tf.Template
		}
		templatePtr := templateMap[tf.Template.URN]
		// check existing field
		if !templatePtr.Fields.isIDExist(tf.Field.ID) {
			templatePtr.Fields = append(templatePtr.Fields, tf.Field)
			templateMap[tf.Template.URN] = templatePtr
		}
	}

	for _, t := range templateMap {
		templates = append(templates, t)
	}

	return
}

func (tfs TemplateFields) toDomainTemplates() (templates []tag.Template) {
	templatesMap := map[string]Template{}
	for _, tf := range tfs {
		// build template
		if _, ok := templatesMap[tf.Template.URN]; !ok {
			templatesMap[tf.Template.URN] = tf.Template
		}

		templatePtr := templatesMap[tf.Template.URN]
		// check existing field
		if !templatePtr.Fields.isIDExist(tf.Field.ID) {
			templatePtr.Fields = append(templatePtr.Fields, tf.Field)
			templatesMap[tf.Template.URN] = templatePtr
		}
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
	return
}

// TemplateField is a placeholder for joined template and field
type TemplateField struct {
	Template Template `db:"templates"`
	Field    Field    `db:"fields"`
}

// TemplateTagFields is a slice of placeholder for joined template, tag, and field
type TemplateTagFields []TemplateTagField

func (ttfs TemplateTagFields) toModelTemplatesAndTags() (templates Templates, tags Tags) {
	tmpltsMap := make(map[string]Template, 0) // template urn as key
	tagsMap := make(map[uint]Tag, 0)

	for _, ttf := range ttfs {
		// build template
		if _, ok := tmpltsMap[ttf.Template.URN]; !ok {
			tmpltsMap[ttf.Template.URN] = ttf.Template
		}

		templatePtr := tmpltsMap[ttf.Template.URN]
		// check existing field
		if !templatePtr.Fields.isIDExist(ttf.Field.ID) {
			templatePtr.Fields = append(templatePtr.Fields, ttf.Field)
			tmpltsMap[ttf.Template.URN] = templatePtr
		}

		if _, ok := tagsMap[ttf.Tag.ID]; !ok {
			ttf.Tag.Field = ttf.Field
			tagsMap[ttf.Tag.ID] = ttf.Tag
		}
	}

	for _, tmp := range tmpltsMap {
		templates = append(templates, tmp)
	}

	for _, tg := range tagsMap {
		tags = append(tags, tg)
	}
	return
}

// TemplateField is a placeholder for joined template, tag, and field
type TemplateTagField struct {
	Template Template `db:"templates"`
	Tag      Tag      `db:"tags"`
	Field    Field    `db:"fields"`
}
