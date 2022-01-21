package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/odpf/columbus/api/handlers"
	libmock "github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/tag"
	"github.com/odpf/columbus/tag/mocks"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TagHandlerTestSuite struct {
	suite.Suite
	handler            *handlers.TagHandler
	tagRepository      *mocks.TagRepository
	templateRepository *mocks.TemplateRepository
	recorder           *httptest.ResponseRecorder
}

func (s *TagHandlerTestSuite) TestNewHandler() {
	s.Run("should return handler and nil if service is not nil", func() {
		actualHandler := handlers.NewTagHandler(new(libmock.Logger), &tag.Service{})

		s.NotNil(actualHandler)
	})
}

func (s *TagHandlerTestSuite) Setup() {
	s.tagRepository = new(mocks.TagRepository)
	s.templateRepository = new(mocks.TemplateRepository)
	templateService := tag.NewTemplateService(s.templateRepository)
	service := tag.NewService(s.tagRepository, templateService)

	s.handler = handlers.NewTagHandler(new(libmock.Logger), service)
	s.recorder = httptest.NewRecorder()
}

func (s *TagHandlerTestSuite) TestCreate() {
	s.Run("should return status bad request and error if body cannot be unmarshalled", func() {
		s.Setup()
		body := "invalid_body"
		request, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"invalid character 'i' looking for beginning of value\"}\n"

		s.handler.Create(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return 404 if template does not exist", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Create", mock.Anything, &t).Return(tag.ErrTemplateNotFound{URN: t.TemplateURN})

		body, err := json.Marshal(t)
		s.Require().NoError(err)
		request, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))

		s.handler.Create(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return 422 if there is validation error", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Create", mock.Anything, &t).Return(tag.ErrValidation{Err: errors.New("validation error")})

		body, err := json.Marshal(t)
		s.Require().NoError(err)
		request, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))

		s.handler.Create(s.recorder, request)
		s.Equal(http.StatusUnprocessableEntity, s.recorder.Result().StatusCode)
	})

	s.Run("should return 500 if found error during insert", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Create", mock.Anything, &t).Return(errors.New("unexpected error during insert"))

		body, err := json.Marshal(t)
		s.Require().NoError(err)
		request, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))

		s.handler.Create(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return 409 if found duplicated record during insert", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Create", mock.Anything, &t).Return(tag.ErrDuplicate{})

		body, err := json.Marshal(t)
		s.Require().NoError(err)
		request, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))

		s.handler.Create(s.recorder, request)
		s.Equal(http.StatusConflict, s.recorder.Result().StatusCode)
	})

	s.Run("should return status created and domain is inserted if found no error", func() {
		s.Setup()
		originalDomainTag := s.buildTag()

		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Create", mock.Anything, &originalDomainTag).Return(nil)

		body, err := json.Marshal(originalDomainTag)
		s.Require().NoError(err)
		request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)

		s.handler.Create(s.recorder, request)

		rsp, err := json.Marshal(originalDomainTag)
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.Equal(http.StatusCreated, s.recorder.Result().StatusCode)
		s.Equal(expectedResponseBody, s.recorder.Body.String())
	})

}

func (s *TagHandlerTestSuite) TestGetByRecord() {
	s.Run("should return status bad request error and its message if record urn is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":       "sample-type",
			"record_urn": "",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"record urn is empty\"}\n"

		s.handler.GetByRecord(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
	s.Run("should return status bad request error and its message if type is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":       "",
			"record_urn": "sample-urn",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"type is empty\"}\n"

		s.handler.GetByRecord(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status unprocessible entity and error if found unexpected error", func() {
		s.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":       recordType,
			"record_urn": recordURN,
		})

		s.tagRepository.On("Read", mock.Anything, tag.Tag{RecordType: recordType, RecordURN: recordURN}).Return(nil, errors.New("unexpected error"))

		s.handler.GetByRecord(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status ok and tags for the specified record", func() {
		s.Setup()
		t := s.buildTag()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":       recordType,
			"record_urn": recordURN,
		})
		s.tagRepository.On("Read", mock.Anything, tag.Tag{RecordType: recordType, RecordURN: recordURN}).Return([]tag.Tag{t}, nil)

		expectedStatusCode := http.StatusOK
		rsp, err := json.Marshal([]tag.Tag{t})
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.GetByRecord(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *TagHandlerTestSuite) TestFindByRecordAndTemplate() {
	s.Run("should return status bad request error and its message if record urn is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         "sample-type",
			"record_urn":   "",
			"template_urn": "sample-template",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"record urn is empty\"}\n"

		s.handler.FindByRecordAndTemplate(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if type is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         "",
			"record_urn":   "sample-urn",
			"template_urn": "sample-template",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"type is empty\"}\n"

		s.handler.FindByRecordAndTemplate(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if template urn is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         "sample-type",
			"record_urn":   "sample-urn",
			"template_urn": "",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"template urn is empty\"}\n"

		s.handler.FindByRecordAndTemplate(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return 404 if template does not exist", func() {
		s.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		template := s.buildTemplate()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         recordType,
			"record_urn":   recordURN,
			"template_urn": template.URN,
		})

		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{}, tag.ErrTemplateNotFound{URN: template.URN})

		s.handler.FindByRecordAndTemplate(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return 404 if tag does not exist", func() {
		s.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		template := s.buildTemplate()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         recordType,
			"record_urn":   recordURN,
			"template_urn": template.URN,
		})

		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil)

		s.tagRepository.On("Read", mock.Anything, tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: template.URN,
		}).Return(nil, tag.ErrNotFound{
			URN:      recordURN,
			Type:     recordType,
			Template: template.URN,
		})

		s.handler.FindByRecordAndTemplate(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return 500 if found unexpected error", func() {
		s.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "template-urn"
		template := s.buildTemplate()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         recordType,
			"record_urn":   recordURN,
			"template_urn": templateURN,
		})

		s.templateRepository.On("Read", mock.Anything, templateURN).Return([]tag.Template{template}, nil)

		s.tagRepository.On("Read", mock.Anything, tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}).Return(nil, errors.New("unexpected error"))

		s.handler.FindByRecordAndTemplate(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status ok and tag", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         t.RecordType,
			"record_urn":   t.RecordURN,
			"template_urn": t.TemplateURN,
		})

		s.templateRepository.On("Read", mock.Anything, t.TemplateURN).Return([]tag.Template{template}, nil)

		s.tagRepository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)

		expectedStatusCode := http.StatusOK
		rsp, err := json.Marshal(t)
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.FindByRecordAndTemplate(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *TagHandlerTestSuite) TestUpdate() {
	s.Run("should return status internal server error and its message if service is nil", func() {
		s.Setup()
		handler := &handlers.TagHandler{}
		t := s.buildTag()
		body, _ := json.Marshal(t)
		var recordURN string = "sample-urn"
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"record_urn": recordURN,
		})

		handler.Update(s.recorder, request)

		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status bad request error and its message if record urn is empty", func() {
		s.Setup()
		t := s.buildTag()
		body, _ := json.Marshal(t)
		request, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		request = mux.SetURLVars(request, map[string]string{
			"type":         t.RecordType,
			"record_urn":   "",
			"template_urn": t.TemplateURN,
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"record urn is empty\"}\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if type is empty", func() {
		s.Setup()
		t := s.buildTag()
		body, _ := json.Marshal(t)
		request, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		request = mux.SetURLVars(request, map[string]string{
			"type":         "",
			"record_urn":   t.RecordURN,
			"template_urn": t.TemplateURN,
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"type is empty\"}\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if record urn is empty", func() {
		s.Setup()
		t := s.buildTag()
		body, _ := json.Marshal(t)
		request, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		request = mux.SetURLVars(request, map[string]string{
			"type":         t.RecordType,
			"record_urn":   t.RecordURN,
			"template_urn": "",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"template urn is empty\"}\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request and error if body cannot be unmarshalled", func() {
		s.Setup()
		body := "invalid_body"
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         "sample-type",
			"record_urn":   "sample-urn",
			"template_urn": "template-urn",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"invalid character 'i' looking for beginning of value\"}\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status not found error if tag could not be found", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, t.TemplateURN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{}, nil)
		s.tagRepository.On("Update", mock.Anything, &t).Return(errors.New("unexpected error during update"))

		body, err := json.Marshal(t)
		s.Require().NoError(err)
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         t.RecordType,
			"record_urn":   t.RecordURN,
			"template_urn": t.TemplateURN,
		})

		s.handler.Update(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return 500 if found error during update", func() {
		s.Setup()
		t := s.buildTag()
		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, t.TemplateURN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Read", mock.Anything, tag.Tag{
			RecordType:  t.RecordType,
			RecordURN:   t.RecordURN,
			TemplateURN: t.TemplateURN,
		}).Return([]tag.Tag{t}, nil)
		s.tagRepository.On("Update", mock.Anything, &t).Return(errors.New("unexpected error during update"))

		body, err := json.Marshal(t)
		s.Require().NoError(err)
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         t.RecordType,
			"record_urn":   t.RecordURN,
			"template_urn": t.TemplateURN,
		})

		s.handler.Update(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status created and domain is updated if found no error", func() {
		s.Setup()
		originalDomainTag := s.buildTag()
		template := s.buildTemplate()
		s.templateRepository.On("Read", mock.Anything, template.URN).Return([]tag.Template{template}, nil)
		s.tagRepository.On("Read", mock.Anything, tag.Tag{
			RecordType:  originalDomainTag.RecordType,
			RecordURN:   originalDomainTag.RecordURN,
			TemplateURN: originalDomainTag.TemplateURN,
		}).Return([]tag.Tag{originalDomainTag}, nil)
		s.tagRepository.On("Update", mock.Anything, &originalDomainTag).Return(nil)

		body, err := json.Marshal(originalDomainTag)
		s.Require().NoError(err)
		request, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(string(body)))
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         originalDomainTag.RecordType,
			"record_urn":   originalDomainTag.RecordURN,
			"template_urn": originalDomainTag.TemplateURN,
		})

		expectedStatusCode := http.StatusOK
		rsp, err := json.Marshal(originalDomainTag)
		s.Require().NoError(err)
		expectedResponseBody := string(rsp) + "\n"

		s.handler.Update(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})
}

func (s *TagHandlerTestSuite) TestDelete() {
	s.Run("should return status internal server error and its message if service is nil", func() {
		s.Setup()
		handler := &handlers.TagHandler{}
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "template-urn"
		request, err := http.NewRequest(http.MethodDelete, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         recordType,
			"record_urn":   recordURN,
			"template_urn": templateURN,
		})

		expectedStatusCode := http.StatusInternalServerError
		expectedResponseBody := "{\"reason\":\"tag service is nil\"}\n"

		handler.Delete(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if type is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodDelete, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         "",
			"record_urn":   "sample-urn",
			"template_urn": "template-urn",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"type is empty\"}\n"

		s.handler.Delete(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if record urn is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodDelete, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"record_urn":   "",
			"type":         "sample-type",
			"template_urn": "template-urn",
		})

		expectedStatusCode := http.StatusBadRequest
		expectedResponseBody := "{\"reason\":\"record urn is empty\"}\n"

		s.handler.Delete(s.recorder, request)
		actualStatusCode := s.recorder.Result().StatusCode
		actualResponseBody := s.recorder.Body.String()

		s.Equal(expectedStatusCode, actualStatusCode)
		s.Equal(expectedResponseBody, actualResponseBody)
	})

	s.Run("should return status bad request error and its message if template urn is empty", func() {
		s.Setup()
		request, err := http.NewRequest(http.MethodDelete, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         "sample-type",
			"record_urn":   "sample-urn",
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

	s.Run("should return status not found error and its message if template does not exist", func() {
		s.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "template-urn"
		request, _ := http.NewRequest(http.MethodDelete, "/", nil)
		request = mux.SetURLVars(request, map[string]string{
			"type":         recordType,
			"record_urn":   recordURN,
			"template_urn": templateURN,
		})
		s.templateRepository.On("Read", mock.Anything, templateURN).Return([]tag.Template{{}}, nil)
		s.tagRepository.On("Delete", mock.Anything, tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}).Return(tag.ErrTemplateNotFound{})

		s.handler.Delete(s.recorder, request)
		s.Equal(http.StatusNotFound, s.recorder.Result().StatusCode)
	})

	s.Run("should return 500 if found unexpected error", func() {
		s.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "template-urn"
		request, err := http.NewRequest(http.MethodDelete, "/", nil)
		s.Require().NoError(err)
		request = mux.SetURLVars(request, map[string]string{
			"type":         recordType,
			"record_urn":   recordURN,
			"template_urn": templateURN,
		})
		s.templateRepository.On("Read", mock.Anything, templateURN).Return([]tag.Template{{}}, nil)
		s.tagRepository.On("Delete", mock.Anything, tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}).Return(errors.New("unexpected error"))

		s.handler.Delete(s.recorder, request)
		s.Equal(http.StatusInternalServerError, s.recorder.Result().StatusCode)
	})

	s.Run("should return status no content and empty if delete success", func() {
		s.Setup()
		var recordType string = "sample-type"
		var recordURN string = "sample-urn"
		var templateURN string = "template-urn"
		request, _ := http.NewRequest(http.MethodDelete, "/", nil)
		request = mux.SetURLVars(request, map[string]string{
			"type":         recordType,
			"record_urn":   recordURN,
			"template_urn": templateURN,
		})
		s.templateRepository.On("Read", mock.Anything, templateURN).Return([]tag.Template{{}}, nil)
		s.tagRepository.On("Delete", mock.Anything, tag.Tag{
			RecordType:  recordType,
			RecordURN:   recordURN,
			TemplateURN: templateURN,
		}).Return(nil)

		s.handler.Delete(s.recorder, request)
		s.Equal(http.StatusNoContent, s.recorder.Result().StatusCode)
	})
}

func (s *TagHandlerTestSuite) buildTemplate() tag.Template {
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

func (s *TagHandlerTestSuite) buildTag() tag.Tag {
	return tag.Tag{
		RecordURN:           "sample-urn",
		RecordType:          "sample-type",
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

func TestTagHandler(t *testing.T) {
	suite.Run(t, &TagHandlerTestSuite{})
}
