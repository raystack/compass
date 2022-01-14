package postgres

import (
	"context"
	"strings"

	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
)

const (
	fieldOptionSeparator = ","
	insertTemplateSQL    = `INSERT INTO templates (urn, display_name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	insertFieldSQL       = `INSERT INTO 
								fields (urn, display_name, description, data_type, options, required, template_urn, created_at, updated_at) 
							VALUES 
								($1, $2, $3, $4, $5, $6, $7, $8, $9)`
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
	updateTemplateSQL = `UPDATE 
							templates 
						 SET 
							(display_name, description, created_at, updated_at) =
							(:display_name, :description, :created_at, :updated_at)
						 WHERE 
							urn = :urn`
	updateFieldSQL = `  UPDATE
							fields
						SET
							(urn, display_name, description, data_type, options, required, template_urn, created_at, updated_at) =
							(:urn, :display_name, :description, :data_type, :options, :required, :template_urn, :created_at, :updated_at)
						WHERE
							id = :id AND template_urn = :template_urn`

	upsertFieldSQL = `INSERT INTO 
							fields (urn, display_name, description, data_type, options, required, template_urn, created_at, updated_at) 
						VALUES 
							(:urn, :display_name,:description, :data_type, :options, :required, :template_urn, :created_at, :updated_at)
		 				ON CONFLICT (urn, template_urn) DO 
						UPDATE SET
							(urn, display_name, description, data_type, options, required, template_urn, created_at, updated_at) =
							(:urn, :display_name, :description, :data_type, :options, :required, :template_urn, :created_at, :updated_at)`

	// cascade deletion
	deleteTemplateSQL = `DELETE FROM templates WHERE urn = $1`
)

var (
	errNilDomainTemplate = errors.New("domain template is nil")
)

// TemplateRepository is a type that manages template operation to the primary database
type TemplateRepository struct {
	dbClient *Client
}

// Create inserts template to database
func (r *TemplateRepository) Create(ctx context.Context, domainTemplate *tag.Template) error {
	if domainTemplate == nil {
		return errNilDomainTemplate
	}
	if r.dbClient == nil {
		return errNilDBClient
	}
	modelTemplate := r.toModelTemplate(*domainTemplate)
	err := r.createTemplateTx(ctx, &modelTemplate)
	if err != nil {
		return err
	}

	templates, err := r.readTemplates(ctx, domainTemplate.URN)
	if err != nil || len(templates) < 1 {
		return errors.Wrap(err, "failed to reflect newly created template")
	}
	r.updateDomainTemplate(domainTemplate, templates[0])
	return nil
}

// Read reads template from database
func (r *TemplateRepository) Read(ctx context.Context, filter tag.Template) ([]tag.Template, error) {
	if r.dbClient == nil {
		return nil, errNilDBClient
	}
	output := []tag.Template{}
	templates, err := r.readTemplates(ctx, filter.URN)
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
	if r.dbClient == nil {
		return errNilDBClient
	}
	templateToUpdateWith := r.toModelTemplate(*domainTemplate)
	err := r.updateTemplateTx(ctx, &templateToUpdateWith)
	if err != nil {
		return errors.Wrap(err, "error updating template")
	}

	return nil
}

// Delete deletes template and its fields from database
func (r *TemplateRepository) Delete(ctx context.Context, filter tag.Template) error {
	if r.dbClient == nil {
		return errNilDBClient
	}

	err := r.deleteTemplate(ctx, filter.URN)
	if err != nil {
		return errors.Wrap(err, "error deleting template")
	}

	return nil
}

func (r *TemplateRepository) createTemplateTx(ctx context.Context, modelTemplate *Template) (err error) {

	tx, txErr := r.dbClient.Conn.BeginTxx(ctx, nil)
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

	_, txErr = tx.ExecContext(ctx, insertTemplateSQL,
		modelTemplate.URN,
		modelTemplate.DisplayName,
		modelTemplate.Description,
		modelTemplate.CreatedAt,
		modelTemplate.UpdatedAt)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to insert a template")
		return
	}

	for _, field := range modelTemplate.Fields {
		_, txErr = tx.ExecContext(ctx, insertFieldSQL,
			field.URN,
			field.DisplayName,
			field.Description,
			field.DataType,
			field.Options,
			field.Required,
			modelTemplate.URN,
			field.CreatedAt,
			field.UpdatedAt)
		if txErr != nil {
			err = errors.Wrap(txErr, "failed to insert a field")
			return
		}
	}

	return
}

func (r *TemplateRepository) readTemplates(ctx context.Context, templateURN string) (templates []Template, err error) {
	tmplts := make(map[string]Template, 0) // template urn as key
	flds := make(map[uint]Field, 0)        // field id as key

	rows, err := r.dbClient.Conn.QueryxContext(ctx, readTemplatesSQL, templateURN)
	if err != nil {
		err = errors.Wrap(err, "failed to create db transaction when reading templates")
		return
	}

	for rows.Next() {
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

func (r *TemplateRepository) updateTemplateTx(ctx context.Context, modelTemplate *Template) (err error) {
	tx, txErr := r.dbClient.Conn.BeginTxx(ctx, nil)
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

	res, txErr := tx.NamedExecContext(ctx, updateTemplateSQL, modelTemplate)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed updating templates")
		return
	}

	rowsAffected, txErr := res.RowsAffected()
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to get row affected in updating template")
		return
	}

	// fields cannot be exist if template does not exist
	if rowsAffected == 0 {
		err = tag.TemplateNotFoundError{URN: modelTemplate.URN}
		return
	}

	for _, field := range modelTemplate.Fields {
		field.TemplateURN = modelTemplate.URN
		if field.ID == 0 {
			_, txErr = tx.ExecContext(ctx, insertFieldSQL,
				field.URN,
				field.DisplayName,
				field.Description,
				field.DataType,
				field.Options,
				field.Required,
				modelTemplate.URN,
				field.CreatedAt,
				field.UpdatedAt)
			if txErr != nil {
				err = errors.Wrapf(txErr, "failed to upsert field of template: %s", field.TemplateURN)
				return
			}
			continue
		}

		res, txErr = tx.NamedExecContext(ctx, updateFieldSQL, field)
		if txErr != nil {
			err = errors.Wrap(txErr, "failed updating fields")
			return
		}

		rowsAffected, txErr = res.RowsAffected()
		if txErr != nil {
			err = errors.Wrap(txErr, "failed to get row affected when updating fields")
			return
		}
		if rowsAffected == 0 {
			err = errors.New("field not found when updating fields")
			return
		}
	}
	return
}

func (r *TemplateRepository) deleteTemplate(ctx context.Context, templateURN string) error {
	res, err := r.dbClient.Conn.ExecContext(ctx, deleteTemplateSQL, templateURN)
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
func NewTemplateRepository(dbClient *Client) *TemplateRepository {
	if dbClient == nil {
		panic(errNilDBClient.Error())
	}
	return &TemplateRepository{
		dbClient: dbClient,
	}
}
