package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/star"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetStarredAssetsWithHeader(t *testing.T) {
	type testCase struct {
		Description   string
		ExpectStatus  int
		Setup         func(tc *testCase, sr *mocks.StarRepository)
		MutateRequest func(req *http.Request) *http.Request
		PostCheck     func(t *testing.T, tc *testCase, resp *http.Response) error
	}

	userID := "dummy-user-id"
	offset := 10
	size := 20

	var testCases = []testCase{
		{
			Description:  "should return 400 status code if user id not found in context",
			ExpectStatus: http.StatusBadRequest,
			Setup:        func(tc *testCase, sr *mocks.StarRepository) {},
		},
		{
			Description:  "should return 400 status code if star repository return invalid error",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, userID).Return(nil, star.InvalidError{})
			},
		},
		{
			Description:  "should return 500 status code if failed to fetch starred",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, userID).Return(nil, errors.New("failed to fetch starred"))
			},
		},
		{
			Description:  "should return 404 status code if starred assets not found",
			ExpectStatus: http.StatusNotFound,
			MutateRequest: func(req *http.Request) *http.Request {
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, userID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return 200 starred assets of a user if no error",
			ExpectStatus: http.StatusOK,
			MutateRequest: func(req *http.Request) *http.Request {
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, userID).Return([]asset.Asset{
					{ID: "1", URN: "asset-urn-1", Type: "asset-type"},
					{ID: "2", URN: "asset-urn-2", Type: "asset-type"},
					{ID: "3", URN: "asset-urn-3", Type: "asset-type"},
				}, nil)

			},
			PostCheck: func(t *testing.T, tc *testCase, resp *http.Response) error {
				actual, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				expected, err := json.Marshal([]asset.Asset{
					{ID: "1", URN: "asset-urn-1", Type: "asset-type"},
					{ID: "2", URN: "asset-urn-2", Type: "asset-type"},
					{ID: "3", URN: "asset-urn-3", Type: "asset-type"},
				})
				require.NoError(t, err)

				assert.JSONEq(t, string(expected), string(actual))

				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			logger := log.NewNoop()
			defer sr.AssertExpectations(t)
			tc.Setup(&tc, sr)

			handler := handlers.NewUserHandler(logger, sr)
			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			handler.GetStarredAssetsWithHeader(rw, rr)
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
}

func TestGetStarredWithPath(t *testing.T) {
	type testCase struct {
		Description   string
		ExpectStatus  int
		Setup         func(tc *testCase, er *mocks.StarRepository)
		MutateRequest func(req *http.Request) *http.Request
		PostCheck     func(t *testing.T, tc *testCase, resp *http.Response) error
	}

	pathUserID := "a-path-user-id"
	offset := 10
	size := 20

	var testCases = []testCase{
		{
			Description:  "should return 500 status code if failed to fetch starred",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s/starred", pathUserID)
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				return req
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, pathUserID).Return(nil, errors.New("failed to fetch starred"))
			},
		},
		{
			Description:  "should return 400 status code if star repository return invalid error",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s/starred", pathUserID)
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				return req
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, pathUserID).Return(nil, star.InvalidError{})
			},
		},
		{
			Description:  "should return 404 status code if starred not found",
			ExpectStatus: http.StatusNotFound,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s/starred", pathUserID)
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				return req
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, pathUserID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return 200 starred assets of a user if no error",
			ExpectStatus: http.StatusOK,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s/starred", pathUserID)
				params := url.Values{}
				params.Add("offset", strconv.Itoa(offset))
				params.Add("size", strconv.Itoa(size))
				req.URL.RawQuery = params.Encode()
				return req
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAllAssetsByUserID", mock.AnythingOfType("*context.valueCtx"), star.Config{Offset: offset, Size: size}, pathUserID).Return([]asset.Asset{
					{ID: "1", URN: "asset-urn-1", Type: "asset-type"},
					{ID: "2", URN: "asset-urn-2", Type: "asset-type"},
					{ID: "3", URN: "asset-urn-3", Type: "asset-type"},
				}, nil)

			},
			PostCheck: func(t *testing.T, tc *testCase, resp *http.Response) error {
				actual, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				expected, err := json.Marshal([]asset.Asset{
					{ID: "1", URN: "asset-urn-1", Type: "asset-type"},
					{ID: "2", URN: "asset-urn-2", Type: "asset-type"},
					{ID: "3", URN: "asset-urn-3", Type: "asset-type"},
				})
				require.NoError(t, err)

				assert.JSONEq(t, string(expected), string(actual))

				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			logger := log.NewNoop()
			defer sr.AssertExpectations(t)
			tc.Setup(&tc, sr)

			handler := handlers.NewUserHandler(logger, sr)
			router := mux.NewRouter()
			router.Path("/v1beta1/{user_id}/starred").Methods("GET").HandlerFunc(handler.GetStarredAssetsWithPath)
			rr := httptest.NewRequest("GET", "/v1beta1", nil)
			rw := httptest.NewRecorder()

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			router.ServeHTTP(rw, rr)
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
}

func TestStarAsset(t *testing.T) {
	type testCase struct {
		Description   string
		ExpectStatus  int
		Setup         func(tc *testCase, sr *mocks.StarRepository)
		MutateRequest func(req *http.Request) *http.Request
	}

	userID := "dummy-user-id"
	assetID := "dummy-asset-id"

	var testCases = []testCase{
		{
			Description:  "should return 400 status code if user id not found in context",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				return req
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {},
		},
		{
			Description:  "should return 400 status code if asset id in param is invalid",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("Create", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return("", star.ErrEmptyAssetID)
			},
		},
		{
			Description:  "should return 400 status code if star repository return invalid error",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("Create", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return("", star.InvalidError{})
			},
		},
		{
			Description:  "should return 404 status code if user not found",
			ExpectStatus: http.StatusNotFound,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("Create", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return("", star.UserNotFoundError{UserID: userID})
			},
		},
		{
			Description:  "should return 500 status code if failed to star an asset",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("Create", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return("", errors.New("failed to star an asset"))
			},
		},
		{
			Description:  "should return 204 if starring success",
			ExpectStatus: http.StatusNoContent,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("Create", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return("1234", nil)
			},
		},
		{
			Description:  "should return 204 if asset is already starred",
			ExpectStatus: http.StatusNoContent,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("Create", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return("", star.DuplicateRecordError{})
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			logger := log.NewNoop()
			defer sr.AssertExpectations(t)
			tc.Setup(&tc, sr)

			handler := handlers.NewUserHandler(logger, sr)
			router := mux.NewRouter()
			router.Path("/user/starred/{asset_id}").Methods("PUT").HandlerFunc(handler.StarAsset)
			rr := httptest.NewRequest("PUT", "/user/starred", nil)
			rw := httptest.NewRecorder()

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			router.ServeHTTP(rw, rr)
			if rw.Code != tc.ExpectStatus {
				t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
				return
			}
		})
	}
}

func TestGetStarredAsset(t *testing.T) {
	type testCase struct {
		Description   string
		ExpectStatus  int
		Setup         func(tc *testCase, sr *mocks.StarRepository)
		MutateRequest func(req *http.Request) *http.Request
		PostCheck     func(t *testing.T, tc *testCase, resp *http.Response) error
	}

	assetType := "an-asset-type"
	assetURN := "dummy-asset-urn"
	userID := "dummy-user-id"
	assetID := "an-asset-id"

	var testCases = []testCase{
		{
			Description:  "should return 400 status code if user id not found in context",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				return req
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {},
		},
		{
			Description:  "should return 400 status code if asset id in param is invalid",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(nil, star.ErrEmptyAssetID)
			},
		},
		{
			Description:  "should return 400 status code if star repository return invalid error",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(nil, star.InvalidError{})
			},
		},
		{
			Description:  "should return 404 status code if a star not found",
			ExpectStatus: http.StatusNotFound,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return 500 status code if failed to fetch a starred asset",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(nil, errors.New("failed to fetch starred"))
			},
		},
		{
			Description:  "should return 200 starred assets of a user if no error",
			ExpectStatus: http.StatusOK,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(&asset.Asset{Type: asset.Type(assetType), URN: assetURN}, nil)
			},
			PostCheck: func(t *testing.T, tc *testCase, resp *http.Response) error {
				actual, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)

				expected, err := json.Marshal(asset.Asset{URN: assetURN, Type: asset.Type(assetType)})
				require.NoError(t, err)

				assert.JSONEq(t, string(expected), string(actual))

				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			logger := log.NewNoop()
			defer sr.AssertExpectations(t)
			tc.Setup(&tc, sr)

			handler := handlers.NewUserHandler(logger, sr)
			router := mux.NewRouter()
			router.Path("/user/starred/{asset_id}").Methods("GET").HandlerFunc(handler.GetStarredAsset)
			rr := httptest.NewRequest("GET", "/user/starred", nil)
			rw := httptest.NewRecorder()

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			router.ServeHTTP(rw, rr)
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
}

func TestUnstarAsset(t *testing.T) {
	type testCase struct {
		Description   string
		ExpectStatus  int
		Setup         func(tc *testCase, sr *mocks.StarRepository)
		MutateRequest func(req *http.Request) *http.Request
	}

	userID := "dummy-user-id"
	assetID := "dummy-asset-id"

	var testCases = []testCase{
		{
			Description:  "should return 400 status code if user id not found in context",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				return req
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {},
		},
		{
			Description:  "should return 400 status code if asset id is empty",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("Delete", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(star.ErrEmptyAssetID)
			},
		},
		{
			Description:  "should return 400 status code if star repository return invalid error",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("Delete", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(star.InvalidError{})
			},
		},
		{
			Description:  "should return 500 status code if failed to unstar an asset",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("Delete", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(errors.New("failed to star an asset"))
			},
		},
		{
			Description:  "should return 204 if unstarring success",
			ExpectStatus: http.StatusNoContent,
			MutateRequest: func(req *http.Request) *http.Request {
				req.URL.Path += fmt.Sprintf("/%s", assetID)
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, sr *mocks.StarRepository) {
				sr.On("Delete", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			sr := new(mocks.StarRepository)
			logger := log.NewNoop()
			defer sr.AssertExpectations(t)
			tc.Setup(&tc, sr)

			handler := handlers.NewUserHandler(logger, sr)
			router := mux.NewRouter()
			router.Path("/user/starred/{asset_id}").Methods("DELETE").HandlerFunc(handler.UnstarAsset)
			rr := httptest.NewRequest("DELETE", "/user/starred", nil)
			rw := httptest.NewRecorder()

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			router.ServeHTTP(rw, rr)
			if rw.Code != tc.ExpectStatus {
				t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
				return
			}
		})
	}
}
