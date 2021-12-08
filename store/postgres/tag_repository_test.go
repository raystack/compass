package postgres_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TagRepositoryTestSuite struct {
	suite.Suite
	dbClient   *gorm.DB
	repository *postgres.TagRepository
}

func (r *TagRepositoryTestSuite) TestNewTagRepository() {
	r.Run("should return repository and nil if db client is not nil", func() {
		dbClient := &gorm.DB{}

		actualTagRepository := postgres.NewTagRepository(dbClient)

		r.NotNil(actualTagRepository)
	})
}

func (r *TagRepositoryTestSuite) Setup() {
	r.dbClient, _ = newTestClient("file::memory:")
	r.dbClient.AutoMigrate(&postgres.Template{})
	r.dbClient.AutoMigrate(&postgres.Field{})
	r.dbClient.AutoMigrate(&postgres.Tag{})
	repository := postgres.NewTagRepository(r.dbClient)
	r.repository = repository
}

func (r *TagRepositoryTestSuite) TestCreate() {
	r.Run("should return error if domain tag is nil", func() {
		r.Setup()
		var domainTag *tag.Tag = nil

		expectedErrorMsg := "domain tag is nil"

		actualError := r.repository.Create(domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTag := &tag.Tag{}
		repository := &postgres.TagRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Create(domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if template is not found", func() {
		r.Setup()
		t := r.getDomainTag()

		err := r.repository.Create(&t)
		r.ErrorIs(err, tag.TemplateNotFoundError{URN: t.TemplateURN})
	})

	r.Run("should return nil and create tag if no error found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		domainTag := r.getDomainTag()

		actualError := r.repository.Create(&domainTag)
		r.NoError(actualError)

		for _, value := range domainTag.TagValues {
			var actualRecord postgres.Tag
			queryResult := r.dbClient.Where(postgres.Tag{
				RecordURN: domainTag.RecordURN,
				FieldID:   value.FieldID,
			}).First(&actualRecord)

			r.NoError(queryResult.Error)
		}
	})

	r.Run("should return nil and update domain tag if no error found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		domainTag := r.getDomainTag()

		actualError := r.repository.Create(&domainTag)
		r.NoError(actualError)

		for _, value := range domainTag.TagValues {
			r.NotZero(value.CreatedAt)
		}
	})
}

func (r *TagRepositoryTestSuite) TestRead() {
	r.Run("should return nil and error if db client is nil", func() {
		var recordURN string = "sample-urn"
		repository := &postgres.TagRepository{}
		paramDomainTag := tag.Tag{
			RecordURN: recordURN,
		}

		expectedErrorMsg := "db client is nil"

		actualTag, actualError := repository.Read(paramDomainTag)

		r.Nil(actualTag)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record type is empty", func() {
		r.Setup()
		paramDomainTag := tag.Tag{
			RecordType: "",
			RecordURN:  "sample-urn",
		}

		expectedErrorMsg := "record type should not be empty"

		actualTag, actualError := r.repository.Read(paramDomainTag)

		r.Nil(actualTag)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return nil and error if record urn is empty", func() {
		r.Setup()
		var recordURN string = ""
		paramDomainTag := tag.Tag{
			RecordType: "sample-type",
			RecordURN:  recordURN,
		}

		expectedErrorMsg := "record urn should not be empty"

		actualTag, actualError := r.repository.Read(paramDomainTag)

		r.Nil(actualTag)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return empty and nil if no record found for the specified record", func() {
		r.Setup()
		paramDomainTag := tag.Tag{
			RecordType: "sample-type",
			RecordURN:  "sample-urn",
		}

		actualTag, actualError := r.repository.Read(paramDomainTag)
		r.NoError(actualError)
		r.Empty(actualTag)
	})

	r.Run("should return record if found for the specified record", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		domainTag := r.getDomainTag()
		err := r.repository.Create(&domainTag)
		r.Require().NoError(err)

		tags, err := r.repository.Read(tag.Tag{
			RecordType: domainTag.RecordType,
			RecordURN:  domainTag.RecordURN,
		})
		r.NoError(err)
		r.NotEmpty(tags)
	})

	r.Run("should return nil and error if template urn is not empty but template is not found", func() {
		r.Setup()
		paramDomainTag := tag.Tag{
			RecordURN:   "sample-urn",
			RecordType:  "sample-type",
			TemplateURN: "governance_policy",
		}

		_, err := r.repository.Read(paramDomainTag)
		r.ErrorIs(err, tag.TemplateNotFoundError{URN: "governance_policy"})
	})

	r.Run("should return nil and not found error if no record found for the specified record and template", func() {
		r.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "governance_policy"
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		paramDomainTag := tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}

		expectedErrorMsg := fmt.Sprintf("could not find tag with record type: \"%s\", record: \"%s\", template: \"%s\"",
			recordType, recordURN, templateURN,
		)

		actualTag, actualError := r.repository.Read(paramDomainTag)
		r.ErrorAs(actualError, new(tag.NotFoundError))
		r.EqualError(actualError, expectedErrorMsg)
		r.Nil(actualTag)
	})

	r.Run("should return maximum of one domain tag for the specified record and template", func() {
		r.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "governance_policy"
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		domainTag := r.getDomainTag()
		if err := r.repository.Create(&domainTag); err != nil {
			panic(err)
		}
		paramDomainTag := tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}

		expectedLength := 1

		actualTag, actualError := r.repository.Read(paramDomainTag)

		r.Len(actualTag, expectedLength)
		r.NoError(actualError)
	})
}

func (r *TagRepositoryTestSuite) TestUpdate() {
	r.Run("should return error if domain tag is nil", func() {
		r.Setup()
		var domainTag *tag.Tag = nil

		expectedErrorMsg := "domain tag is nil"

		actualError := r.repository.Update(domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTag := &tag.Tag{}
		repository := &postgres.TagRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Update(domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if template is not found", func() {
		r.Setup()
		t := r.getDomainTag()

		err := r.repository.Update(&t)
		r.ErrorIs(err, tag.TemplateNotFoundError{URN: t.TemplateURN})
	})

	r.Run("should return nil and update tag if no error found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		domainTag := r.getDomainTag()
		err := r.repository.Create(&domainTag)
		r.Require().NoError(err)

		domainTag.TagValues[0].FieldValue = "Restricted"
		actualError := r.repository.Update(&domainTag)
		r.Require().NoError(actualError)

		for _, value := range domainTag.TagValues {
			var actualRecord postgres.Tag
			queryResult := r.dbClient.Where(
				"record_urn = ? and field_id = ?", domainTag.RecordURN, value.FieldID,
			).First(&actualRecord)

			r.NoError(queryResult.Error)
			r.EqualValues(value.UpdatedAt, actualRecord.UpdatedAt)
		}
	})

	r.Run("should return nil and update domain tag if no error found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		domainTag := r.getDomainTag()
		if err := r.repository.Create(&domainTag); err != nil {
			panic(err)
		}
		domainTag.TagValues = domainTag.TagValues[:1]

		actualError := r.repository.Update(&domainTag)

		r.NoError(actualError)
		r.Len(domainTag.TagValues, 2)
	})
}

func (r *TagRepositoryTestSuite) TestDelete() {
	r.Run("should return error if db client is nil", func() {
		var recordURN string = "sample-urn"
		repository := &postgres.TagRepository{}
		paramDomainTag := tag.Tag{
			RecordURN: recordURN,
		}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Delete(paramDomainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record urn is empty", func() {
		r.Setup()
		var recordURN string = ""
		paramDomainTag := tag.Tag{
			RecordURN: recordURN,
		}

		expectedErrorMsg := "record urn should not be empty"

		actualError := r.repository.Delete(paramDomainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should delete tags related to the record and return error if record has one", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		domainTag := r.getDomainTag()
		if err := r.repository.Create(&domainTag); err != nil {
			r.T().Fatal(err)
		}
		paramDomainTag := tag.Tag{
			RecordType: domainTag.RecordType,
			RecordURN:  domainTag.RecordURN,
		}

		actualError := r.repository.Delete(paramDomainTag)
		var listOfRecordModelTag []postgres.Tag
		err := r.dbClient.Where("record_type = ? and record_urn = ?", domainTag.RecordType, domainTag.RecordURN).
			Find(&listOfRecordModelTag).Error
		if err != nil {
			r.T().Fatal(err)
		}

		r.NoError(actualError)
		r.Empty(listOfRecordModelTag)
	})

	r.Run("should return error if template is not found", func() {
		r.Setup()
		var recordURN string = "sample-urn"
		var templateURN string = "random-urn"
		paramDomainTag := tag.Tag{
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}

		err := r.repository.Delete(paramDomainTag)
		r.ErrorIs(err, tag.TemplateNotFoundError{URN: templateURN})
	})

	r.Run("should delete only the tag for record and template and return nil if record has one", func() {
		r.Setup()
		var recordURN string = "sample-urn"
		domainTemplate := r.getDomainTemplate()
		r.createDomainTemplate(domainTemplate)
		paramDomainTag := tag.Tag{
			RecordURN:   recordURN,
			TemplateURN: domainTemplate.URN,
		}

		actualError := r.repository.Delete(paramDomainTag)
		var listOfRecordModelTag []postgres.Tag
		err := r.dbClient.Where("record_urn = ?", recordURN).Find(&listOfRecordModelTag).Error
		if err != nil {
			panic(err)
		}

		r.NoError(actualError)
		r.Empty(listOfRecordModelTag)
	})
}

func (r *TagRepositoryTestSuite) createDomainTemplate(domainTemplate *tag.Template) {
	listOfModelField := make([]postgres.Field, len(domainTemplate.Fields))
	for i, domainField := range domainTemplate.Fields {
		var options *string
		if len(domainField.Options) > 0 {
			var concatenatedOption string
			strings.Join(domainField.Options, ",")
			options = &concatenatedOption
		}
		listOfModelField[i] = postgres.Field{
			ID:          domainField.ID,
			URN:         domainField.URN,
			DisplayName: domainField.DisplayName,
			Description: domainField.Description,
			DataType:    domainField.DataType,
			Options:     options,
			Required:    domainField.Required,
		}
	}
	modelTemplate := postgres.Template{
		URN:         domainTemplate.URN,
		DisplayName: domainTemplate.DisplayName,
		Description: domainTemplate.Description,
		Fields:      listOfModelField,
	}
	createResult := r.dbClient.Create(&modelTemplate)
	if createResult.Error != nil {
		r.T().Fatal(createResult.Error)
	}
}

func (r *TagRepositoryTestSuite) getDomainTemplate() *tag.Template {
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

func (r *TagRepositoryTestSuite) getDomainTag() tag.Tag {
	return tag.Tag{
		RecordType:          "sample-type",
		RecordURN:           "sample-urn",
		TemplateURN:         "governance_policy",
		TemplateDisplayName: "Governance Policy",
		TemplateDescription: "Template that is mandatory to be used.",
		TagValues: []tag.TagValue{
			{
				FieldID:          1,
				FieldValue:       "Public",
				FieldURN:         "classification",
				FieldDisplayName: "classification",
				FieldDescription: "The classification of this record",
				FieldDataType:    "enumerated",
				FieldRequired:    true,
				FieldOptions:     []string{"Public", "Restricted"},
			},
			{
				FieldID:          2,
				FieldValue:       "dexter@odpf.io",
				FieldURN:         "admin_email",
				FieldDisplayName: "Admin Email",
				FieldDescription: "Email of the admin of therecord.",
				FieldDataType:    "string",
				FieldRequired:    true,
			},
		},
	}
}

func TestTagRepository(t *testing.T) {
	suite.Run(t, &TagRepositoryTestSuite{})
}
