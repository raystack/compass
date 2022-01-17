package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
)

func insertTemplateFieldToDBTx(ctx context.Context, tx *sqlx.Tx, modelTemplate *Template) error {
	insertedTemplate := *modelTemplate
	if err := insertTemplateToDBTx(ctx, tx, &insertedTemplate); err != nil {
		return err
	}

	for _, field := range modelTemplate.Fields {
		field.CreatedAt = modelTemplate.CreatedAt
		field.UpdatedAt = modelTemplate.UpdatedAt
		field.TemplateURN = modelTemplate.URN

		if err := insertFieldToDBTx(ctx, tx, &field); err != nil {
			return err
		}

		insertedTemplate.Fields = append(insertedTemplate.Fields, field)
	}
	*modelTemplate = insertedTemplate
	return nil
}

func insertTemplateToDBTx(ctx context.Context, tx *sqlx.Tx, modelTemplate *Template) error {
	var insertedTemplate Template
	if err := tx.QueryRowxContext(ctx, `
					INSERT INTO 
					templates 
						(urn,display_name,description,created_at,updated_at) 
					VALUES 
						($1,$2,$3,$4,$5)
					RETURNING *
				`,
		modelTemplate.URN, modelTemplate.DisplayName, modelTemplate.Description, modelTemplate.CreatedAt, modelTemplate.UpdatedAt).
		StructScan(&insertedTemplate); err != nil {
		return errors.Wrap(err, "failed to insert a template")
	}

	*modelTemplate = insertedTemplate
	return nil
}

func insertFieldToDBTx(ctx context.Context, tx *sqlx.Tx, field *Field) error {
	var insertedField Field
	if err := tx.QueryRowxContext(ctx, `
					INSERT INTO 
					fields 
						(urn, display_name, description, data_type, options, required, template_urn, created_at, updated_at)
					VALUES 
						($1,$2,$3,$4,$5,$6,$7,$8,$9)
					RETURNING *
					`,
		field.URN, field.DisplayName, field.Description, field.DataType, field.Options, field.Required, field.TemplateURN, field.CreatedAt, field.UpdatedAt).
		StructScan(&insertedField); err != nil {
		return errors.Wrap(err, "failed to insert a field")
	}
	field = &insertedField
	return nil
}

func readTemplatesJoinFieldsFromDB(ctx context.Context, db *sqlx.DB, templateURN string) (templates TemplateFields, err error) {
	if txErr := db.Select(&templates, `
		SELECT 
			t.urn as "templates.urn", t.display_name as "templates.display_name", t.description as "templates.description", 
			t.created_at as "templates.created_at", t.updated_at as "templates.updated_at", 
			f.id as "fields.id", f.urn as "fields.urn", f.display_name as "fields.display_name", f.description as "fields.description",
			f.data_type as "fields.data_type", f.options as "fields.options", f.required as "fields.required", f.template_urn as "fields.template_urn", 
			f.created_at as "fields.created_at", f.updated_at as "fields.updated_at"
		FROM 
			templates t
		JOIN 
			fields f
		ON 
			f.template_urn = t.urn 
		WHERE 
			t.urn = $1`, templateURN); txErr != nil {
		err = errors.Wrap(txErr, "failed reading templates")
		return
	}

	if len(templates) == 0 {
		err = &tag.TemplateNotFoundError{URN: templateURN}
		return
	}

	return
}

func updateTemplateFieldToDBTx(ctx context.Context, tx *sqlx.Tx, targetURN string, modelTemplate *Template) error {

	updatedModelTemplate := *modelTemplate
	if err := updateTemplateToDBTx(ctx, tx, targetURN, &updatedModelTemplate); err != nil {
		return errors.Wrap(err, "failed to update a field")
	}

	for _, field := range modelTemplate.Fields {
		field.TemplateURN = modelTemplate.URN
		field.UpdatedAt = modelTemplate.UpdatedAt

		if field.ID == 0 {
			field.CreatedAt = modelTemplate.UpdatedAt
			field.UpdatedAt = modelTemplate.UpdatedAt
			field.TemplateURN = modelTemplate.URN

			if err := insertFieldToDBTx(ctx, tx, &field); err != nil {
				return errors.Wrap(err, "failed to insert a field")
			}

			updatedModelTemplate.Fields = append(updatedModelTemplate.Fields, field)
			continue
		}

		if err := updateFieldToDBTx(ctx, tx, &field); err != nil {
			return errors.Wrap(err, "failed to update a field")
		}
		updatedModelTemplate.Fields = append(updatedModelTemplate.Fields, field)
	}
	*modelTemplate = updatedModelTemplate
	return nil
}

func updateTemplateToDBTx(ctx context.Context, tx *sqlx.Tx, targetTemplateURN string, modelTemplate *Template) error {
	var updatedTemplate Template
	if err := tx.QueryRowxContext(ctx, `
					UPDATE
						templates 
					SET
						urn = $1, display_name = $2, description = $3, updated_at = $4
					WHERE
						urn = $5
					RETURNING *`,
		modelTemplate.URN, modelTemplate.DisplayName, modelTemplate.Description, modelTemplate.UpdatedAt, targetTemplateURN).
		StructScan(&updatedTemplate); err != nil {
		// scan returns sql.ErrNoRows if no rows
		if errors.Is(err, sql.ErrNoRows) {
			return tag.TemplateNotFoundError{URN: modelTemplate.URN}
		}
		return errors.Wrap(err, "failed building update template sql")
	}

	*modelTemplate = updatedTemplate
	return nil
}

func updateFieldToDBTx(ctx context.Context, tx *sqlx.Tx, field *Field) error {
	var updatedField Field

	if err := tx.QueryRowxContext(ctx, `
					UPDATE
						fields
					SET
						urn = $1, display_name = $2, description = $3, data_type = $4, options = $5, 
						required = $6, template_urn = $7, updated_at = $8
					WHERE
						id = $9 AND template_urn = $7
					RETURNING *`,
		field.URN, field.DisplayName, field.Description, field.DataType, field.Options, field.Required,
		field.TemplateURN, field.UpdatedAt, field.ID).
		StructScan(&updatedField); err != nil {
		return errors.Wrap(err, "failed updating fields")
	}

	if updatedField.ID == 0 {
		return errors.New("field not found when updating fields")
	}

	return nil
}

func deleteTemplateFromDB(ctx context.Context, db *sqlx.DB, templateURN string) error {
	res, err := db.ExecContext(ctx, `
					DELETE FROM
						templates 
					WHERE
						urn = $1`, templateURN)
	if err != nil {
		return errors.Wrapf(err, "failed to delete template with urn: %s", templateURN)
	}
	tmpRowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get row affected in deleting template")
	}
	if tmpRowsAffected == 0 {
		return tag.TemplateNotFoundError{URN: templateURN}
	}
	return nil
}
