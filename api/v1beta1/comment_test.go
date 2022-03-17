package v1beta1_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/odpf/columbus/api"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateComment(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "11111"
		validRequest = &compassv1beta1.CreateCommentRequest{
			DiscussionId: discussionID,
			Body:         "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		}
	)

	type TestCase struct {
		Description string
		Request     *compassv1beta1.CreateCommentRequest
		StatusCode  codes.Code
		Result      string
		Setup       func(context.Context, *mocks.DiscussionRepository)
	}

	var testCases = []TestCase{
		{
			Description: "should return invalid request if empty request",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: discussionID,
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "should return invalid request if discussion_id is not integer",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: "test",
				Body:         validRequest.Body,
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "should return invalid request if discussion_id is < 1",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: "0",
				Body:         validRequest.Body,
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "should return internal server error if the comment creation failed",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: validRequest.GetDiscussionId(),
				Body:         validRequest.GetBody(),
			},
			StatusCode: codes.Internal,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				expectedErr := errors.New("unknown error")
				dr.EXPECT().CreateComment(ctx, mock.AnythingOfType("*discussion.Comment")).Return("", expectedErr)
			},
		},
		{
			Description: "should return OK and comment ID if the comment is successfully created",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: validRequest.GetDiscussionId(),
				Body:         validRequest.GetBody(),
			},
			StatusCode: codes.OK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().CreateComment(ctx, mock.AnythingOfType("*discussion.Comment")).Return("", nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockRepo)
			}
			defer mockRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				DiscussionRepository: mockRepo,
			})

			got, err := handler.CreateComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.StatusCode {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.StatusCode.String(), code.String())
				return
			}
			if got.GetId() != tc.Result {
				t.Errorf("expected result to return id %s, returned id %s instead", tc.Result, got.Id)
				return
			}
		})
	}
}

func TestGetAllComments(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "11111"
	)
	type testCase struct {
		Description string
		Request     *compassv1beta1.GetAllCommentsRequest
		StatusCode  codes.Code
		Setup       func(context.Context, *mocks.DiscussionRepository)
		PostCheck   func(resp *compassv1beta1.GetAllCommentsResponse) error
	}
	var testCases = []testCase{
		{
			Description: `should return invalid argument if discussion_id is not integer`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: "test",
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if discussion_id is < 1`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: "0",
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: `should return internal server error if fetching fails`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: discussionID,
			},
			StatusCode: codes.Internal,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAllComments(ctx, discussionID, discussion.Filter{
					Type:          "all",
					State:         "open",
					SortBy:        "created_at",
					SortDirection: "desc",
				}).Return([]discussion.Comment{}, errors.New("unknown error"))
			},
		},
		{
			Description: `should successfully parse querystring to get filter`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: discussionID,
				Sort:         "updated_at",
				Direction:    "asc",
				Size:         30,
				Offset:       50,
			},
			StatusCode: codes.OK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAllComments(ctx, discussionID, discussion.Filter{
					Type:          "all",
					State:         discussion.StateOpen.String(),
					SortBy:        "updated_at",
					SortDirection: "asc",
					Size:          30,
					Offset:        50,
				}).Return([]discussion.Comment{}, nil)
			},
		},
		{
			Description: "should return status OK along with list of comments",
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: discussionID,
			},
			StatusCode: codes.OK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAllComments(ctx, discussionID, discussion.Filter{
					Type:          "all",
					State:         discussion.StateOpen.String(),
					SortBy:        "created_at",
					SortDirection: "desc",
				}).Return([]discussion.Comment{
					{ID: "1122"},
					{ID: "2233"},
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllCommentsResponse) error {
				expected := []discussion.Comment{
					{ID: "1122"},
					{ID: "2233"},
				}

				var actual []discussion.Comment
				for _, cmt := range resp.GetData() {
					actual = append(actual, discussion.NewCommentFromProto(cmt))
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
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockRepo)
			}
			defer mockRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				DiscussionRepository: mockRepo,
			})

			got, err := handler.GetAllComments(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.StatusCode {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.StatusCode.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestGetComment(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
	)
	type testCase struct {
		Description string
		Request     *compassv1beta1.GetCommentRequest
		StatusCode  codes.Code
		Setup       func(context.Context, *mocks.DiscussionRepository)
		PostCheck   func(resp *compassv1beta1.GetCommentResponse) error
	}
	var testCases = []testCase{
		{
			Description: `should return internal server error if fetching fails`,
			StatusCode:  codes.Internal,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{}, errors.New("unknown error"))
			},
		},
		{
			Description: `should return invalid argument if discussion id not integer`,
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
		},
		{
			Description: `should return invalid argument if discussion id < 0`,
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
		},
		{
			Description: `should return invalid argument if comment id not integer`,
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
		},
		{
			Description: `should return invalid argument if comment id < 0`,
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
		},
		{
			Description: `should return Not Found if comment or discussion not found`,
			StatusCode:  codes.NotFound,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{}, discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID})
			},
		},
		{
			Description: "should return status OK along with comment of a discussion",
			StatusCode:  codes.OK,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{ID: commentID, DiscussionID: discussionID}, nil)
			},
			PostCheck: func(r *compassv1beta1.GetCommentResponse) error {
				expected := discussion.Comment{
					ID:           commentID,
					DiscussionID: discussionID,
				}

				actual := discussion.NewCommentFromProto(r.GetData())
				if reflect.DeepEqual(actual, expected) == false {
					return fmt.Errorf("expected payload to be to be %+v, was %+v", expected, actual)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockRepo)
			}
			defer mockRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				DiscussionRepository: mockRepo,
			})

			got, err := handler.GetComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.StatusCode {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.StatusCode.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
	)
	var validRequest = &compassv1beta1.UpdateCommentRequest{
		Id:           commentID,
		DiscussionId: discussionID,
		Body:         "lorem ipsum",
	}
	testCases := []struct {
		Description string
		Request     *compassv1beta1.UpdateCommentRequest
		StatusCode  codes.Code
		Setup       func(context.Context, *mocks.DiscussionRepository)
	}{
		{
			Description: "discussion id is not integer return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "discussion id is < 0 return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "comment id is not integer return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "comment id is < 0 return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "empty object return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "empty body return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
				Body:         "",
			},
			StatusCode: codes.InvalidArgument,
		},
		{
			Description: "should return internal server error if the update comment failed",
			Request:     validRequest,
			StatusCode:  codes.Internal,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				cmt := &discussion.Comment{
					ID:           validRequest.Id,
					DiscussionID: validRequest.DiscussionId,
					Body:         validRequest.Body,
					UpdatedBy:    user.User{ID: userID},
				}
				expectedErr := errors.New("unknown error")
				dr.EXPECT().UpdateComment(ctx, cmt).Return(expectedErr)
			},
		},
		{
			Description: "should return Not Found if the discussion id or comment id not found",
			Request:     validRequest,
			StatusCode:  codes.NotFound,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				cmt := &discussion.Comment{
					ID:           validRequest.Id,
					DiscussionID: validRequest.DiscussionId,
					Body:         validRequest.Body,
					UpdatedBy:    user.User{ID: userID},
				}
				expectedErr := discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID}
				dr.EXPECT().UpdateComment(ctx, cmt).Return(expectedErr)
			},
		},
		{
			Description: "should return status OK if the comment is successfully updated",
			Request:     validRequest,
			StatusCode:  codes.OK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				cmt := &discussion.Comment{
					ID:           validRequest.Id,
					DiscussionID: validRequest.DiscussionId,
					Body:         validRequest.Body,
					UpdatedBy:    user.User{ID: userID},
				}
				dr.EXPECT().UpdateComment(ctx, cmt).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockRepo)
			}
			defer mockRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				DiscussionRepository: mockRepo,
			})

			_, err := handler.UpdateComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.StatusCode {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.StatusCode, code.String())
				return
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
	)

	testCases := []struct {
		Description string
		Request     *compassv1beta1.DeleteCommentRequest
		StatusCode  codes.Code
		Setup       func(context.Context, *mocks.DiscussionRepository)
	}{
		{
			Description: "discussion id is not integer return bad request",
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
		},
		{
			Description: "discussion id is < 0 return bad request",
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
		},
		{
			Description: "comment id is not integer return bad request",
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
		},
		{
			Description: "comment id is < 0 return bad request",
			StatusCode:  codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
		},
		{
			Description: "should return internal server error if the delete comment failed",
			StatusCode:  codes.Internal,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				expectedErr := errors.New("unknown error")
				dr.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(expectedErr)
			},
		},
		{
			Description: "should return invalid argument if the discussion id or comment id not found",
			StatusCode:  codes.NotFound,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				expectedErr := discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID}
				dr.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(expectedErr)
			},
		},
		{
			Description: "should return OK if the comment is successfully deleted",
			StatusCode:  codes.OK,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockRepo := new(mocks.DiscussionRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockRepo)
			}
			defer mockRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				DiscussionRepository: mockRepo,
			})

			_, err := handler.DeleteComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.StatusCode {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.StatusCode.String(), code.String())
				return
			}
		})
	}
}
