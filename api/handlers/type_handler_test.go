package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/record"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestTypeHandler(t *testing.T) {
	var (
		daggerType = record.Type{
			Name:           "dagger",
			Classification: record.TypeClassificationResource,
		}
		ctx = tmock.AnythingOfType("*context.valueCtx")
	)

	t.Run("Upsert", func(t *testing.T) {
		validPayloadRaw, err := json.Marshal(daggerType)
		if err != nil {
			t.Fatalf("error preparing request payload: %v", err)
			return
		}
		t.Run("should return HTTP 400 if the JSON document is invalid", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", bytes.NewBufferString("{"))
			rw := httptest.NewRecorder()

			handler := handlers.NewTypeHandler(new(mock.Logger), nil)
			handler.Upsert(rw, rr)

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
				payload        record.Type
				expectedReason string
			}{
				{
					payload:        record.Type{},
					expectedReason: "'name' is required",
				},
				{
					payload: record.Type{
						Name: "foo",
					},
					expectedReason: "'classification' is required",
				},
			}

			for _, testCase := range testCases {

				raw, err := json.Marshal(testCase.payload)
				if err != nil {
					t.Fatalf("error creating test payload: %v", err)
					return
				}
				rr := httptest.NewRequest("PUT", "/", bytes.NewBuffer(raw))
				rw := httptest.NewRecorder()

				handler := handlers.NewTypeHandler(new(mock.Logger), nil)
				handler.Upsert(rw, rr)

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
			rr := httptest.NewRequest("PUT", "/", bytes.NewBuffer(validPayloadRaw))
			rw := httptest.NewRecorder()

			typeRepo := new(mock.TypeRepository)
			typeRepo.On("CreateOrReplace", context.Background(), daggerType).Return(nil)
			defer typeRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo)
			handler.Upsert(rw, rr)

			expectedStatus := http.StatusCreated
			if rw.Code != expectedStatus {
				t.Errorf("expected handler to HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
				return
			}
		})
		t.Run("should return 422 if type name is reserved", func(t *testing.T) {
			expectedErr := record.ErrReservedTypeName{TypeName: daggerType.Name}

			rr := httptest.NewRequest("PUT", "/", bytes.NewBuffer(validPayloadRaw))
			rw := httptest.NewRecorder()

			typeRepo := new(mock.TypeRepository)
			typeRepo.On("CreateOrReplace", context.Background(), daggerType).Return(expectedErr)
			defer typeRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo)
			handler.Upsert(rw, rr)

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
			rr := httptest.NewRequest("PUT", "/", bytes.NewBuffer(validPayloadRaw))
			rw := httptest.NewRecorder()

			creationErr := fmt.Errorf("failed to write to elasticsearch")
			typeRepo := new(mock.TypeRepository)
			typeRepo.On("CreateOrReplace", context.Background(), daggerType).Return(creationErr)
			defer typeRepo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), typeRepo)
			handler.Upsert(rw, rr)

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
			typeWithInvalidClassification := &record.Type{
				Name:           "application",
				Classification: "unknown",
			}
			var payload bytes.Buffer
			err := json.NewEncoder(&payload).Encode(typeWithInvalidClassification)
			if err != nil {
				t.Fatalf("error preparing test data: %v", err)
				return
			}
			rr := httptest.NewRequest("PUT", "/", &payload)
			rw := httptest.NewRecorder()

			handler := handlers.NewTypeHandler(new(mock.Logger), nil)
			handler.Upsert(rw, rr)

			expectedCode := 400
			if rw.Code != expectedCode {
				t.Errorf("expected handler to return HTTP %d, returned %d instead", expectedCode, rw.Code)
			}
		})
		t.Run("should lowercase type name before commiting it to storage", func(t *testing.T) {
			ent := &record.Type{
				Name:           "DAGGER",
				Classification: record.TypeClassificationResource,
			}
			expectEnt := *ent
			expectEnt.Name = strings.ToLower(ent.Name)

			var payload bytes.Buffer
			err := json.NewEncoder(&payload).Encode(ent)
			if err != nil {
				t.Fatalf("error preparing test data: %v", err)
				return
			}

			rr := httptest.NewRequest("PUT", "/", &payload)
			rw := httptest.NewRecorder()

			repo := new(mock.TypeRepository)
			repo.On("CreateOrReplace", context.Background(), expectEnt).Return(nil)
			defer repo.AssertExpectations(t)

			handler := handlers.NewTypeHandler(new(mock.Logger), repo)
			handler.Upsert(rw, rr)
		})
	})
	t.Run("Get", func(t *testing.T) {
		type testCase struct {
			Description  string
			ExpectStatus int
			Setup        func(tc *testCase, er *mock.TypeRepository)
			PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
		}

		var types = []record.Type{
			{
				Name:           "bqtable",
				Classification: "dataset",
			},
			{
				Name:           "dagger",
				Classification: "dataset",
			},
			{
				Name:           "firehose",
				Classification: "dataset",
			},
		}

		var testCases = []testCase{
			{
				Description:  "should return all types",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetAll", context.Background()).Return(types, nil)
				},
				PostCheck: func(t *testing.T, tc *testCase, resp *http.Response) error {
					respBody, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					var actual []record.Type
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
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetAll", context.Background()).Return([]record.Type{}, errors.New("failed to fetch type"))
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				er := new(mock.TypeRepository)
				tc.Setup(&tc, er)

				handler := handlers.NewTypeHandler(new(mock.Logger), er)
				rr := httptest.NewRequest("GET", "/", nil)
				rw := httptest.NewRecorder()

				handler.Get(rw, rr)
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
	t.Run("Delete", func(t *testing.T) {
		type testCase struct {
			Description  string
			TypeName     string
			ExpectStatus int
			Setup        func(tc *testCase, er *mock.TypeRepository)
			PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
		}

		var testCases = []testCase{
			{
				Description:  "should return 204 if delete successes",
				TypeName:     "sample",
				ExpectStatus: http.StatusNoContent,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("Delete", ctx, "sample").Return(nil)
				},
			},
			{
				Description:  "should return 422 status code if type name is reserved",
				TypeName:     "sample",
				ExpectStatus: http.StatusUnprocessableEntity,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("Delete", ctx, "sample").Return(record.ErrReservedTypeName{TypeName: "sample"})
				},
			},
			{
				Description:  "should return 500 status code if delete fails",
				TypeName:     "sample",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("Delete", ctx, "sample").Return(errors.New("failed to delete type"))
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("DELETE", "/", nil)
				rw := httptest.NewRecorder()
				rr = mux.SetURLVars(rr, map[string]string{
					"name": tc.TypeName,
				})

				er := new(mock.TypeRepository)
				tc.Setup(&tc, er)
				defer er.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), er)
				handler.Delete(rw, rr)
				if rw.Code != tc.ExpectStatus {
					t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
					return
				}
			})
		}
	})
	t.Run("Find", func(t *testing.T) {
		type testCase struct {
			Description  string
			TypeName     string
			ExpectStatus int
			Setup        func(tc *testCase, er *mock.TypeRepository)
			PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
		}

		sampleType := record.Type{
			Name:           "sample",
			Classification: "dataset",
		}

		var testCases = []testCase{
			{
				Description:  "should return type with name given from route parameter",
				TypeName:     "sample",
				ExpectStatus: http.StatusOK,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetByName", ctx, "sample").Return(sampleType, nil)
				},
				PostCheck: func(t *testing.T, tc *testCase, resp *http.Response) error {
					respBody, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					var actual record.Type
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
				TypeName:     "sample",
				ExpectStatus: http.StatusInternalServerError,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetByName", ctx, "sample").Return(record.Type{}, errors.New("failed to fetch type"))
				},
			},
			{
				Description:  "should return 404 status code if type could not be found",
				TypeName:     "wrong_type",
				ExpectStatus: http.StatusNotFound,
				Setup: func(tc *testCase, er *mock.TypeRepository) {
					er.On("GetByName", ctx, "wrong_type").Return(record.Type{}, record.ErrNoSuchType{
						TypeName: "wrong_type",
					})
				},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rr := httptest.NewRequest("GET", "/", nil)
				rw := httptest.NewRecorder()
				rr = mux.SetURLVars(rr, map[string]string{
					"name": tc.TypeName,
				})
				er := new(mock.TypeRepository)
				tc.Setup(&tc, er)
				defer er.AssertExpectations(t)

				handler := handlers.NewTypeHandler(new(mock.Logger), er)
				handler.Find(rw, rr)

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
}
