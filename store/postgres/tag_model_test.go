package postgres

import (
	"sort"
	"testing"

	"github.com/odpf/compass/tag"
	"github.com/stretchr/testify/assert"
)

func TestTagModel(t *testing.T) {

	recordType := "sample-type"
	recordURN := "sample-urn"

	templates := getTemplateModels()
	tags := getTagModels()

	t.Run("successfully build map of tags model by template URN", func(t *testing.T) {
		expectedTagsMap := map[string][]TagModel{
			"governance_policy": {tags[0], tags[1]},
		}
		tagsMap := tags.buildMapByTemplateURN()

		assert.EqualValues(t, expectedTagsMap, tagsMap)

	})
	t.Run("successfully build tags from tags model", func(t *testing.T) {

		expectedTagDomains := []tag.Tag{
			{
				RecordType:  "sample-type",
				RecordURN:   "sample-urn",
				TemplateURN: "governance_policy",
				TagValues: []tag.TagValue{
					{
						FieldID:          1,
						FieldValue:       "Public",
						FieldURN:         "classification",
						FieldDisplayName: "classification",
						FieldDescription: "The classification of this record",
						FieldDataType:    "enumerated",
						FieldOptions:     []string{"Public", "Restricted"},
						FieldRequired:    true,
					}, {
						FieldID:          2,
						FieldValue:       "dexter@odpf.io",
						FieldURN:         "admin_email",
						FieldDisplayName: "Admin Email",
						FieldDescription: "Email of the admin of theasset.",
						FieldDataType:    "string",
						FieldRequired:    true,
					}},
				TemplateDisplayName: "Governance Policy",
				TemplateDescription: "Template that is mandatory to be used.",
			}}

		actualTagDomains := tags.toTags(recordType, recordURN, templates)

		assert.EqualValues(t, expectedTagDomains, actualTagDomains)
	})
}

func TestTemplateModel(t *testing.T) {

	templates := getTemplateModels()

	t.Run("successfully convert template model to template", func(t *testing.T) {
		expectedTemplate := tag.Template{
			URN:         "governance_policy",
			DisplayName: "Governance Policy",
			Description: "Template that is mandatory to be used.",
			Fields: []tag.Field{
				{
					ID:          1,
					URN:         "classification",
					DisplayName: "classification",
					Description: "The classification of this record",
					DataType:    "enumerated",
					Options:     []string{"Public", "Restricted"},
					Required:    true,
				},
				{
					ID:          2,
					URN:         "admin_email",
					DisplayName: "Admin Email",
					Description: "Email of the admin of theasset.",
					DataType:    "string",
					Required:    true,
				},
			},
		}
		actualTemplate := templates[0].toTemplate()

		assert.EqualValues(t, expectedTemplate, actualTemplate)
	})

	t.Run("successfully build from template model from template", func(t *testing.T) {

		template := getTemplate()
		option := "Public,Restricted"
		expectedTemplateModel := &TagTemplateModel{
			URN:         "governance_policy",
			DisplayName: "Governance Policy",
			Description: "Template that is mandatory to be used.",
			Fields: TagTemplateFieldModels{
				{
					ID:          1,
					URN:         "classification",
					DisplayName: "classification",
					Description: "The classification of this record",
					DataType:    "enumerated",
					Options:     &option,
					Required:    true,
				},
				{
					ID:          0x2,
					URN:         "admin_email",
					DisplayName: "Admin Email",
					Description: "Email of the admin of theasset.",
					DataType:    "string",
					Required:    true,
				},
			},
		}

		templateModel := newTemplateModel(template)

		assert.EqualValues(t, expectedTemplateModel, templateModel)
	})
}

func TestFieldModels(t *testing.T) {
	fieldModels := getFieldModels()
	t.Run("successfully convert fields model to fields", func(t *testing.T) {
		expectedDomainFields := []tag.Field{
			{
				ID:          1,
				URN:         "classification",
				DisplayName: "classification",
				Description: "The classification of this record",
				DataType:    "enumerated",
				Options:     []string{"Public", "Restricted"},
				Required:    true,
			},
			{
				ID:          2,
				URN:         "admin_email",
				DisplayName: "Admin Email",
				Description: "Email of the admin of theasset.",
				DataType:    "string",
				Required:    true,
			},
		}
		actualDomainFields := fieldModels.toDomainFields()

		assert.EqualValues(t, expectedDomainFields, actualDomainFields)
	})

	t.Run("successfully build fields model from fields", func(t *testing.T) {
		domainFields := fieldModels.toDomainFields()
		option := "Public,Restricted"
		expectedFieldModels := TagTemplateFieldModels{
			{
				ID:          1,
				URN:         "classification",
				DisplayName: "classification",
				Description: "The classification of this record",
				DataType:    "enumerated",
				Options:     &option,
				Required:    true,
			},
			{
				ID:          2,
				URN:         "admin_email",
				DisplayName: "Admin Email",
				Description: "Email of the admin of theasset.",
				DataType:    "string",
				Required:    true,
			},
		}

		actualFieldModels := newSliceOfFieldModel(domainFields)

		assert.EqualValues(t, expectedFieldModels, actualFieldModels)
	})

	t.Run("return true if ID exist in slice of fields model", func(t *testing.T) {
		assert.True(t, fieldModels.isIDExist(1))
	})
	t.Run("return false if ID exist in slice of field model", func(t *testing.T) {
		assert.False(t, fieldModels.isIDExist(100))
	})
}

func TestTemplateFields(t *testing.T) {
	tfs := getTemplateFieldModels()
	templateModels := getTemplateModels()
	templates := getTemplate()
	t.Run("successfully build template models", func(t *testing.T) {

		actualTemplateModels := tfs.toTemplateModels()

		assert.EqualValues(t, templateModels[0], actualTemplateModels[0])
	})
	t.Run("successfully build templates", func(t *testing.T) {
		actualTemplates := tfs.toTemplates()

		assert.EqualValues(t, *templates, actualTemplates[0])
	})
}

func TestTemplateTagFields(t *testing.T) {
	ttfs := getTemplateTagFieldModels()
	t.Run("successfully build templates model and tags model", func(t *testing.T) {

		expectedTemplateModels := getTemplateModels()
		expectedTagModels := getTagModels()
		actualTemplateModels, actualTagModels := ttfs.toTemplateAndTagModels()
		assert.EqualValues(t, expectedTemplateModels[0], actualTemplateModels[0])

		sort.Slice(actualTagModels[:], func(i, j int) bool {
			return actualTagModels[i].ID < actualTagModels[j].ID
		})
		sort.Slice(expectedTagModels[:], func(i, j int) bool {
			return expectedTagModels[i].ID < expectedTagModels[j].ID
		})
		assert.EqualValues(t, expectedTagModels[0], actualTagModels[0])
	})
}

func getFieldModels() TagTemplateFieldModels {
	option := "Public,Restricted"
	return TagTemplateFieldModels{
		{
			ID:          1,
			URN:         "classification",
			DisplayName: "classification",
			Description: "The classification of this record",
			DataType:    "enumerated",
			Options:     &option,
			Required:    true,
			TemplateURN: "governance_policy",
		},
		{
			ID:          2,
			URN:         "admin_email",
			DisplayName: "Admin Email",
			Description: "Email of the admin of theasset.",
			DataType:    "string",
			Options:     nil,
			Required:    true,
			TemplateURN: "governance_policy",
		},
	}
}

func getTemplate() *tag.Template {
	return &tag.Template{
		URN:         "governance_policy",
		DisplayName: "Governance Policy",
		Description: "Template that is mandatory to be used.",
		Fields: []tag.Field{
			{
				ID:          1,
				URN:         "classification",
				DisplayName: "classification",
				Description: "The classification of this record",
				DataType:    "enumerated",
				Required:    true,
				Options:     []string{"Public", "Restricted"},
			},
			{
				ID:          2,
				URN:         "admin_email",
				DisplayName: "Admin Email",
				Description: "Email of the admin of theasset.",
				DataType:    "string",
				Required:    true,
			},
		},
	}
}

func getTemplateModels() TagTemplateModels {
	fields := getFieldModels()
	return TagTemplateModels{
		{
			URN:         "governance_policy",
			DisplayName: "Governance Policy",
			Description: "Template that is mandatory to be used.",
			Fields:      fields,
		},
	}
}

func getTagModels() TagModels {
	fields := getFieldModels()
	return TagModels{
		{
			ID:         1,
			Value:      "Public",
			RecordURN:  "sample-urn",
			RecordType: "sample-type",
			FieldID:    1,
			Field:      fields[0],
		},
		{
			ID:         2,
			Value:      "dexter@odpf.io",
			RecordURN:  "sample-urn",
			RecordType: "sample-type",
			FieldID:    2,
			Field:      fields[1],
		},
	}
}

func getTemplateFieldModels() TagJoinTemplateFieldModels {
	templates := getTemplateModels()
	return TagJoinTemplateFieldModels{
		{
			Template: templates[0],
			Field:    templates[0].Fields[0],
		},
		{
			Template: templates[0],
			Field:    templates[0].Fields[1],
		},
	}
}

func getTemplateTagFieldModels() TagJoinTemplateTagFieldModels {
	templates := getTemplateModels()
	tags := getTagModels()
	return TagJoinTemplateTagFieldModels{
		{
			Template: templates[0],
			Tag:      tags[0],
			Field:    templates[0].Fields[0],
		},
		{
			Template: templates[0],
			Tag:      tags[1],
			Field:    templates[0].Fields[1],
		},
	}
}
