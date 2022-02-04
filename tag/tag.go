package tag

//go:generate mockery --name TagRepository --outpkg mocks --output ../lib/mocks/ --structname TagRepository --filename tag_repository.go

import (
	"context"
	"time"
)

// Tag is the tag to be managed
type Tag struct {
	RecordType          string     `json:"record_type" validate:"required"`
	RecordURN           string     `json:"record_urn" validate:"required"`
	TemplateURN         string     `json:"template_urn" validate:"required"`
	TagValues           []TagValue `json:"tag_values" validate:"required,min=1,dive"`
	TemplateDisplayName string     `json:"template_display_name"`
	TemplateDescription string     `json:"template_description"`
}

// TagValue is one of the value for a tag
type TagValue struct {
	FieldID          uint        `json:"field_id" validate:"required"`
	FieldValue       interface{} `json:"field_value" validate:"required"`
	FieldURN         string      `json:"field_urn"`
	FieldDisplayName string      `json:"field_display_name"`
	FieldDescription string      `json:"field_description"`
	FieldDataType    string      `json:"field_data_type"`
	FieldOptions     []string    `json:"field_options"`
	FieldRequired    bool        `json:"field_required"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// TagRepository is a contract to communicate with the primary store
type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	Read(ctx context.Context, filter Tag) ([]Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, filter Tag) error
}
