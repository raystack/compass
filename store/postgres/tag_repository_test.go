package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type TagRepositoryTestSuite struct {
	suite.Suite
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func (r *TagRepositoryTestSuite) SetupSuite() {
	var err error
	logger := logrus.New()
	// logger.SetLevel(logrus.DebugLevel)
	r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *TagRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	err := testDBClient.Close()
	err = purgeClient(r.pool, r.resource)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *TagRepositoryTestSuite) TestNewRepository() {
	r.Run("should return repository and nil if db client is not nil", func() {
		dbClient := &sqlx.DB{}

		actualTagRepository := postgres.NewTagRepository(dbClient)

		r.NotNil(actualTagRepository)
	})
}

func (r *TagRepositoryTestSuite) TestCreate() {
	ctx := context.TODO()
	repository := postgres.NewTagRepository(testDBClient)

	r.Run("should return error if domain tag is nil", func() {
		err := setup()
		r.NoError(err)

		var domainTag *tag.Tag = nil

		expectedErrorMsg := "domain tag is nil"

		actualError := repository.Create(ctx, domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTag := &tag.Tag{}
		repository := &postgres.TagRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Create(ctx, domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if template is not found", func() {
		err := setup()
		r.NoError(err)
		domain := getDomainTag()

		err = repository.Create(ctx, &domain)

		r.EqualError(err, tag.TemplateNotFoundError{URN: domain.TemplateURN}.Error())
	})

	r.Run("should return nil and create tag if no error found", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		err = repository.Create(ctx, &domainTag)
		r.NoError(err)

		tags, err := repository.Read(ctx, domainTag)
		r.NoError(err)
		r.NotNil(tags)
	})

	r.Run("should return nil and update domain tag if no error found", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		err = repository.Create(ctx, &domainTag)
		r.NoError(err)

		for _, value := range domainTag.TagValues {
			r.NotZero(value.CreatedAt)
		}
	})
}

func (r *TagRepositoryTestSuite) TestRead() {
	ctx := context.TODO()
	repository := postgres.NewTagRepository(testDBClient)

	r.Run("should return nil and error if db client is nil", func() {
		var recordURN string = "sample-urn"
		repository := &postgres.TagRepository{}
		paramDomainTag := tag.Tag{
			RecordURN: recordURN,
		}

		expectedErrorMsg := "db client is nil"

		actualTag, actualError := repository.Read(ctx, paramDomainTag)

		r.Nil(actualTag)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record type is empty", func() {
		err := setup()
		r.NoError(err)

		paramDomainTag := tag.Tag{
			RecordType: "",
			RecordURN:  "sample-urn",
		}

		expectedErrorMsg := "record type should not be empty"

		actualTag, actualError := repository.Read(ctx, paramDomainTag)

		r.Nil(actualTag)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return nil and error if record urn is empty", func() {
		err := setup()
		r.NoError(err)

		var recordURN string = ""
		paramDomainTag := tag.Tag{
			RecordType: "sample-type",
			RecordURN:  recordURN,
		}

		expectedErrorMsg := "record urn should not be empty"

		actualTag, actualError := repository.Read(ctx, paramDomainTag)

		r.Nil(actualTag)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return empty and nil if no record found for the specified record", func() {
		err := setup()
		r.NoError(err)

		paramDomainTag := tag.Tag{
			RecordType: "sample-type",
			RecordURN:  "sample-urn",
		}

		actualTag, err := repository.Read(ctx, paramDomainTag)
		// r.NoError( actualError) //TODO: recheck this behaviour
		r.Empty(actualTag)

		r.True(errors.As(err, new(tag.NotFoundError)))
		r.EqualError(err, tag.NotFoundError{
			URN:      paramDomainTag.RecordURN,
			Type:     paramDomainTag.RecordType,
			Template: paramDomainTag.TemplateURN,
		}.Error())
	})

	r.Run("should return record if found for the specified record", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)

		domainTag := getDomainTag()
		err = repository.Create(ctx, &domainTag)
		r.NoError(err)

		tags, err := repository.Read(ctx, tag.Tag{
			RecordType: domainTag.RecordType,
			RecordURN:  domainTag.RecordURN,
		})

		r.NoError(err)
		r.NotEmpty(tags)
		r.Len(tags[0].TagValues, 2)
	})

	r.Run("should return nil and error if template urn is not empty but template is not found", func() {
		err := setup()
		r.NoError(err)

		paramDomainTag := tag.Tag{
			RecordURN:   "sample-urn",
			RecordType:  "sample-type",
			TemplateURN: "governance_policy",
		}

		_, err = repository.Read(ctx, paramDomainTag)
		r.EqualError(err, tag.NotFoundError{
			URN:      paramDomainTag.RecordURN,
			Type:     paramDomainTag.RecordType,
			Template: paramDomainTag.TemplateURN,
		}.Error())
	})

	r.Run("should return nil and not found error if no record found for the specified record and template", func() {
		err := setup()
		r.NoError(err)

		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "governance_policy"

		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)

		paramDomainTag := tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}

		expectedErrorMsg := fmt.Sprintf("could not find tag with record type: \"%s\", record: \"%s\", template: \"%s\"",
			recordType, recordURN, templateURN,
		)

		actualTag, actualError := repository.Read(ctx, paramDomainTag)
		r.EqualError(actualError, expectedErrorMsg)
		r.True(errors.As(actualError, new(tag.NotFoundError)))
		r.Nil(actualTag)
	})

	r.Run("should return maximum of one domain tag for the specified record and template", func() {
		err := setup()
		r.NoError(err)

		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "governance_policy"

		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		if err := repository.Create(ctx, &domainTag); err != nil {
			panic(err)
		}
		paramDomainTag := tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}

		expectedLength := 1

		actualTag, actualError := repository.Read(ctx, paramDomainTag)

		r.Len(actualTag, expectedLength)
		r.NoError(actualError)
	})
}

func (r *TagRepositoryTestSuite) TestUpdate() {
	ctx := context.TODO()
	repository := postgres.NewTagRepository(testDBClient)

	r.Run("should return error if domain tag is nil", func() {
		err := setup()
		r.NoError(err)

		var domainTag *tag.Tag = nil

		expectedErrorMsg := "domain tag is nil"

		actualError := repository.Update(ctx, domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTag := &tag.Tag{}
		repository := &postgres.TagRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Update(ctx, domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if template is not found", func() {
		err := setup()
		r.NoError(err)
		t := getDomainTag()

		err = repository.Update(ctx, &t)
		r.EqualError(err, tag.TemplateNotFoundError{URN: t.TemplateURN}.Error())
	})

	r.Run("should return nil and update tag if no error found", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := getDomainTemplate()
		createDomainTemplate(ctx, domainTemplate)
		domainTag := getDomainTag()
		err = repository.Create(ctx, &domainTag)
		r.Require().NoError(err)

		domainTag.TagValues[0].FieldValue = "Restricted"
		actualError := repository.Update(ctx, &domainTag)
		r.Require().NoError(actualError)

		for _, value := range domainTag.TagValues {
			tags, err := repository.FindModelTag(ctx, postgres.Tag{
				RecordURN:  domainTag.RecordURN,
				RecordType: domainTag.RecordType,
			})
			r.NoError(err)
			r.EqualValues(value.UpdatedAt, tags[0].UpdatedAt)
		}
	})

	r.Run("should return nil and update domain tag if no error found", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		if err := repository.Create(ctx, &domainTag); err != nil {
			panic(err)
		}
		domainTag.TagValues = domainTag.TagValues[:1]

		actualError := repository.Update(ctx, &domainTag)

		r.NoError(actualError)
		r.Len(domainTag.TagValues, 2)
	})
}

func (r *TagRepositoryTestSuite) TestDelete() {
	ctx := context.TODO()
	repository := postgres.NewTagRepository(testDBClient)

	r.Run("should return error if db client is nil", func() {
		var recordURN string = "sample-urn"
		repository := &postgres.TagRepository{}
		paramDomainTag := tag.Tag{
			RecordURN: recordURN,
		}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Delete(ctx, paramDomainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record urn is empty", func() {
		err := setup()
		r.NoError(err)

		var recordURN string = ""
		paramDomainTag := tag.Tag{
			RecordURN: recordURN,
		}

		expectedErrorMsg := "record urn should not be empty"

		actualError := repository.Delete(ctx, paramDomainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should delete tags related to the record and return no error if record has one", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)

		domainTag := getDomainTag()

		if err := repository.Create(ctx, &domainTag); err != nil {
			r.T().Fatal(err)
		}

		paramDomainTag := tag.Tag{
			RecordType: domainTag.RecordType,
			RecordURN:  domainTag.RecordURN,
		}

		actualError := repository.Delete(ctx, paramDomainTag)

		foundTags, err := repository.FindModelTag(ctx, postgres.Tag{
			RecordURN:  paramDomainTag.RecordURN,
			RecordType: paramDomainTag.RecordType,
		})
		if err != nil {
			r.T().Fatal(err)
		}

		r.NoError(actualError)
		r.Empty(foundTags)
	})

	r.Run("should return error if template is not found", func() {
		err := setup()
		r.NoError(err)

		var recordURN string = "sample-urn"
		var templateURN string = "random-urn"
		paramDomainTag := tag.Tag{
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}

		err = repository.Delete(ctx, paramDomainTag)
		r.EqualError(err, tag.TemplateNotFoundError{URN: templateURN}.Error())
	})

	r.Run("should delete only the tag for record and template and return nil if record has one", func() {
		err := setup()
		r.NoError(err)

		var recordURN string = "sample-urn"
		domainTemplate := getDomainTemplate()
		err = createDomainTemplate(ctx, domainTemplate)
		r.NoError(err)

		paramDomainTag := tag.Tag{
			RecordURN:   recordURN,
			TemplateURN: domainTemplate.URN,
		}

		actualError := repository.Delete(ctx, paramDomainTag)

		listOfRecordModelTag, err := repository.FindModelTag(ctx, postgres.Tag{
			RecordURN:  paramDomainTag.RecordURN,
			RecordType: paramDomainTag.RecordType,
		})
		if err != nil {
			r.T().Fatal(err)
		}

		r.NoError(actualError)
		r.Empty(listOfRecordModelTag)
	})
}

func createDomainTemplate(ctx context.Context, domainTemplate *tag.Template) error {
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
	_, err := postgres.CreateTemplateTx(ctx, testDBClient, &modelTemplate, time.Now().UTC())
	return err
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

func getDomainTag() tag.Tag {
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
