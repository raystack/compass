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
	errNilTag       = errors.New("tag is nil")
	errEmptyAssetID = errors.New("asset id should not be empty")
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
				AssetID:   domainTag.AssetID,
				FieldID:   tv.FieldID,
				Value:     fmt.Sprintf("%v", tv.FieldValue),
				CreatedAt: timestamp,
				UpdatedAt: timestamp,
			}

			if err := tx.QueryRowxContext(ctx, `
						INSERT INTO tags
							(value, asset_id, field_id, created_at, updated_at)
						VALUES
							($1, $2, $3, $4, $5)
						RETURNING *`,
				tagToInsert.Value, tagToInsert.AssetID, tagToInsert.FieldID, tagToInsert.CreatedAt, tagToInsert.UpdatedAt).
				StructScan(&insertedTagValue); err != nil {
				if err := checkPostgresError(err); errors.Is(err, errDuplicateKey) {
					return tag.DuplicateError{
						AssetID:     tagToInsert.AssetID,
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

	if filter.AssetID == "" {
		return nil, errEmptyAssetID
	}

	sqlQuery := `
		SELECT 
			t.urn as "tag_templates.urn", t.display_name as "tag_templates.display_name", t.description as "tag_templates.description",
			t.created_at as "tag_templates.created_at", t.updated_at as "tag_templates.updated_at",
			tg.id as "tags.id", tg.value as "tags.value", tg.asset_id as "tags.asset_id",
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
			tg.asset_id = $1`
	sqlArgs := []interface{}{filter.AssetID}

	if filter.TemplateURN != "" {
		// filter by asset and template
		sqlQuery += " AND t.urn = $2"
		sqlArgs = append(sqlArgs, filter.TemplateURN)
	}

	var templateTagFields TagJoinTemplateTagFieldModels
	if err := r.client.db.SelectContext(ctx, &templateTagFields, sqlQuery, sqlArgs...); err != nil {
		return nil, fmt.Errorf("failed reading domain tag: %w", err)
	}

	// (nil, not found error) if no asset id and template urn = ""
	// (empty, nil) if no asset id and template urn != ""
	if len(templateTagFields) == 0 && filter.TemplateURN != "" {
		return nil, tag.NotFoundError{
			AssetID:  filter.AssetID,
			Template: filter.TemplateURN,
		}
	}

	templates, tags := templateTagFields.toTemplateAndTagModels()

	return tags.toTags(filter.AssetID, templates), nil
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
				Value:     valueStr,
				AssetID:   domainTag.AssetID,
				FieldID:   value.FieldID,
				CreatedAt: timestamp,
				UpdatedAt: timestamp,
			}

			var updatedModelTag TagModel
			if err := tx.QueryRowxContext(ctx, `
							INSERT INTO
							tags 
								(value, asset_id, field_id, created_at, updated_at)
							VALUES
								($1, $2, $3, $4, $5)
							ON CONFLICT 
								(asset_id, field_id)
							DO UPDATE SET 
								(value, asset_id, field_id, created_at, updated_at) = 
								($1, $2, $3, $4, $5) 
							RETURNING *`,
				tagModel.Value, tagModel.AssetID, tagModel.FieldID, tagModel.CreatedAt, tagModel.UpdatedAt).
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
	if domainTag.AssetID == "" {
		return errEmptyAssetID
	}
	deletedModelTags := []TagModel{}
	fieldIDMap := map[uint]bool{}
	if domainTag.TemplateURN != "" {
		assetTemplatesFields, err := readTemplatesByURNFromDB(ctx, r.client.db, domainTag.TemplateURN)
		if err != nil {
			return err
		}
		if len(assetTemplatesFields) < 1 {
			return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
		}
		for _, tf := range assetTemplatesFields {
			fieldIDMap[tf.Field.ID] = true
			deletedModelTags = append(deletedModelTags, TagModel{
				AssetID: domainTag.AssetID,
				FieldID: tf.Field.ID,
			})
		}
	} else {
		deletedModelTags = append(deletedModelTags, TagModel{
			AssetID: domainTag.AssetID,
		})
	}

	for _, tagModel := range deletedModelTags {
		sqlQuery := "DELETE FROM tags WHERE tags.asset_id = $1"
		sqlArgs := []interface{}{tagModel.AssetID}

		if tagModel.FieldID != 0 {
			sqlQuery += " AND tags.field_id = $2"
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
