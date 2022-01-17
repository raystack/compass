package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
)

func insertDomainTagToDBTx(ctx context.Context, tx *sqlx.Tx, domainTag *tag.Tag, timestamp time.Time) (modelTags []Tag, err error) {

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

		if err = tx.QueryRowxContext(ctx, `
					INSERT INTO tags
						(value, record_urn, record_type, field_id, created_at, updated_at)
					VALUES
						($1, $2, $3, $4, $5, $6)
					RETURNING *`,
			tag.Value, tag.RecordURN, tag.RecordType, tag.FieldID, tag.CreatedAt, tag.UpdatedAt).
			StructScan(&insertedTagValue); err != nil {
			err = errors.Wrap(err, "failed to insert a domain tag")
			return
		}

		modelTags = append(modelTags, insertedTagValue)
	}
	return
}

func readTemplatesJoinTagsJoinFieldsFromDB(ctx context.Context, db *sqlx.DB, filterTag tag.Tag) ([]TemplateTagField, error) {
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
	sqlArgs := []interface{}{filterTag.RecordURN, filterTag.RecordType}

	if filterTag.TemplateURN != "" {
		// filter by record and template
		sqlQuery += " AND t.urn = $3"
		sqlArgs = append(sqlArgs, filterTag.TemplateURN)
	}

	var rows []TemplateTagField
	if txErr := db.Select(&rows, sqlQuery, sqlArgs...); txErr != nil {
		return nil, errors.Wrap(txErr, "failed reading tag domain")
	}

	if len(rows) == 0 {
		return nil, tag.NotFoundError{
			URN:      filterTag.RecordURN,
			Type:     filterTag.RecordType,
			Template: filterTag.TemplateURN,
		}
	}

	return rows, nil
}

func updateDomainTagToDBTx(ctx context.Context, tx *sqlx.Tx, tag *tag.Tag, timestamp time.Time) (newModelTags []Tag, err error) {
	for _, value := range tag.TagValues {
		if value.FieldValue == nil || value.FieldValue == "" {
			continue
		}
		valueStr := fmt.Sprintf("%v", value.FieldValue)
		modelTag := &Tag{
			Value:      valueStr,
			RecordURN:  tag.RecordURN,
			RecordType: tag.RecordType,
			FieldID:    value.FieldID,
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
		}

		var updatedModelTag Tag
		if txErr := tx.QueryRowxContext(ctx, `
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
			StructScan(&updatedModelTag); txErr != nil {
			err = errors.Wrap(txErr, "failed to update a domain tag")
			return
		}
		newModelTags = append(newModelTags, updatedModelTag)
	}
	return
}

func deleteTagsFromDBTx(ctx context.Context, db *sqlx.DB, modelTags []Tag) (err error) {

	for _, modelTag := range modelTags {
		sqlQuery := "DELETE FROM tags WHERE tags.record_urn = $1 AND tags.record_type = $2"
		sqlArgs := []interface{}{modelTag.RecordURN, modelTag.RecordType}

		if modelTag.FieldID != 0 {
			sqlQuery += " AND tags.field_id = $3"
			sqlArgs = append(sqlArgs, modelTag.FieldID)
		}

		if _, txErr := db.ExecContext(ctx, sqlQuery, sqlArgs...); txErr != nil {
			err = errors.Wrapf(txErr, "failed to delete tag with record urn: %s, record type: %s, field id: %d", modelTag.RecordURN, modelTag.RecordType, modelTag.FieldID)
			return
		}
	}
	return
}
