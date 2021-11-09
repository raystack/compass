package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/models"

	testifyMock "github.com/stretchr/testify/mock"
)

func TestSearchHandler(t *testing.T) {
	ctx := context.Background()
	// todo: pass testCase to ValidateResponse
	type testCase struct {
		Title            string
		SearchText       string
		ExpectStatus     int
		RequestQuery     map[string][]string
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
		},
	}

	// helper for creating testcase.InitRepo that initialises the mock repo with the given types
	var withTypes = func(ents ...models.Type) func(tc testCase, repo *mock.TypeRepository) {
		return func(tc testCase, repo *mock.TypeRepository) {
			for _, ent := range ents {
				repo.On("GetByName", ctx, ent.Name).Return(ent, nil)
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
				searcher.On("Search", ctx, testifyMock.AnythingOfType("models.SearchConfig")).
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
				searcher.On("Search", ctx, testifyMock.AnythingOfType("models.SearchConfig")).
					Return(results, nil)
			},
			InitRepo: func(tc testCase, repo *mock.TypeRepository) {
				repo.On("GetByName", ctx, testifyMock.AnythingOfType("string")).
					Return(models.Type{}, models.ErrNoSuchType{})
			},
			ExpectStatus: http.StatusInternalServerError,
		},
		{
			Title:      "should pass filter to search config format",
			SearchText: "resource",
			RequestQuery: map[string][]string{
				// "landscape" is not a valid filter key. All filters
				// begin with the "filter." prefix. Adding this here is just a little
				// extra check to make sure that the handler correctly parses the filters.
				"landscape":             {"id", "vn"},
				"filter.data.landscape": {"th"},
				"filter.type":           {"topic"},
				"filter.service":        {"kafka", "rabbitmq"},
			},
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := models.SearchConfig{
					Text:          tc.SearchText,
					TypeWhiteList: tc.RequestQuery["filter.type"],
					Filters: map[string][]string{
						"service":        tc.RequestQuery["filter.service"],
						"data.landscape": tc.RequestQuery["filter.data.landscape"],
					},
				}

				searcher.On("Search", ctx, cfg).Return([]models.SearchResult{}, nil)
				return
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				return nil
			},
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
						Record: models.Record{
							Urn:         "test-resource",
							Name:        "test resource",
							Description: "some description",
							Service:     "test-service",
							Data: map[string]interface{}{
								"id":        "test-resource",
								"title":     "test resource",
								"landscape": "id",
								"entity":    "odpf",
							},
							Labels: map[string]string{
								"entity":    "odpf",
								"landscape": "id",
							},
						},
					},
				}
				searcher.On("Search", ctx, cfg).Return(response, nil)
			},
			InitRepo: withTypes(testdata.Type),
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var response []handlers.SearchResponse
				err := json.NewDecoder(body).Decode(&response)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectResponse := []handlers.SearchResponse{
					{
						ID:          "test-resource",
						Title:       "test resource",
						Description: "some description",
						Service:     "test-service",
						Type:        "test",
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
			Title: "should return the requested number of records",
			RequestQuery: map[string][]string{
				"size": {"15"},
			},
			SearchText: "resource",
			InitRepo:   withTypes(testdata.Type),
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				maxResults, _ := strconv.Atoi(tc.RequestQuery["size"][0])
				cfg := models.SearchConfig{
					Text:       tc.SearchText,
					MaxResults: maxResults,
					Filters:    make(map[string][]string),
				}

				var results []models.SearchResult
				for i := 0; i < cfg.MaxResults; i++ {
					urn := fmt.Sprintf("resource-%d", i+1)
					name := fmt.Sprintf("resource %d", i+1)
					record := models.Record{
						Urn:     urn,
						Name:    name,
						Service: "kafka",
						Data: map[string]interface{}{
							"id":        urn,
							"title":     name,
							"landscape": "id",
							"entity":    "odpf",
						},
						Labels: map[string]string{
							"landscape": "id",
							"entity":    "odpf",
						},
					}
					result := models.SearchResult{
						Record:   record,
						TypeName: testdata.Type.Name,
					}
					results = append(results, result)
				}

				searcher.On("Search", ctx, cfg).Return(results, nil)
				return
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var payload []interface{}
				err := json.NewDecoder(body).Decode(&payload)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}
				expectedResults, _ := strconv.Atoi(tc.RequestQuery["size"][0])
				actualResults := len(payload)
				if expectedResults != actualResults {
					return fmt.Errorf("expected search request to return %d results, returned %d results instead", expectedResults, actualResults)
				}
				return nil
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

			params := url.Values{}
			params.Add("text", testCase.SearchText)
			if testCase.RequestQuery != nil {
				for key, values := range testCase.RequestQuery {
					for _, value := range values {
						params.Add(key, value)
					}
				}
			}
			requestURL := "/?" + params.Encode()
			rr := httptest.NewRequest(http.MethodGet, requestURL, nil)
			rw := httptest.NewRecorder()

			handler := handlers.NewSearchHandler(new(mock.Logger), recordSearcher, typeRepo)
			handler.Search(rw, rr)

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
