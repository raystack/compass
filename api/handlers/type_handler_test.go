package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/models"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestTypeHandler(t *testing.T) {
	var (
		daggerType = models.Type{
			Name:           "dagger",
			Classification: models.TypeClassificationResource,
			Fields: models.TypeFields{
				ID:    "name",
				Title: "urn",
				Labels: []string{
					"team",
				},
			},
		}
		ctx           = tmock.AnythingOfType("*context.valueCtx")
		daggerTypeURI = fmt.Sprintf("/v1/types/%s", daggerType.Name)
	)

	t.Run("PUT /v1/types", func(t *testing.T) {
		const apiURL = "/v1/types"
		validPayloadRaw, err := json.Marshal(daggerType)
		if err != nil {
			t.Fatalf("error preparing request payload: %v", err)
			return
		}
		t.Run("should return HTTP 400 if the JSON document is invalid", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/v1/types", bytes.NewBufferString("{"))
			rw := httptest.NewRecorder()

			handler := handlers.NewTypeHandler(new(mock.Logger), nil, nil)
			handler.ServeHTTP(rw, rr)

			if rw.Code != http.StatusBadRequest {
				t.Errorf("handler returned HTTP %d, expected HTTP %d", rw.Code, http.StatusBadRequest)
				return
			}

			var res handlers.ErrorResponse
			err = json.NewDecoder(rw.Body).Decode(&res)
			if err != nil {
				t.Fatalf("error parsing handler response: %v", err)
				return
			}
			expectedReason := "error parsing request body: unexpected EOF"
			if res.Reason != expectedReason {
				t.Errorf("expected handler to return reason %q, returned %q instead", expectedReason, res.Reason)
				return
			}
		})
		t.Run("should return an error if any of the fields in the payload are empty", func(t *testing.T) {
			testCases := []struct {
				payload        models.Type
				expectedReason string
			}{
				{
					payload:        models.Type{},
					expectedReason: "'name' is required",
				},
				{
					payload: models.Type{
						Name: "foo",
					},
					expectedReason: "'classification' is required",
				},
				{
					payload: models.Type{
						Name:           "foo",
						Classification: models.TypeClassificationResource,
					},
					expectedReason: "'record_attributes.title' is required",
				},
				{
					payload: models.Type{
						Name:           "foo",
						Classification: models.TypeClassificationResource,
						Fields: models.TypeFields{
							Title: "bar",
						},
					},
					expectedReason: "'record_attributes.id' is required",
				},
			}

			for _, testCase := range testCases {

				raw, err := json.Marshal(testCase.payload)
				if err != nil {
					t.Fatalf("error creating test payload: %v", err)
					return
				}
				rr := httptest.NewRequest("PUT", apiURL, bytes.NewBuffer(raw))
				rw := httptest.NewRecorder()

				handler := handlers.NewTypeHandler(new(mock.Logger), nil, nil)
				handler.ServeHTTP(rw, rr)

				if rw.Code != http.StatusBadRequest {
					t.Errorf("handler returned HTTP %d, expected HTTP %d", rw.Code, http.StatusBadRequest)
					return
				}

				var res handlers.ErrorResponse
				err = json.NewDecoder(rw.Body).Decode(&res)
				if err != nil {
					t.Fatalf("error parsing handler response: %v", err)
					return
				}
				if res.Reason != testCase.expectedReason {
					t.Errorf("expected handler to return reason %q, returned %q instead", testCase.expectedReason, res.Reason)
					return
				}
			}
		})
		t.Run("should return HTTP 201 for successful type creation/update", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/v1/types", bytes.NewBuffer(validPayloadRaw))
			rw := httptest.NewRecorder()

			typeRepo := new(mock.TypeRepository)
			typeRepo.On("CreateOrReplace", ctx, daggerType).Return(nil)
			defer typeRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo, nil)
			handler.ServeHTTP(rw, rr)

			expectedStatus := http.StatusCreated
			if rw.Code != expectedStatus {
				t.Errorf("expected handler to HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
				return
			}
		})
		t.Run("should return 422 if type name is reserved", func(t *testing.T) {
			expectedErr := models.ErrReservedTypeName{TypeName: daggerType.Name}

			rr := httptest.NewRequest("PUT", "/v1/types", bytes.NewBuffer(validPayloadRaw))
			rw := httptest.NewRecorder()

			typeRepo := new(mock.TypeRepository)
			typeRepo.On("CreateOrReplace", ctx, daggerType).Return(expectedErr)
			defer typeRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo, nil)
			handler.ServeHTTP(rw, rr)

			assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			if err != nil {
				t.Fatalf("error decoding handler response: %v", err)
				return
			}
			assert.Equal(t, expectedErr.Error(), response.Reason)
		})
		t.Run("should return HTTP 500 if creating/updating the type fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/v1/types", bytes.NewBuffer(validPayloadRaw))
			rw := httptest.NewRecorder()

			creationErr := fmt.Errorf("failed to write to elasticsearch")
			typeRepo := new(mock.TypeRepository)
			typeRepo.On("CreateOrReplace", ctx, daggerType).Return(creationErr)
			defer typeRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo, nil)
			handler.ServeHTTP(rw, rr)

			expectedStatus := http.StatusInternalServerError
			if rw.Code != expectedStatus {
				t.Errorf("expected handler to HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
				return
			}
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			if err != nil {
				t.Fatalf("error decoding handler response: %v", err)
				return
			}
			expectedReason := fmt.Sprintf("error creating type: %v", creationErr)
			if response.Reason != expectedReason {
				t.Errorf("expected handler to return %q reason, returned %q instead", expectedReason, response.Reason)
				return
			}
		})
		t.Run("should return HTTP 400 if classification is invalid", func(t *testing.T) {
			typeWithInvalidClassification := &models.Type{
				Name:           "application",
				Classification: "unknown",
				Fields: models.TypeFields{
					ID:    "urn",
					Title: "title",
					Labels: []string{
						"landscape",
					},
				},
			}
			var payload bytes.Buffer
			err := json.NewEncoder(&payload).Encode(typeWithInvalidClassification)
			if err != nil {
				t.Fatalf("error preparing test data: %v", err)
				return
			}
			rr := httptest.NewRequest("PUT", "/v1/types", &payload)
			rw := httptest.NewRecorder()

			handler := handlers.NewTypeHandler(new(mock.Logger), nil, nil)
			handler.ServeHTTP(rw, rr)

			expectedCode := 400
			if rw.Code != expectedCode {
				t.Errorf("expected handler to return HTTP %d, returned %d instead", expectedCode, rw.Code)
			}
		})
		t.Run("should lowercase type name before commiting it to storage", func(t *testing.T) {
			ent := &models.Type{
				Name:           "DAGGER",
				Classification: models.TypeClassificationResource,
				Fields: models.TypeFields{
					ID:    "urn",
					Title: "title",
					Labels: []string{
						"landscape",
					},
				},
			}
			expectEnt := *ent
			expectEnt.Name = strings.ToLower(ent.Name)

			var payload bytes.Buffer
			err := json.NewEncoder(&payload).Encode(ent)
			if err != nil {
				t.Fatalf("error preparing test data: %v", err)
				return
			}

			rr := httptest.NewRequest("PUT", "/v1/types", &payload)
			rw := httptest.NewRecorder()

			repo := new(mock.TypeRepository)
			repo.On("CreateOrReplace", ctx, expectEnt).Return(nil)
			defer repo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), repo, nil)
			handler.ServeHTTP(rw, rr)
		})
	})
	t.Run("PUT /v1/types/{name}", func(t *testing.T) {
		t.Run("should return HTTP 404 if type doesn't exist", func(t *testing.T) {
			reqURI := "/v1/types/dagger"
			rr := httptest.NewRequest("PUT", reqURI, strings.NewReader("{}"))
			rw := httptest.NewRecorder()

			entRepo := new(mock.TypeRepository)
			entRepo.On("GetByName", ctx, "dagger").Return(models.Type{}, models.ErrNoSuchType{"dagger"})
			defer entRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), entRepo, nil)
			handler.ServeHTTP(rw, rr)

			expectedStatus := http.StatusNotFound
			if rw.Code != expectedStatus {
				t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
				return
			}

			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			if err != nil {
				t.Fatalf("error parsing handler response: %v", err)
				return
			}
			expectedReason := `no such type: "dagger"`
			if response.Reason != expectedReason {
				t.Errorf("expected handler to return reason %q, returnd %q instead", expectedReason, response.Reason)
				return
			}
		})
		t.Run("should return HTTP 400 if the incoming payload doesn't contain the required type fields", func(t *testing.T) {
			testCases := []struct {
				payload string
				valid   bool
			}{
				{
					payload: `[{}]`,
				},
				{
					payload: `[{"urn": "whatever"}]`,
				},
				{
					payload: `[{"urn": "whatever", "team": {}}]`,
				},
				{
					payload: `[{"urn": "whatever", "team": ""}]`,
				},
				{
					payload: `[{"urn": "whatever", "team": "de"}]`,
				},
				{
					payload: `[{"urn": "whatever", "team": "de", "name": ""}]`,
				},
				{
					payload: `[{"urn": "whatever", "team": "de", "name": {}}]`,
				},
			}

			for _, testCase := range testCases {
				rr := httptest.NewRequest("PUT", daggerTypeURI, strings.NewReader(testCase.payload))
				rw := httptest.NewRecorder()

				entRepo := new(mock.TypeRepository)
				entRepo.On("GetByName", ctx, "dagger").Return(daggerType, nil)
				defer entRepo.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), entRepo, nil)
				handler.ServeHTTP(rw, rr)

				expectedStatus := http.StatusBadRequest
				if rw.Code != expectedStatus {
					t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
					return
				}
			}
		})
		t.Run("should return HTTP 500 if the resource creation/update fails", func(t *testing.T) {
			t.Run("RecordRepositoryFactory fails", func(t *testing.T) {
				var payload = `[{"urn": "test dagger", "team": "de", "name": "de-dagger-test", "environment": "test"}]`
				rr := httptest.NewRequest("PUT", daggerTypeURI, strings.NewReader(payload))
				rw := httptest.NewRecorder()

				entRepo := new(mock.TypeRepository)
				entRepo.On("GetByName", ctx, "dagger").Return(daggerType, nil)
				defer entRepo.AssertExpectations(t)

				factoryError := errors.New("unknown error")
				recordRepoFac := new(mock.RecordRepositoryFactory)
				recordRepoFac.On("For", daggerType).Return(new(mock.RecordRepository), factoryError)
				defer recordRepoFac.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), entRepo, recordRepoFac)
				handler.ServeHTTP(rw, rr)

				expectedStatus := http.StatusInternalServerError
				if rw.Code != expectedStatus {
					t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
					return
				}

				var response handlers.ErrorResponse
				json.NewDecoder(rw.Body).Decode(&response)
				expectedReason := "Internal Server Error"
				if response.Reason != expectedReason {
					t.Errorf("expected handler to return reason %q, returned %q instead", expectedReason, response.Reason)
					return
				}
			})
			t.Run("RecordRepository fails", func(t *testing.T) {
				var records = []map[string]interface{}{
					{
						"name":        "de-dagger-test",
						"urn":         "test dagger",
						"team":        "de",
						"environment": "test",
					},
				}
				payload, err := json.Marshal(records)
				if err != nil {
					t.Fatalf("error creating request payload: %v", err)
					return
				}
				rr := httptest.NewRequest("PUT", daggerTypeURI, bytes.NewBuffer(payload))
				rw := httptest.NewRecorder()

				entRepo := new(mock.TypeRepository)
				entRepo.On("GetByName", ctx, "dagger").Return(daggerType, nil)
				defer entRepo.AssertExpectations(t)

				repositoryErr := errors.New("unknown error")
				recordRepository := new(mock.RecordRepository)
				recordRepository.On("CreateOrReplaceMany", ctx, records).Return(repositoryErr)
				defer recordRepository.AssertExpectations(t)

				recordRepoFac := new(mock.RecordRepositoryFactory)
				recordRepoFac.On("For", daggerType).Return(recordRepository, nil)
				defer recordRepoFac.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), entRepo, recordRepoFac)
				handler.ServeHTTP(rw, rr)

				expectedStatus := http.StatusInternalServerError
				if rw.Code != expectedStatus {
					t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
					return
				}

				var response handlers.ErrorResponse
				json.NewDecoder(rw.Body).Decode(&response)
				expectedReason := "Internal Server Error"
				if response.Reason != expectedReason {
					t.Errorf("expected handler to return reason %q, returned %q instead", expectedReason, response.Reason)
					return
				}
			})
		})
		t.Run("should return HTTP 400 if the JSON document is invalid", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/v1/types/dagger", bytes.NewBufferString("{"))
			rw := httptest.NewRecorder()

			typeRepo := new(mock.TypeRepository)
			typeRepo.On("GetByName", ctx, "dagger").Return(daggerType, nil)
			defer typeRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo, nil)
			handler.ServeHTTP(rw, rr)

			if rw.Code != http.StatusBadRequest {
				t.Errorf("handler returned HTTP %d, expected HTTP %d", rw.Code, http.StatusBadRequest)
				return
			}

			var res handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&res)
			if err != nil {
				t.Fatalf("error parsing handler response: %v", err)
				return
			}
			expectedReason := "error parsing request body: unexpected EOF"
			if res.Reason != expectedReason {
				t.Errorf("expected handler to return reason %q, returned %q instead", expectedReason, res.Reason)
				return
			}
		})
		t.Run("should return HTTP 200 if the resource is successfully created/update", func(t *testing.T) {
			var records = []models.Record{
				{
					"name":        "de-dagger-test",
					"urn":         "test dagger",
					"team":        "de",
					"environment": "test",
				},
			}
			payload, err := json.Marshal(records)
			if err != nil {
				t.Fatalf("error creating request payload: %v", err)
				return
			}
			rr := httptest.NewRequest("PUT", daggerTypeURI, bytes.NewBuffer(payload))
			rw := httptest.NewRecorder()
			entRepo := new(mock.TypeRepository)
			entRepo.On("GetByName", ctx, "dagger").Return(daggerType, nil)
			defer entRepo.AssertExpectations(t)

			recordRepo := new(mock.RecordRepository)
			recordRepo.On("CreateOrReplaceMany", ctx, records).Return(nil)
			defer recordRepo.AssertExpectations(t)

			recordRepoFac := new(mock.RecordRepositoryFactory)
			recordRepoFac.On("For", daggerType).Return(recordRepo, nil)
			defer recordRepoFac.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), entRepo, recordRepoFac)
			handler.ServeHTTP(rw, rr)

			expectedStatus := http.StatusOK
			if rw.Code != expectedStatus {
				t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
				return
			}

			var response handlers.StatusResponse
			err = json.NewDecoder(rw.Body).Decode(&response)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
				return
			}
			expectedResponse := handlers.StatusResponse{
				Status: "success",
			}

			if reflect.DeepEqual(response, expectedResponse) == false {
				t.Errorf("expected handler to respond with #%v, responded with %#v", expectedResponse, response)
				return
			}
		})
	})
	t.Run("GET /v1/types", func(t *testing.T) {
		type testCase struct {
			Description  string
			RequestURL   string
			ExpectStatus int
			Setup        func(tc *testCase, er *mock.TypeRepository)
			PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
		}

		var types = []models.Type{
			{
				Name:           "bqtable",
				Classification: "dataset",
				Fields: models.TypeFields{
					ID:          "table_name",
					Title:       "table_name",
					Description: "description-bqtable",
					Labels: []string{
						"dataset",
						"project",
					},
				},
			},
			{
				Name:           "dagger",
				Classification: "dataset",
				Fields: models.TypeFields{
					ID:          "urn-dagger",
					Title:       "urn-dagger",
					Description: "description-dagger",
					Labels: []string{
						"topic",
					},
				},
			},
			{
				Name:           "firehose",
				Classification: "dataset",
				Fields: models.TypeFields{
					ID:          "urn-firehose",
					Title:       "urn-firehose",
					Description: "description-firehose",
					Labels: []string{
						"sink",
					},
				},
			},
		}

		var testCases = []testCase{
			{
				Description:  "should return all types",
				RequestURL:   "/v1/types",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetAll", ctx).Return(types, nil)
				},
				PostCheck: func(t *testing.T, tc *testCase, resp *http.Response) error {
					respBody, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					var actual []models.Type
					err = json.Unmarshal(respBody, &actual)
					if err != nil {
						return err
					}
					assert.Equal(t, types, actual)
					return nil
				},
			},
			{
				Description:  "should return 500 status code if failing to fetch types",
				RequestURL:   "/v1/types",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetAll", ctx).Return([]models.Type{}, errors.New("failed to fetch type"))
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				er := new(mock.TypeRepository)
				tc.Setup(&tc, er)

				handler := handlers.NewTypeHandler(new(mock.Logger), er, new(mock.RecordRepositoryFactory))
				rr := httptest.NewRequest("GET", tc.RequestURL, nil)
				rw := httptest.NewRecorder()

				handler.ServeHTTP(rw, rr)
				if rw.Code != tc.ExpectStatus {
					t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
					return
				}

				if tc.PostCheck != nil {
					if err := tc.PostCheck(t, &tc, rw.Result()); err != nil {
						t.Error(err)
					}
				}
			})
		}
	})
	t.Run("DELETE /v1/types/{name}", func(t *testing.T) {
		type testCase struct {
			Description  string
			RequestURL   string
			ExpectStatus int
			Setup        func(tc *testCase, er *mock.TypeRepository)
			PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
		}

		var testCases = []testCase{
			{
				Description:  "should return 204 if delete successes",
				RequestURL:   "/v1/types/sample",
				ExpectStatus: http.StatusNoContent,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("Delete", ctx, "sample").Return(nil)
				},
			},
			{
				Description:  "should return 422 status code if type name is reserved",
				RequestURL:   "/v1/types/sample",
				ExpectStatus: http.StatusUnprocessableEntity,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("Delete", ctx, "sample").Return(models.ErrReservedTypeName{TypeName: "sample"})
				},
			},
			{
				Description:  "should return 500 status code if delete fails",
				RequestURL:   "/v1/types/sample",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("Delete", ctx, "sample").Return(errors.New("failed to delete type"))
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("DELETE", tc.RequestURL, nil)
				rw := httptest.NewRecorder()

				er := new(mock.TypeRepository)
				tc.Setup(&tc, er)
				defer er.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), er, new(mock.RecordRepositoryFactory))
				handler.ServeHTTP(rw, rr)
				if rw.Code != tc.ExpectStatus {
					t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
					return
				}
			})
		}
	})
	t.Run("DELETE /v1/types/{name}/records/{id}", func(t *testing.T) {
		type testCase struct {
			Description  string
			RequestURL   string
			ExpectStatus int
			Setup        func(er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory, rr *mock.RecordRepository)
			PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
		}

		var testCases = []testCase{
			{
				Description:  "should return 204 on success",
				RequestURL:   "/v1/types/sample/records/id-10",
				ExpectStatus: http.StatusNoContent,
				Setup: func(tr *mock.TypeRepository, rrf *mock.RecordRepositoryFactory, rr *mock.RecordRepository) {
					tr.On("GetByName", ctx, "sample").Return(daggerType, nil)
					rrf.On("For", daggerType).Return(rr, nil)
					rr.On("Delete", ctx, "id-10").Return(nil)
				},
			},
			{
				Description:  "should return 404 if type cannot be found",
				RequestURL:   "/v1/types/sample/records/id-10",
				ExpectStatus: http.StatusNotFound,
				Setup: func(tr *mock.TypeRepository, rrf *mock.RecordRepositoryFactory, rr *mock.RecordRepository) {
					tr.On("GetByName", ctx, "sample").Return(models.Type{}, models.ErrNoSuchType{TypeName: daggerType.Name})
				},
			},
			{
				Description:  "should return 500 on error fetching type",
				RequestURL:   "/v1/types/sample/records/id-10",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tr *mock.TypeRepository, rrf *mock.RecordRepositoryFactory, rr *mock.RecordRepository) {
					tr.On("GetByName", ctx, "sample").Return(models.Type{}, errors.New("error fetching type"))
				},
			},
			{
				Description:  "should return 404 when record cannot be found",
				RequestURL:   "/v1/types/sample/records/id-10",
				ExpectStatus: http.StatusNotFound,
				Setup: func(tr *mock.TypeRepository, rrf *mock.RecordRepositoryFactory, rr *mock.RecordRepository) {
					tr.On("GetByName", ctx, "sample").Return(daggerType, nil)
					rrf.On("For", daggerType).Return(rr, nil)
					rr.On("Delete", ctx, "id-10").Return(models.ErrNoSuchRecord{RecordID: "id-10"})
				},
			},
			{
				Description:  "should return 500 on error deleting record",
				RequestURL:   "/v1/types/sample/records/id-10",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tr *mock.TypeRepository, rrf *mock.RecordRepositoryFactory, rr *mock.RecordRepository) {
					tr.On("GetByName", ctx, "sample").Return(daggerType, nil)
					rrf.On("For", daggerType).Return(rr, nil)
					rr.On("Delete", ctx, "id-10").Return(errors.New("error deleting record"))
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("DELETE", tc.RequestURL, nil)
				rw := httptest.NewRecorder()
				typeRepo := new(mock.TypeRepository)
				recordRepo := new(mock.RecordRepository)
				recordRepoFactory := new(mock.RecordRepositoryFactory)
				tc.Setup(typeRepo, recordRepoFactory, recordRepo)
				defer typeRepo.AssertExpectations(t)
				defer recordRepoFactory.AssertExpectations(t)
				defer recordRepo.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo, recordRepoFactory)

				handler.ServeHTTP(rw, rr)
				if rw.Code != tc.ExpectStatus {
					t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
					return
				}
			})
		}
	})
	t.Run("GET /v1/types/{name}/details", func(t *testing.T) {
		type testCase struct {
			Description  string
			RequestURL   string
			ExpectStatus int
			Setup        func(tc *testCase, er *mock.TypeRepository)
			PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
		}

		sampleType := models.Type{
			Name:           "sample",
			Classification: "dataset",
			Fields: models.TypeFields{
				ID:          "urn-dagger",
				Title:       "urn-dagger",
				Description: "description-dagger",
				Labels: []string{
					"topic",
				},
			},
		}

		var testCases = []testCase{
			{
				Description:  "should return type with name given from route parameter",
				RequestURL:   "/v1/types/sample/details",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetByName", ctx, "sample").Return(sampleType, nil)
				},
				PostCheck: func(t *testing.T, tc *testCase, resp *http.Response) error {
					respBody, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					var actual models.Type
					err = json.Unmarshal(respBody, &actual)
					if err != nil {
						return err
					}
					assert.Equal(t, sampleType, actual)
					return nil
				},
			},
			{
				Description:  "should return 500 status code if failing to fetch type",
				RequestURL:   "/v1/types/sample/details",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetByName", ctx, "sample").Return(models.Type{}, errors.New("failed to fetch type"))
				},
			},
			{
				Description:  "should return 404 status code if type could not be found",
				RequestURL:   "/v1/types/wrong_type/details",
				ExpectStatus: http.StatusNotFound,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetByName", ctx, "wrong_type").Return(models.Type{}, models.ErrNoSuchType{
						TypeName: "wrong_type",
					})
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("GET", tc.RequestURL, nil)
				rw := httptest.NewRecorder()
				er := new(mock.TypeRepository)
				tc.Setup(&tc, er)
				defer er.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), er, new(mock.RecordRepositoryFactory))
				handler.ServeHTTP(rw, rr)
				if rw.Code != tc.ExpectStatus {
					t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
					return
				}

				if tc.PostCheck != nil {
					if err := tc.PostCheck(t, &tc, rw.Result()); err != nil {
						t.Error(err)
					}
				}
			})
		}
	})
	t.Run("GET /v1/types/{name}", func(t *testing.T) {
		type testCase struct {
			Description  string
			RequestURL   string
			ExpectStatus int
			Setup        func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory)
			PostCheck    func(tc *testCase, resp *http.Response) error
		}

		var daggerRecords = []models.Record{
			{
				"urn":         "test-fh-1",
				"owner":       "de",
				"created":     "2020-05-13T08:30:04Z",
				"environment": "test",
			},
			{
				"urn":         "test-fh-2",
				"owner":       "de",
				"created":     "2020-05-12T00:00:00Z",
				"environment": "test",
			},
		}

		var testCases = []testCase{
			{
				Description:  "should return an http 404 if the type doesn't exist",
				RequestURL:   "/v1/types/invalid?filter.environment=test",
				ExpectStatus: http.StatusNotFound,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, "invalid").Return(models.Type{}, models.ErrNoSuchType{"invalid"})
				},
			},
			{
				Description:  "should return an http 200 irrespective of environment value",
				RequestURL:   "/v1/types/dagger?filter.environment=nonexisting",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, daggerType.Name).Return(daggerType, nil)
					rr := new(mock.RecordRepository)
					rr.On("GetAll", ctx, map[string][]string{"environment": {"nonexisting"}}).Return(daggerRecords, nil)
					rrf.On("For", daggerType).Return(rr, nil)
				},
			},
			{
				Description:  "should return an http 200 even if the environment is not provided",
				RequestURL:   "/v1/types/dagger",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, daggerType.Name).Return(daggerType, nil)
					rr := new(mock.RecordRepository)
					rr.On("GetAll", ctx, map[string][]string{}).Return(daggerRecords, nil)
					rrf.On("For", daggerType).Return(rr, nil)
				},
			},
			{
				Description:  "should return all records for an type",
				RequestURL:   "/v1/types/dagger?filter.environment=test",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, daggerType.Name).Return(daggerType, nil)
					rr := new(mock.RecordRepository)
					rr.On("GetAll", ctx, map[string][]string{"environment": {"test"}}).Return(daggerRecords, nil)
					rrf.On("For", daggerType).Return(rr, nil)
				},
				PostCheck: func(tc *testCase, resp *http.Response) error {
					var response []models.Record
					err := json.NewDecoder(resp.Body).Decode(&response)
					if err != nil {
						return fmt.Errorf("error parsing response payload: %v", err)
					}
					// TODO: more useful error messages
					if reflect.DeepEqual(response, daggerRecords) == false {
						return fmt.Errorf("expected handler to return %v, returned %v instead", daggerRecords, response)
					}
					return nil
				},
			},
			{
				Description:  "should return the subset of fields specified via select parameter",
				RequestURL:   "/v1/types/dagger?filter.environment=test&select=" + url.QueryEscape("urn,owner"),
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, daggerType.Name).Return(daggerType, nil)
					rr := new(mock.RecordRepository)
					rr.On("GetAll", ctx, map[string][]string{"environment": {"test"}}).Return(daggerRecords, nil)
					rrf.On("For", daggerType).Return(rr, nil)
				},
				PostCheck: func(tc *testCase, resp *http.Response) error {
					var expectRecords []models.Record
					var fields = []string{
						"urn",
						"owner",
					}
					for _, record := range daggerRecords {
						var expectRecord = make(models.Record)
						for _, field := range fields {
							expectRecord[field] = record[field]
						}
						expectRecords = append(expectRecords, expectRecord)
					}
					var response []models.Record
					err := json.NewDecoder(resp.Body).Decode(&response)
					if err != nil {
						return fmt.Errorf("error parsing response payload: %v", err)
					}
					if reflect.DeepEqual(response, expectRecords) == false {
						return fmt.Errorf("expected handler to return %v, returned %v instead", expectRecords, response)
					}
					return nil
				},
			},
			{
				Description:  "should support landscape and entity filters",
				RequestURL:   "/v1/types/dagger?filter.environment=test&filter.landscape=id&filter.entity=odpf",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, daggerType.Name).Return(daggerType, nil)
					rr := new(mock.RecordRepository)
					filters := map[string][]string{
						"landscape":   {"id"},
						"entity":      {"odpf"},
						"environment": {"test"},
					}
					rr.On("GetAll", ctx, filters).Return(daggerRecords, nil)
					rrf.On("For", daggerType).Return(rr, nil)
				},
				PostCheck: func(tc *testCase, resp *http.Response) error {
					var response []models.Record
					err := json.NewDecoder(resp.Body).Decode(&response)
					if err != nil {
						return fmt.Errorf("error parsing response payload: %v", err)
					}
					if reflect.DeepEqual(response, daggerRecords) == false {
						return fmt.Errorf("expected handler to return %v, returned %v instead", daggerRecords, response)
					}
					return nil
				},
			},
			{
				Description:  "(internal) should return http 500 if the handler fails to construct record repository",
				RequestURL:   "/v1/types/dagger?filter.environment=test",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, daggerType.Name).Return(daggerType, nil)
					rr := new(mock.RecordRepository)
					err := fmt.Errorf("something went wrong")
					rrf.On("For", daggerType).Return(rr, err)
				},
			},
			{
				Description:  "(internal) should return an http 500 if calling recordRepository.GetAll fails",
				RequestURL:   "/v1/types/dagger?filter.environment=test",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, daggerType.Name).Return(daggerType, nil)
					rr := new(mock.RecordRepository)
					err := fmt.Errorf("temporarily unavailable")
					rr.On("GetAll", ctx, map[string][]string{"environment": {"test"}}).Return([]models.Record{}, err)
					rrf.On("For", daggerType).Return(rr, nil)
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("GET", tc.RequestURL, nil)
				rw := httptest.NewRecorder()
				er := new(mock.TypeRepository)
				rrf := new(mock.RecordRepositoryFactory)
				tc.Setup(&tc, er, rrf)

				handler := handlers.NewTypeHandler(new(mock.Logger), er, rrf)
				handler.ServeHTTP(rw, rr)
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
	t.Run("GET /v1/types/{name}/{id}", func(t *testing.T) {
		var deployment01 = map[string]interface{}{
			"contents": "data",
		}
		type testCase struct {
			Description  string
			RequestURL   string
			ExpectStatus int
			Setup        func(er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory)
			PostCheck    func(resp *http.Response) error
		}

		var testCases = []testCase{
			{
				Description:  `should return http 404 if the record doesn't exist`,
				RequestURL:   "/v1/types/dagger/record01",
				ExpectStatus: http.StatusNotFound,
				Setup: func(er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, "dagger").Return(daggerType, nil)
					recordRepo := new(mock.RecordRepository)
					recordRepo.On("GetByID", ctx, "record01").Return(map[string]interface{}{}, models.ErrNoSuchRecord{"record01"})
					rrf.On("For", daggerType).Return(recordRepo, nil)
				},
			},
			{
				Description:  `should return http 404 if the type doesn't exist`,
				RequestURL:   "/v1/types/nonexistant/record",
				ExpectStatus: http.StatusNotFound,
				Setup: func(er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, "nonexistant").Return(models.Type{}, models.ErrNoSuchType{"nonexistant"})
				},
			},
			{
				Description:  "(internal) should return an http 500 if the handler fails to construct recordRepository",
				RequestURL:   "/v1/types/dagger/record",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					errSomethingBadHappened := fmt.Errorf("something bad happened")
					er.On("GetByName", ctx, "dagger").Return(daggerType, nil)
					rrf.On("For", daggerType).Return(new(mock.RecordRepository), errSomethingBadHappened)
				},
			},
			{
				Description:  "should return http 200 status along with the record, if found",
				RequestURL:   "/v1/types/dagger/deployment01",
				ExpectStatus: http.StatusOK,
				Setup: func(er *mock.TypeRepository, rrf *mock.RecordRepositoryFactory) {
					er.On("GetByName", ctx, "dagger").Return(daggerType, nil)
					recordRepo := new(mock.RecordRepository)
					recordRepo.On("GetByID", ctx, "deployment01").Return(deployment01, nil)
					rrf.On("For", daggerType).Return(recordRepo, nil)
				},
				PostCheck: func(r *http.Response) error {
					var record models.Record
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
				rr := httptest.NewRequest("GET", tc.RequestURL, nil)
				rw := httptest.NewRecorder()
				typeRepo := new(mock.TypeRepository)
				recordRepoFac := new(mock.RecordRepositoryFactory)
				if tc.Setup != nil {
					tc.Setup(typeRepo, recordRepoFac)
				}

				handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo, recordRepoFac)
				handler.ServeHTTP(rw, rr)

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
