package tag

//go:generate mockery --name=TagTemplateRepository -r --case underscore --with-expecter --structname TagTemplateRepository --filename tag_template_repository.go --output=./mocks

import (
	"context"
	"github.com/raystack/compass/core/namespace"
	"time"
)

// TagTemplateRepository is a contract to communicate with the primary store
type TagTemplateRepository interface {
	Create(ctx context.Context, ns *namespace.Namespace, template *Template) error
	Read(ctx context.Context, templateURN string) ([]Template, error)
	ReadAll(ctx context.Context) ([]Template, error)
	Update(ctx context.Context, ns *namespace.Namespace, templateURN string, template *Template) error
	Delete(ctx context.Context, templateURN string) error
}

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
