package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
)

const (
	fieldOptionSeparator = ","

	updateTemplateSQL = `UPDATE 
							templates 
						 SET (display_name, description, updated_at) = ($1, $2, $3)
						 WHERE urn = $4 RETURNING *`
	updateFieldSQL = `  UPDATE
							fields
						SET
							(urn, display_name, description, data_type, options, required, template_urn, updated_at) =
							($1, $2, $3, $4, $5, $6, $7, $8)
						WHERE id = $9 AND template_urn = $7 RETURNING *`
	// cascade deletion
	deleteTemplateSQL = `DELETE FROM templates WHERE urn = $1`
)

var (
	errNilDomainTemplate = errors.New("domain template is nil")
)

// TemplateRepository is a type that manages template operation to the primary database
type TemplateRepository struct {
	db *sqlx.DB
}

// Create inserts template to database
func (r *TemplateRepository) Create(ctx context.Context, domainTemplate *tag.Template) error {
	if domainTemplate == nil {
		return errNilDomainTemplate
	}
	if r.db == nil {
		return errNilDBClient
	}
	modelTemplate := r.toModelTemplate(*domainTemplate)
	createdTemplate, err := CreateTemplateTx(ctx, r.db, &modelTemplate, time.Now().UTC())
	if err != nil {
		return err
	}

	r.updateDomainTemplate(domainTemplate, *createdTemplate)
	return nil
}

// Read reads template from database
func (r *TemplateRepository) Read(ctx context.Context, filter tag.Template) ([]tag.Template, error) {
	if r.db == nil {
		return nil, errNilDBClient
	}
	output := []tag.Template{}
	templates, err := readTemplates(ctx, r.db, filter.URN)
	if err != nil {
		return output, errors.Wrap(err, "error fetching templates")
	}
	for _, record := range templates {
		tmplt := r.toDomainTemplate(record)
		output = append(output, tmplt)
	}
	return output, nil
}

// Update updates template into database
func (r *TemplateRepository) Update(ctx context.Context, domainTemplate *tag.Template) error {
	if domainTemplate == nil {
		return errNilDomainTemplate
	}
	if r.db == nil {
		return errNilDBClient
	}

	templateToUpdateWith := r.toModelTemplate(*domainTemplate)
	updatedTemplate, err := r.updateTemplateTx(ctx, &templateToUpdateWith, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, "error updating template")
	}

	*domainTemplate = r.toDomainTemplate(*updatedTemplate)
	return nil
}

// Delete deletes template and its fields from database
func (r *TemplateRepository) Delete(ctx context.Context, filter tag.Template) error {
	if r.db == nil {
		return errNilDBClient
	}

	err := r.deleteTemplate(ctx, filter.URN)
	if err != nil {
		return errors.Wrap(err, "error deleting template")
	}

	return nil
}

func (r *TemplateRepository) updateTemplateTx(ctx context.Context, modelTemplate *Template, timestamp time.Time) (modelTemplateOutput *Template, err error) {
	tx, txErr := r.db.BeginTxx(ctx, nil)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to begin db transaction when updating template")
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.Wrap(err, "update template error")
				err = errors.Wrap(txErr, "error during rollback")
			}
		} else {
			txErr := tx.Commit()
			if txErr != nil {
				err = errors.Wrap(txErr, "error during commit")
			}
		}
	}()

	var updatedTemplate Template
	txErr = tx.QueryRowxContext(ctx, updateTemplateSQL,
		modelTemplate.DisplayName,
		modelTemplate.Description,
		timestamp,
		modelTemplate.URN).StructScan(&updatedTemplate)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed updating templates")
		return
	}

	// fields cannot be exist if template does not exist
	if updatedTemplate.URN == "" {
		err = tag.TemplateNotFoundError{URN: modelTemplate.URN}
		return
	}

	for _, field := range modelTemplate.Fields {
		var updatedField Field
		field.TemplateURN = modelTemplate.URN
		if field.ID == 0 {
			txErr = tx.QueryRowxContext(ctx, insertFieldSQL,
				field.URN,
				field.DisplayName,
				field.Description,
				field.DataType,
				field.Options,
				field.Required,
				modelTemplate.URN,
				timestamp,
				timestamp).StructScan(&updatedField)
			if txErr != nil {
				err = errors.Wrapf(txErr, "failed to upsert field of template: %s", field.TemplateURN)
				return
			}
			updatedTemplate.Fields = append(updatedTemplate.Fields, updatedField)
			continue
		}
		txErr = tx.QueryRowxContext(ctx, updateFieldSQL,
			field.URN,
			field.DisplayName,
			field.Description,
			field.DataType,
			field.Options,
			field.Required,
			field.TemplateURN,
			timestamp,
			field.ID).StructScan(&updatedField)
		if txErr != nil {
			err = errors.Wrap(txErr, "failed updating fields")
			return
		}

		if updatedField.ID == 0 {
			err = errors.New("field not found when updating fields")
			return
		}

		updatedTemplate.Fields = append(updatedTemplate.Fields, updatedField)
	}

	modelTemplateOutput = &updatedTemplate
	return
}

func (r *TemplateRepository) deleteTemplate(ctx context.Context, templateURN string) error {
	res, err := r.db.ExecContext(ctx, deleteTemplateSQL, templateURN)
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

func (r *TemplateRepository) updateDomainTemplate(target *tag.Template, source Template) {
	target.URN = source.URN
	target.DisplayName = source.DisplayName
	target.Description = source.Description
	target.CreatedAt = source.CreatedAt
	target.UpdatedAt = source.UpdatedAt
	target.Fields = r.toDomainField(source.Fields)
}

func (r *TemplateRepository) toDomainTemplate(modelTemplate Template) tag.Template {
	return tag.Template{
		URN:         modelTemplate.URN,
		DisplayName: modelTemplate.DisplayName,
		Description: modelTemplate.Description,
		Fields:      r.toDomainField(modelTemplate.Fields),
		CreatedAt:   modelTemplate.CreatedAt,
		UpdatedAt:   modelTemplate.UpdatedAt,
	}
}

func (r *TemplateRepository) toDomainField(listOfModelField []Field) []tag.Field {
	output := make([]tag.Field, len(listOfModelField))
	for i, field := range listOfModelField {
		var options []string
		if field.Options != nil {
			options = strings.Split(*field.Options, fieldOptionSeparator)
		}
		output[i] = tag.Field{
			ID:          field.ID,
			URN:         field.URN,
			DisplayName: field.DisplayName,
			Description: field.Description,
			DataType:    field.DataType,
			Options:     options,
			Required:    field.Required,
			CreatedAt:   field.CreatedAt,
			UpdatedAt:   field.UpdatedAt,
		}
	}
	return output
}

func (r *TemplateRepository) toModelTemplate(domainTemplate tag.Template) Template {
	return Template{
		URN:         domainTemplate.URN,
		DisplayName: domainTemplate.DisplayName,
		Description: domainTemplate.Description,
		Fields:      r.toModelField(domainTemplate.Fields),
	}
}

func (r *TemplateRepository) toModelField(listOfDomainField []tag.Field) []Field {
	var output []Field
	for _, field := range listOfDomainField {
		var options *string
		if len(field.Options) > 0 {
			joinedOptions := strings.Join(field.Options, fieldOptionSeparator)
			options = &joinedOptions
		}
		output = append(output, Field{
			ID:          field.ID,
			URN:         field.URN,
			DisplayName: field.DisplayName,
			Description: field.Description,
			DataType:    field.DataType,
			Options:     options,
			Required:    field.Required,
		})
	}
	return output
}

// NewTemplateRepository initializes template repository clients
func NewTemplateRepository(db *sqlx.DB) *TemplateRepository {
	if db == nil {
		panic(errNilDBClient.Error())
	}
	return &TemplateRepository{
		db: db,
	}
}
