package postgres

import (
	"time"
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

// Template is a model for template database table
type Template struct {
	URN         string    `db:"urn"`
	DisplayName string    `db:"display_name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Fields      []Field   `db:"-"`
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

type TemplateFields []TemplateField

func (tfs TemplateFields) toModelTemplates() (templates []Template) {
	templateMap := make(map[string]Template, 0)

	for _, tf := range tfs {
		if _, ok := templateMap[tf.Template.URN]; !ok {
			templateMap[tf.Template.URN] = tf.Template
		}

		templatePtr := templateMap[tf.Template.URN]
		templatePtr.Fields = append(templatePtr.Fields, tf.Field)
		templateMap[tf.Template.URN] = templatePtr
	}

	for _, t := range templateMap {
		templates = append(templates, t)
	}

	return
}

type TemplateField struct {
	Template Template `db:"templates"`
	Field    Field    `db:"fields"`
}

type TemplateTagField struct {
	Template Template `db:"templates"`
	Tag      Tag      `db:"tags"`
	Field    Field    `db:"fields"`
}
