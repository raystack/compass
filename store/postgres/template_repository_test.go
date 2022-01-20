package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type TemplateRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	repository *postgres.TemplateRepository
	pool       *dockertest.Pool
	resource   *dockertest.Resource
}

func (r *TemplateRepositoryTestSuite) SetupSuite() {
	var err error

	logger := logrus.New()
	// logger.SetLevel(logrus.DebugLevel)
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		logger.Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewTemplateRepository(r.client)
	if err != nil {
		logger.Fatal(err)
	}

}

func (r *TemplateRepositoryTestSuite) TearDownSuite() {
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

func (r *TemplateRepositoryTestSuite) TestNewRepository() {
	r.Run("should return repository and nil if postgres client is not nil", func() {
		pgClient := &postgres.Client{}

		actualRepository, err := postgres.NewTemplateRepository(pgClient)
		r.NotNil(actualRepository)
		r.Nil(err)
	})
}

func (r *TemplateRepositoryTestSuite) TestCreate() {

	r.Run("should return error if domain template is nil", func() {
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := r.repository.Create(r.ctx, domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return nil and insert new record if no error found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()

		actualError := r.repository.Create(r.ctx, &domainTemplate)
		r.NoError(actualError)

		var actualRecord tag.Template
		templates, err := r.repository.Read(r.ctx, domainTemplate)
		r.NoError(err)

		actualRecord = templates[0]
		r.Equal(domainTemplate.URN, actualRecord.URN)
		r.Equal(domainTemplate.DisplayName, actualRecord.DisplayName)
		r.Equal(domainTemplate.Description, actualRecord.Description)
		r.Equal(len(domainTemplate.Fields), len(actualRecord.Fields))
		r.Equal(domainTemplate.Fields[0].DisplayName, actualRecord.Fields[0].DisplayName)
		r.Equal(domainTemplate.Fields[0].URN, actualRecord.Fields[0].URN)
	})

	r.Run("should return nil and update domain template", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		originalDomainTemplate := r.getDomainTemplate()
		referenceDomainTemplate := r.getDomainTemplate()

		actualError := r.repository.Create(r.ctx, &originalDomainTemplate)
		r.NoError(actualError)

		r.Equal(referenceDomainTemplate.URN, originalDomainTemplate.URN)
		r.NotEqual(referenceDomainTemplate.CreatedAt, originalDomainTemplate.CreatedAt)
		r.NotEqual(referenceDomainTemplate.UpdatedAt, originalDomainTemplate.UpdatedAt)
		r.Equal(referenceDomainTemplate.Fields[0].ID, originalDomainTemplate.Fields[0].ID)
		r.Equal(referenceDomainTemplate.Fields[0].URN, originalDomainTemplate.Fields[0].URN)
		r.NotEqual(referenceDomainTemplate.Fields[0].CreatedAt, originalDomainTemplate.Fields[0].CreatedAt)
		r.NotEqual(referenceDomainTemplate.Fields[0].UpdatedAt, originalDomainTemplate.Fields[0].UpdatedAt)
	})

	r.Run("should return error if encountered uncovered error", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()

		r.repository.Create(r.ctx, &domainTemplate)
		actualError := r.repository.Create(r.ctx, &domainTemplate)

		r.Error(actualError)
	})
}

func (r *TemplateRepositoryTestSuite) TestRead() {
	r.Run("should return empty and nil if no record found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)
		domainTemplate := r.getDomainTemplate()

		actualTemplate, actualError := r.repository.Read(r.ctx, domainTemplate)

		r.Empty(actualTemplate)
		r.EqualError(actualError, "error fetching templates: could not find template \"governance_policy\"")
	})

	r.Run("should return domain templates and nil if found any", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()
		if err := r.repository.Create(r.ctx, &domainTemplate); err != nil {
			panic(err)
		}
		now := time.Now()

		expectedTemplate := []tag.Template{domainTemplate}
		r.updateTimeForDomainTemplate(&expectedTemplate[0], now)

		actualTemplate, actualError := r.repository.Read(r.ctx, domainTemplate)
		r.updateTimeForDomainTemplate(&actualTemplate[0], now)

		r.EqualValues(expectedTemplate, actualTemplate)
		r.NoError(actualError)
	})
	r.Run("should return template with multiple fields if exist", func() {
		err := setup(r.ctx, r.client)
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

		err = r.repository.Create(r.ctx, &domainTemplate)
		r.NoError(err)

		templates, err := r.repository.Read(r.ctx, domainTemplate)

		r.NoError(err)
		r.Len(templates[0].Fields, 2)
		r.Equal(domainTemplate.DisplayName, templates[0].DisplayName)
		r.Equal(domainTemplate.UpdatedAt, templates[0].UpdatedAt)
	})
}

func (r *TemplateRepositoryTestSuite) TestUpdate() {
	r.Run("should return error if domain template is nil", func() {
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := r.repository.Update(r.ctx, "", domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record not found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()

		actualError := r.repository.Update(r.ctx, domainTemplate.URN, &domainTemplate)

		r.Error(actualError)
	})

	r.Run("should return nil and updated domain template if update is success", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()
		err = r.repository.Create(r.ctx, &domainTemplate)
		r.NoError(err)

		domainTemplate.DisplayName = "Random Display"
		domainTemplate.Fields[0].DisplayName = "Another Random Display"
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		actualError := r.repository.Update(r.ctx, domainTemplate.URN, &domainTemplate)
		r.NoError(actualError)

		templates, err := r.repository.Read(r.ctx, domainTemplate)
		r.NoError(err)

		recordModelTemplate := templates[0]
		r.Len(recordModelTemplate.Fields, 2)
		r.Equal(domainTemplate.DisplayName, recordModelTemplate.DisplayName)
		r.True(domainTemplate.UpdatedAt.Equal(recordModelTemplate.UpdatedAt))

		expectedFields, err := json.Marshal(recordModelTemplate.Fields)
		r.NoError(err)

		actualFields, err := json.Marshal(domainTemplate.Fields)
		r.NoError(err)

		r.JSONEq(string(expectedFields), string(actualFields))
	})

	r.Run("should return error if trying to update with conflicting existing template", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate1 := r.getDomainTemplate()
		domainTemplate1.URN = "hello1"
		if err := r.repository.Create(r.ctx, &domainTemplate1); err != nil {
			panic(err)
		}
		domainTemplate2 := r.getDomainTemplate()
		domainTemplate2.URN = "hello2"
		if err := r.repository.Create(r.ctx, &domainTemplate2); err != nil {
			panic(err)
		}
		targetURN := domainTemplate2.URN
		domainTemplate2.URN = "hello1"

		actualError := r.repository.Update(r.ctx, targetURN, &domainTemplate2)

		r.Error(actualError)
	})

	r.Run("should return error if trying to update with unrelated field", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()
		if err := r.repository.Create(r.ctx, &domainTemplate); err != nil {
			panic(err)
		}
		domainTemplate.Fields[0].ID = 2

		actualError := r.repository.Update(r.ctx, domainTemplate.URN, &domainTemplate)

		r.Error(actualError)
	})

	r.Run("should return error if trying to update with duplicated field", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "second_field",
			DisplayName: "Second Field",
			Description: "Random description for the second field.",
			DataType:    "string",
			Required:    false,
		})

		if err := r.repository.Create(r.ctx, &domainTemplate); err != nil {
			panic(err)
		}

		domainTemplate.Fields[1].URN = domainTemplate.Fields[0].URN

		actualError := r.repository.Update(r.ctx, domainTemplate.URN, &domainTemplate)

		r.Error(actualError)
	})
}

func (r *TemplateRepositoryTestSuite) TestDelete() {
	r.Run("should return error if record not found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()

		err = r.repository.Delete(r.ctx, domainTemplate)
		r.EqualError(err, "could not find template \"governance_policy\"")
	})

	r.Run("should return nil if record is deleted", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := r.getDomainTemplate()
		if err := r.repository.Create(r.ctx, &domainTemplate); err != nil {
			panic(err)
		}

		actualError := r.repository.Delete(r.ctx, domainTemplate)
		r.NoError(actualError)

		templates, err := r.repository.Read(r.ctx, domainTemplate)
		r.Error(err)
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
