package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/odpf/salt/log"

	"github.com/odpf/columbus/api/httpapi/handlers"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	logger = log.NewNoop()
)

func TestAssetHandlerUpsert(t *testing.T) {
	var userID = uuid.NewString()
	var validPayload = `{
		"urn": "test dagger",
		"type": "table",
		"name": "de-dagger-test",
		"service": "kafka",
		"data": {},
		"upstreams": [
			{"urn": "upstream-1", "type": "job", "service": "optimus"}
		],
		"downstreams": [
			{"urn": "downstream-1", "type": "dashboard", "service": "metabase"},
			{"urn": "downstream-2", "type": "dashboard", "service": "tableau"}
		]
	}`

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
				ctx := user.NewContext(rr.Context(), userID)
				rr = rr.WithContext(ctx)

				handler := handlers.NewAssetHandler(logger, nil, nil, nil, nil)
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
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mocks.AssetRepository)
			ar.On("Upsert", rr.Context(), mock.AnythingOfType("*asset.Asset")).Return("1234-5678", expectedErr)
			defer ar.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, nil, nil, nil)
			handler.Upsert(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
		t.Run("DiscoveryRepository fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mocks.AssetRepository)
			ar.On("Upsert", rr.Context(), mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil, nil)
			defer ar.AssertExpectations(t)

			dr := new(mocks.DiscoveryRepository)
			dr.On("Upsert", rr.Context(), mock.AnythingOfType("asset.Asset")).Return(expectedErr)
			defer dr.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, dr, nil, nil)
			handler.Upsert(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
		t.Run("LineageRepository fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mocks.AssetRepository)
			ar.On("Upsert", rr.Context(), mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil, nil)
			defer ar.AssertExpectations(t)

			dr := new(mocks.DiscoveryRepository)
			dr.On("Upsert", rr.Context(), mock.AnythingOfType("asset.Asset")).Return(nil)
			defer dr.AssertExpectations(t)

			lr := new(mocks.LineageRepository)
			lr.On("Upsert", rr.Context(),
				mock.AnythingOfType("lineage.Node"),
				mock.AnythingOfType("[]lineage.Node"),
				mock.AnythingOfType("[]lineage.Node"),
			).Return(expectedErr)
			defer lr.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, dr, nil, lr)
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
			URN:       "test dagger",
			Type:      asset.TypeTable,
			Name:      "de-dagger-test",
			Service:   "kafka",
			UpdatedBy: user.User{ID: userID},
			Data:      map[string]interface{}{},
		}
		upstreams := []lineage.Node{
			{URN: "upstream-1", Type: asset.TypeJob, Service: "optimus"},
		}
		downstreams := []lineage.Node{
			{URN: "downstream-1", Type: asset.TypeDashboard, Service: "metabase"},
			{URN: "downstream-2", Type: asset.TypeDashboard, Service: "tableau"},
		}

		assetWithID := ast
		assetWithID.ID = uuid.New().String()

		rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rw := httptest.NewRecorder()

		ar := new(mocks.AssetRepository)
		ar.On("Upsert", rr.Context(), &ast).Return(assetWithID.ID, nil, nil).Run(func(args mock.Arguments) {
			argAsset := args.Get(1).(*asset.Asset)
			argAsset.ID = assetWithID.ID
		})
		defer ar.AssertExpectations(t)

		dr := new(mocks.DiscoveryRepository)
		dr.On("Upsert", rr.Context(), assetWithID).Return(nil)
		defer dr.AssertExpectations(t)

		lr := new(mocks.LineageRepository)
		lr.On("Upsert", rr.Context(),
			lineage.Node{
				URN:     ast.URN,
				Type:    ast.Type,
				Service: ast.Service,
			},
			upstreams,
			downstreams,
		).Return(nil)
		defer lr.AssertExpectations(t)

		handler := handlers.NewAssetHandler(logger, ar, dr, nil, lr)
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

func TestAssetHandlerUpsertPatch(t *testing.T) {
	var userID = uuid.NewString()
	var validPayload = `{
		"asset": {
			"urn": "test dagger",
			"type": "table",
			"name": "new-name",
			"service": "kafka",
			"data": {}
		},
		"upstreams": [
			{"urn": "upstream-1", "type": "job", "service": "optimus"}
		],
		"downstreams": [
			{"urn": "downstream-1", "type": "dashboard", "service": "metabase"},
			{"urn": "downstream-2", "type": "dashboard", "service": "tableau"}
		]
	}`
	var currentAsset = asset.Asset{
		URN:       "test dagger",
		Type:      asset.TypeTable,
		Name:      "old-name", // this value will be updated
		Service:   "kafka",
		UpdatedBy: user.User{ID: userID},
		Data:      map[string]interface{}{},
	}

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
				description: "empty asset",
				payload:     `{"asset": {}}`,
			},
			{
				description: "empty urn",
				payload:     `{"asset": {"urn": "", "name": "some-name", "data": {}, "service": "some-service", "type": "table"}}`,
			},
			{
				description: "empty service",
				payload:     `{"asset": {"urn": "some-urn", "name": "some-name", "data": {}, "service": "", "type": "table"}}`,
			},
			{
				description: "empty type",
				payload:     `{"asset": {"urn": "some-urn", "name": "some-name", "data": {}, "service": "some-service", "type": ""}}`,
			},
			{
				description: "invalid type",
				payload:     `{"asset": {"urn": "some-urn", "name": "some-name", "data": {}, "service": "some-service", "type": "invalid_type"}}`,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.description, func(t *testing.T) {
				rw := httptest.NewRecorder()

				rr := httptest.NewRequest("PUT", "/", strings.NewReader(testCase.payload))
				ctx := user.NewContext(rr.Context(), userID)
				rr = rr.WithContext(ctx)

				handler := handlers.NewAssetHandler(logger, nil, nil, nil, nil)
				handler.UpsertPatch(rw, rr)

				expectedStatus := http.StatusBadRequest
				if rw.Code != expectedStatus {
					t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
					return
				}
			})
		}
	})

	t.Run("should return HTTP 500 if the asset creation/update fails", func(t *testing.T) {
		t.Run("Finding asset fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mocks.AssetRepository)
			ar.On("Find", rr.Context(), "test dagger", asset.TypeTable, "kafka").Return(currentAsset, expectedErr)
			defer ar.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, nil, nil, nil)
			handler.UpsertPatch(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
		t.Run("AssetRepository Upsert fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mocks.AssetRepository)
			ar.On("Find", rr.Context(), "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
			ar.On("Upsert", rr.Context(), mock.AnythingOfType("*asset.Asset")).Return("1234-5678", expectedErr)
			defer ar.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, nil, nil, nil)
			handler.UpsertPatch(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
		t.Run("DiscoveryRepository fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mocks.AssetRepository)
			ar.On("Find", rr.Context(), "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
			ar.On("Upsert", rr.Context(), mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil, nil)
			defer ar.AssertExpectations(t)

			dr := new(mocks.DiscoveryRepository)
			dr.On("Upsert", rr.Context(), mock.AnythingOfType("asset.Asset")).Return(expectedErr)
			defer dr.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, dr, nil, nil)
			handler.UpsertPatch(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
		t.Run("LineageRepository fails", func(t *testing.T) {
			rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			expectedErr := errors.New("unknown error")

			ar := new(mocks.AssetRepository)
			ar.On("Find", rr.Context(), "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
			ar.On("Upsert", rr.Context(), mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil, nil)
			defer ar.AssertExpectations(t)

			dr := new(mocks.DiscoveryRepository)
			dr.On("Upsert", rr.Context(), mock.AnythingOfType("asset.Asset")).Return(nil)
			defer dr.AssertExpectations(t)

			lr := new(mocks.LineageRepository)
			lr.On("Upsert", rr.Context(),
				mock.AnythingOfType("lineage.Node"),
				mock.AnythingOfType("[]lineage.Node"),
				mock.AnythingOfType("[]lineage.Node"),
			).Return(expectedErr)
			defer lr.AssertExpectations(t)

			rr.Context()
			handler := handlers.NewAssetHandler(logger, ar, dr, nil, lr)
			handler.UpsertPatch(rw, rr)

			assert.Equal(t, http.StatusInternalServerError, rw.Code)
			var response handlers.ErrorResponse
			err := json.NewDecoder(rw.Body).Decode(&response)
			require.NoError(t, err)
			assert.Contains(t, response.Reason, "Internal Server Error")
		})
	})

	t.Run("should return HTTP 200 and asset's ID if the asset is successfully patched", func(t *testing.T) {
		patchedAsset := asset.Asset{
			URN:       "test dagger",
			Type:      asset.TypeTable,
			Name:      "new-name",
			Service:   "kafka",
			UpdatedBy: user.User{ID: userID},
			Data:      map[string]interface{}{},
		}
		upstreams := []lineage.Node{
			{URN: "upstream-1", Type: asset.TypeJob, Service: "optimus"},
		}
		downstreams := []lineage.Node{
			{URN: "downstream-1", Type: asset.TypeDashboard, Service: "metabase"},
			{URN: "downstream-2", Type: asset.TypeDashboard, Service: "tableau"},
		}

		assetWithID := patchedAsset
		assetWithID.ID = uuid.New().String()

		rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rw := httptest.NewRecorder()

		ar := new(mocks.AssetRepository)
		ar.On("Find", rr.Context(), "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
		ar.On("Upsert", rr.Context(), &patchedAsset).Return(assetWithID.ID, nil, nil).Run(func(args mock.Arguments) {
			argAsset := args.Get(1).(*asset.Asset)
			argAsset.ID = assetWithID.ID
		})
		defer ar.AssertExpectations(t)

		dr := new(mocks.DiscoveryRepository)
		dr.On("Upsert", rr.Context(), assetWithID).Return(nil)
		defer dr.AssertExpectations(t)

		lr := new(mocks.LineageRepository)
		lr.On("Upsert", rr.Context(),
			lineage.Node{
				URN:     patchedAsset.URN,
				Type:    patchedAsset.Type,
				Service: patchedAsset.Service,
			},
			upstreams,
			downstreams,
		).Return(nil)
		defer lr.AssertExpectations(t)

		handler := handlers.NewAssetHandler(logger, ar, dr, nil, lr)
		handler.UpsertPatch(rw, rr)

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
		Setup        func(context.Context, *testCase, *mocks.AssetRepository, *mocks.DiscoveryRepository)
		PostCheck    func(t *testing.T, tc *testCase, resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  "should return 400 when asset id is not uuid",
			AssetID:      "not-uuid",
			ExpectStatus: http.StatusBadRequest,
			Setup: func(ctx context.Context, tc *testCase, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.On("Delete", ctx, tc.AssetID).Return(asset.InvalidError{AssetID: tc.AssetID})
			},
		},
		{
			Description:  "should return 404 when asset cannot be found",
			AssetID:      uuid.NewString(),
			ExpectStatus: http.StatusNotFound,
			Setup: func(ctx context.Context, tc *testCase, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.On("Delete", ctx, tc.AssetID).Return(asset.NotFoundError{AssetID: tc.AssetID})
			},
		},
		{
			Description:  "should return 500 on error deleting asset",
			AssetID:      uuid.NewString(),
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, tc *testCase, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.On("Delete", ctx, tc.AssetID).Return(errors.New("error deleting asset"))
			},
		},
		{
			Description:  "should return 500 on error deleting asset from discovery",
			AssetID:      uuid.NewString(),
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, tc *testCase, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.On("Delete", ctx, tc.AssetID).Return(nil)
				dr.On("Delete", ctx, tc.AssetID).Return(asset.NotFoundError{AssetID: tc.AssetID})
			},
		},
		{
			Description:  "should return 204 on success",
			AssetID:      uuid.NewString(),
			ExpectStatus: http.StatusNoContent,
			Setup: func(ctx context.Context, tc *testCase, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.On("Delete", ctx, tc.AssetID).Return(nil)
				dr.On("Delete", ctx, tc.AssetID).Return(nil)
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

			ar := new(mocks.AssetRepository)
			dr := new(mocks.DiscoveryRepository)
			tc.Setup(rr.Context(), &tc, ar, dr)
			defer ar.AssertExpectations(t)

			handler := handlers.NewAssetHandler(logger, ar, dr, nil, nil)
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
		assetID = uuid.NewString()
		ast     = asset.Asset{
			ID: assetID,
		}
	)

	type testCase struct {
		Description  string
		ExpectStatus int
		Setup        func(context.Context, *mocks.AssetRepository)
		PostCheck    func(resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  `should return http 400 if asset id is not uuid`,
			ExpectStatus: http.StatusBadRequest,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return http 404 if asset doesn't exist`,
			ExpectStatus: http.StatusNotFound,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: assetID})
			},
		},
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  "should return http 200 status along with the asset, if found",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(ast, nil, nil)
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
			ar := new(mocks.AssetRepository)
			tc.Setup(rr.Context(), ar)

			handler := handlers.NewAssetHandler(logger, ar, nil, nil, nil)
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
		Setup        func(context.Context, *mocks.AssetRepository)
		PostCheck    func(resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{}).Return([]asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should return http 500 if fetching total fails`,
			Querystring:  "?with_total=1",
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{}).Return([]asset.Asset{}, nil, nil)
				ar.On("GetCount", ctx, asset.Config{}).Return(0, errors.New("unknown error"))
			},
		},
		{
			Description:  `should parse querystring to get config`,
			Querystring:  "?types=table&services=bigquery&size=30&offset=50&sort=created_at&direction=desc&filter[data.dataset]=booking&filter[data.project]=p-godata-id&q=internal&q_fields=name,urn",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{
					Types:         []asset.Type{"table"},
					Services:      []string{"bigquery"},
					Size:          30,
					Offset:        50,
					SortDirection: "desc",
					SortBy:        "created_at",
					Data: map[string]string{
						"dataset": "booking",
						"project": "p-godata-id",
					},
					Query:       "internal",
					QueryFields: []string{"name", "urn"},
				}).Return([]asset.Asset{}, nil, nil)
			},
		},
		{
			Description:  `should parse data and query fields querystring to get config`,
			Querystring:  "?filter[data.dataset]=booking&filter[data.project]=p-godata-id&q=internal&q_fields=name,urn,description,services",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{
					Data: map[string]string{
						"dataset": "booking",
						"project": "p-godata-id",
					},
					Query:       "internal",
					QueryFields: []string{"name", "urn", "description", "services"},
				}).Return([]asset.Asset{}, nil, nil)
			},
		},
		{
			Description:  `should parse data fields querystring to get config`,
			Querystring:  "?filter[data.dataset]=booking&filter[data.project]=p-godata-id",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{
					Data: map[string]string{
						"dataset": "booking",
						"project": "p-godata-id",
					},
				}).Return([]asset.Asset{}, nil, nil)
			},
		},
		{
			Description:  `should parse query fields querystring to get config`,
			Querystring:  "?q=internal&q_fields=name,urn,description,services",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{
					Query:       "internal",
					QueryFields: []string{"name", "urn", "description", "services"},
				}).Return([]asset.Asset{}, nil, nil)
			},
		},
		{
			Description:  "should convert multiple types and services from querystring to config",
			Querystring:  "?types=table,job&services=bigquery,kafka",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{
					Types:    []asset.Type{"table", "job"},
					Services: []string{"bigquery", "kafka"},
				}).Return([]asset.Asset{}, nil, nil)
			},
		},
		{
			Description:  "should return http 200 status along with list of assets",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{}).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
				}, nil, nil)
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
			Querystring:  "?with_total=true&types=job&services=kafka&size=10&offset=5",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Config{
					Types:    []asset.Type{"job"},
					Services: []string{"kafka"},
					Size:     10,
					Offset:   5,
				}).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
					{ID: "testid-3"},
				}, nil, nil)
				ar.On("GetCount", ctx, asset.Config{
					Size:     10,
					Offset:   5,
					Types:    []asset.Type{"job"},
					Services: []string{"kafka"},
				}).Return(150, nil, nil)
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

			ar := new(mocks.AssetRepository)
			tc.Setup(rr.Context(), ar)

			handler := handlers.NewAssetHandler(logger, ar, nil, nil, nil)
			handler.GetAll(rw, rr)

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

func TestAssetHandlerGetStargazers(t *testing.T) {
	type testCase struct {
		Description   string
		ExpectStatus  int
		Setup         func(tc *testCase, sr *mocks.StarRepository)
		MutateRequest func(req *http.Request) *http.Request
	}

	offset := 10
	size := 20
	defaultStarCfg := star.Config{Offset: offset, Size: size}
	var assetID = uuid.NewString()

	var testCases = []testCase{
		{
			Description:  "should return 500 status code if failed to fetch star repository",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s/stargazers", assetID)
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				return req
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("GetStargazers", mock.Anything, defaultStarCfg, assetID).Return(nil, errors.New("some error"))
			},
		},
		{
			Description:  "should return 404 status code if star repository return not found error",
			ExpectStatus: http.StatusNotFound,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s/stargazers", assetID)
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				return req
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("GetStargazers", mock.Anything, defaultStarCfg, assetID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return 200 ok if star repository return nil error",
			ExpectStatus: http.StatusOK,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s/stargazers", assetID)
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				return req
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("GetStargazers", mock.Anything, defaultStarCfg, assetID).Return([]user.User{{ID: "1"}, {ID: "2"}, {ID: "3"}}, nil, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			logger := log.NewNoop()
			defer sr.AssertExpectations(t)
			tc.Setup(&tc, sr)

			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"id": assetID,
			})
			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			handler := handlers.NewAssetHandler(logger, nil, nil, sr, nil)
			handler.GetStargazers(rw, rr)

		})
	}
}

func TestAssetHandlerGetVersionHistory(t *testing.T) {
	var assetID = uuid.NewString()

	type testCase struct {
		Description  string
		Querystring  string
		ExpectStatus int
		Setup        func(context.Context, *mocks.AssetRepository)
		PostCheck    func(resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  `should return http 400 if asset id is not uuid`,
			ExpectStatus: http.StatusBadRequest,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetVersionHistory", ctx, asset.Config{}, assetID).Return([]asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetVersionHistory", ctx, asset.Config{}, assetID).Return([]asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should parse querystring to get config`,
			Querystring:  "?size=30&offset=50",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetVersionHistory", ctx, asset.Config{
					Size:   30,
					Offset: 50,
				}, assetID).Return([]asset.Asset{}, nil, nil)
			},
		},
		{
			Description:  "should return http 200 status along with list of asset versions",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetVersionHistory", ctx, asset.Config{}, assetID).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
				}, nil, nil)
			},
			PostCheck: func(r *http.Response) error {
				expected := []asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
				}
				var actual []asset.Asset
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
			rr = mux.SetURLVars(rr, map[string]string{
				"id": assetID,
			})
			rw := httptest.NewRecorder()
			ar := new(mocks.AssetRepository)
			tc.Setup(rr.Context(), ar)

			handler := handlers.NewAssetHandler(logger, ar, nil, nil, nil)
			handler.GetVersionHistory(rw, rr)

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

func TestAssetHandlerGetByVersion(t *testing.T) {
	var (
		assetID = uuid.NewString()
		version = "0.2"
		ast     = asset.Asset{
			ID:      assetID,
			Version: version,
		}
	)

	type testCase struct {
		Description  string
		ExpectStatus int
		Setup        func(context.Context, *mocks.AssetRepository)
		PostCheck    func(resp *http.Response) error
	}

	var testCases = []testCase{
		{
			Description:  `should return http 400 if asset id is not uuid`,
			ExpectStatus: http.StatusBadRequest,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByVersion", ctx, assetID, version).Return(asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return http 404 if asset doesn't exist`,
			ExpectStatus: http.StatusNotFound,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByVersion", ctx, assetID, version).Return(asset.Asset{}, asset.NotFoundError{AssetID: assetID})
			},
		},
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByVersion", ctx, assetID, version).Return(asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  "should return http 200 status along with the asset, if found",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByVersion", ctx, assetID, version).Return(ast, nil, nil)
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
				"id":      assetID,
				"version": version,
			})
			ar := new(mocks.AssetRepository)
			tc.Setup(rr.Context(), ar)

			handler := handlers.NewAssetHandler(logger, ar, nil, nil, nil)
			handler.GetByVersion(rw, rr)

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
