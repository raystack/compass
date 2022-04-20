package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/compass/tag"
)

var (
	errNilTag          = errors.New("tag is nil")
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
		return errNilTag
	}

	templateFieldModels, err := readTemplatesByURNFromDB(ctx, r.client.db, domainTag.TemplateURN)
	if err != nil {
		return err
	}

	if len(templateFieldModels) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}

	templates := templateFieldModels.toTemplates()

	var insertedModelTags []TagModel
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		timestamp := time.Now().UTC()

		for _, tv := range domainTag.TagValues {
			var insertedTagValue TagModel
			if tv.FieldValue == nil {
				continue
			}
			tagToInsert := &TagModel{
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
				tagToInsert.Value, tagToInsert.RecordURN, tagToInsert.RecordType, tagToInsert.FieldID, tagToInsert.CreatedAt, tagToInsert.UpdatedAt).
				StructScan(&insertedTagValue); err != nil {
				if err := checkPostgresError(err); errors.Is(err, errDuplicateKey) {
					return tag.DuplicateError{
						RecordURN:   tagToInsert.RecordURN,
						RecordType:  tagToInsert.RecordType,
						TemplateURN: domainTag.TemplateURN,
					}
				}
				return fmt.Errorf("failed to insert a domain tag: %w", err)
			}

			insertedModelTags = append(insertedModelTags, insertedTagValue)
		}
		return nil
	}); err != nil {
		return err
	}

	return r.complementTag(domainTag, templates[0], insertedModelTags)
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
			t.urn as "tag_templates.urn", t.display_name as "tag_templates.display_name", t.description as "tag_templates.description",
			t.created_at as "tag_templates.created_at", t.updated_at as "tag_templates.updated_at",
			tg.id as "tags.id", tg.value as "tags.value", tg.record_urn as "tags.record_urn", tg.record_type as "tags.record_type",
			tg.field_id as "tags.field_id", tg.created_at as "tags.created_at", tg.updated_at as "tags.updated_at",
			f.id as "tag_template_fields.id", f.urn as "tag_template_fields.urn", f.display_name as "tag_template_fields.display_name", f.description as "tag_template_fields.description",
			f.data_type as "tag_template_fields.data_type", f.options as "tag_template_fields.options", f.required as "tag_template_fields.required", f.template_urn as "tag_template_fields.template_urn",
			f.created_at as "tag_template_fields.created_at", f.updated_at as "tag_template_fields.updated_at"
		FROM 
			tag_templates t
		JOIN 
			tag_template_fields f ON f.template_urn = t.urn 
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

	var templateTagFields TagJoinTemplateTagFieldModels
	if err := r.client.db.SelectContext(ctx, &templateTagFields, sqlQuery, sqlArgs...); err != nil {
		return nil, fmt.Errorf("failed reading domain tag: %w", err)
	}

	// (nil, not found error) if no record and template urn = ""
	// (empty, nil) if no record and template urn != ""
	if len(templateTagFields) == 0 && filter.TemplateURN != "" {
		return nil, tag.NotFoundError{
			URN:      filter.RecordURN,
			Type:     filter.RecordType,
			Template: filter.TemplateURN,
		}
	}

	templates, tags := templateTagFields.toTemplateAndTagModels()

	return tags.toTags(filter.RecordType, filter.RecordURN, templates), nil
}

// Update updates tags in the database
func (r *TagRepository) Update(ctx context.Context, domainTag *tag.Tag) error {
	if domainTag == nil {
		return errNilTag
	}

	templateFieldModels, err := readTemplatesByURNFromDB(ctx, r.client.db, domainTag.TemplateURN)
	if err != nil {
		return err
	}
	if len(templateFieldModels) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}

	templates := templateFieldModels.toTemplates()

	var updatedModelTags []TagModel
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		timestamp := time.Now().UTC()

		for _, value := range domainTag.TagValues {
			if value.FieldValue == nil || value.FieldValue == "" {
				continue
			}
			valueStr := fmt.Sprintf("%v", value.FieldValue)
			tagModel := &TagModel{
				Value:      valueStr,
				RecordURN:  domainTag.RecordURN,
				RecordType: domainTag.RecordType,
				FieldID:    value.FieldID,
				CreatedAt:  timestamp,
				UpdatedAt:  timestamp,
			}

			var updatedModelTag TagModel
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
				tagModel.Value, tagModel.RecordURN, tagModel.RecordType, tagModel.FieldID, tagModel.CreatedAt, tagModel.UpdatedAt).
				StructScan(&updatedModelTag); err != nil {
				return err
			}
			updatedModelTags = append(updatedModelTags, updatedModelTag)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to update a domain tag: %w", err)
	}

	return r.complementTag(domainTag, templates[0], updatedModelTags)
}

// Delete deletes tags from database
func (r *TagRepository) Delete(ctx context.Context, domainTag tag.Tag) error {
	if domainTag.RecordURN == "" {
		return errEmptyRecordURN
	}
	deletedModelTags := []TagModel{}
	fieldIDMap := map[uint]bool{}
	if domainTag.TemplateURN != "" {
		recordTemplatesFields, err := readTemplatesByURNFromDB(ctx, r.client.db, domainTag.TemplateURN)
		if err != nil {
			return err
		}
		if len(recordTemplatesFields) < 1 {
			return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
		}
		for _, tf := range recordTemplatesFields {
			fieldIDMap[tf.Field.ID] = true
			deletedModelTags = append(deletedModelTags, TagModel{
				RecordURN:  domainTag.RecordURN,
				RecordType: domainTag.RecordType,
				FieldID:    tf.Field.ID,
			})
		}
	} else {
		deletedModelTags = append(deletedModelTags, TagModel{
			RecordURN:  domainTag.RecordURN,
			RecordType: domainTag.RecordType,
		})
	}

	for _, tagModel := range deletedModelTags {
		sqlQuery := "DELETE FROM tags WHERE tags.record_urn = $1 AND tags.record_type = $2"
		sqlArgs := []interface{}{tagModel.RecordURN, tagModel.RecordType}

		if tagModel.FieldID != 0 {
			sqlQuery += " AND tags.field_id = $3"
			sqlArgs = append(sqlArgs, tagModel.FieldID)
		}

		if _, err := r.client.db.ExecContext(ctx, sqlQuery, sqlArgs...); err != nil {
			return fmt.Errorf("failed to delete a domain tag: %w", err)
		}
	}
	return nil
}

func (r *TagRepository) complementTag(domainTag *tag.Tag, template tag.Template, tagModels []TagModel) error {
	tagByFieldID := make(map[uint]TagModel)
	for _, t := range tagModels {
		tagByFieldID[t.FieldID] = t
	}
	var listOfTagValue []tag.TagValue
	for _, field := range template.Fields {
		t := tagByFieldID[field.ID]
		parsedValue, _ := tag.ParseTagValue(domainTag.TemplateURN, field.ID, field.DataType, t.Value, field.Options)
		listOfTagValue = append(listOfTagValue, tag.TagValue{
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
	domainTag.TemplateURN = template.URN
	domainTag.TemplateDescription = template.Description
	domainTag.TemplateDisplayName = template.DisplayName
	domainTag.TagValues = listOfTagValue
	return nil
}

// NewTagRepository initializes tag repository
// all methods in tag repository uses passed by reference
// which will mutate the reference variable in method's argument
func NewTagRepository(client *Client) (*TagRepository, error) {
	if client == nil {
		return nil, errNilPostgresClient
	}
	return &TagRepository{
		client: client,
	}, nil
}
