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
)

var (
	errNilDomainTemplate = errors.New("domain template is nil")
)

// TemplateRepository is a type that manages template operation to the primary database
type TemplateRepository struct {
	client *Client
}

// Create inserts template to database
func (r *TemplateRepository) Create(ctx context.Context, domainTemplate *tag.Template) error {
	if domainTemplate == nil {
		return errNilDomainTemplate
	}

	modelTemplate := r.toModelTemplate(*domainTemplate)

	timestamp := time.Now().UTC()
	modelTemplate.CreatedAt = timestamp
	modelTemplate.UpdatedAt = timestamp

	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		return insertTemplateFieldToDBTx(ctx, tx, &modelTemplate)
	}); err != nil {
		return errors.New("failed to insert template")
	}

	r.updateDomainTemplate(domainTemplate, modelTemplate)
	return nil
}

// Read reads template from database
func (r *TemplateRepository) Read(ctx context.Context, filter tag.Template) (output []tag.Template, err error) {
	templatesFields, err := readTemplatesJoinFieldsFromDB(ctx, r.client.db, filter.URN)
	if err != nil {
		err = errors.Wrap(err, "error fetching templates")
		return
	}

	templates := templatesFields.toModelTemplates()

	for _, record := range templates {
		domainTemplate := r.toDomainTemplate(record)
		output = append(output, domainTemplate)
	}
	return
}

// Update updates template into database
func (r *TemplateRepository) Update(ctx context.Context, targetURN string, domainTemplate *tag.Template) error {
	if domainTemplate == nil {
		return errNilDomainTemplate
	}
	modelTemplate := r.toModelTemplate(*domainTemplate)
	modelTemplate.UpdatedAt = time.Now().UTC()
	if err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		return updateTemplateFieldToDBTx(ctx, tx, targetURN, &modelTemplate)
	}); err != nil {
		return errors.Wrap(err, "failed to update template")
	}
	*domainTemplate = r.toDomainTemplate(modelTemplate)
	return nil
}

// Delete deletes template and its fields from database
func (r *TemplateRepository) Delete(ctx context.Context, filter tag.Template) error {
	if err := deleteTemplateFromDB(ctx, r.client.db, filter.URN); err != nil {
		return errors.Wrap(err, "error deleting template")
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
func NewTemplateRepository(c *Client) (*TemplateRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &TemplateRepository{
		client: c,
	}, nil
}
