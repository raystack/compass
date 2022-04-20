package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/odpf/salt/log"

	"github.com/odpf/compass/api/httpapi/handlers"
	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/discovery"
	"github.com/odpf/compass/lib/mocks"
	"github.com/stretchr/testify/mock"
)

func TestRecordHandler(t *testing.T) {
	var (
		ctx      = mock.AnythingOfType("*context.emptyCtx")
		typeName = asset.TypeTable.String()
		logger   = log.NewNoop()
	)

	tr := new(mocks.TypeRepository)

	t.Run("GetByType", func(t *testing.T) {
		type testCase struct {
			Description  string
			Type         string
			QueryStrings string
			ExpectStatus int
			Setup        func(tc *testCase, rrf *mocks.RecordRepositoryFactory)
			PostCheck    func(tc *testCase, resp *http.Response) error
		}

		var assets = []asset.Asset{
			{
				URN: "test-fh-1",
				Data: map[string]interface{}{
					"urn":         "test-fh-1",
					"owner":       "de",
					"created":     "2020-05-13T08:30:04Z",
					"environment": "test",
				},
			},
			{
				URN: "test-fh-2",
				Data: map[string]interface{}{
					"urn":         "test-fh-2",
					"owner":       "de",
					"created":     "2020-05-12T00:00:00Z",
					"environment": "test",
				},
			},
		}

		var testCases = []testCase{
			{
				Description:  "should return an http 404 if the type doesn't exist",
				Type:         "invalid",
				QueryStrings: "filter.environment=test",
				ExpectStatus: http.StatusNotFound,
				Setup:        func(tc *testCase, rrf *mocks.RecordRepositoryFactory) {},
			},
			{
				Description:  "should get from and size from querystring and pass it to repo",
				Type:         typeName,
				QueryStrings: "from=5&size=10",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, rrf *mocks.RecordRepositoryFactory) {
					rr := new(mocks.RecordRepository)
					rr.On("GetAll", ctx, discovery.GetConfig{
						Filters: map[string][]string{},
						From:    5,
						Size:    10,
					}).Return(discovery.RecordList{Data: assets}, nil)
					rrf.On("For", typeName).Return(rr, nil)
				},
			},
			{
				Description:  "should create filter from querystring",
				Type:         typeName,
				QueryStrings: "filter.service=kafka,rabbitmq&filter.data.company=appel",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, rrf *mocks.RecordRepositoryFactory) {
					rr := new(mocks.RecordRepository)
					rr.On("GetAll", ctx, discovery.GetConfig{
						Filters: map[string][]string{
							"service":      {"kafka", "rabbitmq"},
							"data.company": {"appel"},
						}}).Return(discovery.RecordList{Data: assets}, nil)
					rrf.On("For", typeName).Return(rr, nil)
				},
			},
			{
				Description:  "should return http 500 if the handler fails to construct record repository",
				Type:         typeName,
				QueryStrings: "filter.data.environment=test",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, rrf *mocks.RecordRepositoryFactory) {
					rr := new(mocks.RecordRepository)
					err := fmt.Errorf("something went wrong")
					rrf.On("For", typeName).Return(rr, err)
				},
			},
			{
				Description:  "should return an http 500 if calling recordRepository.GetAll fails",
				Type:         typeName,
				QueryStrings: "filter.data.environment=test",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, rrf *mocks.RecordRepositoryFactory) {
					rr := new(mocks.RecordRepository)
					err := fmt.Errorf("temporarily unavailable")
					rr.On("GetAll", ctx, discovery.GetConfig{
						Filters: map[string][]string{"data.environment": {"test"}},
					}).Return(discovery.RecordList{Data: []asset.Asset{}}, err)
					rrf.On("For", typeName).Return(rr, nil)
				},
			},
			{
				Description:  "should return 200 on success and RecordList",
				Type:         typeName,
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, rrf *mocks.RecordRepositoryFactory) {
					rr := new(mocks.RecordRepository)
					rr.On("GetAll", ctx, discovery.GetConfig{
						Filters: map[string][]string{},
					}).Return(discovery.RecordList{Data: assets}, nil)
					rrf.On("For", typeName).Return(rr, nil)
				},
				PostCheck: func(tc *testCase, resp *http.Response) error {
					var response discovery.RecordList
					err := json.NewDecoder(resp.Body).Decode(&response)
					if err != nil {
						return fmt.Errorf("error parsing response payload: %w", err)
					}

					expected := discovery.RecordList{
						Data: assets,
					}

					if reflect.DeepEqual(response, expected) == false {
						return fmt.Errorf("expected handler to return %v, returned %v instead", expected, response)
					}
					return nil
				},
			},
			{
				Description:  "should return the subset of fields specified via select parameter",
				Type:         typeName,
				QueryStrings: "filter.data.environment=test&select=" + url.QueryEscape("urn,owner"),
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, rrf *mocks.RecordRepositoryFactory) {
					rr := new(mocks.RecordRepository)
					rr.On("GetAll", ctx, discovery.GetConfig{
						Filters: map[string][]string{"data.environment": {"test"}},
					}).Return(discovery.RecordList{Data: assets}, nil)
					rrf.On("For", typeName).Return(rr, nil)
				},
				PostCheck: func(tc *testCase, resp *http.Response) error {
					var response discovery.RecordList
					err := json.NewDecoder(resp.Body).Decode(&response)
					if err != nil {
						return fmt.Errorf("error parsing response payload: %w", err)
					}

					expected := discovery.RecordList{
						Data: []asset.Asset{
							{
								URN: "test-fh-1",
								Data: map[string]interface{}{
									"urn":   "test-fh-1",
									"owner": "de",
								},
							},
							{
								URN: "test-fh-2",
								Data: map[string]interface{}{
									"urn":   "test-fh-2",
									"owner": "de",
								},
							},
						},
					}

					if reflect.DeepEqual(response, expected) == false {
						return fmt.Errorf("expected handler to return %v, returned %v instead", expected, response)
					}

					return nil
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("GET", "/?"+tc.QueryStrings, nil)
				rw := httptest.NewRecorder()
				rrf := new(mocks.RecordRepositoryFactory)
				tc.Setup(&tc, rrf)

				service := discovery.NewService(rrf, nil)
				handler := handlers.NewRecordHandler(logger, tr, service, rrf)

				handler.GetByType(rw, rr, map[string]string{
					"name": tc.Type,
				})

				if rw.Code != tc.ExpectStatus {
					t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
					return
				}

				if tc.PostCheck != nil {
					if err := tc.PostCheck(&tc, rw.Result()); err != nil {
						t.Error(err)
					}
				}
			})
		}
	})
	t.Run("GetOneByType", func(t *testing.T) {
		var deployment01 = asset.Asset{
			URN: "id-1",
			Data: map[string]interface{}{
				"contents": "data",
			},
		}
		type testCase struct {
			Description  string
			Type         string
			AssetID      string
			ExpectStatus int
			Setup        func(rrf *mocks.RecordRepositoryFactory)
			PostCheck    func(resp *http.Response) error
		}

		var testCases = []testCase{
			{
				Description:  `should return http 404 if the record doesn't exist`,
				Type:         typeName,
				AssetID:      "record01",
				ExpectStatus: http.StatusNotFound,
				Setup: func(rrf *mocks.RecordRepositoryFactory) {
					recordRepo := new(mocks.RecordRepository)
					recordRepo.On("GetByID", ctx, "record01").Return(asset.Asset{}, asset.NotFoundError{AssetID: "record01"})
					rrf.On("For", typeName).Return(recordRepo, nil)
				},
			},
			{
				Description:  `should return http 404 if the type doesn't exist`,
				Type:         "invalid",
				AssetID:      "record",
				ExpectStatus: http.StatusNotFound,
				Setup:        func(rrf *mocks.RecordRepositoryFactory) {},
			},
			{
				Description:  "(internal) should return an http 500 if the handler fails to construct recordRepository",
				Type:         typeName,
				AssetID:      "record",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(rrf *mocks.RecordRepositoryFactory) {
					rrf.On("For", typeName).Return(new(mocks.RecordRepository), fmt.Errorf("something bad happened"))
				},
			},
			{
				Description:  "should return http 200 status along with the record, if found",
				Type:         typeName,
				AssetID:      "deployment01",
				ExpectStatus: http.StatusOK,
				Setup: func(rrf *mocks.RecordRepositoryFactory) {
					recordRepo := new(mocks.RecordRepository)
					recordRepo.On("GetByID", ctx, "deployment01").Return(deployment01, nil)
					rrf.On("For", typeName).Return(recordRepo, nil)
				},
				PostCheck: func(r *http.Response) error {
					var record asset.Asset
					err := json.NewDecoder(r.Body).Decode(&record)
					if err != nil {
						return fmt.Errorf("error reading response body: %w", err)
					}
					if reflect.DeepEqual(record, deployment01) == false {
						return fmt.Errorf("expected returned record to be to be %+v, was %+v", deployment01, record)
					}
					return nil
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("GET", "/", nil)
				rw := httptest.NewRecorder()
				recordRepoFac := new(mocks.RecordRepositoryFactory)
				if tc.Setup != nil {
					tc.Setup(recordRepoFac)
				}

				service := discovery.NewService(recordRepoFac, nil)
				handler := handlers.NewRecordHandler(logger, tr, service, recordRepoFac)
				handler.GetOneByType(rw, rr, map[string]string{
					"name": tc.Type,
					"id":   tc.AssetID,
				})

				if rw.Code != tc.ExpectStatus {
					t.Errorf("expected handler to return http %d, returned %d instead", tc.ExpectStatus, rw.Code)
					return
				}
				if tc.PostCheck != nil {
					if err := tc.PostCheck(rw.Result()); err != nil {
						t.Error(err)
						return
					}
				}
			})
		}
	})

}
