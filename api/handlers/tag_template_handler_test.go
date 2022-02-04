package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/odpf/salt/log"

	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/tag"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TagTemplateHandlerTestSuite struct {
	suite.Suite
	handler            *handlers.TagTemplateHandler
	templateRepository *mocks.TagTemplateRepository
	recorder           *httptest.ResponseRecorder
	logger             log.Noop
}

func (s *TagTemplateHandlerTestSuite) TestNewHandler() {
	s.Run("should return handler and nil if service is not nil", func() {
		service := &tag.TemplateService{}

		actualHandler := handlers.NewTagTemplateHandler(&s.logger, service)
		s.NotNil(actualHandler)
	})
}

func (s *TagTemplateHandlerTestSuite) Setup() {
	s.templateRepository = new(mocks.TagTemplateRepository)
	service := tag.NewTemplateService(s.templateRepository)

	s.handler = handlers.NewTagTemplateHandler(&s.logger, service)
	s.recorder = httptest.NewRecorder()
}

func (s *TagTemplateHandlerTestSuite) TestCreate() {
	s.Run("should return status bad request and error if body cannot be unmarshalled", func() {
		s.Setup()
		body := "invalid_body"
		request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		s.Require().NoError(err)

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"invalid character 'i' looking for beginning of value\"}\n"

		s.handler.Create(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return 409 for duplicate template", func() {
		s.Setup()
		template := s.buildTemplate()
		body, err := json.Marshal(template)
		s.Require().NoError(err)
		s.templateRepository.On("Read", mock.Anything, template.URN).Return(nil, nil)
		s.templateRepository.On("Create", mock.Anything, &template).Return(tag.DuplicateTemplateError{URN: template.URN})
		request, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))

		s.handler.Create(s.recorder, request)
		s.Equal(http.StatusConflict, s.recorder.Result().StatusCode)
	})

	s.Run("should return 500 if found error during insert", func() {
		s.Setup()
		template := s.buildTemplate()
		body, err := json.Marshal(template)
		s.Require().NoError(err)
		s.templateRepository.On("Read", mock.Anything, template.URN).Return(nil, nil)
		s.templateRepository.On("Create", mock.Anything, &template).Return(errors.New("unexpected error during insert"))
		request, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))

		s.handler.Create(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status created and domain is inserted if found no error", func() {
		s.Setup()
		originalDomainTemplate := s.buildTemplate()
		body, err := json.Marshal(originalDomainTemplate)
		s.Require().NoError(err)
		s.templateRepository.On("Read", mock.Anything, originalDomainTemplate.URN).Return(nil, nil)
		s.templateRepository.On("Create", mock.Anything, &originalDomainTemplate).Return(nil)
		request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)

		expectedStatusCode := http.StatusCreated
		rsp, err := json.Marshal(originalDomainTemplate)
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.Create(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *TagTemplateHandlerTestSuite) TestIndex() {
	s.Run("should return status unprocessible entity and error if found unexpected error", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/templates?urn=governance_policy", nil)
		s.Require().NoError(err)
		s.templateRepository.On("Read", mock.Anything, "governance_policy").Return(nil, errors.New("unexpected error"))

		s.handler.Index(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status ok and templates if found template based on the query", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/templates?urn=governance_policy", nil)
		s.Require().NoError(err)
		recordDomainTemplate := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, "governance_policy").Return([]tag.Template{recordDomainTemplate}, nil)

		expectedStatusCode := http.StatusOK
		rsp, err := json.Marshal([]tag.Template{recordDomainTemplate})
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.Index(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return all templates if no urn query params", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/templates", nil)
		s.Require().NoError(err)
		recordDomainTemplate := s.buildTemplate()
		s.templateRepository.On("ReadAll", mock.Anything).Return([]tag.Template{recordDomainTemplate}, nil)

		expectedStatusCode := http.StatusOK
		rsp, err := json.Marshal([]tag.Template{recordDomainTemplate})
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.Index(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

}

func (s *TagTemplateHandlerTestSuite) TestUpdate() {
	s.Run("should return status bad request error and its message if urn is empty", func() {
		s.Setup()
		template := s.buildTemplate()
		body, err := json.Marshal(template)
		s.Require().NoError(err)
		var templateURN string = ""
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"template urn is empty\"}\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if body cannot be unmarshalled", func() {
		s.Setup()
		body := "invalid_body"
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"invalid character 'i' looking for beginning of value\"}\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return 404 if template does not exist", func() {
		s.Setup()
		template := s.buildTemplate()
		body, err := json.Marshal(template)
		s.Require().NoError(err)
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})
		s.templateRepository.On("Read", mock.Anything, templateURN).Return([]tag.Template{}, nil)

		s.handler.Update(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return 422 if there is validation error", func() {
		s.Setup()
		template := s.buildTemplate()
		body, err := json.Marshal(template)
		s.Require().NoError(err)
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})
		s.templateRepository.On("Read", mock.Anything, templateURN).Return(nil, tag.ValidationError{Err: errors.New("validation error")})

		s.handler.Update(s.recorder, request)
		s.Equal(http.StatusUnprocessableEntity, s.recorder.Result().StatusCode)
	})

	s.Run("should return 500 if encountered error during update", func() {
		s.Setup()
		template := s.buildTemplate()
		body, err := json.Marshal(template)
		s.Require().NoError(err)
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})
		s.templateRepository.On("Read", mock.Anything, templateURN).Return(nil, errors.New("unexpected error"))

		s.handler.Update(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status ok and its message if successfully updated", func() {
		s.Setup()
		template := s.buildTemplate()
		body, err := json.Marshal(template)
		s.Require().NoError(err)
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": template.URN,
		})

		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil).Once()

		s.templateRepository.On("Update", mock.Anything, template.URN, &template).Run(func(args mock.Arguments) {
			template.UpdatedAt = time.Now()
		}).Return(nil)

		expectedStatusCode := http.StatusOK
		rsp, err := json.Marshal(template)
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *TagTemplateHandlerTestSuite) TestFind() {
	s.Run("should return status bad request error and its message if urn is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": "",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"template urn is empty\"}\n"

		s.handler.Find(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if urn is empty", func() {
		s.Setup()
		var templateURN string = ""
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"template urn is empty\"}\n"

		s.handler.Find(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status 404 and its message if template is not found", func() {
		s.Setup()
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})
		s.templateRepository.On("Read", mock.Anything, templateURN).Return([]tag.Template{}, nil)

		s.handler.Find(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return status ok and template if domain template is found", func() {
		s.Setup()
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})
		template := s.buildTemplate()

		s.templateRepository.On("Read", mock.Anything, templateURN).Return([]tag.Template{template}, nil)

		expectedStatusCode := http.StatusOK
		rsp, err := json.Marshal(template)
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.Find(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *TagTemplateHandlerTestSuite) TestDelete() {
	s.Run("should return status bad request error and its message if urn is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": "",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"template urn is empty\"}\n"

		s.handler.Delete(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return 404 if template is not found", func() {
		s.Setup()
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})
		s.templateRepository.On("Delete", mock.Anything, templateURN).Return(tag.TemplateNotFoundError{URN: templateURN})

		s.handler.Delete(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return status no content and template if domain template is found", func() {
		s.Setup()
		var templateURN string = "governance_policy"
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"template_urn": templateURN,
		})
		s.templateRepository.On("Delete", mock.Anything, templateURN).Return(nil)

		s.handler.Delete(s.recorder, request)
		s.Equal(http.StatusNoContent, s.recorder.Result().StatusCode)
	})
}

func (s *TagTemplateHandlerTestSuite) buildTemplate() tag.Template {
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

func TestTagTemplateHandler(t *testing.T) {
	suite.Run(t, &TagTemplateHandlerTestSuite{})
}
