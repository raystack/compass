package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/tag"
	"github.com/pkg/errors"
)

const (
	insertTagSQL = `INSERT INTO 
						tags (value, record_urn, record_type, field_id, created_at, updated_at) 
					VALUES 
						($1, $2, $3, $4, $5, $6)
					RETURNING *`
	findDomainTagsSQL = `			SELECT
								templates.urn, 
								templates.display_name, 
								templates.description, 
								templates.created_at,
								templates.updated_at,
								tags.id,
								tags.value,
								tags.record_urn,
								tags.record_type,
								tags.field_id,
								tags.created_at,
								tags.updated_at,
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
							FROM templates 
							INNER JOIN fields ON templates.urn = fields.template_urn
							INNER JOIN tags ON fields.id = tags.field_id`
	findModelTagSQL = `	SELECT
								id,
								value,
								record_urn,
								record_type,
								field_id,
								created_at,
								updated_at
							FROM tags WHERE tags.record_type = $1 AND tags.record_urn = $2`
	filterByRecordAndTemplate = ` WHERE tags.record_urn = $1 AND tags.record_type = $2 AND templates.urn = $3`
	filterByRecord            = ` WHERE tags.record_urn = $1 AND tags.record_type = $2`

	readTagsWithFieldSQL = `	SELECT 
									tags.value,
									tags.record_urn,
									tags.record_type,
									tags.field_id,
									tags.created_at,
									tags.updated_at,
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
									tags JOIN fields 
								ON fields.id = tags.field_id
								WHERE tags.record_type = $1 and tags.record_urn = $2 and tags.field_id = $3`
	getTagsJoinFieldsSQL = `	SELECT 
									tags.value,
									tags.record_urn,
									tags.record_type,
									tags.field_id,
									tags.created_at,
									tags.updated_at,
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
									tags JOIN fields 
								ON fields.id = tags.field_id`

	readTagsByRecordSQL = `	SELECT * 
								value,
								record_urn,
								record_type,
								field_id,
								created_at,
								updated_at
							FROM 
								tags
							WHERE record_type = $1 and record_urn = $2`

	upsertTagSQL = `	INSERT INTO tags (value, record_urn, record_type, field_id, created_at, updated_at)` +
		`VALUES ($1, $2, $3, $4, $5, $6)` +
		`ON CONFLICT (record_urn, record_type, field_id) DO ` +
		`UPDATE SET (value, record_urn, record_type, field_id, created_at, updated_at) = ($1, $2, $3, $4, $5, $6)` +
		`RETURNING *`

	deleteTagSql             = `DELETE FROM tags`
	filterByRecordAndFieldID = ` WHERE tags.record_urn = $1 AND tags.record_type = $2 AND tags.field_id = $3`
)

var (
	errNilDomainTag    = errors.New("domain tag is nil")
	errEmptyRecordType = errors.New("record type should not be empty")
	errEmptyRecordURN  = errors.New("record urn should not be empty")
)

// TagRepository is a type that manages tag operation ot the primary database
type TagRepository struct {
	db *sqlx.DB
}

// Create inserts tag to database
func (r *TagRepository) Create(ctx context.Context, domainTag *tag.Tag) error {
	if r.db == nil {
		return errNilDBClient
	}
	if domainTag == nil {
		return errNilDomainTag
	}
	recordModelTemplates, err := readTemplates(ctx, r.db, domainTag.TemplateURN)
	if err != nil {
		return err
	}
	if len(recordModelTemplates) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}
	domainTemplate := r.convertToDomainTemplate(recordModelTemplates[0])
	insertedModelTags, err := r.insertDomainTagToDB(ctx, domainTag, time.Now().UTC())
	if err != nil {
		return err
	}
	return r.complementDomainTag(domainTag, domainTemplate, insertedModelTags)
}

// Read reads tags grouped by its template
func (r *TagRepository) Read(ctx context.Context, filter tag.Tag) ([]tag.Tag, error) {
	if r.db == nil {
		return nil, errNilDBClient
	}
	if filter.RecordType == "" {
		return nil, errEmptyRecordType
	}
	if filter.RecordURN == "" {
		return nil, errEmptyRecordURN
	}

	tags, err := r.findDomainTags(ctx, filter.RecordType, filter.RecordURN, filter.TemplateURN)
	if err != nil {
		return nil, err
	}

	return tags, nil
	// if filter.TemplateURN != "" {
	// 	// get templates by urn
	// 	recordModelTemplates, err := readTemplates(ctx, r.db, filter.TemplateURN)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if len(recordModelTemplates) < 1 {
	// 		return nil, tag.TemplateNotFoundError{URN: filter.TemplateURN}
	// 	}
	// 	domainTag, err := r.readByRecordAndTemplateFromDB(ctx, filter.RecordType, filter.RecordURN, recordModelTemplates[0])
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return []tag.Tag{domainTag}, nil
	// }
	// return r.readByRecordFromDB(ctx, filter.RecordType, filter.RecordURN)
}

// Update updates tags in the database
func (r *TagRepository) Update(ctx context.Context, domainTag *tag.Tag) error {
	if r.db == nil {
		return errNilDBClient
	}
	if domainTag == nil {
		return errNilDomainTag
	}
	recordModelTemplates, err := readTemplates(ctx, r.db, domainTag.TemplateURN)

	if err != nil {
		return err
	}
	if len(recordModelTemplates) < 1 {
		return tag.TemplateNotFoundError{URN: domainTag.TemplateURN}
	}
	domainTemplate := r.convertToDomainTemplate(recordModelTemplates[0])
	updatedModelTags, err := r.updateDomainTagToDB(ctx, domainTag, time.Now().UTC())
	if err != nil {
		return err
	}
	return r.complementDomainTag(domainTag, domainTemplate, updatedModelTags)
	// 	recordModelTemplate, err := r.getTemplateWithFields(domainTag.TemplateURN)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	domainTemplate := r.convertToDomainTemplate(recordModelTemplate)
	// 	err = r.dbClient.Transaction(func(tx *gorm.DB) error {
	// 		return r.updateDomainTagToDB(tx, domainTag)
	// 	})
	// 	if err == nil {
	// 		listOfRecordModelTag, err := r.readListOfModelTagFromDB(r.dbClient, domainTag.RecordType, domainTag.RecordURN, domainTemplate)
	// 		if err == nil {
	// 			err = r.complementDomainTag(domainTag, domainTemplate, listOfRecordModelTag)
	// 		}
	// 	}
	// 	return err
	// }
}

// Delete deletes tags from database
func (r *TagRepository) Delete(ctx context.Context, domainTag tag.Tag) error {
	if r.db == nil {
		return errNilDBClient
	}
	if domainTag.RecordURN == "" {
		return errEmptyRecordURN
	}
	deletedModelTags := []Tag{}
	if domainTag.TemplateURN != "" {
		recordModelTemplates, err := readTemplates(ctx, r.db, domainTag.TemplateURN)
		if err != nil {
			return err
		}
		for _, template := range recordModelTemplates {
			for _, field := range template.Fields {
				deletedModelTags = append(deletedModelTags, Tag{
					RecordURN:  domainTag.RecordURN,
					RecordType: domainTag.RecordType,
					FieldID:    field.ID,
				})
			}
		}
	} else {
		deletedModelTags = append(deletedModelTags, Tag{
			RecordURN:  domainTag.RecordURN,
			RecordType: domainTag.RecordType,
		})
	}
	return r.deleteTagsTx(ctx, deletedModelTags)
}

func (r *TagRepository) updateDomainTagToDB(ctx context.Context, tag *tag.Tag, timestamp time.Time) (newModelTags []Tag, err error) {

	tx, txErr := r.db.BeginTxx(ctx, nil)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to create db transaction when creating tag")
		return
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.Wrap(err, "update domain tag error")
				err = errors.Wrap(txErr, "error during rollback")
			}
		} else {
			txErr := tx.Commit()
			if txErr != nil {
				err = errors.Wrap(txErr, "error during commit")
			}
		}
	}()

	for _, value := range tag.TagValues {
		if value.FieldValue == nil || value.FieldValue == "" {
			continue
		}
		var updatedModelTag Tag
		txErr = tx.QueryRowxContext(ctx, upsertTagSQL,
			value.FieldValue,
			tag.RecordURN,
			tag.RecordType,
			value.FieldID,
			timestamp,
			timestamp).StructScan(&updatedModelTag)
		if txErr != nil {
			err = errors.Wrap(txErr, "failed to update a domain tag")
			return
		}
		newModelTags = append(newModelTags, updatedModelTag)
	}
	return
}

// func (r *TagRepository) updateDomainTagToDB(tx *gorm.DB, domainTag *tag.Tag) error {
// 	for _, value := range domainTag.TagValues {
// 		if value.FieldValue == nil || value.FieldValue == "" {
// 			continue
// 		}
// 		var recordModelTag Tag
// 		queryResult := tx.Where("record_type = ? and record_urn = ? and field_id = ?", domainTag.RecordType, domainTag.RecordURN, value.FieldID).First(&recordModelTag)
// 		recordModelTag.Value = fmt.Sprintf("%v", value.FieldValue)
// 		if queryResult.Error != nil {
// 			recordModelTag.RecordURN = domainTag.RecordURN
// 			recordModelTag.FieldID = value.FieldID
// 			createResult := tx.Create(&recordModelTag)
// 			if createResult.Error != nil {
// 				return createResult.Error
// 			}
// 			continue
// 		}
// 		updateResult := tx.Updates(&recordModelTag)
// 		if updateResult.Error != nil {
// 			return updateResult.Error
// 		}
// 	}
// 	return nil
// }

func (r *TagRepository) convertToDomainTemplate(modelTemplate Template) tag.Template {
	listOfDomainField := []tag.Field{}
	for _, field := range modelTemplate.Fields {
		var options []string
		if field.Options != nil {
			options = strings.Split(*field.Options, ",")
		}
		listOfDomainField = append(listOfDomainField, tag.Field{
			ID:          field.ID,
			URN:         field.URN,
			DisplayName: field.DisplayName,
			Description: field.Description,
			DataType:    field.DataType,
			Required:    field.Required,
			Options:     options,
			CreatedAt:   field.CreatedAt,
			UpdatedAt:   field.UpdatedAt,
		})
	}
	return tag.Template{
		URN:         modelTemplate.URN,
		DisplayName: modelTemplate.DisplayName,
		Description: modelTemplate.Description,
		Fields:      listOfDomainField,
		CreatedAt:   modelTemplate.CreatedAt,
		UpdatedAt:   modelTemplate.UpdatedAt,
	}
}
func (r *TagRepository) buildDeleteTagSQL(tag Tag) (sql string, filterArgs []interface{}) {
	filterSQL := ` WHERE`
	sequenceCount := 0
	if tag.RecordURN != "" {
		sequenceCount += 1
		filterArgs = append(filterArgs, tag.RecordURN)
		filterSQLParam := ` tags.record_urn = $` + strconv.Itoa(sequenceCount)
		filterSQL = filterSQL + filterSQLParam
	}
	if tag.RecordType != "" {
		sequenceCount += 1
		filterArgs = append(filterArgs, tag.RecordType)
		filterSQLParam := ` AND tags.record_type = $` + strconv.Itoa(sequenceCount)
		filterSQL = filterSQL + filterSQLParam
	}
	if tag.FieldID != 0 {
		sequenceCount += 1
		filterArgs = append(filterArgs, tag.FieldID)
		filterSQLParam := ` AND tags.field_id = $` + strconv.Itoa(sequenceCount)
		filterSQL = filterSQL + filterSQLParam
	}

	return deleteTagSql + filterSQL, filterArgs
}

func (r *TagRepository) deleteTagsTx(ctx context.Context, modelTags []Tag) (err error) {
	tx, txErr := r.db.BeginTxx(ctx, nil)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to create db transaction when deleting model tag")
		return
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.Wrap(err, "delete model tag error")
				err = errors.Wrap(txErr, "error during rollback")
			}
		} else {
			txErr := tx.Commit()
			if txErr != nil {
				err = errors.Wrap(txErr, "error during commit")
			}
		}
	}()

	for _, modelTag := range modelTags {
		// sqlQuery := deleteTagSql + filterByRecord
		// filterArgs := []interface{}{modelTag.RecordURN, modelTag.RecordType}
		// if modelTag.FieldID != 0 {
		// 	sqlQuery = deleteTagSql + filterByRecordAndFieldID
		// 	filterArgs = append(filterArgs, modelTag.FieldID)
		// }

		// var id uint

		sqlQuery, filterArgs := r.buildDeleteTagSQL(modelTag)
		fmt.Println(sqlQuery)
		fmt.Println(filterArgs...)
		_, txErr = r.db.ExecContext(ctx, sqlQuery, filterArgs...)
		if txErr != nil {
			err = errors.Wrapf(txErr, "failed to delete tag with record urn: %s, record type: %s, field id: %s", modelTag.RecordURN, modelTag.RecordType, modelTag.FieldID)
			return
		}

		// if id == 0 {
		// 	// not found
		// 	err = tag.NotFoundError{
		// 		URN:  modelTag.RecordURN,
		// 		Type: modelTag.RecordType,
		// 	}
		// 	return

		// }
	}
	return
}

// func (r *TagRepository) deleteByRecordAndTemplate(tx *gorm.DB, recordType, recordURN string, template Template) error {
// 	var listOfFieldID []uint
// 	for _, field := range template.Fields {
// 		listOfFieldID = append(listOfFieldID, field.ID)
// 	}
// 	return tx.Where("record_type = ? and record_urn = ? and field_id in ?", recordType, recordURN, listOfFieldID).Delete(&Tag{}).Error
// }

// func (r *TagRepository) readByRecordFromDB(ctx context.Context, recordType, recordURN string) ([]tag.Tag, error) {
// 	var listOfRecordModelTag []Tag
// 	err := tx.Preload("Field").
// 		Where("record_type = ? AND record_urn = ?", recordType, recordURN).
// 		Find(&listOfRecordModelTag).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	listOfTemplateURN := r.getListOfTemplateURN(listOfRecordModelTag)
// 	listOfTemplate, err := r.getListOfTemplateByListOfURNFromDB(ctx, listOfTemplateURN)
// 	if err != nil {
// 		return nil, err
// 	}
// 	templateByURN := r.mapTemplateByTemplateURN(listOfTemplate)
// 	tagsByTemplateURN := r.mapTagsByTemplateURN(listOfRecordModelTag)
// 	output := []tag.Tag{}
// 	for templateURN, tags := range tagsByTemplateURN {
// 		domainTag := r.convertToDomainTag(recordType, recordURN, templateByURN[templateURN], tags)
// 		output = append(output, domainTag)
// 	}
// 	return output, nil
// }

func (r *TagRepository) mapTagsByTemplateURN(tags []Tag) map[string][]Tag {
	tagsByTemplateURN := make(map[string][]Tag)
	for _, t := range tags {
		key := t.Field.TemplateURN
		if tagsByTemplateURN[key] == nil {
			tagsByTemplateURN[key] = []Tag{}
		}
		tagsByTemplateURN[key] = append(tagsByTemplateURN[key], t)
	}
	return tagsByTemplateURN
}

func (r *TagRepository) mapTemplateByTemplateURN(templates []Template) map[string]Template {
	templateByURN := make(map[string]Template)
	for _, t := range templates {
		templateByURN[t.URN] = t
	}
	return templateByURN
}

func (r *TagRepository) getListOfTemplateURN(tags []Tag) []string {
	isURNRegistered := make(map[string]bool)
	for _, t := range tags {
		key := t.Field.TemplateURN
		if !isURNRegistered[key] {
			isURNRegistered[key] = true
		}
	}
	var output []string
	for urn := range isURNRegistered {
		output = append(output, urn)
	}
	return output
}

func (r *TagRepository) getListOfTemplateByListOfURNFromDB(ctx context.Context, listOfTemplateURN []string) ([]Template, error) {
	var listOfTemplate []Template
	for _, templateURN := range listOfTemplateURN {
		templates, err := readTemplates(ctx, r.db, templateURN)
		if err != nil {
			return nil, err
		}
		listOfTemplate = append(listOfTemplate, templates...)
	}

	return listOfTemplate, nil
}

func (r *TagRepository) findDomainTags(ctx context.Context, recordType, recordURN string, templateURN string) ([]tag.Tag, error) {
	var listOfRecordModelTag []Tag
	var templates []Template

	fieldsMap := make(map[uint]Field, 0)         // field id as key
	tagsMap := make(map[uint]Tag, 0)             // tag id as key
	templatesMap := make(map[string]Template, 0) // template_urn as key

	sqlQuery := findDomainTagsSQL + filterByRecord
	filterArgs := []interface{}{recordURN, recordType}

	if templateURN != "" {
		sqlQuery = findDomainTagsSQL + filterByRecordAndTemplate
		filterArgs = append(filterArgs, templateURN)
	}

	rows, err := r.db.QueryxContext(ctx, sqlQuery, filterArgs...)
	if err != nil {
		return []tag.Tag{}, fmt.Errorf("error looking for record urn [%s] and template urn [%s]: %w", recordURN, templateURN, err)
	}

	// if rows.Next() == false {
	// 	return []tag.Tag{}, tag.NotFoundError{
	// 		URN:      recordURN,
	// 		Type:     recordType,
	// 		Template: templateURN,
	// 	}
	// }
	// if rows.Next() == false {
	// 	return []tag.Tag{}, nil
	// }
	found := false
	for rows.Next() {
		found = true
		tg := Tag{}
		tmplt := Template{}
		fld := Field{}
		rows.Scan(
			&tmplt.URN,
			&tmplt.DisplayName,
			&tmplt.Description,
			&tmplt.CreatedAt,
			&tmplt.UpdatedAt,
			&tg.ID,
			&tg.Value,
			&tg.RecordURN,
			&tg.RecordType,
			&tg.FieldID,
			&tg.CreatedAt,
			&tg.UpdatedAt,
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

		if _, ok := tagsMap[tg.ID]; !ok {
			tagsMap[tg.ID] = tg
		}
		if _, ok := fieldsMap[fld.ID]; !ok {
			fieldsMap[fld.ID] = fld
		}
		if _, ok := templatesMap[tmplt.URN]; !ok {
			templatesMap[tmplt.URN] = tmplt
		}
	}

	if !found {
		return []tag.Tag{}, tag.NotFoundError{
			URN:      recordURN,
			Type:     recordType,
			Template: templateURN,
		}
	}

	// build template(s)
	for _, field := range fieldsMap {
		tmp := templatesMap[field.TemplateURN]
		tmp.Fields = append(tmp.Fields, field)
		templatesMap[field.TemplateURN] = tmp
	}

	for _, tmp := range templatesMap {
		templates = append(templates, tmp)
	}

	// build tags
	for idx, tag := range tagsMap {
		tag.Field = fieldsMap[tag.FieldID]
		tagsMap[idx] = tag
	}

	for _, tag := range tagsMap {
		listOfRecordModelTag = append(listOfRecordModelTag, tag)
	}

	templateByURN := r.mapTemplateByTemplateURN(templates)
	tagsByTemplateURN := r.mapTagsByTemplateURN(listOfRecordModelTag)

	output := []tag.Tag{}
	for templateURN, tags := range tagsByTemplateURN {
		domainTag := r.convertToDomainTag(recordType, recordURN, templateByURN[templateURN], tags)
		output = append(output, domainTag)
	}
	return output, nil
}

func (r *TagRepository) FindModelTag(ctx context.Context, tag Tag) ([]Tag, error) {
	var modelTagsDB []Tag

	rows, err := r.db.QueryxContext(ctx, findModelTagSQL, tag.RecordType, tag.RecordURN)
	if err != nil {
		return []Tag{}, errors.Wrapf(err, "error looking for record urn [%s] and type [%s]: %w", tag.RecordURN, tag.RecordType)
	}
	for rows.Next() {
		var modelTagDB Tag
		if err = rows.StructScan(&modelTagDB); err != nil {
			return nil, errors.Wrapf(err, "error scanning record urn [%s] and type [%s]", tag.RecordURN, tag.RecordType)
		}
		modelTagsDB = append(modelTagsDB, modelTagDB)
	}

	return modelTagsDB, nil
}

// func (r *TagRepository) readByRecordAndTemplateFromDB(ctx context.Context, recordType, recordURN string, template Template) (tag.Tag, error) {
// 	var listOfRecordModelTag []Tag
// 	fieldsMap := make(map[uint]Field, 0) // template urn as key
// 	tagsMap := make(map[uint]Tag, 0)     // field id as key

// 	for _, field := range template.Fields {
// 		rows, err := r.db.QueryxContext(ctx, readTagsWithFieldSQL, recordType, recordURN, field.ID)
// 		if err != nil {
// 			return tag.Tag{}, fmt.Errorf("error looking for record urn [%s] and template urn [%s]: %w", recordURN, template.URN, err)
// 		}

// 		if rows.Next() == false {
// 			return tag.Tag{}, tag.NotFoundError{
// 				URN:      recordURN,
// 				Type:     recordType,
// 				Template: field.TemplateURN,
// 			}
// 		}

// 		for rows.Next() {
// 			tg := Tag{}
// 			fld := Field{}
// 			rows.Scan(
// 				&tg.Value,
// 				&tg.RecordURN,
// 				&tg.RecordType,
// 				&tg.FieldID,
// 				&tg.CreatedAt,
// 				&tg.UpdatedAt,
// 				&fld.ID,
// 				&fld.URN,
// 				&fld.DisplayName,
// 				&fld.Description,
// 				&fld.DataType,
// 				&fld.Options,
// 				&fld.Required,
// 				&fld.TemplateURN,
// 				&fld.CreatedAt,
// 				&fld.UpdatedAt,
// 			)
// 			if _, ok := tagsMap[tg.ID]; !ok {
// 				tagsMap[tg.ID] = tg
// 			}
// 			if _, ok := fieldsMap[fld.ID]; !ok {
// 				fieldsMap[fld.ID] = fld
// 			}
// 		}
// 	}

// 	for idx, tag := range tagsMap {
// 		tag.Field = fieldsMap[tag.FieldID]
// 		tagsMap[idx] = tag
// 	}

// 	for _, tag := range tagsMap {
// 		listOfRecordModelTag = append(listOfRecordModelTag, tag)
// 	}

// 	return r.convertToDomainTag(recordType, recordURN, template, listOfRecordModelTag), nil
// }

func (r *TagRepository) convertToDomainTag(recordType, recordURN string, template Template, tags []Tag) tag.Tag {
	var listOfDomainTagValue []tag.TagValue
	for _, t := range tags {
		var options []string
		if t.Field.Options != nil {
			options = strings.Split(*t.Field.Options, ",")
		}
		parsedValue, _ := tag.ParseTagValue(template.URN, t.FieldID, t.Field.DataType, t.Value, options)
		listOfDomainTagValue = append(listOfDomainTagValue, tag.TagValue{
			FieldID:          t.FieldID,
			FieldValue:       parsedValue,
			FieldURN:         t.Field.URN,
			FieldDisplayName: t.Field.DisplayName,
			FieldDescription: t.Field.Description,
			FieldDataType:    t.Field.DataType,
			FieldOptions:     options,
			FieldRequired:    t.Field.Required,
			CreatedAt:        t.CreatedAt,
			UpdatedAt:        t.UpdatedAt,
		})
	}
	return tag.Tag{
		RecordType:          recordType,
		RecordURN:           recordURN,
		TemplateURN:         template.URN,
		TagValues:           listOfDomainTagValue,
		TemplateDisplayName: template.DisplayName,
		TemplateDescription: template.Description,
	}
}

func (r *TagRepository) complementDomainTag(domainTag *tag.Tag, domainTemplate tag.Template, listOfRecordModelTag []Tag) error {
	tagByFieldID := make(map[uint]Tag)
	for _, t := range listOfRecordModelTag {
		tagByFieldID[t.FieldID] = t
	}
	var listOfDomainTagValue []tag.TagValue
	for _, field := range domainTemplate.Fields {
		t := tagByFieldID[field.ID]
		parsedValue, _ := tag.ParseTagValue(domainTag.TemplateURN, field.ID, field.DataType, t.Value, field.Options)
		listOfDomainTagValue = append(listOfDomainTagValue, tag.TagValue{
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
	domainTag.TemplateURN = domainTemplate.URN
	domainTag.TemplateDescription = domainTemplate.Description
	domainTag.TemplateDisplayName = domainTemplate.DisplayName
	domainTag.TagValues = listOfDomainTagValue
	return nil
}

// func (r *TagRepository) getTemplateWithFields(templateURN string) (Template, error) {
// 	var template Template
// 	queryResult := r.dbClient.Preload("Fields").First(&template, "urn = ?", templateURN)
// 	if queryResult.Error == gorm.ErrRecordNotFound {
// 		return template, tag.TemplateNotFoundError{URN: templateURN}
// 	}

// 	return template, queryResult.Error
// }

func (r *TagRepository) readTagsByRecord(ctx context.Context, recordType, recordURN string) ([]Tag, error) {
	var tags []Tag

	rows, err := r.db.QueryxContext(ctx, readTagsByRecordSQL, recordType, recordURN)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find tags by record_type and record_urn")
	}

	for rows.Next() {
		tg := Tag{}
		rows.Scan(
			&tg.Value,
			&tg.RecordURN,
			&tg.RecordType,
			&tg.FieldID,
			&tg.CreatedAt,
			&tg.UpdatedAt,
		)
		tags = append(tags, tg)
	}

	return tags, nil
}

func (r *TagRepository) readTagsWithFields(ctx context.Context, recordType, recordURN string, domainTemplate tag.Template) ([]Tag, error) {
	var tags []Tag
	fieldsMap := make(map[uint]Field, 0) // template urn as key
	tagsMap := make(map[uint]Tag, 0)     // field id as key

	for _, field := range domainTemplate.Fields {
		rows, err := r.db.QueryxContext(ctx, readTagsWithFieldSQL, recordType, recordURN, field.ID)
		if err != nil {
			// TODO: might need to log something here
			continue
		}

		for rows.Next() {
			tg := Tag{}
			fld := Field{}
			rows.Scan(
				&tg.Value,
				&tg.RecordURN,
				&tg.RecordType,
				&tg.FieldID,
				&tg.CreatedAt,
				&tg.UpdatedAt,
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
			if _, ok := tagsMap[tg.ID]; !ok {
				tagsMap[tg.ID] = tg
			}
			if _, ok := fieldsMap[fld.ID]; !ok {
				fieldsMap[fld.ID] = fld
			}
		}
	}

	for idx, tag := range tagsMap {
		tag.Field = fieldsMap[tag.FieldID]
		tagsMap[idx] = tag
	}

	for _, tag := range tags {
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *TagRepository) insertDomainTagToDB(ctx context.Context, domainTag *tag.Tag, timestamp time.Time) (newModelTags []Tag, err error) {
	tx, txErr := r.db.BeginTxx(ctx, nil)
	if txErr != nil {
		err = errors.Wrap(txErr, "failed to create db transaction when creating tag")
		return
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.Wrap(err, "create domain tag error")
				err = errors.Wrap(txErr, "error during rollback")
			}
		} else {
			txErr := tx.Commit()
			if txErr != nil {
				err = errors.Wrap(txErr, "error during commit")
			}
		}
	}()

	for _, tv := range domainTag.TagValues {
		var insertedTagValue Tag
		if tv.FieldValue == nil {
			continue
		}
		modelTag := Tag{
			RecordType: domainTag.RecordType,
			RecordURN:  domainTag.RecordURN,
			FieldID:    tv.FieldID,
			Value:      fmt.Sprintf("%v", tv.FieldValue),
		}
		txErr = tx.QueryRowxContext(ctx, insertTagSQL,
			modelTag.Value,
			modelTag.RecordURN,
			modelTag.RecordType,
			modelTag.FieldID,
			timestamp,
			timestamp).StructScan(&insertedTagValue)
		if txErr != nil {
			err = errors.Wrap(txErr, "failed to insert a domain tag")
			return
		}

		newModelTags = append(newModelTags, insertedTagValue)
	}
	return
}

// NewTagRepository initializes tag repository
func NewTagRepository(db *sqlx.DB) *TagRepository {
	if db == nil {
		panic(errNilDBClient.Error())
	}
	return &TagRepository{
		db: db,
	}
}
