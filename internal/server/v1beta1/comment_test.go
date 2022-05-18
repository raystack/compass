package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/discussion"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
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
		Description  string
		Request      *compassv1beta1.CreateCommentRequest
		ExpectStatus codes.Code
		Result       string
		Setup        func(context.Context, *mocks.DiscussionService)
	}

	var testCases = []TestCase{
		{
			Description: "should return invalid request if empty request",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: discussionID,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid request if discussion_id is not integer",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: "test",
				Body:         validRequest.Body,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid request if discussion_id is < 1",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: "0",
				Body:         validRequest.Body,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return internal server error if the comment creation failed",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: validRequest.GetDiscussionId(),
				Body:         validRequest.GetBody(),
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				expectedErr := errors.New("unknown error")
				ds.EXPECT().CreateComment(ctx, mock.AnythingOfType("*discussion.Comment")).Return("", expectedErr)
			},
		},
		{
			Description: "should return OK and comment ID if the comment is successfully created",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: validRequest.GetDiscussionId(),
				Body:         validRequest.GetBody(),
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().CreateComment(ctx, mock.AnythingOfType("*discussion.Comment")).Return("", nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockSvc.AssertExpectations(t)

			handler := NewAPIServer(logger, nil, nil, mockSvc, nil, nil, nil)

			got, err := handler.CreateComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
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
		Description  string
		Request      *compassv1beta1.GetAllCommentsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.DiscussionService)
		PostCheck    func(resp *compassv1beta1.GetAllCommentsResponse) error
	}
	var testCases = []testCase{
		{
			Description: `should return invalid argument if discussion_id is not integer`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: "test",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return invalid argument if discussion_id is < 1`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: "0",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: `should return internal server error if fetching fails`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: discussionID,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetComments(ctx, discussionID, discussion.Filter{
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
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetComments(ctx, discussionID, discussion.Filter{
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
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetComments(ctx, discussionID, discussion.Filter{
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
				expected := &compassv1beta1.GetAllCommentsResponse{
					Data: []*compassv1beta1.Comment{
						{Id: "1122"},
						{Id: "2233"},
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockSvc.AssertExpectations(t)

			handler := NewAPIServer(logger, nil, nil, mockSvc, nil, nil, nil)

			got, err := handler.GetAllComments(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
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
		Description  string
		Request      *compassv1beta1.GetCommentRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.DiscussionService)
		PostCheck    func(resp *compassv1beta1.GetCommentResponse) error
	}
	var testCases = []testCase{
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should return invalid argument if discussion id not integer`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
		},
		{
			Description:  `should return invalid argument if discussion id < 0`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
		},
		{
			Description:  `should return invalid argument if comment id not integer`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
		},
		{
			Description:  `should return invalid argument if comment id < 0`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
		},
		{
			Description:  `should return Not Found if comment or discussion not found`,
			ExpectStatus: codes.NotFound,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{}, discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID})
			},
		},
		{
			Description:  "should return status OK along with comment of a discussion",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{ID: commentID, DiscussionID: discussionID}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetCommentResponse) error {
				expected := &compassv1beta1.GetCommentResponse{
					Data: &compassv1beta1.Comment{
						Id:           commentID,
						DiscussionId: discussionID,
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockSvc.AssertExpectations(t)

			handler := NewAPIServer(logger, nil, nil, mockSvc, nil, nil, nil)

			got, err := handler.GetComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
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
		Description  string
		Request      *compassv1beta1.UpdateCommentRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.DiscussionService)
	}{
		{
			Description: "discussion id is not integer return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "discussion id is < 0 return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "comment id is not integer return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "comment id is < 0 return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty object return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty body return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
				Body:         "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  "should return internal server error if the update comment failed",
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				cmt := &discussion.Comment{
					ID:           validRequest.Id,
					DiscussionID: validRequest.DiscussionId,
					Body:         validRequest.Body,
					UpdatedBy:    user.User{ID: userID},
				}
				expectedErr := errors.New("unknown error")
				ds.EXPECT().UpdateComment(ctx, cmt).Return(expectedErr)
			},
		},
		{
			Description:  "should return Not Found if the discussion id or comment id not found",
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				cmt := &discussion.Comment{
					ID:           validRequest.Id,
					DiscussionID: validRequest.DiscussionId,
					Body:         validRequest.Body,
					UpdatedBy:    user.User{ID: userID},
				}
				expectedErr := discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID}
				ds.EXPECT().UpdateComment(ctx, cmt).Return(expectedErr)
			},
		},
		{
			Description:  "should return status OK if the comment is successfully updated",
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				cmt := &discussion.Comment{
					ID:           validRequest.Id,
					DiscussionID: validRequest.DiscussionId,
					Body:         validRequest.Body,
					UpdatedBy:    user.User{ID: userID},
				}
				ds.EXPECT().UpdateComment(ctx, cmt).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockSvc.AssertExpectations(t)

			handler := NewAPIServer(logger, nil, nil, mockSvc, nil, nil, nil)

			_, err := handler.UpdateComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus, code.String())
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
		Description  string
		Request      *compassv1beta1.DeleteCommentRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.DiscussionService)
	}{
		{
			Description:  "discussion id is not integer return bad request",
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
		},
		{
			Description:  "discussion id is < 0 return bad request",
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
		},
		{
			Description:  "comment id is not integer return bad request",
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
		},
		{
			Description:  "comment id is < 0 return bad request",
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
		},
		{
			Description:  "should return internal server error if the delete comment failed",
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				expectedErr := errors.New("unknown error")
				ds.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(expectedErr)
			},
		},
		{
			Description:  "should return invalid argument if the discussion id or comment id not found",
			ExpectStatus: codes.NotFound,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				expectedErr := discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID}
				ds.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(expectedErr)
			},
		},
		{
			Description:  "should return OK if the comment is successfully deleted",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockSvc.AssertExpectations(t)

			handler := NewAPIServer(logger, nil, nil, mockSvc, nil, nil, nil)

			_, err := handler.DeleteComment(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}
