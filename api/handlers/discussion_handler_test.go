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
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDiscussionHandlerGetAll(t *testing.T) {
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
				dr.On("GetAll", ctx, discussion.Filter{}).Return([]discussion.Discussion{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should parse querystring to get filter`,
			Querystring:  "?labels=label1,label2,label4&assignee=646130cf-3dde-4d61-99e9-6070dd369597&asset=e5d81dcd-3046-4d33-b1ac-efdd221e621d&owner=62326386-dc9d-4ae5-9448-e54c720f856d&type=issues&state=closed&sort=updated_at&direction=asc&size=30&offset=50",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.DiscussionRepository) {
				ar.On("GetAll", ctx, discussion.Filter{
					Type:          "issues",
					State:         "closed",
					Assignees:     []string{"646130cf-3dde-4d61-99e9-6070dd369597"},
					Assets:        []string{"e5d81dcd-3046-4d33-b1ac-efdd221e621d"},
					Owner:         "62326386-dc9d-4ae5-9448-e54c720f856d",
					Labels:        []string{"label1", "label2", "label4"},
					SortBy:        "updated_at",
					SortDirection: "asc",
					Size:          30,
					Offset:        50,
				}).Return([]discussion.Discussion{}, nil)
			},
		},
		{
			Description:  "should return http 200 status along with list of discussions",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, ar *mocks.DiscussionRepository) {
				ar.On("GetAll", ctx, discussion.Filter{}).Return([]discussion.Discussion{
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

			handler := handlers.NewDiscussionHandler(logger, dr)
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

func TestDiscussionHandlerCreate(t *testing.T) {
	var userID = uuid.NewString()
	var validPayload = `{"title": "Lorem Ipsum", "body": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", "type": "qanda"}`
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
				description: "empty title",
				payload:     `{"title": "", "body": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", "type": "qanda"}`,
			},
			{
				description: "empty body",
				payload:     `{"title": "Lorem Ipsum", "body": "", "type": "qanda"}`,
			},
			{
				description: "empty type",
				payload:     `{"title": "Lorem Ipsum", "body": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", "type": ""}`,
			},
			{
				description: "wrong type",
				payload:     `{"title": "Lorem Ipsum", "body": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", "type": "wrongtype"}`,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.description, func(t *testing.T) {
				rw := httptest.NewRecorder()

				rr := httptest.NewRequest("POST", "/", strings.NewReader(testCase.payload))
				ctx := user.NewContext(rr.Context(), userID)
				rr = rr.WithContext(ctx)

				dr := new(mocks.DiscussionRepository)

				handler := handlers.NewDiscussionHandler(logger, dr)
				handler.Create(rw, rr)

				expectedStatus := http.StatusBadRequest
				if rw.Code != expectedStatus {
					t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
					return
				}
			})
		}
	})

	t.Run("should return HTTP 500 if the discussion creation fails", func(t *testing.T) {
		rr := httptest.NewRequest("POST", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rw := httptest.NewRecorder()

		expectedErr := errors.New("unknown error")

		dr := new(mocks.DiscussionRepository)
		dr.On("Create", rr.Context(), mock.AnythingOfType("*discussion.Discussion")).Return("1234-5678", expectedErr)
		defer dr.AssertExpectations(t)

		rr.Context()
		handler := handlers.NewDiscussionHandler(logger, dr)
		handler.Create(rw, rr)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		var response handlers.ErrorResponse
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response.Reason, "Internal Server Error")
	})

	t.Run("should return HTTP 201 and discussion ID if the discussion is successfully created", func(t *testing.T) {
		dsc := discussion.Discussion{
			Title: "Lorem Ipsum",
			Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			Type:  "qanda",
			Owner: user.User{ID: userID},
		}
		discussionWithID := dsc
		discussionWithID.ID = "12"

		rr := httptest.NewRequest("POST", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rw := httptest.NewRecorder()

		dr := new(mocks.DiscussionRepository)
		dr.On("Create", rr.Context(), &dsc).Return(discussionWithID.ID, nil).Run(func(args mock.Arguments) {
			argDiscussion := args.Get(1).(*discussion.Discussion)
			argDiscussion.ID = discussionWithID.ID
		})
		defer dr.AssertExpectations(t)

		handler := handlers.NewDiscussionHandler(logger, dr)
		handler.Create(rw, rr)

		assert.Equal(t, http.StatusCreated, rw.Code)
		var response map[string]interface{}
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)

		discussionID, exists := response["id"]
		assert.True(t, exists)
		assert.Equal(t, discussionWithID.ID, discussionID)
	})
}

func TestDiscussionHandlerGet(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
	)
	type testCase struct {
		Description  string
		Querystring  string
		ExpectStatus int
		DiscussionID string
		Setup        func(context.Context, *mocks.DiscussionRepository)
		PostCheck    func(resp *http.Response) error
	}
	var testCases = []testCase{
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			DiscussionID: discussionID,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.On("Get", ctx, discussionID).Return(discussion.Discussion{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should return http 400 if discussion id not integer`,
			ExpectStatus: http.StatusBadRequest,
			DiscussionID: "random",
		},
		{
			Description:  `should return http 400 if discussion id < 0`,
			ExpectStatus: http.StatusBadRequest,
			DiscussionID: "-1",
		},
		{
			Description:  `should return http 404 if discussion not found`,
			ExpectStatus: http.StatusNotFound,
			DiscussionID: discussionID,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.On("Get", ctx, discussionID).Return(discussion.Discussion{}, discussion.NotFoundError{DiscussionID: discussionID})
			},
		},
		{
			Description:  "should return http 200 status along with discussions",
			ExpectStatus: http.StatusOK,
			DiscussionID: discussionID,
			Setup: func(ctx context.Context, ar *mocks.DiscussionRepository) {
				ar.On("Get", ctx, discussionID).Return(discussion.Discussion{ID: discussionID}, nil)
			},
			PostCheck: func(r *http.Response) error {
				expected := discussion.Discussion{
					ID: discussionID,
				}

				var actual discussion.Discussion
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
			rr = mux.SetURLVars(rr, map[string]string{
				"id": tc.DiscussionID,
			})

			rw := httptest.NewRecorder()

			dr := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(rr.Context(), dr)
			}
			handler := handlers.NewDiscussionHandler(logger, dr)
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

func TestDiscussionHandlerPatch(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
	)
	var validPayload = `{"title": "Lorem Ipsum"}`
	t.Run("should check payload", func(t *testing.T) {
		testCases := []struct {
			Description  string
			Payload      string
			StatusCode   int
			DiscussionID string
		}{
			{
				Description:  "discussion id is not integer return bad request",
				DiscussionID: "random",
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "discussion id is < 0 return bad request",
				DiscussionID: "-1",
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "empty object return no content",
				Payload:      `{}`,
				DiscussionID: discussionID,
				StatusCode:   http.StatusNoContent,
			},
			{
				Description: "wrong payload return bad request",
				Payload:     `{,..`,
				StatusCode:  http.StatusBadRequest,
			},
			{
				Description: "invalid type return bad request",
				Payload:     `{"type": "random"}`,
				StatusCode:  http.StatusBadRequest,
			},
			{
				Description: "invalid state return bad request",
				Payload:     `{"state": "random"}`,
				StatusCode:  http.StatusBadRequest,
			},
			{
				Description: "assignees more than limit should return bad request",
				Payload:     `{"assignees": ["1","2","3","4","5","6","7","8","9","10","11"]}`,
				StatusCode:  http.StatusBadRequest,
			},
			{
				Description: "assets more than limit should return bad request",
				Payload:     `{"assets": ["1","2","3","4","5","6","7","8","9","10","11"]}`,
				StatusCode:  http.StatusBadRequest,
			},
			{
				Description: "labels more than limit should return bad request",
				Payload:     `{"labels": ["1","2","3","4","5","6","7","8","9","10","11"]}`,
				StatusCode:  http.StatusBadRequest,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Description, func(t *testing.T) {
				rw := httptest.NewRecorder()

				rr := httptest.NewRequest("PATCH", "/", strings.NewReader(testCase.Payload))
				ctx := user.NewContext(rr.Context(), userID)
				rr = rr.WithContext(ctx)
				rr = mux.SetURLVars(rr, map[string]string{
					"id": testCase.DiscussionID,
				})

				dr := new(mocks.DiscussionRepository)

				handler := handlers.NewDiscussionHandler(logger, dr)
				handler.Patch(rw, rr)

				assert.Equal(t, testCase.StatusCode, rw.Code)
			})
		}
	})

	t.Run("should return HTTP 500 if the discussion patch fails", func(t *testing.T) {
		rr := httptest.NewRequest("PATCH", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"id": discussionID,
		})

		rw := httptest.NewRecorder()

		expectedErr := errors.New("unknown error")

		dr := new(mocks.DiscussionRepository)
		dr.On("Patch", rr.Context(), mock.AnythingOfType("*discussion.Discussion")).Return(expectedErr)
		defer dr.AssertExpectations(t)

		rr.Context()
		handler := handlers.NewDiscussionHandler(logger, dr)
		handler.Patch(rw, rr)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		var response handlers.ErrorResponse
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response.Reason, "Internal Server Error")
	})

	t.Run("should return HTTP 204 if the discussion is successfully patched", func(t *testing.T) {
		dsc := &discussion.Discussion{
			ID:    discussionID,
			Title: "Lorem Ipsum",
		}
		rr := httptest.NewRequest("PATCH", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"id": discussionID,
		})

		rw := httptest.NewRecorder()

		dr := new(mocks.DiscussionRepository)
		dr.On("Patch", rr.Context(), dsc).Return(nil).Run(func(args mock.Arguments) {
			argDiscussion := args.Get(1).(*discussion.Discussion)
			argDiscussion.ID = dsc.ID
		})
		defer dr.AssertExpectations(t)

		handler := handlers.NewDiscussionHandler(logger, dr)
		handler.Patch(rw, rr)

		assert.Equal(t, http.StatusNoContent, rw.Code)
	})
}
