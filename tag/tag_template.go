package tag

//go:generate mockery --name TagTemplateRepository --outpkg mocks --output ../lib/mocks/ --with-expecter  --structname TagTemplateRepository --filename tag_template_repository.go

import (
	"context"
	"time"

	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TagTemplateRepository is a contract to communicate with the primary store
type TagTemplateRepository interface {
	Create(ctx context.Context, template *Template) error
	Read(ctx context.Context, templateURN string) ([]Template, error)
	ReadAll(ctx context.Context) ([]Template, error)
	Update(ctx context.Context, templateURN string, template *Template) error
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

// ToProto convert domain to protobuf
func (t Template) ToProto() *compassv1beta1.TagTemplate {
	var templateFieldsPB []*compassv1beta1.TagTemplateField
	for _, tf := range t.Fields {
		templateFieldsPB = append(templateFieldsPB, tf.ToProto())
	}

	var createdAtPB *timestamppb.Timestamp
	if !t.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(t.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !t.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(t.UpdatedAt)
	}

	return &compassv1beta1.TagTemplate{
		Urn:         t.URN,
		DisplayName: t.DisplayName,
		Description: t.Description,
		Fields:      templateFieldsPB,
		CreatedAt:   createdAtPB,
		UpdatedAt:   updatedAtPB,
	}
}

// NewTemplateFromProto converts proto to tag.Template
func NewTemplateFromProto(pb *compassv1beta1.TagTemplate) Template {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var fields []Field
	if pb.GetFields() != nil {
		for _, tfPB := range pb.GetFields() {
			fields = append(fields, NewTemplateFieldFromProto(tfPB))
		}
	}

	return Template{
		URN:         pb.GetUrn(),
		DisplayName: pb.GetDisplayName(),
		Description: pb.GetDescription(),
		Fields:      fields,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
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

// ToProto convert domain to protobuf
func (f Field) ToProto() *compassv1beta1.TagTemplateField {
	var createdAtPB *timestamppb.Timestamp
	if !f.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(f.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !f.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(f.UpdatedAt)
	}

	return &compassv1beta1.TagTemplateField{
		Id:          uint32(f.ID),
		Urn:         f.URN,
		DisplayName: f.DisplayName,
		Description: f.Description,
		DataType:    f.DataType,
		Options:     f.Options,
		Required:    f.Required,
		CreatedAt:   createdAtPB,
		UpdatedAt:   updatedAtPB,
	}
}

// NewTemplateFieldFromProto converts proto to tag.Field
func NewTemplateFieldFromProto(pb *compassv1beta1.TagTemplateField) Field {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	return Field{
		ID:          uint(pb.GetId()),
		URN:         pb.GetUrn(),
		DisplayName: pb.GetDisplayName(),
		Description: pb.GetDescription(),
		DataType:    pb.GetDataType(),
		Options:     pb.GetOptions(),
		Required:    pb.GetRequired(),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
