package postgres

import (
	"strings"
	"time"

	"github.com/odpf/compass/tag"
)

// TagModel is a model for tag value in database table
type TagModel struct {
	ID        uint                  `db:"id"`
	Value     string                `db:"value"`
	AssetID   string                `db:"asset_id"`
	FieldID   uint                  `db:"field_id"`
	CreatedAt time.Time             `db:"created_at"`
	UpdatedAt time.Time             `db:"updated_at"`
	Field     TagTemplateFieldModel `db:"-"`
}

type TagModels []TagModel

func (ts TagModels) buildMapByTemplateURN() map[string][]TagModel {
	tagsByTemplateURN := make(map[string][]TagModel)
	for _, t := range ts {
		key := t.Field.TemplateURN
		if tagsByTemplateURN[key] == nil {
			tagsByTemplateURN[key] = []TagModel{}
		}
		tagsByTemplateURN[key] = append(tagsByTemplateURN[key], t)
	}
	return tagsByTemplateURN
}

func (ts TagModels) toTags(assetID string, templates TagTemplateModels) []tag.Tag {
	templateByURN := templates.buildMapByURN()
	tagsByTemplateURN := ts.buildMapByTemplateURN()

	output := []tag.Tag{}
	for templateURN, tagModels := range tagsByTemplateURN {
		var listOfTagValue []tag.TagValue
		templateModel := templateByURN[templateURN]
		for _, t := range tagModels {
			var options []string
			if t.Field.Options != nil {
				options = strings.Split(*t.Field.Options, ",")
			}
			parsedValue, _ := tag.ParseTagValue(templateModel.URN, t.FieldID, t.Field.DataType, t.Value, options)
			listOfTagValue = append(listOfTagValue, tag.TagValue{
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
			AssetID:             assetID,
			TemplateURN:         templateModel.URN,
			TagValues:           listOfTagValue,
			TemplateDisplayName: templateModel.DisplayName,
			TemplateDescription: templateModel.Description,
		})
	}
	return output
}

// TagTemplateModel is a model for template database table
type TagTemplateModel struct {
	URN         string                 `db:"urn"`
	DisplayName string                 `db:"display_name"`
	Description string                 `db:"description"`
	CreatedAt   time.Time              `db:"created_at"`
	UpdatedAt   time.Time              `db:"updated_at"`
	Fields      TagTemplateFieldModels `db:"-"`
}

func (tmp *TagTemplateModel) toTemplate() tag.Template {
	return tag.Template{
		URN:         tmp.URN,
		DisplayName: tmp.DisplayName,
		Description: tmp.Description,
		Fields:      tmp.Fields.toDomainFields(),
		CreatedAt:   tmp.CreatedAt,
		UpdatedAt:   tmp.UpdatedAt,
	}
}

func newTemplateModel(template *tag.Template) *TagTemplateModel {
	fieldModels := newSliceOfFieldModel(template.Fields)

	return &TagTemplateModel{
		URN:         template.URN,
		DisplayName: template.DisplayName,
		Description: template.Description,
		Fields:      fieldModels,
	}
}

type TagTemplateModels []TagTemplateModel

func (tmps TagTemplateModels) buildMapByURN() map[string]TagTemplateModel {
	templateByURN := make(map[string]TagTemplateModel)
	for _, t := range tmps {
		templateByURN[t.URN] = t
	}
	return templateByURN
}

// TagTemplateFieldModel is a model for field tag in database table
type TagTemplateFieldModel struct {
	ID          uint             `db:"id"`
	URN         string           `db:"urn"`
	DisplayName string           `db:"display_name"`
	Description string           `db:"description"`
	DataType    string           `db:"data_type"`
	Options     *string          `db:"options"`
	Required    bool             `db:"required"`
	TemplateURN string           `db:"template_urn"`
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"`
	Template    TagTemplateModel `db:"-"`
}

type TagTemplateFieldModels []TagTemplateFieldModel

func (fs *TagTemplateFieldModels) isIDExist(id uint) bool {
	for _, field := range *fs {
		if field.ID == id {
			return true
		}
	}
	return false
}

func (fs *TagTemplateFieldModels) toDomainFields() []tag.Field {
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

func newSliceOfFieldModel(listOfDomainField []tag.Field) TagTemplateFieldModels {
	newFields := TagTemplateFieldModels{}
	for _, field := range listOfDomainField {
		var options *string
		if len(field.Options) > 0 {
			joinedOptions := strings.Join(field.Options, fieldOptionSeparator)
			options = &joinedOptions
		}
		newFields = append(newFields, TagTemplateFieldModel{
			ID:          field.ID,
			URN:         field.URN,
			DisplayName: field.DisplayName,
			Description: field.Description,
			DataType:    field.DataType,
			Options:     options,
			Required:    field.Required,
		})
	}
	return newFields
}

// TagJoinTemplateFieldModel is a placeholder for joined template and field
type TagJoinTemplateFieldModel struct {
	Template TagTemplateModel      `db:"tag_templates"`
	Field    TagTemplateFieldModel `db:"tag_template_fields"`
}

// TagJoinTemplateFieldModels is a slice of placeholder for joined template and field
type TagJoinTemplateFieldModels []TagJoinTemplateFieldModel

func (tfs TagJoinTemplateFieldModels) toTemplateModels() (templates []TagTemplateModel) {
	templateMap := make(map[string]TagTemplateModel)

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

func (tfs TagJoinTemplateFieldModels) toTemplates() (templates []tag.Template) {
	templatesMap := map[string]TagTemplateModel{}
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

// TemplateField is a placeholder for joined template, tag, and field
type TagJoinTemplateTagFieldModel struct {
	Template TagTemplateModel      `db:"tag_templates"`
	Tag      TagModel              `db:"tags"`
	Field    TagTemplateFieldModel `db:"tag_template_fields"`
}

// TagJoinTemplateTagFieldModels is a slice of placeholder for joined template, tag, and field
type TagJoinTemplateTagFieldModels []TagJoinTemplateTagFieldModel

func (ttfs TagJoinTemplateTagFieldModels) toTemplateAndTagModels() (templates TagTemplateModels, tags TagModels) {
	tmpltsMap := make(map[string]TagTemplateModel) // template urn as key
	tagsMap := make(map[uint]TagModel)

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
