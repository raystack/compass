package postgres

import (
	"sort"
	"testing"

	"github.com/odpf/columbus/tag"
	"github.com/stretchr/testify/assert"
)

func TestModelTag(t *testing.T) {

	recordType := "sample-type"
	recordURN := "sample-urn"

	templates := getModelTemplates()
	tags := getModelTags()

	t.Run("successfully build map of tags by template URN", func(t *testing.T) {
		expectedTagsMap := map[string][]Tag{
			"governance_policy": {tags[0], tags[1]},
		}
		tagsMap := tags.buildMapByTemplateURN()

		assert.EqualValues(t, expectedTagsMap, tagsMap)

	})
	t.Run("successfully build domain tags from model tags", func(t *testing.T) {

		expectedDomainTags := []tag.Tag{
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
						FieldDescription: "Email of the admin of therecord.",
						FieldDataType:    "string",
						FieldRequired:    true,
					}},
				TemplateDisplayName: "Governance Policy",
				TemplateDescription: "Template that is mandatory to be used.",
			}}

		actualDomainTags := tags.toDomainTags(recordType, recordURN, templates)

		assert.EqualValues(t, expectedDomainTags, actualDomainTags)
	})
}

func TestModelTemplate(t *testing.T) {

	templates := getModelTemplates()

	t.Run("successfully convert model template to domain template", func(t *testing.T) {
		expectedDomainTemplate := tag.Template{
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
					Description: "Email of the admin of therecord.",
					DataType:    "string",
					Required:    true,
				},
			},
		}
		actualDomainTemplate := templates[0].toDomainTemplate()

		assert.EqualValues(t, expectedDomainTemplate, actualDomainTemplate)
	})

	t.Run("successfully build from model template from domain template", func(t *testing.T) {

		domainTemplate := getDomainTemplate()
		option := "Public,Restricted"
		expectedModelTemplate := Template{
			URN:         "governance_policy",
			DisplayName: "Governance Policy",
			Description: "Template that is mandatory to be used.",
			Fields: Fields{
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
					Description: "Email of the admin of therecord.",
					DataType:    "string",
					Required:    true,
				},
			},
		}

		modelTemplate := Template{}
		modelTemplate.buildFromDomainTemplate(*domainTemplate)

		assert.EqualValues(t, expectedModelTemplate, modelTemplate)
	})
}

func TestModelFields(t *testing.T) {
	modelFields := getModelFields()
	t.Run("successfully convert model fields to domain fields", func(t *testing.T) {
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
				Description: "Email of the admin of therecord.",
				DataType:    "string",
				Required:    true,
			},
		}
		actualDomainFields := modelFields.toDomainFields()

		assert.EqualValues(t, expectedDomainFields, actualDomainFields)
	})

	t.Run("successfully build model fields from domain fields", func(t *testing.T) {
		domainFields := modelFields.toDomainFields()
		option := "Public,Restricted"
		expectedModelFields := &Fields{
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
				Description: "Email of the admin of therecord.",
				DataType:    "string",
				Required:    true,
			},
		}

		actualModelFields := &Fields{}
		actualModelFields.buildFromDomainFields(domainFields)

		assert.EqualValues(t, expectedModelFields, actualModelFields)
	})

	t.Run("return true if ID exist in fields", func(t *testing.T) {
		assert.True(t, modelFields.isIDExist(1))
	})
	t.Run("return false if ID exist in fields", func(t *testing.T) {
		assert.False(t, modelFields.isIDExist(100))
	})
}

func TestTemplateFields(t *testing.T) {
	tfs := getTemplateFields()
	modelTemplates := getModelTemplates()
	domainTemplates := getDomainTemplate()
	t.Run("successfully build model templates", func(t *testing.T) {

		actualModelTemplates := tfs.toModelTemplates()

		assert.EqualValues(t, modelTemplates[0], actualModelTemplates[0])
	})
	t.Run("successfully build domain templates", func(t *testing.T) {
		actualDomainTemplates := tfs.toDomainTemplates()

		assert.EqualValues(t, *domainTemplates, actualDomainTemplates[0])
	})
}

func TestTemplateTagFields(t *testing.T) {
	ttfs := getTemplateTagFields()
	t.Run("successfully build model templates and tags", func(t *testing.T) {

		expectedModelTemplates := getModelTemplates()
		expectedModelTags := getModelTags()
		actualModelTemplates, actualModelTags := ttfs.toModelTemplatesAndTags()
		assert.EqualValues(t, expectedModelTemplates[0], actualModelTemplates[0])

		sort.Slice(actualModelTags[:], func(i, j int) bool {
			return actualModelTags[i].ID < actualModelTags[j].ID
		})
		sort.Slice(expectedModelTags[:], func(i, j int) bool {
			return expectedModelTags[i].ID < expectedModelTags[j].ID
		})
		assert.EqualValues(t, expectedModelTags[0], actualModelTags[0])
	})
}

func getModelFields() Fields {
	option := "Public,Restricted"
	return Fields{
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
			Description: "Email of the admin of therecord.",
			DataType:    "string",
			Options:     nil,
			Required:    true,
			TemplateURN: "governance_policy",
		},
	}
}

func getDomainTemplate() *tag.Template {
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
				Description: "Email of the admin of therecord.",
				DataType:    "string",
				Required:    true,
			},
		},
	}
}

func getModelTemplates() Templates {
	fields := getModelFields()
	return Templates{
		{
			URN:         "governance_policy",
			DisplayName: "Governance Policy",
			Description: "Template that is mandatory to be used.",
			Fields:      fields,
		},
	}
}

func getModelTags() Tags {
	fields := getModelFields()
	return Tags{
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

func getTemplateFields() TemplateFields {
	templates := getModelTemplates()
	return TemplateFields{
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

func getTemplateTagFields() TemplateTagFields {
	templates := getModelTemplates()
	tags := getModelTags()
	return TemplateTagFields{
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
