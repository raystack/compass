package postgres

import (
	"context"
	"fmt"
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

	domainTemplates := modelTemplatesFields.toDomainTemplates()
	if len(domainTemplates) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}

	var insertedModelTags []Tag
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		timestamp := time.Now().UTC()

		for _, tv := range domainTag.TagValues {
			var insertedTagValue Tag
			if tv.FieldValue == nil {
				continue
			}
			tag := &Tag{
				RecordType: domainTag.RecordType,
				RecordURN:  domainTag.RecordURN,
				FieldID:    tv.FieldID,
				Value:      fmt.Sprintf("%v", tv.FieldValue),
				CreatedAt:  timestamp,
				UpdatedAt:  timestamp,
			}

			if err := tx.QueryRowxContext(ctx, `
						INSERT INTO tags
							(value, record_urn, record_type, field_id, created_at, updated_at)
						VALUES
							($1, $2, $3, $4, $5, $6)
						RETURNING *`,
				tag.Value, tag.RecordURN, tag.RecordType, tag.FieldID, tag.CreatedAt, tag.UpdatedAt).
				StructScan(&insertedTagValue); err != nil {
				return errors.Wrap(err, "failed to insert a domain tag")
			}

			insertedModelTags = append(insertedModelTags, insertedTagValue)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to create domain tag")
	}

	return r.complementDomainTag(domainTag, domainTemplates[0], insertedModelTags)
}

// Read reads tags grouped by its template
func (r *TagRepository) Read(ctx context.Context, filter tag.Tag) ([]tag.Tag, error) {

	if filter.RecordType == "" {
		return nil, errEmptyRecordType
	}
	if filter.RecordURN == "" {
		return nil, errEmptyRecordURN
	}

	sqlQuery := `
		SELECT 
			t.urn as "templates.urn", t.display_name as "templates.display_name", t.description as "templates.description",
			t.created_at as "templates.created_at", t.updated_at as "templates.updated_at",
			tg.id as "tags.id", tg.value as "tags.value", tg.record_urn as "tags.record_urn", tg.record_type as "tags.record_type",
			tg.field_id as "tags.field_id", tg.created_at as "tags.created_at", tg.updated_at as "tags.updated_at",
			f.id as "fields.id", f.urn as "fields.urn", f.display_name as "fields.display_name", f.description as "fields.description",
			f.data_type as "fields.data_type", f.options as "fields.options", f.required as "fields.required", f.template_urn as "fields.template_urn",
			f.created_at as "fields.created_at", f.updated_at as "fields.updated_at"
		FROM 
			templates t
		JOIN 
			fields f ON f.template_urn = t.urn 
		JOIN
			tags tg ON f.id = tg.field_id
		WHERE
			tg.record_urn = $1 AND tg.record_type = $2`
	sqlArgs := []interface{}{filter.RecordURN, filter.RecordType}

	if filter.TemplateURN != "" {
		// filter by record and template
		sqlQuery += " AND t.urn = $3"
		sqlArgs = append(sqlArgs, filter.TemplateURN)
	}

	var templateTagFields TemplateTagFields
	if txErr := r.client.db.Select(&templateTagFields, sqlQuery, sqlArgs...); txErr != nil {
		return nil, errors.Wrap(txErr, "failed reading tag domain")
	}

	if len(templateTagFields) == 0 {
		return nil, tag.NotFoundError{
			URN:      filter.RecordURN,
			Type:     filter.RecordType,
			Template: filter.TemplateURN,
		}
	}

	templates, tags := templateTagFields.toModelTemplatesAndTags()

	return tags.toDomainTags(filter.RecordType, filter.RecordURN, templates), nil
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

	domainTemplates := modelTemplatesFields.toDomainTemplates()
	if len(domainTemplates) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}

	var updatedModelTags []Tag
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		timestamp := time.Now().UTC()

		for _, value := range domainTag.TagValues {
			if value.FieldValue == nil || value.FieldValue == "" {
				continue
			}
			valueStr := fmt.Sprintf("%v", value.FieldValue)
			modelTag := &Tag{
				Value:      valueStr,
				RecordURN:  domainTag.RecordURN,
				RecordType: domainTag.RecordType,
				FieldID:    value.FieldID,
				CreatedAt:  timestamp,
				UpdatedAt:  timestamp,
			}

			var updatedModelTag Tag
			if err := tx.QueryRowxContext(ctx, `
							INSERT INTO
							tags 
								(value, record_urn, record_type, field_id, created_at, updated_at)
							VALUES
								($1, $2, $3, $4, $5, $6)
							ON CONFLICT 
								(record_urn, record_type, field_id)
							DO UPDATE SET 
								(value, record_urn, record_type, field_id, created_at, updated_at) = 
								($1, $2, $3, $4, $5, $6) 
							RETURNING *`,
				modelTag.Value, modelTag.RecordURN, modelTag.RecordType, modelTag.FieldID, modelTag.CreatedAt, modelTag.UpdatedAt).
				StructScan(&updatedModelTag); err != nil {
				return errors.Wrap(err, "failed to update a domain tag")
			}
			updatedModelTags = append(updatedModelTags, updatedModelTag)
		}
		return nil
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

	for _, modelTag := range deletedModelTags {
		sqlQuery := "DELETE FROM tags WHERE tags.record_urn = $1 AND tags.record_type = $2"
		sqlArgs := []interface{}{modelTag.RecordURN, modelTag.RecordType}

		if modelTag.FieldID != 0 {
			sqlQuery += " AND tags.field_id = $3"
			sqlArgs = append(sqlArgs, modelTag.FieldID)
		}

		if _, txErr := r.client.db.ExecContext(ctx, sqlQuery, sqlArgs...); txErr != nil {
			return errors.Wrapf(txErr, "failed to delete tag with record urn: %s, record type: %s, field id: %d", modelTag.RecordURN, modelTag.RecordType, modelTag.FieldID)
		}
	}
	return nil
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
