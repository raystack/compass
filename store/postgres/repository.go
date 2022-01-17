package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
)

const (
	readTemplatesSQL = `	SELECT 
								templates.urn, 
								templates.display_name, 
								templates.description, 
								templates.created_at,
								templates.updated_at,
								fields.id, 
								fields.urn, 
								fields.display_name,
								fields.description,
								fields.data_type,
								fields.options,
								fields.required,
								fields.template_urn, 
								fields.created_at,
								fields.updated_at 
							FROM 
								templates JOIN fields 
							ON fields.template_urn = templates.urn
							WHERE templates.urn = $1`
	insertTemplateSQL = `INSERT INTO templates (urn, display_name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING *`
	insertFieldSQL    = `	INSERT INTO 
								fields (urn, display_name, description, data_type, options, required, template_urn, created_at, updated_at) 
							VALUES 
								($1, $2, $3, $4, $5, $6, $7, $8, $9)
							RETURNING *`
)

func readTemplates(ctx context.Context, dbConn *sqlx.DB, templateURN string) (templates []Template, err error) {
	tmplts := make(map[string]Template, 0) // template urn as key
	flds := make(map[uint]Field, 0)        // field id as key

	rows, err := dbConn.QueryxContext(ctx, readTemplatesSQL, templateURN)
	if err != nil {
		err = errors.Wrap(err, "failed to reading templates")
		return
	}

	found := false
	for rows.Next() {
		found = true
		tmp := Template{}
		fld := Field{}
		rows.Scan(
			&tmp.URN,
			&tmp.DisplayName,
			&tmp.Description,
			&tmp.CreatedAt,
			&tmp.UpdatedAt,
			&fld.ID,
			&fld.URN,
			&fld.DisplayName,
			&fld.Description,
			&fld.DataType,
			&fld.Options,
			&fld.Required,
			&fld.TemplateURN,
			&fld.CreatedAt,
			&fld.UpdatedAt,
		)
		if _, ok := tmplts[tmp.URN]; !ok {
			tmplts[tmp.URN] = tmp
		}
		if _, ok := flds[fld.ID]; !ok {
			flds[fld.ID] = fld
		}
	}

	if !found {
		return nil, &tag.TemplateNotFoundError{URN: templateURN}
	}

	for _, field := range flds {
		tmp := tmplts[field.TemplateURN]
		tmp.Fields = append(tmp.Fields, field)
		tmplts[field.TemplateURN] = tmp
	}

	for _, tmp := range tmplts {
		templates = append(templates, tmp)
	}
	return
}

func CreateTemplateTx(ctx context.Context, db *sqlx.DB, modelTemplate *Template, timestamp time.Time) (modelTemplateOutput *Template, err error) {
	tx, txErr := db.BeginTxx(ctx, nil)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to create db transaction when creating template")
		return
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.Wrap(err, "create template error")
				err = errors.Wrap(txErr, "error during rollback")
			}
		} else {
			txErr := tx.Commit()
			if txErr != nil {
				err = errors.Wrap(txErr, "error during commit")
			}
		}
	}()

	var insertedTemplate Template
	txErr = tx.QueryRowxContext(ctx, insertTemplateSQL,
		modelTemplate.URN,
		modelTemplate.DisplayName,
		modelTemplate.Description,
		timestamp,
		timestamp).StructScan(&insertedTemplate)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to insert a template")
		return
	}

	for _, field := range modelTemplate.Fields {
		var insertedField Field
		txErr = tx.QueryRowxContext(ctx, insertFieldSQL,
			field.URN,
			field.DisplayName,
			field.Description,
			field.DataType,
			field.Options,
			field.Required,
			modelTemplate.URN,
			timestamp,
			timestamp).StructScan(&insertedField)
		if txErr != nil {
			err = errors.Wrap(txErr, "failed to insert a field")
			return
		}
		insertedTemplate.Fields = append(insertedTemplate.Fields, insertedField)
	}
	modelTemplateOutput = &insertedTemplate
	return
}
