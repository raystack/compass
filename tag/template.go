package tag

import (
	"time"
)

// Template is a template of a tag for a resource
type Template struct {
	URN         string    `json:"urn" validate:"required"`
	DisplayName string    `json:"display_name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Fields      []Field   `json:"fields" validate:"required,min=1,dive"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Field is a field for a single template
type Field struct {
	ID          uint      `json:"id"`
	URN         string    `json:"urn" validate:"required"`
	DisplayName string    `json:"display_name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	DataType    string    `json:"data_type" validate:"oneof=string double boolean enumerated datetime"`
	Options     []string  `json:"options"`
	Required    bool      `json:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TemplateRepository is a contract to communicate with the primary store
type TemplateRepository interface {
	Create(template *Template) error
	Read(filter Template) ([]Template, error)
	Update(template *Template) error
	Delete(filter Template) error
}
