package web_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/models"
	"github.com/odpf/columbus/web"

	testifyMock "github.com/stretchr/testify/mock"
)

func TestSearchHandler(t *testing.T) {
	baseURL := &url.URL{
		Path: "/v1/search",
	}

	// todo: pass testCase to ValidateResponse
	type testCase struct {
		Title            string
		SearchText       string
		ExpectStatus     int
		RequestParams    map[string][]string
		InitRepo         func(testCase, *mock.TypeRepository)
		InitSearcher     func(testCase, *mock.RecordSearcher)
		ValidateResponse func(testCase, io.Reader) error
	}

	var testdata = struct {
		Type models.Type
	}{
		Type: models.Type{
			Name:           "test",
			Classification: models.TypeClassificationResource,
			Fields: models.TypeFields{
				ID:    "id",
				Title: "title",
				Labels: []string{
					"landscape",
					"entity",
				},
			},
		},
	}

	// helper for creating testcase.InitRepo that initialises the mock repo with the given types
	var withTypes = func(ents ...models.Type) func(tc testCase, repo *mock.TypeRepository) {
		return func(tc testCase, repo *mock.TypeRepository) {
			for _, ent := range ents {
				repo.On("GetByName", ent.Name).Return(ent, nil)
			}
			return
		}
	}

	var testCases = []testCase{
		{
			Title:            "should return HTTP 400 if 'text' parameter is empty or missing",
			ExpectStatus:     http.StatusBadRequest,
			ValidateResponse: func(tc testCase, body io.Reader) error { return nil },
			SearchText:       "",
		},
		{
			Title:      "should report HTTP 500 if record searcher fails",
			SearchText: "test",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				err := fmt.Errorf("service unavailable")
				searcher.On("Search", testifyMock.AnythingOfType("models.SearchConfig")).
					Return([]models.SearchResult{}, err)
			},
			ExpectStatus: http.StatusInternalServerError,
		},
		{
			Title:      "should return an error if looking up an type detail fails",
			SearchText: "test",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				results := []models.SearchResult{
					{
						TypeName: "test",
						Record:   models.Record{},
					},
				}
				searcher.On("Search", testifyMock.AnythingOfType("models.SearchConfig")).
					Return(results, nil)
			},
			InitRepo: func(tc testCase, repo *mock.TypeRepository) {
				repo.On("GetByName", testifyMock.AnythingOfType("string")).
					Return(models.Type{}, models.ErrNoSuchType{})
			},
			ExpectStatus: http.StatusInternalServerError,
		},
		{
			Title:      "should return the matched documents",
			SearchText: "test",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text:    tc.SearchText,
					Filters: make(map[string][]string),
				}
				response := []models.SearchResult{
					{
						TypeName: "test",
						Record: map[string]interface{}{
							"id":        "test-resource",
							"title":     "test resource",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}
				searcher.On("Search", cfg).Return(response, nil)
			},
			InitRepo: withTypes(testdata.Type),
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var response []web.SearchResponse
				err := json.NewDecoder(body).Decode(&response)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectResponse := []web.SearchResponse{
					{
						Title:          "test resource",
						ID:             "test-resource",
						Classification: string(models.TypeClassificationResource),
						Type:           "test",
						Labels: map[string]string{
							"entity":    "odpf",
							"landscape": "id",
						},
					},
				}

				if reflect.DeepEqual(response, expectResponse) == false {
					return fmt.Errorf("expected handler response to be %#v, was %#v", expectResponse, response)
				}
				return nil
			},
		},
		{
			Title:      "should drop records that have mandatory fields missing",
			SearchText: "test",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text:    tc.SearchText,
					Filters: make(map[string][]string),
				}
				results := []models.SearchResult{
					{
						TypeName: "test",
						Record: models.Record{
							"id":        "test-resource-1",
							"title":     "test resource 1",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
					{
						TypeName: "test",
						Record: models.Record{
							"id":        "test-resource-2",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
					{
						TypeName: "test",
						Record: models.Record{
							"title":     "test resource 3",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
					{
						TypeName: "test",
						Record: models.Record{
							"id":     "test-resource-4",
							"title":  "test resource 4",
							"entity": "odpf",
						},
					},
				}
				searcher.On("Search", cfg).Return(results, nil)
			},
			InitRepo: withTypes(testdata.Type),
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var response []web.SearchResponse
				err := json.NewDecoder(body).Decode(&response)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectResponse := []web.SearchResponse{
					{
						Title:          "test resource 1",
						ID:             "test-resource-1",
						Classification: string(models.TypeClassificationResource),
						Type:           "test",
						Labels: map[string]string{
							"landscape": "id",
							"entity":    "odpf",
						},
					},
					{
						Title:          "test resource 4",
						ID:             "test-resource-4",
						Classification: string(models.TypeClassificationResource),
						Type:           "test",
						Labels: map[string]string{
							"entity": "odpf",
						},
					},
				}

				if reflect.DeepEqual(response, expectResponse) == false {
					return fmt.Errorf("expected handler response to be %#v, was %#v", expectResponse, response)
				}
				return nil
			},
		},
		{
			Title: "should return the requested number of records",
			RequestParams: map[string][]string{
				"size": []string{"15"},
			},
			SearchText: "resource",
			InitRepo:   withTypes(testdata.Type),
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				maxResults, _ := strconv.Atoi(tc.RequestParams["size"][0])
				cfg := models.SearchConfig{
					Text:       tc.SearchText,
					MaxResults: maxResults,
					Filters:    make(map[string][]string),
				}

				var results []models.SearchResult
				for i := 0; i < cfg.MaxResults; i++ {
					record := models.Record{
						"id":        fmt.Sprintf("resource-%d", i+1),
						"title":     fmt.Sprintf("resource %d", i+1),
						"landscape": "id",
						"entity":    "odpf",
					}
					result := models.SearchResult{
						Record:   record,
						TypeName: testdata.Type.Name,
					}
					results = append(results, result)
				}

				searcher.On("Search", cfg).Return(results, nil)
				return
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var payload []interface{}
				err := json.NewDecoder(body).Decode(&payload)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectedResults, _ := strconv.Atoi(tc.RequestParams["size"][0])
				actualResults := len(payload)
				if expectedResults != actualResults {
					return fmt.Errorf("expected search request to return %d results, returned %d results instead", expectedResults, actualResults)
				}
				return nil
			},
		},
		{
			Title:      "should filter results for landscape",
			SearchText: "resource",
			RequestParams: map[string][]string{
				// "landscape" is not a valid filter key. All filters
				// begin with the "filter." prefix. Adding this here is just a little
				// extra check to make sure that the handler correctly parses the filters.
				"landscape":        []string{"id", "vn"},
				"filter.landscape": []string{"th"},
			},
			InitRepo: withTypes(testdata.Type),
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text: tc.SearchText,
					Filters: map[string][]string{
						"landscape": tc.RequestParams["filter.landscape"],
					},
				}

				results := []models.SearchResult{
					{
						TypeName: testdata.Type.Name,
						Record: models.Record{
							"id":        "test-1",
							"title":     "test 1",
							"landscape": "th",
							"entity":    "odpf",
						},
					},
				}
				searcher.On("Search", cfg).Return(results, nil)
				return
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var actualResults []web.SearchResponse
				err := json.NewDecoder(body).Decode(&actualResults)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectResults := []web.SearchResponse{
					{
						Title:          "test 1",
						ID:             "test-1",
						Type:           testdata.Type.Name,
						Classification: string(testdata.Type.Classification),
						Labels: map[string]string{
							"landscape": "th",
							"entity":    "odpf",
						},
					},
				}

				if reflect.DeepEqual(actualResults, expectResults) == false {
					return fmt.Errorf("expected handler response to be %#v, was %#v", expectResults, actualResults)
				}
				return nil
			},
		},
		{
			Title:      "should filter results for entity",
			SearchText: "resource",
			RequestParams: map[string][]string{
				"filter.entity": []string{"odpf"},
			},
			InitRepo: withTypes(testdata.Type),
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text: tc.SearchText,
					Filters: map[string][]string{
						"entity": tc.RequestParams["filter.entity"],
					},
				}

				results := []models.SearchResult{
					{
						TypeName: testdata.Type.Name,
						Record: models.Record{
							"id":        "test-1",
							"title":     "test 1",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}
				searcher.On("Search", cfg).Return(results, nil)
				return
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var actualResults []web.SearchResponse
				err := json.NewDecoder(body).Decode(&actualResults)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectResults := []web.SearchResponse{
					{
						Title:          "test 1",
						ID:             "test-1",
						Type:           testdata.Type.Name,
						Classification: string(testdata.Type.Classification),
						Labels: map[string]string{
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}

				if reflect.DeepEqual(actualResults, expectResults) == false {
					return fmt.Errorf("expected handler response to be %#v, was %#v", expectResults, actualResults)
				}
				return nil
			},
		},
		{
			Title:      "should return description, if available",
			SearchText: "resource",
			InitRepo: func(tc testCase, repo *mock.TypeRepository) {
				// create a copy of the testdata type
				// and override the description field
				recordType := testdata.Type
				recordType.Fields.Description = "description_text"
				initFunc := withTypes(recordType)
				initFunc(tc, repo)
			},
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text:    tc.SearchText,
					Filters: make(map[string][]string),
				}
				results := []models.SearchResult{
					{
						TypeName: "test",
						Record: models.Record{
							"title":            "test",
							"id":               "test",
							"description_text": "this is a test record",
						},
					},
				}
				searcher.On("Search", cfg).Return(results, nil)
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var actualResults []web.SearchResponse
				err := json.NewDecoder(body).Decode(&actualResults)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}

				expectResults := []web.SearchResponse{
					{
						Title:          "test",
						ID:             "test",
						Description:    "this is a test record",
						Type:           testdata.Type.Name,
						Classification: string(testdata.Type.Classification),
						Labels:         map[string]string{},
					},
				}

				if reflect.DeepEqual(actualResults, expectResults) == false {
					return fmt.Errorf("expected handler response to be %#v, was %#v", expectResults, actualResults)
				}
				return nil
			},
		},
		{
			Title:      "should filter results for environment",
			SearchText: "resource",
			RequestParams: map[string][]string{
				"filter.environment": {"test"},
			},
			InitRepo: withTypes(testdata.Type),
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text: tc.SearchText,
					Filters: map[string][]string{
						"environment": {"test"},
					},
				}
				results := []models.SearchResult{
					{
						TypeName: testdata.Type.Name,
						Record: models.Record{
							"id":        "test-1",
							"title":     "test 1",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}
				searcher.On("Search", cfg).Return(results, nil)
				return
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var actualResults []web.SearchResponse
				err := json.NewDecoder(body).Decode(&actualResults)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectResults := []web.SearchResponse{
					{
						Title:          "test 1",
						ID:             "test-1",
						Type:           testdata.Type.Name,
						Classification: string(testdata.Type.Classification),
						Labels: map[string]string{
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}

				if reflect.DeepEqual(actualResults, expectResults) == false {
					return fmt.Errorf("expected handler response to be %#v, was %#v", expectResults, actualResults)
				}
				return nil
			},
		},
		{
			Title:         "should filter results even when environment is not provided",
			SearchText:    "resource",
			RequestParams: map[string][]string{},
			InitRepo:      withTypes(testdata.Type),
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text:    tc.SearchText,
					Filters: map[string][]string{},
				}
				results := []models.SearchResult{
					{
						TypeName: testdata.Type.Name,
						Record: models.Record{
							"id":        "test-1",
							"title":     "test 1",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}
				searcher.On("Search", cfg).Return(results, nil)
				return
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var actualResults []web.SearchResponse
				err := json.NewDecoder(body).Decode(&actualResults)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectResults := []web.SearchResponse{
					{
						Title:          "test 1",
						ID:             "test-1",
						Type:           testdata.Type.Name,
						Classification: string(testdata.Type.Classification),
						Labels: map[string]string{
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}

				if reflect.DeepEqual(actualResults, expectResults) == false {
					return fmt.Errorf("expected handler response to be %#v, was %#v", expectResults, actualResults)
				}
				return nil
			},
		},
		{
			Title:      "should filter results on the basis of types",
			SearchText: "text",
			RequestParams: map[string][]string{
				"filter.type": {"dagger"},
			},
			ExpectStatus: http.StatusOK,
			InitRepo:     withTypes(testdata.Type),
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text:          tc.SearchText,
					TypeWhiteList: tc.RequestParams["filter.type"],
					Filters:       map[string][]string{},
				}
				results := []models.SearchResult{
					{
						TypeName: testdata.Type.Name,
						Record: models.Record{
							"id":        "test-1",
							"title":     "test 1",
							"landscape": "id",
							"entity":    "odpf",
						},
					},
				}
				searcher.On("Search", cfg).Return(results, nil)
				return
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			var (
				recordSearcher = new(mock.RecordSearcher)
				typeRepo       = new(mock.TypeRepository)
			)
			if testCase.InitRepo != nil {
				testCase.InitRepo(testCase, typeRepo)
			}
			if testCase.InitSearcher != nil {
				testCase.InitSearcher(testCase, recordSearcher)
			}

			defer recordSearcher.AssertExpectations(t)
			defer typeRepo.AssertExpectations(t)

			handler := web.NewSearchHandler(new(mock.Logger), recordSearcher, typeRepo)

			params := url.Values{}
			params.Add("text", testCase.SearchText)
			if testCase.RequestParams != nil {
				for key, values := range testCase.RequestParams {
					for _, value := range values {
						params.Add(key, value)
					}
				}
			}
			requestURL := baseURL.ResolveReference(&url.URL{RawQuery: params.Encode()})
			rr := httptest.NewRequest(http.MethodGet, requestURL.String(), nil)
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, rr)

			expectStatus := testCase.ExpectStatus
			if expectStatus == 0 {
				expectStatus = http.StatusOK
			}
			if rw.Code != expectStatus {
				t.Errorf("expected handler to return http status %d, was %d instead", expectStatus, rw.Code)
				return
			}
			if testCase.ValidateResponse != nil {
				if err := testCase.ValidateResponse(testCase, rw.Body); err != nil {
					t.Errorf("error validating handler response: %v", err)
					return
				}
			}
		})
	}
}
