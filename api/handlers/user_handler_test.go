package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discussion"
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

			handler := handlers.NewUserHandler(logger, sr, nil)
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

	pathUserID := uuid.NewString()
	offset := 10
	size := 20

	var testCases = []testCase{
		{
			Description:  "should return 500 status code if failed to fetch starred",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
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

			handler := handlers.NewUserHandler(logger, sr, nil)
			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"user_id": pathUserID,
			})

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			handler.GetStarredAssetsWithPath(rw, rr)
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
			Setup:        func(tc *testCase, sr *mocks.StarRepository) {},
		},
		{
			Description:  "should return 400 status code if asset id in param is invalid",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
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

			handler := handlers.NewUserHandler(logger, sr, nil)
			rr := httptest.NewRequest("PUT", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"asset_id": assetID,
			})

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			handler.StarAsset(rw, rr)
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
			Setup:        func(tc *testCase, sr *mocks.StarRepository) {},
		},
		{
			Description:  "should return 400 status code if asset id in param is invalid",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(asset.Asset{}, star.ErrEmptyAssetID)
			},
		},
		{
			Description:  "should return 400 status code if star repository return invalid error",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(asset.Asset{}, star.InvalidError{})
			},
		},
		{
			Description:  "should return 404 status code if a star not found",
			ExpectStatus: http.StatusNotFound,
			MutateRequest: func(req *http.Request) *http.Request {
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(asset.Asset{}, star.NotFoundError{})
			},
		},
		{
			Description:  "should return 500 status code if failed to fetch a starred asset",
			ExpectStatus: http.StatusInternalServerError,
			MutateRequest: func(req *http.Request) *http.Request {
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(asset.Asset{}, errors.New("failed to fetch starred"))
			},
		},
		{
			Description:  "should return 200 starred assets of a user if no error",
			ExpectStatus: http.StatusOK,
			MutateRequest: func(req *http.Request) *http.Request {
				ctx := user.NewContext(req.Context(), userID)
				return req.WithContext(ctx)
			},
			Setup: func(tc *testCase, er *mocks.StarRepository) {
				er.On("GetAssetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, assetID).Return(asset.Asset{Type: asset.Type(assetType), URN: assetURN}, nil)
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

			handler := handlers.NewUserHandler(logger, sr, nil)
			rr := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"asset_id": assetID,
			})

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			handler.GetStarredAsset(rw, rr)
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
			Setup:        func(tc *testCase, sr *mocks.StarRepository) {},
		},
		{
			Description:  "should return 400 status code if asset id is empty",
			ExpectStatus: http.StatusBadRequest,
			MutateRequest: func(req *http.Request) *http.Request {
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

			handler := handlers.NewUserHandler(logger, sr, nil)
			rr := httptest.NewRequest("DELETE", "/", nil)
			rw := httptest.NewRecorder()
			rr = mux.SetURLVars(rr, map[string]string{
				"asset_id": assetID,
			})

			if tc.MutateRequest != nil {
				rr = tc.MutateRequest(rr)
			}

			handler.UnstarAsset(rw, rr)
			if rw.Code != tc.ExpectStatus {
				t.Errorf("expected handler to return %d status, was %d instead", tc.ExpectStatus, rw.Code)
				return
			}
		})
	}
}

func TestUserGetDiscussions(t *testing.T) {
	var userID = uuid.NewString()
	type testCase struct {
		Description  string
		Querystring  string
		ExpectStatus int
		Setup        func(context.Context, *mocks.DiscussionRepository)
		PostCheck    func(resp *http.Response) error
	}
	var testCases = []testCase{
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{
					Type:                  "all",
					State:                 discussion.StateOpen.String(),
					Assignees:             []string{userID},
					SortBy:                "created_at",
					SortDirection:         "desc",
					DisjointAssigneeOwner: true,
				}).Return([]discussion.Discussion{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should parse querystring to get filter`,
			Querystring:  "?labels=label1,label2,label4&asset=e5d81dcd-3046-4d33-b1ac-efdd221e621d&type=issues&state=closed&sort=updated_at&direction=asc&size=30&offset=50",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{
					Type:                  "issues",
					State:                 "closed",
					Assignees:             []string{userID},
					Assets:                []string{"e5d81dcd-3046-4d33-b1ac-efdd221e621d"},
					Labels:                []string{"label1", "label2", "label4"},
					SortBy:                "updated_at",
					SortDirection:         "asc",
					Size:                  30,
					Offset:                50,
					DisjointAssigneeOwner: true,
				}).Return([]discussion.Discussion{}, nil)
			},
		},
		{
			Description:  `should set filter to default if empty`,
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{
					Type:                  "all",
					State:                 "open",
					Assignees:             []string{userID},
					SortBy:                "created_at",
					SortDirection:         "desc",
					Size:                  0,
					Offset:                0,
					DisjointAssigneeOwner: true,
				}).Return([]discussion.Discussion{}, nil)
			},
		},
		{
			Description:  "should return http 200 status along with list of discussions",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{
					Type:                  "all",
					State:                 discussion.StateOpen.String(),
					Assignees:             []string{userID},
					SortBy:                "created_at",
					SortDirection:         "desc",
					DisjointAssigneeOwner: true,
				}).Return([]discussion.Discussion{
					{ID: "1122"},
					{ID: "2233"},
				}, nil)
			},
			PostCheck: func(r *http.Response) error {
				expected := []discussion.Discussion{
					{ID: "1122"},
					{ID: "2233"},
				}

				var actual []discussion.Discussion
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
			ctx := user.NewContext(rr.Context(), userID)
			rr = rr.WithContext(ctx)
			rw := httptest.NewRecorder()

			dr := new(mocks.DiscussionRepository)
			tc.Setup(rr.Context(), dr)

			handler := handlers.NewUserHandler(logger, nil, dr)
			handler.GetDiscussions(rw, rr)

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
