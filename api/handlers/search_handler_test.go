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
	"testing"

	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/record"

	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSearchHandler(t *testing.T) {
	ctx := context.Background()
	// todo: pass testCase to ValidateResponse
	type testCase struct {
		Title            string
		ExpectStatus     int
		Querystring      string
		InitSearcher     func(testCase, *mock.RecordSearcher)
		ValidateResponse func(testCase, io.Reader) error
	}

	var testCases = []testCase{
		{
			Title:            "should return HTTP 400 if 'text' parameter is empty or missing",
			ExpectStatus:     http.StatusBadRequest,
			ValidateResponse: func(tc testCase, body io.Reader) error { return nil },
			Querystring:      "",
		},
		{
			Title:        "should return 400 if type is invalid",
			Querystring:  "text=test&filter.type=random",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {},
			ExpectStatus: http.StatusBadRequest,
		},
		{
			Title:       "should report HTTP 500 if record searcher fails",
			Querystring: "text=test",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				err := fmt.Errorf("service unavailable")
				searcher.On("Search", ctx, testifyMock.AnythingOfType("discovery.SearchConfig")).
					Return([]record.Record{}, err)
			},
			ExpectStatus: http.StatusInternalServerError,
		},
		{
			Title:       "should pass filter to search config format",
			Querystring: "text=resource&landscape=id,vn&filter.data.landscape=th&filter.type=topic&filter.service=kafka,rabbitmq",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := discovery.SearchConfig{
					Text:          "resource",
					TypeWhiteList: []record.Type{record.TypeTopic},
					Filters: map[string][]string{
						"service":        {"kafka", "rabbitmq"},
						"data.landscape": {"th"},
					},
				}

				searcher.On("Search", ctx, cfg).Return([]record.Record{}, nil)
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				return nil
			},
		},
		{
			Title:       "should return the matched documents",
			Querystring: "text=test",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := discovery.SearchConfig{
					Text:    "test",
					Filters: make(map[string][]string),
				}
				response := []record.Record{
					{
						Type:        "test",
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
				}
				searcher.On("Search", ctx, cfg).Return(response, nil)
			},
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
			Title:       "should return the requested number of records",
			Querystring: "text=resource&size=10",
			InitSearcher: func(tc testCase, searcher *mock.RecordSearcher) {
				cfg := discovery.SearchConfig{
					Text:       "resource",
					MaxResults: 10,
					Filters:    make(map[string][]string),
				}

				var results []record.Record
				for i := 0; i < cfg.MaxResults; i++ {
					urn := fmt.Sprintf("resource-%d", i+1)
					name := fmt.Sprintf("resource %d", i+1)
					r := record.Record{
						Urn:     urn,
						Type:    record.TypeTable,
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

					results = append(results, r)
				}

				searcher.On("Search", ctx, cfg).Return(results, nil)
			},
			ValidateResponse: func(tc testCase, body io.Reader) error {
				var payload []interface{}
				err := json.NewDecoder(body).Decode(&payload)
				if err != nil {
					return fmt.Errorf("error reading response body: %v", err)
				}

				expectedSize := 10
				actualSize := len(payload)
				if expectedSize != actualSize {
					return fmt.Errorf("expected search request to return %d results, returned %d results instead", expectedSize, actualSize)
				}
				return nil
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			var (
				recordSearcher = new(mock.RecordSearcher)
			)
			if testCase.InitSearcher != nil {
				testCase.InitSearcher(testCase, recordSearcher)
			}
			defer recordSearcher.AssertExpectations(t)

			params, err := url.ParseQuery(testCase.Querystring)
			require.NoError(t, err)

			requestURL := "/?" + params.Encode()
			rr := httptest.NewRequest(http.MethodGet, requestURL, nil)
			rw := httptest.NewRecorder()

			service := discovery.NewService(nil, recordSearcher)
			handler := handlers.NewSearchHandler(new(mock.Logger), service)
			handler.Search(rw, rr)

			expectStatus := testCase.ExpectStatus
			if expectStatus == 0 {
				expectStatus = http.StatusOK
			}
			assert.Equal(t, expectStatus, rw.Code)
			if testCase.ValidateResponse != nil {
				if err := testCase.ValidateResponse(testCase, rw.Body); err != nil {
					t.Errorf("error validating handler response: %v", err)
					return
				}
			}
		})
	}
}
