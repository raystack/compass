package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/odpf/compass/core/tag"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type TagTemplateRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	repository *postgres.TagTemplateRepository
	pool       *dockertest.Pool
	resource   *dockertest.Resource
}

func (r *TagTemplateRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewNoop()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewTagTemplateRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *TagTemplateRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	err := r.client.Close()
	if err != nil {
		r.T().Fatal(err)
	}
	err = purgeDocker(r.pool, r.resource)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *TagTemplateRepositoryTestSuite) TestNewRepository() {
	r.Run("should return repository and nil if postgres client is not nil", func() {
		pgClient := &postgres.Client{}

		actualRepository, err := postgres.NewTagTemplateRepository(pgClient)
		r.NotNil(actualRepository)
		r.Nil(err)
	})
}

func (r *TagTemplateRepositoryTestSuite) TestCreate() {

	r.Run("should return error if template is nil", func() {
		var template *tag.Template = nil

		expectedErrorMsg := "template is nil"

		actualError := r.repository.Create(r.ctx, template)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return nil and insert new record if no error found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()

		actualError := r.repository.Create(r.ctx, &template)
		r.NoError(actualError)

		var actualRecord tag.Template
		templates, err := r.repository.Read(r.ctx, template.URN)
		r.NoError(err)

		actualRecord = templates[0]
		r.Equal(template.URN, actualRecord.URN)
		r.Equal(template.DisplayName, actualRecord.DisplayName)
		r.Equal(template.Description, actualRecord.Description)
		r.Equal(len(template.Fields), len(actualRecord.Fields))
		r.Equal(template.Fields[0].DisplayName, actualRecord.Fields[0].DisplayName)
		r.Equal(template.Fields[0].URN, actualRecord.Fields[0].URN)
	})

	r.Run("should return nil and update template", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		originalTemplate := r.getTemplate()
		referenceTemplate := r.getTemplate()

		actualError := r.repository.Create(r.ctx, &originalTemplate)
		r.NoError(actualError)

		r.Equal(referenceTemplate.URN, originalTemplate.URN)
		r.NotEqual(referenceTemplate.CreatedAt, originalTemplate.CreatedAt)
		r.NotEqual(referenceTemplate.UpdatedAt, originalTemplate.UpdatedAt)
		r.Equal(referenceTemplate.Fields[0].ID, originalTemplate.Fields[0].ID)
		r.Equal(referenceTemplate.Fields[0].URN, originalTemplate.Fields[0].URN)
		r.NotEqual(referenceTemplate.Fields[0].CreatedAt, originalTemplate.Fields[0].CreatedAt)
		r.NotEqual(referenceTemplate.Fields[0].UpdatedAt, originalTemplate.Fields[0].UpdatedAt)
	})

	r.Run("should return error if encountered uncovered error", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()

		if err := r.repository.Create(r.ctx, &template); err != nil {
			r.Error(err)
		}
		actualError := r.repository.Create(r.ctx, &template)

		r.Error(actualError)
	})
}

func (r *TagTemplateRepositoryTestSuite) TestRead() {
	r.Run("should return empty and no error if no record found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)
		template := r.getTemplate()

		actualTemplate, actualError := r.repository.Read(r.ctx, template.URN)

		r.Empty(actualTemplate)
		r.NoError(actualError)
	})

	r.Run("should return templates and nil if found any", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()
		if err := r.repository.Create(r.ctx, &template); err != nil {
			panic(err)
		}
		now := time.Now()

		expectedTemplate := []tag.Template{template}
		r.updateTimeForTemplate(&expectedTemplate[0], now)

		actualTemplate, actualError := r.repository.Read(r.ctx, template.URN)
		r.updateTimeForTemplate(&actualTemplate[0], now)

		r.EqualValues(expectedTemplate, actualTemplate)
		r.NoError(actualError)
	})
	r.Run("should return template with multiple fields if exist", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()

		template.DisplayName = "Random Display"
		template.Fields[0].DisplayName = "Another Random Display"
		template.Fields = append(template.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		err = r.repository.Create(r.ctx, &template)
		r.NoError(err)

		templates, err := r.repository.Read(r.ctx, template.URN)

		r.NoError(err)
		r.Len(templates[0].Fields, 2)
		r.Equal(template.DisplayName, templates[0].DisplayName)
		r.Equal(template.UpdatedAt, templates[0].UpdatedAt)
	})
}

func (r *TagTemplateRepositoryTestSuite) TestReadAll() {
	r.Run("should return empty and no error if no record found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)
		actualTemplate, actualError := r.repository.ReadAll(r.ctx)

		r.Empty(actualTemplate)
		r.NoError(actualError)
	})

	r.Run("should return templates and nil if found any", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()
		if err := r.repository.Create(r.ctx, &template); err != nil {
			panic(err)
		}
		now := time.Now()

		expectedTemplate := []tag.Template{template}
		r.updateTimeForTemplate(&expectedTemplate[0], now)

		actualTemplate, actualError := r.repository.ReadAll(r.ctx)
		r.updateTimeForTemplate(&actualTemplate[0], now)

		r.EqualValues(expectedTemplate, actualTemplate)
		r.NoError(actualError)
	})
	r.Run("should return template with multiple fields if exist", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()

		template.DisplayName = "Random Display"
		template.Fields[0].DisplayName = "Another Random Display"
		template.Fields = append(template.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		err = r.repository.Create(r.ctx, &template)
		r.NoError(err)

		templates, err := r.repository.ReadAll(r.ctx)

		r.NoError(err)
		r.Len(templates[0].Fields, 2)
		r.Equal(template.DisplayName, templates[0].DisplayName)
		r.Equal(template.UpdatedAt, templates[0].UpdatedAt)
	})
}

func (r *TagTemplateRepositoryTestSuite) TestUpdate() {
	r.Run("should return error if template is nil", func() {
		var template *tag.Template = nil

		expectedErrorMsg := "template is nil"

		actualError := r.repository.Update(r.ctx, "", template)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record not found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()

		actualError := r.repository.Update(r.ctx, template.URN, &template)

		r.Error(actualError)
	})

	r.Run("should return nil and updated template if update is success", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()
		err = r.repository.Create(r.ctx, &template)
		r.NoError(err)

		template.DisplayName = "Random Display"
		template.Fields[0].DisplayName = "Another Random Display"
		template.Fields = append(template.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		actualError := r.repository.Update(r.ctx, template.URN, &template)
		r.NoError(actualError)

		templates, err := r.repository.Read(r.ctx, template.URN)
		r.NoError(err)

		recordModelTemplate := templates[0]
		r.Len(recordModelTemplate.Fields, 2)
		r.Equal(template.DisplayName, recordModelTemplate.DisplayName)
		r.True(template.UpdatedAt.Equal(recordModelTemplate.UpdatedAt))

		expectedFields, err := json.Marshal(recordModelTemplate.Fields)
		r.NoError(err)

		actualFields, err := json.Marshal(template.Fields)
		r.NoError(err)

		r.JSONEq(string(expectedFields), string(actualFields))
	})

	r.Run("should return error if trying to update with conflicting existing template", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template1 := r.getTemplate()
		template1.URN = "hello1"
		if err := r.repository.Create(r.ctx, &template1); err != nil {
			panic(err)
		}
		template2 := r.getTemplate()
		template2.URN = "hello2"
		if err := r.repository.Create(r.ctx, &template2); err != nil {
			panic(err)
		}
		targetURN := template2.URN
		template2.URN = "hello1"

		actualError := r.repository.Update(r.ctx, targetURN, &template2)

		r.Error(actualError)
	})

	r.Run("should return error if trying to update with unrelated field", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()
		if err := r.repository.Create(r.ctx, &template); err != nil {
			panic(err)
		}
		template.Fields[0].ID = 2

		actualError := r.repository.Update(r.ctx, template.URN, &template)

		r.Error(actualError)
	})

	r.Run("should return error if trying to update with duplicated field", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()
		template.Fields = append(template.Fields, tag.Field{
			URN:         "second_field",
			DisplayName: "Second Field",
			Description: "Random description for the second field.",
			DataType:    "string",
			Required:    false,
		})

		if err := r.repository.Create(r.ctx, &template); err != nil {
			panic(err)
		}

		template.Fields[1].URN = template.Fields[0].URN

		actualError := r.repository.Update(r.ctx, template.URN, &template)

		r.Error(actualError)
	})
}

func (r *TagTemplateRepositoryTestSuite) TestDelete() {
	r.Run("should return error if record not found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()

		err = r.repository.Delete(r.ctx, template.URN)
		r.EqualError(err, "could not find template \"governance_policy\"")
	})

	r.Run("should return nil if record is deleted", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		template := r.getTemplate()
		if err := r.repository.Create(r.ctx, &template); err != nil {
			panic(err)
		}

		actualError := r.repository.Delete(r.ctx, template.URN)
		r.NoError(actualError)

		templates, err := r.repository.Read(r.ctx, template.URN)
		r.NoError(err)
		r.Empty(templates)
	})
}

func (r *TagTemplateRepositoryTestSuite) getTemplate() tag.Template {
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

func (r *TagTemplateRepositoryTestSuite) updateTimeForTemplate(template *tag.Template, t time.Time) {
	template.CreatedAt = t
	template.UpdatedAt = t
	for i := 0; i < len(template.Fields); i++ {
		template.Fields[i].CreatedAt = t
		template.Fields[i].UpdatedAt = t
	}
}

func TestTagTemplateRepository(t *testing.T) {
	suite.Run(t, &TagTemplateRepositoryTestSuite{})
}
