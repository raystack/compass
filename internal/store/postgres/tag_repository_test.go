package postgres_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/odpf/compass/core/tag"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

var domainAssetID = uuid.NewString()

type TagRepositoryTestSuite struct {
	suite.Suite
	ctx                context.Context
	client             *postgres.Client
	repository         *postgres.TagRepository
	templateRepository *postgres.TagTemplateRepository
	pool               *dockertest.Pool
	resource           *dockertest.Resource
}

func (r *TagRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewNoop()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewTagRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
	r.templateRepository, err = postgres.NewTagTemplateRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *TagRepositoryTestSuite) TearDownSuite() {
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

func (r *TagRepositoryTestSuite) TestNewRepository() {
	r.Run("should return repository and nil if postgres client is not nil", func() {
		pgClient := &postgres.Client{}

		actualTagRepository, err := postgres.NewTagRepository(pgClient)

		r.NotNil(actualTagRepository)
		r.Nil(err)
	})
}

func (r *TagRepositoryTestSuite) TestCreate() {
	r.Run("should return error if tag is nil", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		var domainTag *tag.Tag = nil

		expectedErrorMsg := "tag is nil"

		actualError := r.repository.Create(r.ctx, domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if template is not found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)
		domain := getDomainTag()

		err = r.repository.Create(r.ctx, &domain)

		r.EqualError(err, tag.TemplateNotFoundError{URN: domain.TemplateURN}.Error())
	})

	r.Run("should return nil and create tag if no error found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		err = r.repository.Create(r.ctx, &domainTag)
		r.NoError(err)

		tags, err := r.repository.Read(r.ctx, domainTag)
		r.NoError(err)

		r.Equal(domainTag.AssetID, tags[0].AssetID)
		r.Equal(domainTag.TemplateDescription, tags[0].TemplateDescription)
		r.Equal(domainTag.TemplateDisplayName, tags[0].TemplateDisplayName)
		r.Equal(domainTag.TemplateURN, tags[0].TemplateURN)
		r.NotEmpty(domainTag.TagValues)
		r.NotEmpty(tags[0].TagValues)

		expectedTagValues := domainTag.TagValues
		actualTagValues := tags[0].TagValues

		sort.Slice(expectedTagValues[:], func(i, j int) bool {
			return expectedTagValues[i].FieldID < expectedTagValues[j].FieldID
		})
		sort.Slice(actualTagValues[:], func(i, j int) bool {
			return actualTagValues[i].FieldID < actualTagValues[j].FieldID
		})
		r.EqualValues(expectedTagValues, actualTagValues)
	})

	r.Run("should return nil and update tag if no error found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		err = r.repository.Create(r.ctx, &domainTag)
		r.NoError(err)

		for _, value := range domainTag.TagValues {
			r.NotZero(value.CreatedAt)
		}
	})
}

func (r *TagRepositoryTestSuite) TestRead() {
	r.Run("should return error if asset id is empty", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		paramDomainTag := tag.Tag{
			AssetID: "",
		}

		expectedErrorMsg := "asset id should not be empty"

		actualTag, actualError := r.repository.Read(r.ctx, paramDomainTag)

		r.Nil(actualTag)
		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return empty and nil if no tags found for the specified asset", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		paramDomainTag := tag.Tag{
			AssetID: uuid.NewString(),
		}

		actualTag, actualError := r.repository.Read(r.ctx, paramDomainTag)
		r.NoError(actualError)
		r.Empty(actualTag)
	})

	r.Run("should return tags if found for the specified asset", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.Require().NoError(err)

		domainTag := getDomainTag()
		err = r.repository.Create(r.ctx, &domainTag)
		r.Require().NoError(err)

		tags, err := r.repository.Read(r.ctx, tag.Tag{
			AssetID: domainTag.AssetID,
		})

		r.NoError(err)
		r.NotEmpty(tags)
		r.Len(tags[0].TagValues, 2)
	})

	r.Run("should return nil and not found error if no tags found for the specified asset id and template", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		var assetID = uuid.NewString()
		var templateURN = "governance_policy"

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.NoError(err)

		paramDomainTag := tag.Tag{
			AssetID:     assetID,
			TemplateURN: templateURN,
		}

		expectedErrorMsg := fmt.Sprintf("could not find tag with asset id: \"%s\", template: \"%s\"",
			assetID, templateURN,
		)

		actualTag, actualError := r.repository.Read(r.ctx, paramDomainTag)
		r.ErrorAs(actualError, new(tag.NotFoundError))
		r.EqualError(actualError, expectedErrorMsg)
		r.Nil(actualTag)
	})

	r.Run("should return maximum of one tag for the specified asset id and template", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		var assetID = domainAssetID
		var templateURN = "governance_policy"

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		if err := r.repository.Create(r.ctx, &domainTag); err != nil {
			panic(err)
		}
		paramDomainTag := tag.Tag{
			AssetID:     assetID,
			TemplateURN: templateURN,
		}

		expectedLength := 1

		actualTag, actualError := r.repository.Read(r.ctx, paramDomainTag)

		r.Len(actualTag, expectedLength)
		r.NoError(actualError)
	})
}

func (r *TagRepositoryTestSuite) TestUpdate() {
	r.Run("should return error if tag is nil", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		var domainTag *tag.Tag = nil

		expectedErrorMsg := "tag is nil"

		actualError := r.repository.Update(r.ctx, domainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should return error if template is not found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)
		t := getDomainTag()

		err = r.repository.Update(r.ctx, &t)
		r.EqualError(err, tag.TemplateNotFoundError{URN: t.TemplateURN}.Error())
	})

	r.Run("should return nil and update tag if no error found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.Require().NoError(err)

		domainTag := getDomainTag()
		err = r.repository.Create(r.ctx, &domainTag)
		r.Require().NoError(err)

		domainTag.TagValues[0].FieldValue = "Restricted"
		actualError := r.repository.Update(r.ctx, &domainTag)
		r.Require().NoError(actualError)

		updatedTags, err := r.repository.Read(r.ctx, domainTag)
		r.NoError(err)

		for _, updatedTag := range updatedTags {
			for idx, value := range updatedTag.TagValues {
				r.NoError(err)
				r.EqualValues(value.UpdatedAt, domainTag.TagValues[idx].UpdatedAt)
			}
		}
	})

	r.Run("should return nil and update tag if no error found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.NoError(err)
		domainTag := getDomainTag()

		if err := r.repository.Create(r.ctx, &domainTag); err != nil {
			panic(err)
		}
		domainTag.TagValues = domainTag.TagValues[:1]

		actualError := r.repository.Update(r.ctx, &domainTag)

		r.NoError(actualError)
		r.Len(domainTag.TagValues, 2)
	})
}

func (r *TagRepositoryTestSuite) TestDelete() {
	r.Run("should return error if asset id is empty", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		paramDomainTag := tag.Tag{
			AssetID: "",
		}

		expectedErrorMsg := "asset id should not be empty"

		actualError := r.repository.Delete(r.ctx, paramDomainTag)

		r.EqualError(actualError, expectedErrorMsg)
	})

	r.Run("should delete tags related to the asset id and return no error if the asset id has one", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.NoError(err)

		domainTag := getDomainTag()

		if err := r.repository.Create(r.ctx, &domainTag); err != nil {
			r.T().Fatal(err)
		}

		paramDomainTag := tag.Tag{
			AssetID: domainTag.AssetID,
		}

		actualError := r.repository.Delete(r.ctx, paramDomainTag)
		r.NoError(actualError)

		foundTags, err := r.repository.Read(r.ctx, tag.Tag{
			AssetID: paramDomainTag.AssetID,
		})
		if err != nil {
			r.T().Fatal(err)
		}

		r.Empty(foundTags)
	})

	r.Run("should return error if template is not found", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		var assetID = uuid.NewString()
		var templateURN = "random-urn"
		paramDomainTag := tag.Tag{
			AssetID:     assetID,
			TemplateURN: templateURN,
		}

		err = r.repository.Delete(r.ctx, paramDomainTag)
		r.ErrorIs(err, tag.TemplateNotFoundError{URN: templateURN})
	})

	r.Run("should delete only the tag for asset id and template and return error if asset id has none", func() {
		err := setup(r.ctx, r.client)
		r.NoError(err)

		var assetID = uuid.NewString()
		domainTemplate := getTemplate()
		err = r.templateRepository.Create(r.ctx, domainTemplate)
		r.NoError(err)

		paramDomainTag := tag.Tag{
			AssetID:     assetID,
			TemplateURN: domainTemplate.URN,
		}

		actualError := r.repository.Delete(r.ctx, paramDomainTag)
		r.Error(actualError)
	})
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
				Description: "The classification of this asset",
				DataType:    "enumerated",
				Required:    true,
				Options:     []string{"Public", "Restricted"},
			},
			{
				ID:          2,
				URN:         "admin_email",
				DisplayName: "Admin Email",
				Description: "Email of the admin of the asset.",
				DataType:    "string",
				Required:    true,
			},
		},
	}
}

func getDomainTag() tag.Tag {
	return tag.Tag{
		AssetID:             domainAssetID,
		TemplateURN:         "governance_policy",
		TemplateDisplayName: "Governance Policy",
		TemplateDescription: "Template that is mandatory to be used.",
		TagValues: []tag.TagValue{
			{
				FieldID:          1,
				FieldValue:       "Public",
				FieldURN:         "classification",
				FieldDisplayName: "classification",
				FieldDescription: "The classification of this asset",
				FieldDataType:    "enumerated",
				FieldRequired:    true,
				FieldOptions:     []string{"Public", "Restricted"},
			},
			{
				FieldID:          2,
				FieldValue:       "dexter@odpf.io",
				FieldURN:         "admin_email",
				FieldDisplayName: "Admin Email",
				FieldDescription: "Email of the admin of the asset.",
				FieldDataType:    "string",
				FieldRequired:    true,
			},
		},
	}
}

func TestTagRepository(t *testing.T) {
	suite.Run(t, &TagRepositoryTestSuite{})
}
