package postgres_test

import (
	"context"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type TemplateRepositoryTestSuite struct {
	suite.Suite
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func (r *TemplateRepositoryTestSuite) SetupSuite() {
	var err error
	logger := logrus.New()
	// logger.SetLevel(logrus.DebugLevel)
	r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *TemplateRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	err := testDBClient.Close()
	err = purgeClient(r.pool, r.resource)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *TemplateRepositoryTestSuite) TestNewRepository() {
	r.Run("should return repository and nil if db client is not nil", func() {
		dummyDBClient := &sqlx.DB{}

		actualRepository := postgres.NewTemplateRepository(dummyDBClient)
		r.NotNil(actualRepository)
	})
}

func (r *TemplateRepositoryTestSuite) TestCreate() {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	r.Run("should return error if domain template is nil", func() {
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := repository.Create(ctx, domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTemplate := tag.Template{
			URN: "governance_policy",
		}
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Create(ctx, &domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return nil and insert new record if no error found", func() {
		domainTemplate := r.getDomainTemplate()

		err := repository.Create(ctx, &domainTemplate)
		r.NoError(err)

		var actualRecord tag.Template
		templates, err := repository.Read(ctx, domainTemplate)
		r.Nil(err)

		actualRecord = templates[0]
		r.Equal(domainTemplate.URN, actualRecord.URN)
		r.Equal(domainTemplate.DisplayName, actualRecord.DisplayName)
		r.Equal(domainTemplate.Description, actualRecord.Description)
		r.Equal(len(domainTemplate.Fields), len(actualRecord.Fields))
		r.Equal(domainTemplate.Fields[0].DisplayName, actualRecord.Fields[0].DisplayName)
		r.Equal(domainTemplate.Fields[0].URN, actualRecord.Fields[0].URN)
	})

	r.Run("should return error if encountered uncovered error", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()

		err = repository.Create(ctx, &domainTemplate)
		err = repository.Create(ctx, &domainTemplate)

		r.EqualError(errors.Unwrap(err), "failed to insert a template: ERROR: duplicate key value violates unique constraint \"templates_pkey\" (SQLSTATE 23505)")
	})
}

func (r *TemplateRepositoryTestSuite) TestRead() {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	r.Run("should return nil and error if db client is nil", func() {
		domainTemplate := r.getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualTemplate, actualError := repository.Read(ctx, domainTemplate)

		r.Nil(actualTemplate)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return empty and nil if no record found", func() {
		err := setup()
		r.NoError(err)
		domainTemplate := r.getDomainTemplate()

		actualTemplate, actualError := repository.Read(ctx, domainTemplate)

		r.Empty(actualTemplate)
		r.NoError(actualError)
	})

	r.Run("should return template with multiple fields if exist", func() {
		err := setup()
		r.NoError(err)
		domainTemplate := r.getDomainTemplate()

		domainTemplate.DisplayName = "Random Display"
		domainTemplate.Fields[0].DisplayName = "Another Random Display"
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		err = repository.Create(ctx, &domainTemplate)
		r.NoError(err)

		templates, err := repository.Read(ctx, domainTemplate)

		r.NoError(err)
		r.Len(templates[0].Fields, 2)
		r.Equal(domainTemplate.DisplayName, templates[0].DisplayName)
		r.Equal(domainTemplate.UpdatedAt, templates[0].UpdatedAt)
	})
}

func (r *TemplateRepositoryTestSuite) TestUpdate() {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	r.Run("should return error if domain template is nil", func() {
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := repository.Update(ctx, domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTemplate := r.getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Update(ctx, &domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record not found", func() {
		err := setup()
		r.NoError(err)
		domainTemplate := r.getDomainTemplate()

		actualError := repository.Update(ctx, &domainTemplate)

		r.Error(actualError)
	})

	r.Run("should return nil and updated domain template if update is success", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()
		err = repository.Create(ctx, &domainTemplate)
		r.NoError(err)

		domainTemplate.DisplayName = "Random Display"
		domainTemplate.Fields[0].DisplayName = "Another Random Display"
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		actualError := repository.Update(ctx, &domainTemplate)
		r.NoError(actualError)

		templates, err := repository.Read(ctx, domainTemplate)
		r.NoError(err)
		recordModelTemplate := templates[0]

		r.Len(recordModelTemplate.Fields, 2)

		r.Equal(domainTemplate.DisplayName, recordModelTemplate.DisplayName)
		r.Equal(domainTemplate.UpdatedAt, recordModelTemplate.UpdatedAt)

		reflect.DeepEqual(domainTemplate.Fields, recordModelTemplate.Fields)
	})

	r.Run("should return error if trying to update with unrelated field", func() {
		err := setup()
		r.NoError(err)
		domainTemplate := r.getDomainTemplate()
		if err := repository.Create(ctx, &domainTemplate); err != nil {
			panic(err)
		}
		domainTemplate.Fields[0].ID = 2

		actualError := repository.Update(ctx, &domainTemplate)

		r.EqualError(actualError, "error updating template: failed updating fields: sql: no rows in result set")
	})

	r.Run("should return error if trying to update with duplicated field", func() {
		err := setup()
		r.NoError(err)
		domainTemplate := r.getDomainTemplate()
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "second_field",
			DisplayName: "Second Field",
			Description: "Random description for the second field.",
			DataType:    "string",
			Required:    false,
		})

		if err := repository.Create(ctx, &domainTemplate); err != nil {
			panic(err)
		}

		domainTemplate.Fields[1].URN = domainTemplate.Fields[0].URN

		actualError := repository.Update(ctx, &domainTemplate)

		r.EqualError(actualError, "error updating template: failed updating fields: ERROR: duplicate key value violates unique constraint \"fields_idx_urn_template_urn\" (SQLSTATE 23505)")
	})
}

func (r *TemplateRepositoryTestSuite) TestDelete() {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	r.Run("should return error if db client is nil", func() {
		domainTemplate := r.getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Delete(ctx, domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record not found", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()

		err = repository.Delete(ctx, domainTemplate)
		r.EqualError(err, "error deleting template: could not find template \"governance_policy\"")
	})

	r.Run("should return nil if record is deleted", func() {
		err := setup()
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()
		if err := repository.Create(ctx, &domainTemplate); err != nil {
			panic(err)
		}

		err = repository.Delete(ctx, domainTemplate)
		r.NoError(err)

		templates, err := repository.Read(ctx, domainTemplate)
		r.NoError(err)

		r.Empty(templates)
	})
}

func (r *TemplateRepositoryTestSuite) getDomainTemplate() tag.Template {
	return tag.Template{
		URN:         "governance_policy",
		DisplayName: "Governance Policy",
		Description: "Template that is mandatory to be used.",
		Fields: []tag.Field{
			{
				ID:          1,
				URN:         "team_owner",
				DisplayName: "Team Owner",
				Description: "Owner of the resource.",
				DataType:    "enumerated",
				Required:    true,
				Options:     []string{"PIC", "Escalated"},
			},
		},
	}
}

func (r *TemplateRepositoryTestSuite) updateTimeForDomainTemplate(domainTemplate *tag.Template, t time.Time) {
	domainTemplate.CreatedAt = t
	domainTemplate.UpdatedAt = t
	for i := 0; i < len(domainTemplate.Fields); i++ {
		domainTemplate.Fields[i].CreatedAt = t
		domainTemplate.Fields[i].UpdatedAt = t
	}
}

func TestTemplateRepository(t *testing.T) {
	suite.Run(t, &TemplateRepositoryTestSuite{})
}
