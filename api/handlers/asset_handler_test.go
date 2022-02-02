package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/odpf/salt/log"

	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/lib/mock"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	logger = log.NewNoop()
)

func TestAssetHandlerUpsert(t *testing.T) {
	var validPayload = `{"urn": "test dagger", "type": "table", "name": "de-dagger-test", "service": "kafka", "data": {}}`

	t.Run("should return HTTP 400 for invalid payload", func(t *testing.T) {
		testCases := []struct {
			description string
			payload     string
		}{
			{
				description: "empty object",
				payload:     `{}`,
			},
			{
				description: "empty urn",
				payload:     `{"urn": "", "name": "some-name", "data": {}, "service": "some-service", "type": "table"}`,
			},
			{
				description: "empty name",
				payload:     `{"urn": "some-urn", "name": "", "data": {}, "service": "some-service", "type": "table"}`,
			},
			{
				description: "empty data",
				payload:     `{"urn": "some-urn", "name": "some-name", "data": null, "service": "some-service", "type": "table"}`,
			},
			{
				description: "empty service",
				payload:     `{"urn": "some-urn", "name": "some-name", "data": {}, "service": "", "type": "table"}`,
			},
			{
				description: "empty type",
				payload:     `{"urn": "some-urn", "name": "some-name", "data": {}, "service": "some-service", "type": ""}`,
			},
			{
				description: "invalid type",
				payload:     `{"urn": "some-urn", "name": "some-name", "data": {}, "service": "some-service", "type": "invalid_type"}`,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.description, func(t *testing.T) {
				rw := httptest.NewRecorder()
				rr := httptest.NewRequest("PUT", "/", strings.NewReader(testCase.payload))

				handler := handlers.NewAssetHandler(logger, nil, nil)
				handler.Upsert(rw, rr)

				expectedStatus := http.StatusBadRequest
				if rw.Code != expectedStatus {
					t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
					return
				}
			})
		}
	})
	t.Run("should return HTTP 500 if the asset creation/update fails", func(t *testing.T) {
		t.Run("AssetRepository fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mock.AssetRepository)
			ar.On("Upsert", rr.Context(), tmock.AnythingOfType("*asset.Asset")).Return(expectedErr)
			defer ar.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, nil)
			handler.Upsert(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
		t.Run("DiscoveryRepository fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mock.AssetRepository)
			ar.On("Upsert", rr.Context(), tmock.AnythingOfType("*asset.Asset")).Return(nil)
			defer ar.AssertExpectations(t)

			dr := new(mock.DiscoveryRepository)
			dr.On("Upsert", rr.Context(), tmock.AnythingOfType("asset.Asset")).Return(expectedErr)
			defer dr.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, dr)
			handler.Upsert(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
	})
	t.Run("should return HTTP 200 and asset's ID if the asset is successfully created/updated", func(t *testing.T) {
		ast := asset.Asset{
			URN:     "test dagger",
			Type:    asset.TypeTable,
			Name:    "de-dagger-test",
			Service: "kafka",
			Data:    map[string]interface{}{},
		}
		assetWithID := ast
		assetWithID.ID = uuid.New().String()

		rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
		rw := httptest.NewRecorder()

		ar := new(mock.AssetRepository)
		ar.On("Upsert", rr.Context(), &ast).Return(nil).Run(func(args tmock.Arguments) {
			argAsset := args.Get(1).(*asset.Asset)
			argAsset.ID = assetWithID.ID
		})
		defer ar.AssertExpectations(t)

		dr := new(mock.DiscoveryRepository)
		dr.On("Upsert", rr.Context(), assetWithID).Return(nil)
		defer dr.AssertExpectations(t)

		handler := handlers.NewAssetHandler(logger, ar, dr)
		handler.Upsert(rw, rr)

		assert.Equal(t, http.StatusOK, rw.Code)
		var response map[string]interface{}
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)

		assetID, exists := response["id"]
		assert.True(t, exists)
		assert.Equal(t, assetWithID.ID, assetID)
	})
}

func TestAssetHandlerDelete(t *testing.T) {
	type testCase struct {
		Description  string
		AssetID      string
		ExpectStatus int
		Setup        func(context.Context, *mock.AssetRepository, *mock.DiscoveryRepository)
		PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  "should return 404 when asset cannot be found",
			AssetID:      "id-10",
			ExpectStatus: http.StatusNotFound,
			Setup: func(ctx context.Context, ar *mock.AssetRepository, dr *mock.DiscoveryRepository) {
				ar.On("Delete", ctx, "id-10").Return(asset.NotFoundError{AssetID: "id-10"})
			},
		},
		{
			Description:  "should return 500 on error deleting asset",
			AssetID:      "id-10",
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mock.AssetRepository, dr *mock.DiscoveryRepository) {
				ar.On("Delete", ctx, "id-10").Return(errors.New("error deleting asset"))
			},
		},
		{
			Description:  "should return 500 on error deleting asset from discovery",
			AssetID:      "id-10",
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mock.AssetRepository, dr *mock.DiscoveryRepository) {
				ar.On("Delete", ctx, "id-10").Return(nil)
				dr.On("Delete", ctx, "id-10").Return(asset.NotFoundError{AssetID: "id-10"})
			},
		},
		{
			Description:  "should return 204 on success",
			AssetID:      "id-10",
			ExpectStatus: http.StatusNoContent,
			Setup: func(ctx context.Context, ar *mock.AssetRepository, dr *mock.DiscoveryRepository) {
				ar.On("Delete", ctx, "id-10").Return(nil)
				dr.On("Delete", ctx, "id-10").Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			rr := httptest.NewRequest("DELETE", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"id": tc.AssetID,
			})

			ar := new(mock.AssetRepository)
			dr := new(mock.DiscoveryRepository)
			tc.Setup(rr.Context(), ar, dr)
			defer ar.AssertExpectations(t)

			handler := handlers.NewAssetHandler(logger, ar, dr)
			handler.Delete(rw, rr)

			if rw.Code != tc.ExpectStatus {
				t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
				return
			}
		})
	}
}

func TestAssetHandlerGetByID(t *testing.T) {
	var (
		assetID = "id-1"
		ast     = asset.Asset{
			ID: assetID,
		}
	)

	type testCase struct {
		Description  string
		ExpectStatus int
		Setup        func(context.Context, *mock.AssetRepository)
		PostCheck    func(resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  `should return http 404 if asset doesn't exist`,
			ExpectStatus: http.StatusNotFound,
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: assetID})
			},
		},
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  "should return http 200 status along with the asset, if found",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(ast, nil)
			},
			PostCheck: func(r *http.Response) error {
				var responsePayload asset.Asset
				err := json.NewDecoder(r.Body).Decode(&responsePayload)
				if err != nil {
					return fmt.Errorf("error reading response body: %w", err)
				}
				if reflect.DeepEqual(responsePayload, ast) == false {
					return fmt.Errorf("expected returned asset to be to be %+v, was %+v", ast, responsePayload)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"id": assetID,
			})
			ar := new(mock.AssetRepository)
			tc.Setup(rr.Context(), ar)

			handler := handlers.NewAssetHandler(logger, ar, nil)
			handler.GetByID(rw, rr)

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
}

func TestAssetHandlerGet(t *testing.T) {
	type testCase struct {
		Description  string
		Querystring  string
		ExpectStatus int
		Setup        func(context.Context, *mock.AssetRepository)
		PostCheck    func(resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("Get", ctx, asset.Config{}).Return([]asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should return http 500 if fetching total fails`,
			Querystring:  "?with_total=1",
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("Get", ctx, asset.Config{}).Return([]asset.Asset{}, nil)
				ar.On("GetCount", ctx, asset.Config{}).Return(0, errors.New("unknown error"))
			},
		},
		{
			Description:  `should parse querystring to get config`,
			Querystring:  "?text=asd&type=table&service=bigquery&size=30&offset=50",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("Get", ctx, asset.Config{
					Text:    "asd",
					Type:    "table",
					Service: "bigquery",
					Size:    30,
					Offset:  50,
				}).Return([]asset.Asset{}, nil)
			},
		},
		{
			Description:  "should return http 200 status along with list of assets",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("Get", ctx, asset.Config{}).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
				}, nil)
			},
			PostCheck: func(r *http.Response) error {
				type responsePayload struct {
					Data []asset.Asset `json:"data"`
				}
				expected := responsePayload{
					Data: []asset.Asset{
						{ID: "testid-1"},
						{ID: "testid-2"},
					},
				}

				var actual responsePayload
				err := json.NewDecoder(r.Body).Decode(&actual)
				if err != nil {
					return fmt.Errorf("error reading response body: %w", err)
				}
				if reflect.DeepEqual(actual, expected) == false {
					return fmt.Errorf("expected payload to be to be %+v, was %+v", expected, actual)
				}
				return nil
			},
		},
		{
			Description:  "should return total in the payload if with_total flag is given",
			ExpectStatus: http.StatusOK,
			Querystring:  "?with_total=true&text=dsa&type=job&service=kafka&size=10&offset=5",
			Setup: func(ctx context.Context, ar *mock.AssetRepository) {
				ar.On("Get", ctx, asset.Config{
					Text:    "dsa",
					Type:    "job",
					Service: "kafka",
					Size:    10,
					Offset:  5,
				}).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
					{ID: "testid-3"},
				}, nil)
				ar.On("GetCount", ctx, asset.Config{
					Text:    "dsa",
					Type:    "job",
					Service: "kafka",
				}).Return(150, nil)
			},
			PostCheck: func(r *http.Response) error {
				type responsePayload struct {
					Total int           `json:"total"`
					Data  []asset.Asset `json:"data"`
				}
				expected := responsePayload{
					Total: 150,
					Data: []asset.Asset{
						{ID: "testid-1"},
						{ID: "testid-2"},
						{ID: "testid-3"},
					},
				}

				var actual responsePayload
				err := json.NewDecoder(r.Body).Decode(&actual)
				if err != nil {
					return fmt.Errorf("error reading response body: %w", err)
				}
				if reflect.DeepEqual(actual, expected) == false {
					return fmt.Errorf("expected payload to be to be %+v, was %+v", expected, actual)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			rr := httptest.NewRequest("GET", "/"+tc.Querystring, nil)
			rw := httptest.NewRecorder()

			ar := new(mock.AssetRepository)
			tc.Setup(rr.Context(), ar)

			handler := handlers.NewAssetHandler(logger, ar, nil)
			handler.Get(rw, rr)

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
}
