package postgres

import (
	"strings"

	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const fieldOptionSeparator = ","

var (
	errNilDomainTemplate = errors.New("domain template is nil")
)

// TemplateRepository is a type that manages template operation to the primary database
type TemplateRepository struct {
	dbClient *gorm.DB
}

// Create inserts template to database
func (r *TemplateRepository) Create(domainTemplate *tag.Template) error {
	if domainTemplate == nil {
		return errNilDomainTemplate
	}
	if r.dbClient == nil {
		return errNilDBClient
	}
	modelTemplate := r.toModelTemplate(*domainTemplate)
	if result := r.dbClient.Create(&modelTemplate); result.Error != nil {
		return result.Error
	}
	r.updateDomainTemplate(domainTemplate, modelTemplate)
	return nil
}

// Read reads template from database
func (r *TemplateRepository) Read(filter tag.Template) ([]tag.Template, error) {
	if r.dbClient == nil {
		return nil, errNilDBClient
	}
	output := []tag.Template{}
	modelTemplate := r.toModelTemplate(filter)
	var rows []Template
	res := r.dbClient.Where(modelTemplate).Preload("Fields").Find(&rows)
	if res.Error != nil {
		return output, errors.Wrap(res.Error, "error fetching templates")
	}
	for _, record := range rows {
		tmplt := r.toDomainTemplate(record)
		output = append(output, tmplt)
	}
	return output, nil
}

// Update updates template into database
func (r *TemplateRepository) Update(domainTemplate *tag.Template) error {
	if domainTemplate == nil {
		return errNilDomainTemplate
	}
	if r.dbClient == nil {
		return errNilDBClient
	}
	err := r.dbClient.Transaction(func(tx *gorm.DB) error {
		templateToUpdateWith := r.toModelTemplate(*domainTemplate)
		if err := r.updateModelTemplateToDB(tx, domainTemplate.URN, templateToUpdateWith); err != nil {
			return err
		}
		return r.updateModelFieldsToDB(tx, templateToUpdateWith.URN, templateToUpdateWith.Fields)
	})
	if err != nil {
		return errors.Wrap(err, "error updating template")
	}
	var recordModelTemplate Template
	res := r.dbClient.Preload("Fields").First(&recordModelTemplate, "urn = ?", domainTemplate.URN)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return tag.TemplateNotFoundError{URN: domainTemplate.URN}
	}
	if res.Error != nil {
		return errors.Wrap(res.Error, "error finding template")
	}

	r.updateDomainTemplate(domainTemplate, recordModelTemplate)
	return nil
}

// Delete deletes template and its fields from database
func (r *TemplateRepository) Delete(filter tag.Template) error {
	if r.dbClient == nil {
		return errNilDBClient
	}
	var recordModelTemplate Template
	res := r.dbClient.Preload("Fields").First(&recordModelTemplate, "urn = ?", filter.URN)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return tag.TemplateNotFoundError{URN: filter.URN}
	}
	if res.Error != nil {
		return errors.Wrap(res.Error, "error finding template")
	}
	deleteResult := r.dbClient.Transaction(func(tx *gorm.DB) error {
		for _, field := range recordModelTemplate.Fields {
			if err := tx.Delete(&field).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&recordModelTemplate).Error
	})
	return deleteResult
}

func (r *TemplateRepository) updateModelTemplateToDB(tx *gorm.DB, targetURN string, templateToUpdateWith Template) error {
	var recordModelTemplate Template
	if err := tx.First(&recordModelTemplate, "urn = ?", targetURN).Error; err != nil {
		return err
	}
	result := tx.Model(&recordModelTemplate).Updates(&templateToUpdateWith)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TemplateRepository) updateModelFieldsToDB(tx *gorm.DB, templateURN string, fields []Field) error {
	for _, field := range fields {
		field.TemplateURN = templateURN
		if field.ID == 0 {
			createResult := tx.Create(&field)
			if createResult.Error != nil {
				return createResult.Error
			}
		}
		var recordModelFieldByID Field
		queryResult := tx.Where("id = ? and template_urn = ?", field.ID, field.TemplateURN).
			First(&recordModelFieldByID)
		if err := queryResult.Error; err != nil {
			return err
		}
		recordModelFieldByID.URN = field.URN
		recordModelFieldByID.DisplayName = field.DisplayName
		recordModelFieldByID.Description = field.Description
		recordModelFieldByID.DataType = field.DataType
		recordModelFieldByID.Options = field.Options
		recordModelFieldByID.Required = field.Required
		if err := tx.Model(&recordModelFieldByID).Updates(&recordModelFieldByID).Error; err != nil {
			return err
		}
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
func NewTemplateRepository(dbClient *gorm.DB) *TemplateRepository {
	if dbClient == nil {
		panic(errNilDBClient.Error())
	}
	return &TemplateRepository{
		dbClient: dbClient,
	}
}
