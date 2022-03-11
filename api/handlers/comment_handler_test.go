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
	"github.com/odpf/columbus/comment"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommentHandlerCreate(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "11111"
	)
	var validPayload = `{"body": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."}`
	t.Run("should return HTTP 400 for invalid empty payload or wrong discussion_id", func(t *testing.T) {
		testCases := []struct {
			description  string
			discussionID string
			payload      string
		}{
			{
				description:  "empty object",
				payload:      `{}`,
				discussionID: discussionID,
			},
			{
				description:  "if discussion_id is not integer",
				payload:      validPayload,
				discussionID: "test",
			},
			{
				description:  "if discussion_id is < 1",
				payload:      validPayload,
				discussionID: "0",
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.description, func(t *testing.T) {
				rw := httptest.NewRecorder()

				rr := httptest.NewRequest("POST", "/", strings.NewReader(testCase.payload))
				ctx := user.NewContext(rr.Context(), userID)
				rr = rr.WithContext(ctx)
				rr = mux.SetURLVars(rr, map[string]string{
					"discussion_id": testCase.discussionID,
				})

				cr := new(mocks.CommentRepository)

				handler := handlers.NewCommentHandler(logger, cr)
				handler.Create(rw, rr)

				expectedStatus := http.StatusBadRequest
				if rw.Code != expectedStatus {
					t.Errorf("expected handler to return HTTP %d, returned HTTP %d instead", expectedStatus, rw.Code)
					return
				}
			})
		}
	})

	t.Run("should return HTTP 500 if the comment creation fails", func(t *testing.T) {
		rr := httptest.NewRequest("POST", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
		})
		rw := httptest.NewRecorder()

		expectedErr := errors.New("unknown error")

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Create(rr.Context(), mock.AnythingOfType("*comment.Comment")).Return("", expectedErr)
		defer cr.AssertExpectations(t)

		rr.Context()
		handler := handlers.NewCommentHandler(logger, cr)
		handler.Create(rw, rr)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		var response handlers.ErrorResponse
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response.Reason, "Internal Server Error")
	})

	t.Run("should return HTTP 201 and comment ID if the comment is successfully created", func(t *testing.T) {
		cmt := comment.Comment{
			DiscussionID: discussionID,
			Body:         "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			UpdatedBy:    user.User{ID: userID},
			Owner:        user.User{ID: userID},
		}
		commentWithID := cmt
		commentWithID.ID = "12"

		rr := httptest.NewRequest("POST", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
		})
		rw := httptest.NewRecorder()

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Create(rr.Context(), &cmt).
			Run(func(ctx context.Context, cmt *comment.Comment) {
				cmt.ID = commentWithID.ID
			}).
			Return(commentWithID.ID, nil)
		defer cr.AssertExpectations(t)

		handler := handlers.NewCommentHandler(logger, cr)
		handler.Create(rw, rr)

		assert.Equal(t, http.StatusCreated, rw.Code)
		var response map[string]interface{}
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)

		commentID, exists := response["id"]
		assert.True(t, exists)
		assert.Equal(t, commentWithID.ID, commentID)
	})
}

func TestCommentHandlerGetAll(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "11111"
	)
	type testCase struct {
		Description  string
		Querystring  string
		ExpectStatus int
		DiscussionID string
		Setup        func(context.Context, *mocks.CommentRepository)
		PostCheck    func(resp *http.Response) error
	}
	var testCases = []testCase{
		{
			Description:  `should return http 400 if discussion_id is not integer`,
			DiscussionID: "test",
			ExpectStatus: http.StatusBadRequest,
		},
		{
			Description:  `should return http 400 if discussion_id is < 1`,
			DiscussionID: "0",
			ExpectStatus: http.StatusBadRequest,
		},
		{
			Description:  `should return http 500 if fetching fails`,
			DiscussionID: discussionID,
			ExpectStatus: http.StatusInternalServerError,
			Setup: func(ctx context.Context, cr *mocks.CommentRepository) {
				cr.EXPECT().GetAll(ctx, discussionID, comment.Filter{
					SortBy:        "created_at",
					SortDirection: "desc",
				}).Return([]comment.Comment{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should parse querystring to get filter`,
			DiscussionID: discussionID,
			Querystring:  "?sort=updated_at&direction=asc&size=30&offset=50",
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, cr *mocks.CommentRepository) {
				cr.EXPECT().GetAll(ctx, discussionID, comment.Filter{
					SortBy:        "updated_at",
					SortDirection: "asc",
					Size:          30,
					Offset:        50,
				}).Return([]comment.Comment{}, nil)
			},
		},
		{
			Description:  "should return http 200 status along with list of comments",
			DiscussionID: discussionID,
			ExpectStatus: http.StatusOK,
			Setup: func(ctx context.Context, cr *mocks.CommentRepository) {
				cr.EXPECT().GetAll(ctx, discussionID, comment.Filter{
					SortBy:        "created_at",
					SortDirection: "desc",
				}).Return([]comment.Comment{
					{ID: "1122"},
					{ID: "2233"},
				}, nil)
			},
			PostCheck: func(r *http.Response) error {
				expected := []comment.Comment{
					{ID: "1122"},
					{ID: "2233"},
				}

				var actual []comment.Comment
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
				"discussion_id": tc.DiscussionID,
			})

			rw := httptest.NewRecorder()

			cr := new(mocks.CommentRepository)
			defer cr.AssertExpectations(t)

			if tc.Setup != nil {
				tc.Setup(rr.Context(), cr)
			}

			handler := handlers.NewCommentHandler(logger, cr)
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

func TestCommentHandlerGet(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
	)
	type testCase struct {
		Description  string
		Querystring  string
		ExpectStatus int
		DiscussionID string
		CommentID    string
		Setup        func(context.Context, *mocks.CommentRepository)
		PostCheck    func(resp *http.Response) error
	}
	var testCases = []testCase{
		{
			Description:  `should return http 500 if fetching fails`,
			ExpectStatus: http.StatusInternalServerError,
			CommentID:    commentID,
			DiscussionID: discussionID,
			Setup: func(ctx context.Context, cr *mocks.CommentRepository) {
				cr.EXPECT().Get(ctx, commentID, discussionID).Return(comment.Comment{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should return http 400 if discussion id not integer`,
			ExpectStatus: http.StatusBadRequest,
			CommentID:    commentID,
			DiscussionID: "random",
		},
		{
			Description:  `should return http 400 if discussion id < 0`,
			ExpectStatus: http.StatusBadRequest,
			CommentID:    commentID,
			DiscussionID: "-1",
		},
		{
			Description:  `should return http 400 if comment id not integer`,
			ExpectStatus: http.StatusBadRequest,
			CommentID:    "random",
			DiscussionID: discussionID,
		},
		{
			Description:  `should return http 400 if comment id < 0`,
			ExpectStatus: http.StatusBadRequest,
			CommentID:    "-1",
			DiscussionID: discussionID,
		},
		{
			Description:  `should return http 404 if comment or discussion not found`,
			ExpectStatus: http.StatusNotFound,
			CommentID:    commentID,
			DiscussionID: discussionID,
			Setup: func(ctx context.Context, cr *mocks.CommentRepository) {
				cr.EXPECT().Get(ctx, commentID, discussionID).Return(comment.Comment{}, comment.NotFoundError{DiscussionID: discussionID, CommentID: commentID})
			},
		},
		{
			Description:  "should return http 200 status along with comment of a discussion",
			ExpectStatus: http.StatusOK,
			CommentID:    commentID,
			DiscussionID: discussionID,
			Setup: func(ctx context.Context, cr *mocks.CommentRepository) {
				cr.EXPECT().Get(ctx, commentID, discussionID).Return(comment.Comment{ID: commentID, DiscussionID: discussionID}, nil)
			},
			PostCheck: func(r *http.Response) error {
				expected := comment.Comment{
					ID:           commentID,
					DiscussionID: discussionID,
				}

				var actual comment.Comment
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
				"discussion_id": tc.DiscussionID,
				"id":            tc.CommentID,
			})

			rw := httptest.NewRecorder()

			cr := new(mocks.CommentRepository)
			if tc.Setup != nil {
				tc.Setup(rr.Context(), cr)
			}
			handler := handlers.NewCommentHandler(logger, cr)
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

func TestCommentHandlerUpdate(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
	)
	var validPayload = `{"body": "Lorem Ipsum"}`
	t.Run("should check payload", func(t *testing.T) {
		testCases := []struct {
			Description  string
			Payload      string
			StatusCode   int
			DiscussionID string
			CommentID    string
		}{
			{
				Description:  "discussion id is not integer return bad request",
				DiscussionID: "random",
				CommentID:    commentID,
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "discussion id is < 0 return bad request",
				DiscussionID: "-1",
				CommentID:    commentID,
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "comment id is not integer return bad request",
				DiscussionID: discussionID,
				CommentID:    "random",
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "comment id is < 0 return bad request",
				DiscussionID: discussionID,
				CommentID:    "-1",
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "empty object return bad request",
				Payload:      `{}`,
				DiscussionID: discussionID,
				CommentID:    commentID,
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description: "wrong payload return bad request",
				Payload:     `{,..`,
				CommentID:   commentID,
				StatusCode:  http.StatusBadRequest,
			},
			{
				Description:  "empty body return bad request",
				Payload:      `{"body":"    "}`,
				DiscussionID: discussionID,
				CommentID:    commentID,
				StatusCode:   http.StatusBadRequest,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Description, func(t *testing.T) {
				rw := httptest.NewRecorder()

				rr := httptest.NewRequest("PUT", "/", strings.NewReader(testCase.Payload))
				ctx := user.NewContext(rr.Context(), userID)
				rr = rr.WithContext(ctx)
				rr = mux.SetURLVars(rr, map[string]string{
					"discussion_id": testCase.DiscussionID,
					"id":            testCase.CommentID,
				})

				cr := new(mocks.CommentRepository)

				handler := handlers.NewCommentHandler(logger, cr)
				handler.Update(rw, rr)

				assert.Equal(t, testCase.StatusCode, rw.Code)
			})
		}
	})

	t.Run("should return HTTP 500 if the update comment failed", func(t *testing.T) {
		cmt := &comment.Comment{
			ID:           commentID,
			DiscussionID: discussionID,
			Body:         "Lorem Ipsum",
			UpdatedBy:    user.User{ID: userID},
		}
		rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
			"id":            commentID,
		})

		rw := httptest.NewRecorder()

		expectedErr := errors.New("unknown error")

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Update(rr.Context(), cmt).Return(expectedErr)
		defer cr.AssertExpectations(t)

		handler := handlers.NewCommentHandler(logger, cr)
		handler.Update(rw, rr)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		var response handlers.ErrorResponse
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response.Reason, "Internal Server Error")
	})

	t.Run("should return HTTP 404 if the discussion id or comment id not found", func(t *testing.T) {
		cmt := &comment.Comment{
			ID:           commentID,
			DiscussionID: discussionID,
			Body:         "Lorem Ipsum",
			UpdatedBy:    user.User{ID: userID},
		}
		rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
			"id":            commentID,
		})

		rw := httptest.NewRecorder()

		expectedErr := comment.NotFoundError{DiscussionID: discussionID, CommentID: commentID}

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Update(rr.Context(), cmt).Return(expectedErr)
		defer cr.AssertExpectations(t)

		handler := handlers.NewCommentHandler(logger, cr)
		handler.Update(rw, rr)

		assert.Equal(t, http.StatusNotFound, rw.Code)
	})

	t.Run("should return HTTP 204 if the comment is successfully updated", func(t *testing.T) {
		cmt := &comment.Comment{
			ID:           commentID,
			DiscussionID: discussionID,
			Body:         "Lorem Ipsum",
			UpdatedBy:    user.User{ID: userID},
		}
		rr := httptest.NewRequest("PUT", "/", strings.NewReader(validPayload))
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
			"id":            commentID,
		})

		rw := httptest.NewRecorder()

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Update(rr.Context(), cmt).Run(func(ctx context.Context, cmtArg *comment.Comment) {
			cmtArg.ID = cmt.ID
		}).Return(nil)
		defer cr.AssertExpectations(t)

		handler := handlers.NewCommentHandler(logger, cr)
		handler.Update(rw, rr)

		assert.Equal(t, http.StatusNoContent, rw.Code)
	})
}

func TestCommentHandlerDelete(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
	)

	t.Run("should check path params", func(t *testing.T) {
		testCases := []struct {
			Description  string
			Payload      string
			StatusCode   int
			DiscussionID string
			CommentID    string
		}{
			{
				Description:  "discussion id is not integer return bad request",
				DiscussionID: "random",
				CommentID:    commentID,
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "discussion id is < 0 return bad request",
				DiscussionID: "-1",
				CommentID:    commentID,
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "comment id is not integer return bad request",
				DiscussionID: discussionID,
				CommentID:    "random",
				StatusCode:   http.StatusBadRequest,
			},
			{
				Description:  "comment id is < 0 return bad request",
				DiscussionID: discussionID,
				CommentID:    "-1",
				StatusCode:   http.StatusBadRequest,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Description, func(t *testing.T) {
				rw := httptest.NewRecorder()

				rr := httptest.NewRequest("DELETE", "/", nil)
				ctx := user.NewContext(rr.Context(), userID)
				rr = rr.WithContext(ctx)
				rr = mux.SetURLVars(rr, map[string]string{
					"discussion_id": testCase.DiscussionID,
					"id":            testCase.CommentID,
				})

				cr := new(mocks.CommentRepository)

				handler := handlers.NewCommentHandler(logger, cr)
				handler.Delete(rw, rr)

				assert.Equal(t, testCase.StatusCode, rw.Code)
			})
		}
	})

	t.Run("should return HTTP 500 if the delete comment failed", func(t *testing.T) {
		rr := httptest.NewRequest("DELETE", "/", nil)
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
			"id":            commentID,
		})

		rw := httptest.NewRecorder()

		expectedErr := errors.New("unknown error")

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Delete(rr.Context(), commentID, discussionID).Return(expectedErr)
		defer cr.AssertExpectations(t)

		handler := handlers.NewCommentHandler(logger, cr)
		handler.Delete(rw, rr)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		var response handlers.ErrorResponse
		err := json.NewDecoder(rw.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response.Reason, "Internal Server Error")
	})

	t.Run("should return HTTP 404 if the discussion id or comment id not found", func(t *testing.T) {
		rr := httptest.NewRequest("DELETE", "/", nil)
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
			"id":            commentID,
		})

		rw := httptest.NewRecorder()

		expectedErr := comment.NotFoundError{DiscussionID: discussionID, CommentID: commentID}

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Delete(rr.Context(), commentID, discussionID).Return(expectedErr)
		defer cr.AssertExpectations(t)

		handler := handlers.NewCommentHandler(logger, cr)
		handler.Delete(rw, rr)

		assert.Equal(t, http.StatusNotFound, rw.Code)
	})

	t.Run("should return HTTP 204 if the comment is successfully deleted", func(t *testing.T) {
		rr := httptest.NewRequest("DELETE", "/", nil)
		ctx := user.NewContext(rr.Context(), userID)
		rr = rr.WithContext(ctx)
		rr = mux.SetURLVars(rr, map[string]string{
			"discussion_id": discussionID,
			"id":            commentID,
		})

		rw := httptest.NewRecorder()

		cr := new(mocks.CommentRepository)
		cr.EXPECT().Delete(rr.Context(), commentID, discussionID).Return(nil)
		defer cr.AssertExpectations(t)

		handler := handlers.NewCommentHandler(logger, cr)
		handler.Delete(rw, rr)

		assert.Equal(t, http.StatusNoContent, rw.Code)
	})
}
