package tag

//go:generate mockery --name TagTemplateRepository --outpkg mocks --output ../lib/mocks/ --structname TagTemplateRepository --filename tag_template_repository.go

import (
	"context"
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

// TagTemplateRepository is a contract to communicate with the primary store
type TagTemplateRepository interface {
	Create(ctx context.Context, template *Template) error
	Read(ctx context.Context, templateURN string) ([]Template, error)
	ReadAll(ctx context.Context) ([]Template, error)
	Update(ctx context.Context, templateURN string, template *Template) error
	Delete(ctx context.Context, templateURN string) error
}
