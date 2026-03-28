package handler

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/discussion"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/handler/mocks"
	"github.com/raystack/compass/internal/middleware"
	compassv1beta1 "github.com/raystack/compass/proto/gen/raystack/compass/v1beta1"
	log "github.com/raystack/salt/observability/logger"
	"github.com/stretchr/testify/mock"
	
	
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateComment(t *testing.T) {
	var (
		ns = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "tenant",
			State:    namespace.SharedState,
			Metadata: nil,
		}
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		discussionID = "11111"
		validRequest = &compassv1beta1.CreateCommentRequest{
			DiscussionId: discussionID,
			Body:         "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		}
	)
	ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
	ctx = middleware.BuildContextWithNamespace(ctx, ns)
	type TestCase struct {
		Description  string
		Request      *compassv1beta1.CreateCommentRequest
		ExpectStatus connect.Code
		Result       string
		Setup        func(context.Context, *mocks.DiscussionService, *mocks.NamespaceService)
	}

	var testCases = []TestCase{
		{
			Description: "should return invalid request if empty request",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: discussionID,
			},
			ExpectStatus: connect.CodeInvalidArgument,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description: "should return invalid request if discussion_id is not integer",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: "test",
				Body:         validRequest.Body,
			},
			ExpectStatus: connect.CodeInvalidArgument,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description: "should return invalid request if discussion_id is < 1",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: "0",
				Body:         validRequest.Body,
			},
			ExpectStatus: connect.CodeInvalidArgument,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description: "should return internal server error if the comment creation failed",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: validRequest.GetDiscussionId(),
				Body:         validRequest.GetBody(),
			},
			ExpectStatus: connect.CodeInternal,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
				expectedErr := errors.New("unknown error")
				ds.EXPECT().CreateComment(ctx, ns, mock.AnythingOfType("*discussion.Comment")).Return("", expectedErr)
			},
		},
		{
			Description: "should return OK and comment ID if the comment is successfully created",
			Request: &compassv1beta1.CreateCommentRequest{
				DiscussionId: validRequest.GetDiscussionId(),
				Body:         validRequest.GetBody(),
			},
			ExpectStatus: 0,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
				ds.EXPECT().CreateComment(ctx, ns, mock.AnythingOfType("*discussion.Comment")).Return("", nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc, mockNamespaceSvc)
			}
			defer mockNamespaceSvc.AssertExpectations(t)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, ns, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, mockSvc, nil, nil, mockUserSvc)

			got, err := handler.CreateComment(ctx, connect.NewRequest(tc.Request))
			if tc.ExpectStatus == 0 {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}
			} else {
				code := connect.CodeOf(err)
				if code != tc.ExpectStatus {
					t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
					return
				}
			}
			if err == nil && got != nil && got.Msg.GetId() != tc.Result {
				t.Errorf("expected result to return id %s, returned id %s instead", tc.Result, got.Msg.Id)
				return
			}
		})
	}
}

func TestGetAllComments(t *testing.T) {
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		discussionID = "11111"
		ns           = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "tenant",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
	ctx = middleware.BuildContextWithNamespace(ctx, ns)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllCommentsRequest
		ExpectStatus connect.Code
		Setup        func(context.Context, *mocks.DiscussionService, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.GetAllCommentsResponse) error
	}
	var testCases = []testCase{
		{
			Description: `should return invalid argument if discussion_id is not integer`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: "test",
			},
			ExpectStatus: connect.CodeInvalidArgument,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description: `should return invalid argument if discussion_id is < 1`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: "0",
			},
			ExpectStatus: connect.CodeInvalidArgument,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description: `should return internal server error if fetching fails`,
			Request: &compassv1beta1.GetAllCommentsRequest{
				DiscussionId: discussionID,
			},
			ExpectStatus: connect.CodeInternal,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
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
			ExpectStatus: 0,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
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
			ExpectStatus: 0,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
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
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc, mockNamespaceSvc)
			}
			defer mockNamespaceSvc.AssertExpectations(t)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, ns, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, mockSvc, nil, nil, mockUserSvc)

			got, err := handler.GetAllComments(ctx, connect.NewRequest(tc.Request))
			if tc.ExpectStatus == 0 {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}
			} else {
				code := connect.CodeOf(err)
				if code != tc.ExpectStatus {
					t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
					return
				}
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got.Msg); err != nil {
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
		userUUID     = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
		ns           = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "tenant",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
	ctx = middleware.BuildContextWithNamespace(ctx, ns)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetCommentRequest
		ExpectStatus connect.Code
		Setup        func(context.Context, *mocks.DiscussionService, *mocks.NamespaceService)
		PostCheck    func(resp *compassv1beta1.GetCommentResponse) error
	}
	var testCases = []testCase{
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: connect.CodeInternal,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
				ds.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should return invalid argument if discussion id not integer`,
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description:  `should return invalid argument if discussion id < 0`,
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description:  `should return invalid argument if comment id not integer`,
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description:  `should return invalid argument if comment id < 0`,
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description:  `should return Not Found if comment or discussion not found`,
			ExpectStatus: connect.CodeNotFound,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
				ds.EXPECT().GetComment(ctx, commentID, discussionID).Return(discussion.Comment{}, discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID})
			},
		},
		{
			Description:  "should return status OK along with comment of a discussion",
			ExpectStatus: 0,
			Request: &compassv1beta1.GetCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
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
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc, mockNamespaceSvc)
			}
			defer mockNamespaceSvc.AssertExpectations(t)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, ns, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, mockSvc, nil, nil, mockUserSvc)

			got, err := handler.GetComment(ctx, connect.NewRequest(tc.Request))
			if tc.ExpectStatus == 0 {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}
			} else {
				code := connect.CodeOf(err)
				if code != tc.ExpectStatus {
					t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
					return
				}
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got.Msg); err != nil {
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
		userUUID     = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
		ns           = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "tenant",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
	ctx = middleware.BuildContextWithNamespace(ctx, ns)
	var validRequest = &compassv1beta1.UpdateCommentRequest{
		Id:           commentID,
		DiscussionId: discussionID,
		Body:         "lorem ipsum",
	}
	testCases := []struct {
		Description  string
		Request      *compassv1beta1.UpdateCommentRequest
		ExpectStatus connect.Code
		Setup        func(context.Context, *mocks.DiscussionService, *mocks.NamespaceService)
	}{
		{
			Description: "discussion id is not integer return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
			ExpectStatus: connect.CodeInvalidArgument,
		},
		{
			Description: "discussion id is < 0 return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
			ExpectStatus: connect.CodeInvalidArgument,
		},
		{
			Description: "comment id is not integer return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
			ExpectStatus: connect.CodeInvalidArgument,
		},
		{
			Description: "comment id is < 0 return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
			ExpectStatus: connect.CodeInvalidArgument,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
			},
		},
		{
			Description: "empty object return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			ExpectStatus: connect.CodeInvalidArgument,
		},
		{
			Description: "empty body return bad request",
			Request: &compassv1beta1.UpdateCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
				Body:         "",
			},
			ExpectStatus: connect.CodeInvalidArgument,
		},
		{
			Description:  "should return internal server error if the update comment failed",
			Request:      validRequest,
			ExpectStatus: connect.CodeInternal,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
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
			ExpectStatus: connect.CodeNotFound,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
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
			ExpectStatus: 0,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
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
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc, mockNamespaceSvc)
			}
			defer mockNamespaceSvc.AssertExpectations(t)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, ns, userUUID, "").Return(userID, nil)
			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, mockSvc, nil, nil, mockUserSvc)

			_, err := handler.UpdateComment(ctx, connect.NewRequest(tc.Request))
			if tc.ExpectStatus == 0 {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}
			} else {
				code := connect.CodeOf(err)
				if code != tc.ExpectStatus {
					t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
					return
				}
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		discussionID = "123"
		commentID    = "11"
		ns           = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "tenant",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})
	ctx = middleware.BuildContextWithNamespace(ctx, ns)
	testCases := []struct {
		Description  string
		Request      *compassv1beta1.DeleteCommentRequest
		ExpectStatus connect.Code
		Setup        func(context.Context, *mocks.DiscussionService, *mocks.NamespaceService)
	}{
		{
			Description:  "discussion id is not integer return bad request",
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: "random",
			},
		},
		{
			Description:  "discussion id is < 0 return bad request",
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: "-1",
			},
		},
		{
			Description:  "comment id is not integer return bad request",
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           "random",
				DiscussionId: discussionID,
			},
		},
		{
			Description:  "comment id is < 0 return bad request",
			ExpectStatus: connect.CodeInvalidArgument,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           "-1",
				DiscussionId: discussionID,
			},
		},
		{
			Description:  "should return internal server error if the delete comment failed",
			ExpectStatus: connect.CodeInternal,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
				expectedErr := errors.New("unknown error")
				ds.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(expectedErr)
			},
		},
		{
			Description:  "should return invalid argument if the discussion id or comment id not found",
			ExpectStatus: connect.CodeNotFound,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
				expectedErr := discussion.NotFoundError{DiscussionID: discussionID, CommentID: commentID}
				ds.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(expectedErr)
			},
		},
		{
			Description:  "should return OK if the comment is successfully deleted",
			ExpectStatus: 0,
			Request: &compassv1beta1.DeleteCommentRequest{
				Id:           commentID,
				DiscussionId: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService, nss *mocks.NamespaceService) {
				ds.EXPECT().DeleteComment(ctx, commentID, discussionID).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			mockNamespaceSvc := new(mocks.NamespaceService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc, mockNamespaceSvc)
			}
			defer mockNamespaceSvc.AssertExpectations(t)
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, ns, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockNamespaceSvc, nil, nil, mockSvc, nil, nil, mockUserSvc)

			_, err := handler.DeleteComment(ctx, connect.NewRequest(tc.Request))
			if tc.ExpectStatus == 0 {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}
			} else {
				code := connect.CodeOf(err)
				if code != tc.ExpectStatus {
					t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
					return
				}
			}
		})
	}
}

func TestCommentToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Comment     discussion.Comment
		ExpectProto *compassv1beta1.Comment
	}

	var testCases = []testCase{
		{
			Title:       "should return no timestamp pb if timestamp is zero",
			Comment:     discussion.Comment{ID: "id1"},
			ExpectProto: &compassv1beta1.Comment{Id: "id1"},
		},
		{
			Title:       "should return timestamp pb if timestamp is not zero",
			Comment:     discussion.Comment{ID: "id1", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.Comment{Id: "id1", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := commentToProto(tc.Comment)
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestCommentFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title  string
		PB     *compassv1beta1.Comment
		Expect discussion.Comment
	}

	var testCases = []testCase{
		{
			Title:  "should return non empty time.Time, owner, and updated by if pb is not empty or zero",
			PB:     &compassv1beta1.Comment{Id: "id1", Owner: &compassv1beta1.User{Id: "uid1"}, UpdatedBy: &compassv1beta1.User{Id: "uid1"}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			Expect: discussion.Comment{ID: "id1", Owner: user.User{ID: "uid1"}, UpdatedBy: user.User{ID: "uid1"}, CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:  "should return empty time.Time, owner, and updated by if pb is empty or zero",
			PB:     &compassv1beta1.Comment{Id: "id1"},
			Expect: discussion.Comment{ID: "id1"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := commentFromProto(tc.PB)
			if reflect.DeepEqual(got, tc.Expect) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}
