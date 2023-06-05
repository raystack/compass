package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/goto/compass/core/tag"
	"github.com/jmoiron/sqlx"
)

const (
	fieldOptionSeparator = ","
)

var errNilTemplate = errors.New("template is nil")

// TagTemplateRepository is a type that manages template operation to the primary database
type TagTemplateRepository struct {
	client *Client
}

// Create inserts template to database
func (r *TagTemplateRepository) Create(ctx context.Context, templateDomain *tag.Template) error {
	if templateDomain == nil {
		return errNilTemplate
	}

	templateModel := newTemplateModel(templateDomain)

	timestamp := time.Now().UTC()
	templateModel.CreatedAt = timestamp
	templateModel.UpdatedAt = timestamp

	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		insertedTemplate := *templateModel
		if err := insertTemplateToDBTx(ctx, tx, &insertedTemplate); err != nil {
			return err
		}

		for _, field := range templateModel.Fields {
			field.CreatedAt = templateModel.CreatedAt
			field.UpdatedAt = templateModel.UpdatedAt
			field.TemplateURN = templateModel.URN

			if err := insertFieldToDBTx(ctx, tx, &field); err != nil {
				return err
			}

			insertedTemplate.Fields = append(insertedTemplate.Fields, field)
		}
		*templateModel = insertedTemplate
		return nil
	}); err != nil {
		return errors.New("failed to insert template")
	}

	*templateDomain = templateModel.toTemplate()
	return nil
}

// Read reads template from database by URN
func (r *TagTemplateRepository) Read(ctx context.Context, templateURN string) ([]tag.Template, error) {
	templates := []tag.Template{}
	var templatesFieldModels TagJoinTemplateFieldModels

	// return empty with nil error if not found
	templatesFieldModels, err := readTemplatesByURNFromDB(ctx, r.client.db, templateURN)
	if err != nil {
		return nil, err
	}

	templateModels := templatesFieldModels.toTemplateModels()

	for _, template := range templateModels {
		templateDomain := template.toTemplate()
		templates = append(templates, templateDomain)
	}
	return templates, nil
}

// Read reads all template from database
func (r *TagTemplateRepository) ReadAll(ctx context.Context) ([]tag.Template, error) {
	templates := []tag.Template{}
	var templatesFieldModels TagJoinTemplateFieldModels
	// return empty with nil error if not found
	templatesFieldModels, err := readAllTemplatesFromDB(ctx, r.client.db)
	if err != nil {
		return nil, err
	}

	templateModels := templatesFieldModels.toTemplateModels()

	for _, template := range templateModels {
		templateDomain := template.toTemplate()
		templates = append(templates, templateDomain)
	}
	return templates, nil
}

// Update updates template into database
func (r *TagTemplateRepository) Update(ctx context.Context, targetURN string, templateDomain *tag.Template) error {
	if templateDomain == nil {
		return errNilTemplate
	}

	templateModel := newTemplateModel(templateDomain)
	updatedTemplateModel := *templateModel
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		timestamp := time.Now().UTC()

		updatedTemplateModel.UpdatedAt = timestamp
		if err := updateTemplateToDBTx(ctx, tx, targetURN, &updatedTemplateModel); err != nil {
			return fmt.Errorf("failed to update a field: %w", err)
		}

		for _, field := range templateModel.Fields {
			field.TemplateURN = templateModel.URN
			field.UpdatedAt = timestamp

			if field.ID == 0 {
				field.CreatedAt = timestamp
				field.TemplateURN = templateModel.URN

				if err := insertFieldToDBTx(ctx, tx, &field); err != nil {
					return fmt.Errorf("failed to insert a field: %w", err)
				}

				updatedTemplateModel.Fields = append(updatedTemplateModel.Fields, field)
				continue
			}

			if err := updateFieldToDBTx(ctx, tx, &field); err != nil {
				return fmt.Errorf("failed to update a field: %w", err)
			}
			updatedTemplateModel.Fields = append(updatedTemplateModel.Fields, field)
		}

		*templateModel = updatedTemplateModel
		return nil
	}); err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	*templateDomain = updatedTemplateModel.toTemplate()

	return nil
}

// Delete deletes template and its fields from database
func (r *TagTemplateRepository) Delete(ctx context.Context, templateURN string) error {
	res, err := r.client.db.ExecContext(ctx, `
					DELETE FROM
						tag_templates 
					WHERE
						urn = $1`, templateURN)
	if err != nil {
		return fmt.Errorf("failed to delete template with urn: %w", err)
	}

	tmpRowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get row affected in deleting template: %w", err)
	}

	if tmpRowsAffected == 0 {
		return tag.TemplateNotFoundError{URN: templateURN}
	}
	return nil
}

func insertTemplateToDBTx(ctx context.Context, tx *sqlx.Tx, templateModel *TagTemplateModel) error {
	var insertedTemplate TagTemplateModel
	if err := tx.QueryRowxContext(ctx, `
					INSERT INTO 
					tag_templates 
						(urn,display_name,description,created_at,updated_at) 
					VALUES 
						($1,$2,$3,$4,$5)
					RETURNING *
				`,
		templateModel.URN, templateModel.DisplayName, templateModel.Description, templateModel.CreatedAt, templateModel.UpdatedAt).
		StructScan(&insertedTemplate); err != nil {
		return fmt.Errorf("failed to insert a template: %w", err)
	}

	*templateModel = insertedTemplate
	return nil
}

func insertFieldToDBTx(ctx context.Context, tx *sqlx.Tx, field *TagTemplateFieldModel) error {
	var insertedField TagTemplateFieldModel
	if err := tx.QueryRowxContext(ctx, `
					INSERT INTO 
					tag_template_fields 
						(urn, display_name, description, data_type, options, required, template_urn, created_at, updated_at)
					VALUES 
						($1,$2,$3,$4,$5,$6,$7,$8,$9)
					RETURNING *
					`,
		field.URN, field.DisplayName, field.Description, field.DataType, field.Options, field.Required, field.TemplateURN, field.CreatedAt, field.UpdatedAt).
		StructScan(&insertedField); err != nil {
		return fmt.Errorf("failed to insert a field: %w", err)
	}
	*field = insertedField
	return nil
}

func readAllTemplatesFromDB(ctx context.Context, db *sqlx.DB) (TagJoinTemplateFieldModels, error) {
	var templateFields TagJoinTemplateFieldModels
	// return empty with nil error if not found
	if err := db.Select(&templateFields, `
		SELECT
			t.urn as "tag_templates.urn", t.display_name as "tag_templates.display_name", t.description as "tag_templates.description",
			t.created_at as "tag_templates.created_at", t.updated_at as "tag_templates.updated_at",
			f.id as "tag_template_fields.id", f.urn as "tag_template_fields.urn", f.display_name as "tag_template_fields.display_name", f.description as "tag_template_fields.description",
			f.data_type as "tag_template_fields.data_type", f.options as "tag_template_fields.options", f.required as "tag_template_fields.required", f.template_urn as "tag_template_fields.template_urn",
			f.created_at as "tag_template_fields.created_at", f.updated_at as "tag_template_fields.updated_at"
		FROM
			tag_templates t
		JOIN
			tag_template_fields f
		ON
			f.template_urn = t.urn`); err != nil {
		return nil, fmt.Errorf("tag_templates: failed to read all from DB %w", err)
	}

	return templateFields, nil
}

func readTemplatesByURNFromDB(ctx context.Context, db *sqlx.DB, templateURN string) (TagJoinTemplateFieldModels, error) {
	var templateFields TagJoinTemplateFieldModels
	// return empty with nil error if not found
	if err := db.Select(&templateFields, `
		SELECT
			t.urn as "tag_templates.urn", t.display_name as "tag_templates.display_name", t.description as "tag_templates.description",
			t.created_at as "tag_templates.created_at", t.updated_at as "tag_templates.updated_at",
			f.id as "tag_template_fields.id", f.urn as "tag_template_fields.urn", f.display_name as "tag_template_fields.display_name", f.description as "tag_template_fields.description",
			f.data_type as "tag_template_fields.data_type", f.options as "tag_template_fields.options", f.required as "tag_template_fields.required", f.template_urn as "tag_template_fields.template_urn",
			f.created_at as "tag_template_fields.created_at", f.updated_at as "tag_template_fields.updated_at"
		FROM
			tag_templates t
		JOIN
			tag_template_fields f
		ON
			f.template_urn = t.urn
		WHERE
			t.urn = $1`, templateURN); err != nil {
		return nil, fmt.Errorf("tag_templates: failed to read from DB %w", err)
	}

	return templateFields, nil
}

func updateTemplateToDBTx(ctx context.Context, tx *sqlx.Tx, targetTemplateURN string, templateModel *TagTemplateModel) error {
	var updatedTemplate TagTemplateModel
	if err := tx.QueryRowxContext(ctx, `
					UPDATE
						tag_templates 
					SET
						urn = $1, display_name = $2, description = $3, updated_at = $4
					WHERE
						urn = $5
					RETURNING *`,
		templateModel.URN, templateModel.DisplayName, templateModel.Description, templateModel.UpdatedAt, targetTemplateURN).
		StructScan(&updatedTemplate); err != nil {
		// scan returns sql.ErrNoRows if no rows
		if errors.Is(err, sql.ErrNoRows) {
			return tag.TemplateNotFoundError{URN: templateModel.URN}
		}
		return fmt.Errorf("failed building update template sql: %w", err)
	}

	*templateModel = updatedTemplate
	return nil
}

func updateFieldToDBTx(ctx context.Context, tx *sqlx.Tx, field *TagTemplateFieldModel) error {
	var updatedField TagTemplateFieldModel

	if err := tx.QueryRowxContext(ctx, `
					UPDATE
						tag_template_fields
					SET
						urn = $1, display_name = $2, description = $3, data_type = $4, options = $5, 
						required = $6, template_urn = $7, updated_at = $8
					WHERE
						id = $9 AND template_urn = $7
					RETURNING *`,
		field.URN, field.DisplayName, field.Description, field.DataType, field.Options, field.Required,
		field.TemplateURN, field.UpdatedAt, field.ID).
		StructScan(&updatedField); err != nil {
		return fmt.Errorf("failed updating fields: %w", err)
	}

	if updatedField.ID == 0 {
		return errors.New("field not found when updating fields")
	}

	*field = updatedField
	return nil
}

// NewTagTemplateRepository initializes template repository clients
// all methods in template repository uses passed by reference
// which will mutate the reference variable in method's argument
func NewTagTemplateRepository(c *Client) (*TagTemplateRepository, error) {
	if c == nil {
		return nil, errNilPostgresClient
	}
	return &TagTemplateRepository{
		client: c,
	}, nil
}
