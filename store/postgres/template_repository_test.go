package postgres_test

import (
	"testing"
	"time"

	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/tag"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TemplateRepositoryTestSuite struct {
	suite.Suite
	dbClient   *gorm.DB
	repository *postgres.TemplateRepository
}

func (r *TemplateRepositoryTestSuite) Setup() {
	r.dbClient, _ = newTestClient("file::memory:")
	r.dbClient.AutoMigrate(&postgres.Template{})
	r.dbClient.AutoMigrate(&postgres.Field{})
	r.dbClient.AutoMigrate(&postgres.Tag{})
	repository := postgres.NewTemplateRepository(r.dbClient)
	r.repository = repository
}

func (r *TemplateRepositoryTestSuite) TestNewRepository() {
	r.Run("should return repository and nil if db client is not nil", func() {
		dbClient := &gorm.DB{}

		actualRepository := postgres.NewTemplateRepository(dbClient)
		r.NotNil(actualRepository)
	})
}

func (r *TemplateRepositoryTestSuite) TestCreate() {
	r.Run("should return error if domain template is nil", func() {
		r.Setup()
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := r.repository.Create(domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTemplate := tag.Template{
			URN: "governance_policy",
		}
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Create(&domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return nil and insert new record if no error found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()

		actualError := r.repository.Create(&domainTemplate)
		var actualRecord postgres.Template
		result := r.dbClient.Preload("Fields").First(&actualRecord)

		r.NoError(actualError)
		r.NotZero(result.RowsAffected)
		r.Equal(domainTemplate.URN, actualRecord.URN)
		r.Equal(domainTemplate.DisplayName, actualRecord.DisplayName)
		r.Equal(domainTemplate.Description, actualRecord.Description)
		r.Equal(len(domainTemplate.Fields), len(actualRecord.Fields))
		r.Equal(domainTemplate.Fields[0].DisplayName, actualRecord.Fields[0].DisplayName)
		r.Equal(domainTemplate.Fields[0].URN, actualRecord.Fields[0].URN)
	})

	r.Run("should return nil and update domain template", func() {
		r.Setup()
		originalDomainTemplate := r.getDomainTemplate()
		referenceDomainTemplate := r.getDomainTemplate()

		actualError := r.repository.Create(&originalDomainTemplate)
		var actualRecord postgres.Template
		r.dbClient.Preload("Fields").First(&actualRecord)

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
		r.Setup()
		domainTemplate := r.getDomainTemplate()

		r.repository.Create(&domainTemplate)
		actualError := r.repository.Create(&domainTemplate)

		r.Error(actualError)
	})
}

func (r *TemplateRepositoryTestSuite) TestRead() {
	r.Run("should return nil and error if db client is nil", func() {
		domainTemplate := r.getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualTemplate, actualError := repository.Read(domainTemplate)

		r.Nil(actualTemplate)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return empty and nil if no record found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()

		actualTemplate, actualError := r.repository.Read(domainTemplate)

		r.Empty(actualTemplate)
		r.NoError(actualError)
	})

	r.Run("should return domain templates and nil if found any", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		if err := r.repository.Create(&domainTemplate); err != nil {
			panic(err)
		}
		now := time.Now()

		expectedTemplate := []tag.Template{domainTemplate}
		r.updateTimeForDomainTemplate(&expectedTemplate[0], now)

		actualTemplate, actualError := r.repository.Read(domainTemplate)
		r.updateTimeForDomainTemplate(&actualTemplate[0], now)

		r.EqualValues(expectedTemplate, actualTemplate)
		r.NoError(actualError)
	})
}

func (r *TemplateRepositoryTestSuite) TestUpdate() {
	r.Run("should return error if domain template is nil", func() {
		r.Setup()
		var domainTemplate *tag.Template = nil

		expectedErrorMsg := "domain template is nil"

		actualError := r.repository.Update(domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if db client is nil", func() {
		domainTemplate := r.getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Update(&domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record not found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()

		actualError := r.repository.Update(&domainTemplate)

		r.Error(actualError)
	})

	r.Run("should return nil and updated domain template if update is success", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		err := r.repository.Create(&domainTemplate)
		r.NoError(err)

		domainTemplate.DisplayName = "Random Display"
		domainTemplate.Fields[0].DisplayName = "Another Random Display"
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "new_field",
			DisplayName: "New Field",
			Description: "This field is a new addition.",
			DataType:    "string",
		})

		actualError := r.repository.Update(&domainTemplate)
		r.NoError(actualError)

		var recordModelTemplate postgres.Template
		if err := r.dbClient.Preload("Fields").First(&recordModelTemplate).Error; err != nil {
			panic(err)
		}
		r.Equal(domainTemplate.DisplayName, recordModelTemplate.DisplayName)
		r.Equal(domainTemplate.Fields[0].DisplayName, recordModelTemplate.Fields[0].DisplayName)
		r.Equal(domainTemplate.UpdatedAt, recordModelTemplate.UpdatedAt)
		r.Equal(domainTemplate.Fields[0].UpdatedAt, recordModelTemplate.Fields[0].UpdatedAt)
	})

	r.Run("should return error if trying to update with duplicate template", func() {
		r.Setup()
		domainTemplate1 := r.getDomainTemplate()
		domainTemplate1.URN = "hello1"
		if err := r.repository.Create(&domainTemplate1); err != nil {
			panic(err)
		}
		domainTemplate2 := r.getDomainTemplate()
		domainTemplate2.URN = "hello2"
		if err := r.repository.Create(&domainTemplate2); err != nil {
			panic(err)
		}
		domainTemplate2.URN = "hello1"

		actualError := r.repository.Update(&domainTemplate2)

		r.Error(actualError)
	})

	r.Run("should return error if trying to update with unrelated field", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		if err := r.repository.Create(&domainTemplate); err != nil {
			panic(err)
		}
		domainTemplate.Fields[0].ID = 2

		actualError := r.repository.Update(&domainTemplate)

		r.Error(actualError)
	})

	r.Run("should return error if trying to update with duplicate field", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		domainTemplate.Fields = append(domainTemplate.Fields, tag.Field{
			URN:         "second_field",
			DisplayName: "Second Field",
			Description: "Random description for the second field.",
			DataType:    "string",
			Required:    false,
		})
		if err := r.repository.Create(&domainTemplate); err != nil {
			panic(err)
		}
		domainTemplate.Fields[1].URN = domainTemplate.Fields[0].URN

		actualError := r.repository.Update(&domainTemplate)

		r.Error(actualError)
	})
}

func (r *TemplateRepositoryTestSuite) TestDelete() {
	r.Run("should return error if db client is nil", func() {
		domainTemplate := r.getDomainTemplate()
		repository := &postgres.TemplateRepository{}

		expectedErrorMsg := "db client is nil"

		actualError := repository.Delete(domainTemplate)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if record not found", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()

		err := r.repository.Delete(domainTemplate)

		r.Error(err)
		r.ErrorAs(err, new(tag.TemplateNotFoundError))
	})

	r.Run("should return nil if record is deleted", func() {
		r.Setup()
		domainTemplate := r.getDomainTemplate()
		if err := r.repository.Create(&domainTemplate); err != nil {
			panic(err)
		}

		actualError := r.repository.Delete(domainTemplate)
		var actualModelTemplate postgres.Template
		templateResult := r.dbClient.First(&actualModelTemplate)
		var actualModelField postgres.Field
		fieldResult := r.dbClient.First(&actualModelField)

		r.NoError(actualError)
		r.Zero(templateResult.RowsAffected)
		r.Zero(fieldResult.RowsAffected)
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
