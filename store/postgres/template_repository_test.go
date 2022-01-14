package postgres_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logger := logrus.New()
	// logger.SetLevel(logrus.DebugLevel)
	// _, _, err := newTestClient(logger)
	p, r, err := newTestClient(logger)
	if err != nil {
		log.Fatal(err)
	}

	// Run tests
	code := m.Run()

	// Clean tests
	err = testDBClient.Conn.Close()
	err = purgeClient(p, r)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

func TestNewRepository(t *testing.T) {
	t.Run("should return repository and nil if db client is not nil", func(t *testing.T) {
		dummyDBClient := &postgres.Client{}

		actualRepository := postgres.NewTemplateRepository(dummyDBClient)
		assert.NotNil(t, actualRepository)
	})
}

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	t.Run("should return error if domain template is nil", func(t *testing.T) {
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := repository.Create(ctx, domainTemplate)

		assert.EqualError(t, actualError, expectedErrorMsg)
	})

	t.Run("should return error if db client is nil", func(t *testing.T) {
		domainTemplate := tag.Template{
			URN: "governance_policy",
		}
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Create(ctx, &domainTemplate)

		assert.EqualError(t, actualError, expectedErrorMsg)
	})

	t.Run("should return nil and insert new record if no error found", func(t *testing.T) {
		domainTemplate := getDomainTemplate()

		err := repository.Create(ctx, &domainTemplate)
		assert.NoError(t, err)

		var actualRecord tag.Template
		templates, err := repository.Read(ctx, domainTemplate)
		require.Nil(t, err)

		actualRecord = templates[0]
		assert.Equal(t, domainTemplate.URN, actualRecord.URN)
		assert.Equal(t, domainTemplate.DisplayName, actualRecord.DisplayName)
		assert.Equal(t, domainTemplate.Description, actualRecord.Description)
		assert.Equal(t, len(domainTemplate.Fields), len(actualRecord.Fields))
		assert.Equal(t, domainTemplate.Fields[0].DisplayName, actualRecord.Fields[0].DisplayName)
		assert.Equal(t, domainTemplate.Fields[0].URN, actualRecord.Fields[0].URN)
	})

	t.Run("should return error if encountered uncovered error", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)

		domainTemplate := getDomainTemplate()

		err = repository.Create(ctx, &domainTemplate)
		err = repository.Create(ctx, &domainTemplate)

		assert.EqualError(t, errors.Unwrap(err), "failed to insert a template: ERROR: duplicate key value violates unique constraint \"templates_pkey\" (SQLSTATE 23505)")
	})
}

func TestRead(t *testing.T) {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	t.Run("should return nil and error if db client is nil", func(t *testing.T) {
		domainTemplate := getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualTemplate, actualError := repository.Read(ctx, domainTemplate)

		assert.Nil(t, actualTemplate)
		assert.EqualError(t, actualError, expectedErrorMsg)
	})

	t.Run("should return empty and nil if no record found", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)
		domainTemplate := getDomainTemplate()

		actualTemplate, actualError := repository.Read(ctx, domainTemplate)

		assert.Empty(t, actualTemplate)
		assert.NoError(t, actualError)
	})

	t.Run("should return template with multiple fields if exist", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)
		domainTemplate := getDomainTemplate()

		domainTemplate.DisplayName = "Random Display"
		domainTemplate.Fields[0].DisplayName = "Another Random Display"
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		err = repository.Create(ctx, &domainTemplate)
		require.NoError(t, err)

		templates, err := repository.Read(ctx, domainTemplate)

		assert.NoError(t, err)
		assert.Len(t, templates[0].Fields, 2)
		assert.Equal(t, domainTemplate.DisplayName, templates[0].DisplayName)
		assert.Equal(t, domainTemplate.UpdatedAt, templates[0].UpdatedAt)
	})
}

func TestUpdate(t *testing.T) {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	t.Run("should return error if domain template is nil", func(t *testing.T) {
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := repository.Update(ctx, domainTemplate)

		assert.EqualError(t, actualError, expectedErrorMsg)
	})

	t.Run("should return error if db client is nil", func(t *testing.T) {
		domainTemplate := getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Update(ctx, &domainTemplate)

		assert.EqualError(t, actualError, expectedErrorMsg)
	})

	t.Run("should return error if record not found", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)
		domainTemplate := getDomainTemplate()

		actualError := repository.Update(ctx, &domainTemplate)

		assert.Error(t, actualError)
	})

	t.Run("should return nil and updated domain template if update is success", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)
		domainTemplate := getDomainTemplate()
		err = repository.Create(ctx, &domainTemplate)
		require.NoError(t, err)

		domainTemplate.DisplayName = "Random Display"
		domainTemplate.Fields[0].DisplayName = "Another Random Display"
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		actualError := repository.Update(ctx, &domainTemplate)
		assert.NoError(t, actualError)

		templates, err := repository.Read(ctx, domainTemplate)
		require.NoError(t, err)
		recordModelTemplate := templates[0]

		assert.Len(t, recordModelTemplate.Fields, 2)
		assert.Equal(t, domainTemplate.DisplayName, recordModelTemplate.DisplayName)
		assert.Equal(t, domainTemplate.Fields[0].DisplayName, recordModelTemplate.Fields[0].DisplayName)
		assert.Equal(t, domainTemplate.UpdatedAt, recordModelTemplate.UpdatedAt)
		assert.Equal(t, domainTemplate.Fields[0].UpdatedAt, recordModelTemplate.Fields[0].UpdatedAt)
	})

	t.Run("should return error if trying to update with unrelated field", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)
		domainTemplate := getDomainTemplate()
		if err := repository.Create(ctx, &domainTemplate); err != nil {
			panic(err)
		}
		domainTemplate.Fields[0].ID = 2

		actualError := repository.Update(ctx, &domainTemplate)

		assert.EqualError(t, actualError, "error updating template: field not found when updating fields")
	})

	t.Run("should return error if trying to update with duplicated field", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)
		domainTemplate := getDomainTemplate()
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

		assert.EqualError(t, actualError, "error updating template: failed updating fields: ERROR: duplicate key value violates unique constraint \"fields_idx_urn_template_urn\" (SQLSTATE 23505)")
	})
}

func TestDelete(t *testing.T) {
	ctx := context.TODO()
	repository := postgres.NewTemplateRepository(testDBClient)

	t.Run("should return error if db client is nil", func(t *testing.T) {
		domainTemplate := getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Delete(ctx, domainTemplate)

		assert.EqualError(t, actualError, expectedErrorMsg)
	})

	t.Run("should return error if record not found", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)

		domainTemplate := getDomainTemplate()

		err = repository.Delete(ctx, domainTemplate)
		assert.EqualError(t, err, "error deleting template: could not find template \"governance_policy\"")
	})

	t.Run("should return nil if record is deleted", func(t *testing.T) {
		err := setup()
		require.NoError(t, err)

		domainTemplate := getDomainTemplate()
		if err := repository.Create(ctx, &domainTemplate); err != nil {
			panic(err)
		}

		err = repository.Delete(ctx, domainTemplate)
		assert.NoError(t, err)

		templates, err := repository.Read(ctx, domainTemplate)
		require.NoError(t, err)

		assert.Empty(t, templates)
	})
}

func getDomainTemplate() tag.Template {
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

func updateTimeForDomainTemplate(domainTemplate *tag.Template, t time.Time) {
	domainTemplate.CreatedAt = t
	domainTemplate.UpdatedAt = t
	for i := 0; i < len(domainTemplate.Fields); i++ {
		domainTemplate.Fields[i].CreatedAt = t
		domainTemplate.Fields[i].UpdatedAt = t
	}
}
