package tag_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/odpf/compass/core/tag"
	"github.com/odpf/compass/core/tag/mocks"
	"github.com/odpf/compass/core/tag/validator"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TemplateServiceTestSuite struct {
	suite.Suite
	service    *tag.TemplateService
	repository *mocks.TagTemplateRepository
}

func (s *TemplateServiceTestSuite) TestNewTemplateService() {
	s.Run("should return service and nil no error found", func() {
		r := &mocks.TagTemplateRepository{}

		actualService := tag.NewTemplateService(r)
		s.NotNil(actualService)
	})
}

func (s *TemplateServiceTestSuite) Setup() {
	s.repository = &mocks.TagTemplateRepository{}
	var err error
	s.service = tag.NewTemplateService(s.repository)
	s.Require().NoError(err)
}

func (s *TemplateServiceTestSuite) TestValidate() {
	r := &mocks.TagTemplateRepository{}
	service := tag.NewTemplateService(r)

	s.Run("should return error if urn is empty", func() {
		template := s.buildTemplate()
		template.URN = ""

		expectedErrorMsg := "error with [urn : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"urn": "cannot be empty",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if display name is empty", func() {
		template := s.buildTemplate()
		template.DisplayName = ""

		expectedErrorMsg := "error with [display_name : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"display_name": "cannot be empty",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if description is empty", func() {
		template := s.buildTemplate()
		template.Description = ""

		expectedErrorMsg := "error with [description : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"description": "cannot be empty",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields is nil", func() {
		template := s.buildTemplate()
		template.Fields = nil

		expectedErrorMsg := "error with [fields : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields": "cannot be empty",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields is empty", func() {
		template := s.buildTemplate()
		template.Fields = []tag.Field{}

		expectedErrorMsg := "error with [fields : must be at least 1]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields": "must be at least 1",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields urn is empty", func() {
		template := s.buildTemplate()
		template.Fields[0].URN = ""

		expectedErrorMsg := "error with [fields[0].urn : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].urn": "cannot be empty",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields display name is empty", func() {
		template := s.buildTemplate()
		template.Fields[0].DisplayName = ""

		expectedErrorMsg := "error with [fields[0].display_name : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].display_name": "cannot be empty",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields description is empty", func() {
		template := s.buildTemplate()
		template.Fields[0].Description = ""

		expectedErrorMsg := "error with [fields[0].description : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].description": "cannot be empty",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields data type is invalid", func() {
		template := s.buildTemplate()
		template.Fields[0].DataType = "Random_Type"

		expectedErrorMsg := "error with [fields[0].data_type : data_type must be one of [string double boolean enumerated datetime]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].data_type": "data_type must be one of [string double boolean enumerated datetime]",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields data type enumerated but options nil", func() {
		template := s.buildTemplate()
		template.Fields[0].Options = nil

		expectedErrorMsg := "error with [fields[0].options : cannot be empty with data_type [enumerated]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].options": "cannot be empty with data_type [enumerated]",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields data type enumerated but options empty", func() {
		template := s.buildTemplate()
		template.Fields[0].Options = []string{}

		expectedErrorMsg := "error with [fields[0].options : cannot be empty with data_type [enumerated]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].options": "cannot be empty with data_type [enumerated]",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if fields data type enumerated but options contains empty", func() {
		template := s.buildTemplate()
		template.Fields[0].Options = []string{
			"Team Owner", "", "Governor Email",
		}

		expectedErrorMsg := "error with [fields[0].options : cannot contain empty element]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].options": "cannot contain empty element",
			},
		}

		actualError := service.Validate(template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return nil if fields data type not enumerated and options empty", func() {
		template := s.buildTemplate()
		template.Fields[0].Options = nil
		template.Fields[0].DataType = "string"

		actualError := service.Validate(template)

		s.NoError(actualError)
	})
}

func (s *TemplateServiceTestSuite) TestCreate() {
	ctx := context.TODO()

	s.Run("should return error if domain template is nil", func() {
		s.Setup()

		err := s.service.CreateTemplate(ctx, nil)
		s.Error(err)
	})

	s.Run("should return error if error encountered during validation", func() {
		s.Setup()
		template := s.buildTemplate()
		template.Description = ""

		expectedErrorMsg := "error with [description : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"description": "cannot be empty",
			},
		}

		actualError := s.service.CreateTemplate(ctx, &template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if error encountered when checking for duplication", func() {
		s.Setup()
		template := s.buildTemplate()
		s.repository.EXPECT().Read(ctx, template.URN).Return(nil, errors.New("unexpected error"))

		err := s.service.CreateTemplate(ctx, &template)
		s.Error(err)
	})

	s.Run("should return error if template specified by the urn already exists", func() {
		s.Setup()
		template := s.buildTemplate()
		s.repository.EXPECT().Read(ctx, template.URN).Return([]tag.Template{{}}, nil)

		err := s.service.CreateTemplate(ctx, &template)
		s.Equal(tag.DuplicateTemplateError{URN: template.URN}, err)
	})

	s.Run("should return error if found error during create", func() {
		s.Setup()
		now := time.Now()
		originalDomainTemplate := s.buildTemplate()
		referenceDomainTemplate := s.buildTemplate()
		referenceDomainTemplate.CreatedAt = now

		s.repository.EXPECT().Read(ctx, originalDomainTemplate.URN).Return([]tag.Template{}, nil)
		s.repository.EXPECT().Create(ctx, &originalDomainTemplate).Return(errors.New("unexpected error"))

		err := s.service.CreateTemplate(ctx, &originalDomainTemplate)
		s.Error(err)
	})

	s.Run("should return nil if success in create", func() {
		s.Setup()
		now := time.Now()
		originalDomainTemplate := s.buildTemplate()
		referenceDomainTemplate := s.buildTemplate()
		referenceDomainTemplate.CreatedAt = now

		s.repository.EXPECT().Read(ctx, originalDomainTemplate.URN).Return([]tag.Template{}, nil)
		s.repository.EXPECT().Create(ctx, &originalDomainTemplate).Run(func(ctx context.Context, template *tag.Template) {
			template.CreatedAt = now
		}).Return(nil)

		actualError := s.service.CreateTemplate(ctx, &originalDomainTemplate)

		s.NoError(actualError)
		s.EqualValues(referenceDomainTemplate, originalDomainTemplate)
	})
}

func (s *TemplateServiceTestSuite) TestIndex() {
	ctx := context.TODO()

	s.Run("should return nil and error if encountered unexpected error during read", func() {
		s.Setup()
		template := s.buildTemplate()
		s.repository.EXPECT().Read(ctx, template.URN).Return(nil, errors.New("unexpected error"))

		_, err := s.service.GetTemplates(ctx, template.URN)
		s.Error(err)
	})

	s.Run("should return nil and error for error in readall with empty urn", func() {
		s.Setup()
		s.repository.EXPECT().ReadAll(ctx).Return(nil, errors.New("unexpected error"))

		_, err := s.service.GetTemplates(ctx, "")
		s.Error(err)
	})

	s.Run("should return domain templates and nil if no error found", func() {
		s.Setup()
		template := s.buildTemplate()
		s.repository.EXPECT().Read(ctx, template.URN).Return([]tag.Template{template}, nil)

		expectedTemplate := []tag.Template{template}

		actualTemplate, actualError := s.service.GetTemplates(ctx, template.URN)

		s.EqualValues(expectedTemplate, actualTemplate)
		s.NoError(actualError)
	})

	s.Run("should return domain templates and nil if no error found with empty urn", func() {
		s.Setup()
		template := s.buildTemplate()
		s.repository.EXPECT().ReadAll(ctx).Return([]tag.Template{template}, nil)

		expectedTemplate := []tag.Template{template}

		actualTemplate, actualError := s.service.GetTemplates(ctx, "")

		s.EqualValues(expectedTemplate, actualTemplate)
		s.NoError(actualError)
	})
}

func (s *TemplateServiceTestSuite) TestUpdate() {
	ctx := context.TODO()

	s.Run("should return error if domain template is nil", func() {
		s.Setup()
		var template *tag.Template = nil

		err := s.service.UpdateTemplate(ctx, "", template)
		s.EqualError(err, "template is nil")
	})

	s.Run("should return error if error encountered during validation", func() {
		s.Setup()
		template := s.buildTemplate()
		template.Description = ""

		expectedErrorMsg := "error with [description : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"description": "cannot be empty",
			},
		}

		actualError := s.service.UpdateTemplate(ctx, template.URN, &template)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if encountered unexpected error during read for existence", func() {
		s.Setup()
		template := s.buildTemplate()

		s.repository.EXPECT().Read(ctx, template.URN).Return(nil, errors.New("unexpected error"))

		err := s.service.UpdateTemplate(ctx, template.URN, &template)
		s.Error(err)
	})

	s.Run("should return error if field is not part of template", func() {
		s.Setup()
		template := s.buildTemplate()
		newTemplate := s.buildTemplate()
		newTemplate.Fields[0].ID = 99

		s.repository.EXPECT().Read(ctx, newTemplate.URN).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields.[0].id : [99] is not part of the template]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields.[0].id": fmt.Sprintf("[%d] is not part of the template",
					newTemplate.Fields[0].ID,
				),
			},
		}

		actualError := s.service.UpdateTemplate(ctx, newTemplate.URN, &newTemplate)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if trying to update field urn that already exist", func() {
		s.Setup()
		template := s.buildTemplate()
		newTemplate := s.buildTemplate()
		newTemplate.Fields[0].URN = template.Fields[1].URN

		s.repository.EXPECT().Read(ctx, newTemplate.URN).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields.[0].urn : [team_custodianr] already exists within the template]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields.[0].urn": fmt.Sprintf("[%s] already exists within the template",
					newTemplate.Fields[0].URN,
				),
			},
		}

		actualError := s.service.UpdateTemplate(ctx, template.URN, &newTemplate)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if found error during update", func() {
		s.Setup()
		template := s.buildTemplate()
		newTemplate := s.buildTemplate()

		s.repository.EXPECT().Read(ctx, newTemplate.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Update(ctx, newTemplate.URN, &newTemplate).Return(errors.New("unexpected error"))

		err := s.service.UpdateTemplate(ctx, template.URN, &newTemplate)
		s.Error(err)
	})

	s.Run("should return nil if repository update is success", func() {
		s.Setup()
		template := s.buildTemplate()
		newTemplate := s.buildTemplate()

		s.repository.EXPECT().Read(ctx, template.URN).Return([]tag.Template{template}, nil)
		s.repository.EXPECT().Update(ctx, newTemplate.URN, &newTemplate).Run(func(ctx context.Context, templateURN string, template *tag.Template) {
			template.UpdatedAt = time.Now()
		}).Return(nil)

		actualError := s.service.UpdateTemplate(ctx, template.URN, &newTemplate)
		s.NoError(actualError)
	})
}

func (s *TemplateServiceTestSuite) TestFind() {
	ctx := context.TODO()

	s.Run("should return empty and error if found unexpected error", func() {
		s.Setup()
		var urn string = "sample-urn"
		s.repository.EXPECT().Read(ctx, urn).Return(nil, errors.New("unexpected error"))

		_, err := s.service.GetTemplate(ctx, urn)
		s.Error(err)
	})

	s.Run("should return not found error if template is not found", func() {
		s.Setup()
		var urn string = "sample-urn"
		s.repository.EXPECT().Read(ctx, urn).Return([]tag.Template{}, nil)

		_, err := s.service.GetTemplate(ctx, urn)
		s.Error(err)
		s.Equal(err.Error(), tag.TemplateNotFoundError{URN: urn}.Error())
	})

	s.Run("should return domain template and nil if record is found", func() {
		s.Setup()
		var urn string = "sample-urn"
		template := s.buildTemplate()
		s.repository.EXPECT().Read(ctx, urn).Return([]tag.Template{template}, nil)

		expectedTemplate := template

		actualTemplate, actualError := s.service.GetTemplate(ctx, urn)

		s.EqualValues(expectedTemplate, actualTemplate)
		s.NoError(actualError)
	})
}

func (s *TemplateServiceTestSuite) TestDelete() {
	ctx := context.TODO()

	s.Run("should return error if encountered unexpected error during delete", func() {
		s.Setup()
		var urn string = "sample-urn"
		s.repository.EXPECT().Delete(ctx, mock.Anything).Return(errors.New("unexpected error"))

		actualError := s.service.DeleteTemplate(ctx, urn)

		s.Error(actualError)
		s.EqualValues("error deleting template: unexpected error", actualError.Error())
	})

	s.Run("should return delete result from repository", func() {
		s.Setup()
		var urn string = "sample-urn"
		s.repository.EXPECT().Delete(ctx, mock.Anything).Return(nil).Once()
		s.repository.EXPECT().Delete(ctx, mock.Anything).Return(errors.New("unexpected error")).Once()

		actualError1 := s.service.DeleteTemplate(ctx, urn)
		actualError2 := s.service.DeleteTemplate(ctx, urn)

		s.NoError(actualError1)
		s.Error(actualError2)
	})
}

func (s *TemplateServiceTestSuite) buildTemplate() tag.Template {
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
			{
				ID:          2,
				URN:         "team_custodianr",
				DisplayName: "Team Custodian",
				Description: "Custodian of the resource.",
				DataType:    "string",
				Required:    false,
			},
		},
	}
}

func TestTemplateService(t *testing.T) {
	suite.Run(t, &TemplateServiceTestSuite{})
}
