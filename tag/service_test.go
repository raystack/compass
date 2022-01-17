package tag_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/odpf/columbus/tag"
	"github.com/odpf/columbus/tag/mocks"
	"github.com/odpf/columbus/tag/validator"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	tagService   *tag.Service
	templateRepo *mocks.TemplateRepository
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
	s.templateRepo = &mocks.TemplateRepository{}

	templateService := tag.NewTemplateService(s.templateRepo)
	s.tagService = tag.NewService(s.repository, templateService)
}

func (s *ServiceTestSuite) TestValidate() {
	repository := &mocks.TagRepository{}
	templateService := tag.NewTemplateService(&mocks.TemplateRepository{})
	tagService := tag.NewService(repository, templateService)

	s.Run("should return error if record URN is empty", func() {
		t := s.buildTag()
		t.RecordURN = ""

		expectedErrorMsg := "error with [record_urn : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"record_urn": "cannot be empty",
			},
		}

		actualError := tagService.Validate(&t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if record Type is empty", func() {
		t := s.buildTag()
		t.RecordType = ""

		expectedErrorMsg := "error with [record_type : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"record_type": "cannot be empty",
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
		t.RecordURN = ""

		expectedErrorMsg := "error with [record_urn : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"record_urn": "cannot be empty",
			},
		}

		actualError := s.tagService.Create(ctx, &t)
		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if error retrieving template", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return(nil, errors.New("random error"))

		err := s.tagService.Create(ctx, &t)
		s.Error(err)
	})

	s.Run("should return error if specified field is not part of the template", func() {
		s.Setup()
		t := s.buildTag()
		t.TagValues[0].FieldID = 5

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields[0].id : not part of template [governance_policy]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].id": fmt.Sprintf("not part of template [%s]", t.TemplateURN),
			},
		}

		actualError := s.tagService.Create(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if required field is not passed", func() {
		s.Setup()
		t := s.buildTag()
		t.TagValues = t.TagValues[:1]

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields[1].id : required by template [governance_policy]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[1].id": fmt.Sprintf("required by template [%s]", t.TemplateURN),
			},
		}

		actualError := s.tagService.Create(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))

	})

	s.Run("should return error if passed value is not parsable", func() {
		t := s.buildTag()
		t.TagValues[1].FieldValue = "hello"

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)

		expectedErrorMsg := "error with [fields[1].value : template [governance_policy] on field [2] should be boolean]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[1].value": fmt.Sprintf("template [%s] on field [%d] should be boolean", t.TemplateURN, 2),
			},
		}

		actualError := s.tagService.Create(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return repository error if repository met error", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Create", mock.Anything, &t).Return(errors.New("random error"))

		err := s.tagService.Create(ctx, &t)
		s.Error(err)
	})

	s.Run("should return repository nil if repository not error", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Create", mock.Anything, &t).Return(nil)

		actualError := s.tagService.Create(ctx, &t)

		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestGetByRecord() {
	ctx := context.TODO()

	s.Run("should return tags and error based on the repository", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		expectedTag := []tag.Tag{t}

		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType: t.RecordType,
			RecordURN:  t.RecordURN,
		}).Return(expectedTag, nil)

		actualTag, actualError := s.tagService.GetByRecord(ctx, t.RecordType, t.RecordURN)

		s.EqualValues(expectedTag, actualTag)
		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestFindByRecordAndTemplate() {
	ctx := context.TODO()

	s.Run("should return error if error retrieving template", func() {
		s.Setup()

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return(nil, errors.New("random error"))

		_, err := s.tagService.FindByRecordAndTemplate(ctx, "record-type", "record-urn", template.URN)
		s.Error(err)
	})

	s.Run("should return nil and error if tag is not found", func() {
		s.Setup()
		var recordType string = "record-type"
		var urn string = "record-urn"
		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType:  recordType,
			RecordURN:   urn,
			TemplateURN: template.URN,
		}).Return([]tag.Tag{}, nil)

		_, err := s.tagService.FindByRecordAndTemplate(ctx, recordType, urn, template.URN)
		s.ErrorIs(err, tag.NotFoundError{
			Type:     recordType,
			URN:      urn,
			Template: template.URN,
		})
	})

	s.Run("should return tag and nil if tag is found", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)

		expectedTag := t

		actualTag, actualError := s.tagService.FindByRecordAndTemplate(ctx, t.RecordType, t.RecordURN, template.URN)

		s.EqualValues(expectedTag, actualTag)
		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	ctx := context.TODO()

	s.Run("should return error if value validations return error", func() {
		s.Setup()
		t := s.buildTag()
		t.RecordURN = ""

		expectedErrorMsg := "error with [record_urn : cannot be empty]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"record_urn": "cannot be empty",
			},
		}
		actualError := s.tagService.Update(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if error retrieving template", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return(nil, errors.New("random error"))

		err := s.tagService.Update(ctx, &t)
		s.Error(err)
	})

	s.Run("should return not found error if tag could not be found", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()

		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{}, nil)

		err := s.tagService.Update(ctx, &t)

		s.Error(err)
		s.ErrorIs(err, tag.NotFoundError{URN: t.RecordURN, Type: t.RecordType, Template: t.TemplateURN})
	})

	s.Run("should return error if specified field is not part of the template", func() {
		s.Setup()
		t := s.buildTag()
		t.TagValues[0].FieldID = 5

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)

		expectedErrorMsg := "error with [fields[0].id : not part of template [governance_policy]]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[0].id": fmt.Sprintf("not part of template [%s]", t.TemplateURN),
			},
		}

		actualError := s.tagService.Update(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return error if passed value is not parsable", func() {
		t := s.buildTag()
		t.TagValues[1].FieldValue = "hello"

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)

		expectedErrorMsg := "error with [fields[1].value : template [governance_policy] on field [2] should be boolean]"
		expectedFieldError := tag.ValidationError{
			validator.FieldError{
				"fields[1].value": fmt.Sprintf("template [%s] on field [%d] should be boolean", t.TemplateURN, 2),
			},
		}

		actualError := s.tagService.Update(ctx, &t)

		s.EqualError(actualError, expectedErrorMsg)
		s.EqualValues(expectedFieldError, actualError.(tag.ValidationError))
	})

	s.Run("should return repository error if repository met error", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)
		s.repository.On("Update", mock.Anything, &t).Return(errors.New("random error"))

		err := s.tagService.Update(ctx, &t)
		s.Error(err)
	})

	s.Run("should return repository nil if repository not error", func() {
		s.Setup()
		t := s.buildTag()

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)
		s.repository.On("Update", mock.Anything, &t).Return(nil)

		actualError := s.tagService.Update(ctx, &t)

		s.NoError(actualError)
	})
}

func (s *ServiceTestSuite) TestDelete() {
	ctx := context.TODO()

	s.Run("should return error if error retrieving template", func() {
		s.Setup()

		templateURN := "template-urn"
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(templateURN)).Return(nil, errors.New("random error"))

		err := s.tagService.Delete(ctx, "record-type", "record-urn", templateURN)
		s.Error(err)
	})

	s.Run("should return error if repository error", func() {
		s.Setup()
		recordType := "sample-type"
		recordURN := "sample-urn"

		template := s.buildTemplate()
		s.templateRepo.On("Read", mock.Anything, s.templateQuery(template.URN)).Return([]tag.Template{template}, nil)
		s.repository.On("Delete", mock.Anything, tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: template.URN,
		}).Return(errors.New("random error"))

		err := s.tagService.Delete(ctx, recordType, recordURN, template.URN)
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
				Description: "The classification of this record",
				DataType:    "enumerated",
				Required:    true,
				Options:     []string{"Public", "Restricted"},
			},
			{
				ID:          2,
				URN:         "is_encrypted",
				DisplayName: "Is Encrypted?",
				Description: "Specify whether this record is encrypted or not.",
				DataType:    "boolean",
				Required:    true,
			},
		},
	}
}

func (s *ServiceTestSuite) buildTag() tag.Tag {
	return tag.Tag{
		RecordURN:           "record-urn",
		RecordType:          "sample-record-type",
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
				FieldValue:       true,
				FieldURN:         "is_encrypted",
				FieldDisplayName: "Is Encrypted?",
				FieldDescription: "Specify whether this record is encrypted or not.",
				FieldDataType:    "boolean",
				FieldRequired:    true,
			},
		},
	}
}

func (s *ServiceTestSuite) templateQuery(urn string) tag.Template {
	return tag.Template{URN: urn}
}

func TestService(t *testing.T) {
	suite.Run(t, &ServiceTestSuite{})
}
