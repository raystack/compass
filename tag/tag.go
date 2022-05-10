package tag

//go:generate mockery --name TagRepository --outpkg mocks --output ../lib/mocks/ --with-expecter  --structname TagRepository --filename tag_repository.go
import (
	"context"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TagRepository is a contract to communicate with the primary store
type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	Read(ctx context.Context, filter Tag) ([]Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, filter Tag) error
}

// Tag is the tag to be managed
type Tag struct {
	AssetID             string     `json:"asset_id" validate:"required"`
	TemplateURN         string     `json:"template_urn" validate:"required"`
	TagValues           []TagValue `json:"tag_values" validate:"required,min=1,dive"`
	TemplateDisplayName string     `json:"template_display_name"`
	TemplateDescription string     `json:"template_description"`
}

// ToProto convert domain to protobuf
func (t Tag) ToProto() (*compassv1beta1.Tag, error) {
	var tagValuesPB []*compassv1beta1.TagValue
	for _, tv := range t.TagValues {
		tvPB, err := tv.ToProto()
		if err != nil {
			return nil, err
		}
		tagValuesPB = append(tagValuesPB, tvPB)
	}

	return &compassv1beta1.Tag{
		AssetId:             t.AssetID,
		TemplateUrn:         t.TemplateURN,
		TagValues:           tagValuesPB,
		TemplateDisplayName: t.TemplateDisplayName,
		TemplateDescription: t.TemplateDescription,
	}, nil
}

// NewFromProto converts proto to tag.Tag
func NewFromProto(pb *compassv1beta1.Tag) Tag {
	var tagValues []TagValue
	if pb.GetTagValues() != nil {
		for _, tvPB := range pb.GetTagValues() {
			tv := NewTagValueFromProto(tvPB)
			tagValues = append(tagValues, tv)
		}
	}

	return Tag{
		AssetID:             pb.GetAssetId(),
		TemplateURN:         pb.GetTemplateUrn(),
		TagValues:           tagValues,
		TemplateDisplayName: pb.GetTemplateDisplayName(),
		TemplateDescription: pb.GetTemplateDescription(),
	}
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

// ToProto convert domain to protobuf
func (tv TagValue) ToProto() (*compassv1beta1.TagValue, error) {
	var err error
	var createdAtPB *timestamppb.Timestamp
	if !tv.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(tv.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !tv.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(tv.UpdatedAt)
	}

	var fieldValuePB *structpb.Value
	if tv.FieldValue != nil {
		fieldValuePB, err = structpb.NewValue(tv.FieldValue)
		if err != nil {
			return nil, err
		}
	}

	return &compassv1beta1.TagValue{
		FieldId:          uint32(tv.FieldID),
		FieldValue:       fieldValuePB,
		FieldUrn:         tv.FieldURN,
		FieldDisplayName: tv.FieldDisplayName,
		FieldDescription: tv.FieldDescription,
		FieldDataType:    tv.FieldDataType,
		FieldOptions:     tv.FieldOptions,
		FieldRequired:    tv.FieldRequired,
		CreatedAt:        createdAtPB,
		UpdatedAt:        updatedAtPB,
	}, nil
}

// NewTagValueFromProto converts proto to tag.TagValue
func NewTagValueFromProto(pb *compassv1beta1.TagValue) TagValue {
	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var fieldValue interface{}
	if pb.GetFieldValue() != nil {
		fieldValue = pb.GetFieldValue().AsInterface()
	}

	return TagValue{
		FieldID:          uint(pb.GetFieldId()),
		FieldValue:       fieldValue,
		FieldURN:         pb.GetFieldUrn(),
		FieldDisplayName: pb.GetFieldDisplayName(),
		FieldDescription: pb.GetFieldDescription(),
		FieldDataType:    pb.GetFieldDataType(),
		FieldOptions:     pb.GetFieldOptions(),
		FieldRequired:    pb.GetFieldRequired(),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}
