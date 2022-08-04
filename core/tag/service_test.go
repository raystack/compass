package tag_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/odpf/compass/core/tag"
	"github.com/odpf/compass/core/tag/mocks"
	"github.com/odpf/compass/core/tag/validator"

	"github.com/golang-module/carbon/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	tagService   *tag.Service
	templateRepo *mocks.TagTemplateRepository
	repository   *mocks.TagRepository
}

func (s *ServiceTestSuite) TestNewService() {
	s.Run("should return service and nil if repository is not nil", func() {
		repository := &mocks.TagRepository{}
		templateService := tag.NewTemplateService(s.templateRepo)
		actualService := tag.NewService(repository, templateService)

		s.NotNil(actualService)
	})
}

func (s *ServiceTestSuite) Setup() {
	s.repository = &mocks.TagRepository{}
	s.templateRepo = &mocks.TagTemplateRepository{}

	templateService := tag.NewTemplateService(s.templateRepo)
	s.tagService = tag.NewService(s.repository, templateService)
}

func (s *ServiceTestSuite) TestValidate() {
	repository := &mocks.TagRepository{}
	templateService := tag.NewTemplateService(&mocks.TagTemplateRepository{})
	tagService := tag.NewService(repository, templateService)

	s.Run("should return error if asset id is empty", func() {
		t := s.buildTag()
		t.AssetID = ""

		expectedErrorMsg := "error with [asset_id : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"asset_id": "cannot be empty",
			},
		}

		actualError := tagService.Validate(&t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if template URN is empty", func() {
		t := s.buildTag()
		t.TemplateURN = ""

		expectedErrorMsg := "error with [template_urn : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"template_urn": "cannot be empty",
			},
		}

		actualError := tagService.Validate(&t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if tag values are nil", func() {
		t := s.buildTag()
		t.TemplateURN = ""

		expectedErrorMsg := "error with [template_urn : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"template_urn": "cannot be empty",
			},
		}

		actualError := tagService.Validate(&t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if tag values contains zero element", func() {
		t := s.buildTag()
		t.TagValues = []tag.TagValue{}

		expectedErrorMsg := "error with [tag_values : must be at least 1]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"tag_values": "must be at least 1",
			},
		}

		actualError := tagService.Validate(&t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if tag value field ID is zero", func() {
		t := s.buildTag()
		t.TagValues[1].FieldID = 0

		expectedErrorMsg := "error with [tag_values[1].field_id : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"tag_values[1].field_id": "cannot be empty",
			},
		}

		actualError := tagService.Validate(&t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if tag value field value is nil", func() {
		t := s.buildTag()
		t.TagValues[1].FieldValue = nil

		expectedErrorMsg := "error with [tag_values[1].field_value : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"tag_values[1].field_value": "cannot be empty",
			},
		}

		actualError := tagService.Validate(&t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})
}

func (s *ServiceTestSuite) TestCreate() {
	ctx := context.TODO()

	s.Run("should return error if value validations return error", func() {
		s.Setup()
		t := s.buildTag()
		t.AssetID = ""

		expectedErrorMsg := "error with [asset_id : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"asset_id": "cannot be empty",
			},
		}

		actualError := s.tagService.CreateTag(ctx, &t)
		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if error retrieving template", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return(nil, errors.New("random error"))

		err := s.tagService.CreateTag(ctx, &t)
		s.Error(err)
	})

	s.Run("should return error if specified field is not part of the template", func() {
		s.Setup()
		t := s.buildTag()
		t.TagValues[0].FieldID = 50

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields[0].id : not part of template [governance_policy]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].id": fmt.Sprintf("not part of template [%s]", t.TemplateURN),
			},
		}

		actualError := s.tagService.CreateTag(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if required field is not passed", func() {
		s.Setup()
		t := s.buildTag()
		t.TagValues = t.TagValues[:1]

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields[1].id : required by template [governance_policy]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[1].id": fmt.Sprintf("required by template [%s]", t.TemplateURN),
			},
		}

		actualError := s.tagService.CreateTag(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))

	})

	s.Run("should return error if passed value is not parsable", func() {
		t := s.buildTag()
		t.TagValues[1].FieldValue = "hello"

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields[1].value : template [governance_policy] on field [2] should be boolean]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[1].value": fmt.Sprintf("template [%s] on field [%d] should be boolean", t.TemplateURN, 2),
			},
		}

		actualError := s.tagService.CreateTag(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return repository error if repository met error", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Create(mock.Anything, &t).Return(errors.New("random error"))

		err := s.tagService.CreateTag(ctx, &t)
		s.Error(err)
	})

	s.Run("should return repository nil if repository not error", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Create(mock.Anything, &t).Return(nil)

		actualError := s.tagService.CreateTag(ctx, &t)

		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestGetByAsset() {
	ctx := context.TODO()

	s.Run("should return tags and error based on the repository", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		expectedTag := []tag.Tag{t}

		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID: t.AssetID,
		}).Return(expectedTag, nil)

		actualTag, actualError := s.tagService.GetTagsByAssetID(ctx, t.AssetID)

		s.EqualValues(expectedTag, actualTag)
		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestFindByAssetAndTemplate() {
	ctx := context.TODO()

	s.Run("should return error if error retrieving template", func() {
		s.Setup()

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return(nil, errors.New("random error"))

		_, err := s.tagService.FindTagByAssetIDAndTemplateURN(ctx, uuid.NewString(), template.URN)
		s.Error(err)
	})

	s.Run("should return nil and error if tag is not found", func() {
		s.Setup()
		var assetID string = uuid.NewString()
		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID:     assetID,
			TemplateURN: template.URN,
		}).Return([]tag.Tag{}, nil)

		_, err := s.tagService.FindTagByAssetIDAndTemplateURN(ctx, assetID, template.URN)
		s.Equal(err.Error(), tag.NotFoundError{
			AssetID:  assetID,
			Template: template.URN,
		}.Error())
	})

	s.Run("should return tag and nil if tag is found", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID:     t.AssetID,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)

		expectedTag := t

		actualTag, actualError := s.tagService.FindTagByAssetIDAndTemplateURN(ctx, t.AssetID, template.URN)

		s.EqualValues(expectedTag, actualTag)
		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	ctx := context.TODO()

	s.Run("should return error if value validations return error", func() {
		s.Setup()
		t := s.buildTag()
		t.AssetID = ""

		expectedErrorMsg := "error with [asset_id : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"asset_id": "cannot be empty",
			},
		}
		actualError := s.tagService.UpdateTag(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if error retrieving template", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return(nil, errors.New("random error"))

		err := s.tagService.UpdateTag(ctx, &t)
		s.Error(err)
	})

	s.Run("should return not found error if tag could not be found", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()

		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID:     t.AssetID,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{}, nil)

		err := s.tagService.UpdateTag(ctx, &t)

		s.Error(err)
		s.ErrorIs(err, tag.NotFoundError{AssetID: t.AssetID, Template: t.TemplateURN})
	})

	s.Run("should return error if specified field is not part of the template", func() {
		s.Setup()
		t := s.buildTag()
		t.TagValues[0].FieldID = 50

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID:     t.AssetID,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)

		expectedErrorMsg := "error with [fields[0].id : not part of template [governance_policy]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].id": fmt.Sprintf("not part of template [%s]", t.TemplateURN),
			},
		}

		actualError := s.tagService.UpdateTag(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if passed value is not parsable", func() {
		t := s.buildTag()
		t.TagValues[1].FieldValue = "hello"

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID:     t.AssetID,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)

		expectedErrorMsg := "error with [fields[1].value : template [governance_policy] on field [2] should be boolean]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[1].value": fmt.Sprintf("template [%s] on field [%d] should be boolean", t.TemplateURN, 2),
			},
		}

		actualError := s.tagService.UpdateTag(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return repository error if repository met error", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID:     t.AssetID,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)
		s.repository.EXPECT().Update(mock.Anything, &t).Return(errors.New("random error"))

		err := s.tagService.UpdateTag(ctx, &t)
		s.Error(err)
	})

	s.Run("should return repository nil if repository not error", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Read(mock.Anything, tag.Tag{
			AssetID:     t.AssetID,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)
		s.repository.EXPECT().Update(mock.Anything, &t).Return(nil)

		actualError := s.tagService.UpdateTag(ctx, &t)

		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestDelete() {
	ctx := context.TODO()

	s.Run("should return error if error retrieving template", func() {
		s.Setup()

		templateURN := "template-urn"
		s.templateRepo.EXPECT().Read(mock.Anything, templateURN).Return(nil, errors.New("random error"))

		err := s.tagService.DeleteTag(ctx, uuid.NewString(), templateURN)
		s.Error(err)
	})

	s.Run("should return error if repository error", func() {
		s.Setup()
		assetID := uuid.NewString()

		template := s.buildTemplate()
		s.templateRepo.EXPECT().Read(mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Delete(mock.Anything, tag.Tag{
			AssetID:     assetID,
			TemplateURN: template.URN,
		}).Return(errors.New("random error"))

		err := s.tagService.DeleteTag(ctx, assetID, template.URN)
		s.Error(err)
	})
}

func (s *ServiceTestSuite) buildTemplate() tag.Template {
	return tag.Template{
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
				URN:         "is_encrypted",
				DisplayName: "Is Encrypted?",
				Description: "Specify whether this asset is encrypted or not.",
				DataType:    "boolean",
				Required:    true,
			},
			{
				ID:          3,
				URN:         "owner",
				DisplayName: "name of owner",
				Description: "name of owner of this asset",
				DataType:    "string",
				Required:    false,
			},
			{
				ID:          4,
				URN:         "date_created",
				DisplayName: "date of creation?",
				Description: "date when asset was created",
				DataType:    "datetime",
				Required:    false,
			},
			{
				ID:          5,
				URN:         "no_of_records",
				DisplayName: "no of records",
				Description: "record count for the asset",
				DataType:    "double",
				Required:    false,
			},
		},
	}
}

func (s *ServiceTestSuite) buildTag() tag.Tag {
	return tag.Tag{
		AssetID:             uuid.NewString(),
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
				FieldValue:       true,
				FieldURN:         "is_encrypted",
				FieldDisplayName: "Is Encrypted?",
				FieldDescription: "Specify whether this asset is encrypted or not.",
				FieldDataType:    "boolean",
				FieldRequired:    true,
			},
			{
				FieldID:          3,
				FieldValue:       "john doe",
				FieldURN:         "owner",
				FieldDisplayName: "name of owner",
				FieldDescription: "name of owner of this asset",
				FieldDataType:    "string",
				FieldRequired:    false,
			},
			{
				FieldID:          4,
				FieldValue:       carbon.Parse("2020-12-31").ToRfc3339String(),
				FieldURN:         "date_created",
				FieldDisplayName: "date of creation?",
				FieldDescription: "date when asset was created",
				FieldDataType:    "datetime",
				FieldRequired:    false,
			},
			{
				FieldID:          5,
				FieldValue:       "91.0",
				FieldURN:         "no_of_records",
				FieldDisplayName: "no of records",
				FieldDescription: "record count for the asset",
				FieldDataType:    "double",
				FieldRequired:    false,
			},
		},
	}
}

func TestService(t *testing.T) {
	suite.Run(t, &ServiceTestSuite{})
}
