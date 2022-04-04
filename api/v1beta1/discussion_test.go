package v1beta1_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/odpf/columbus/api"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/discussion"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetAllDiscussions(t *testing.T) {
	var userID = uuid.NewString()
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllDiscussionsRequest
		Setup        func(context.Context, *mocks.DiscussionRepository)
		ExpectStatus codes.Code
		PostCheck    func(resp *compassv1beta1.GetAllDiscussionsResponse) error
	}

	var testCases = []testCase{
		{
			Description: `should return internal server error if fetching fails`,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{
					Type:          "all",
					State:         discussion.StateOpen.String(),
					SortBy:        "created_at",
					SortDirection: "desc",
				}).Return([]discussion.Discussion{}, errors.New("unknown error"))
			},
			ExpectStatus: codes.Internal,
		},
		{
			Description: `should parse querystring to get filter`,
			Request: &compassv1beta1.GetAllDiscussionsRequest{
				Type:      discussion.TypeIssues.String(),
				State:     discussion.StateClosed.String(),
				Labels:    "label1,label2,label4",
				Assignee:  "646130cf-3dde-4d61-99e9-6070dd369597",
				Asset:     "e5d81dcd-3046-4d33-b1ac-efdd221e621d",
				Owner:     "62326386-dc9d-4ae5-9448-e54c720f856d",
				Sort:      "updated_at",
				Direction: "asc",
				Size:      30,
				Offset:    50,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{
					Type:          discussion.TypeIssues.String(),
					State:         discussion.StateClosed.String(),
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
			ExpectStatus: codes.OK,
		},
		{
			Description: "should return status OK along with list of discussions",
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().GetAll(ctx, discussion.Filter{
					Type:          "all",
					State:         discussion.StateOpen.String(),
					SortBy:        "created_at",
					SortDirection: "desc",
				}).Return([]discussion.Discussion{
					{ID: "1122"},
					{ID: "2233"},
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllDiscussionsResponse) error {
				expected := &compassv1beta1.GetAllDiscussionsResponse{
					Data: []*compassv1beta1.Discussion{
						{Id: "1122"},
						{Id: "2233"},
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
			ExpectStatus: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockRepo := new(mocks.DiscussionRepository)

			tc.Setup(ctx, mockRepo)
			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				DiscussionRepository: mockRepo,
			})
			got, err := handler.GetAllDiscussions(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", code.String(), tc.ExpectStatus.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					assert.Error(t, err)
					return
				}
			}
		})
	}
}

func TestCreateDiscussion(t *testing.T) {
	var userID = uuid.NewString()
	var validRequest = &compassv1beta1.CreateDiscussionRequest{
		Title: "Lorem Ipsum",
		Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		Type:  discussion.TypeQAndA.String(),
	}

	type testCase struct {
		Description  string
		Request      *compassv1beta1.CreateDiscussionRequest
		Setup        func(context.Context, *mocks.DiscussionRepository)
		ExpectStatus codes.Code
	}

	var testCases = []testCase{
		{
			Description:  "should return invalid argument if empty object",
			Request:      &compassv1beta1.CreateDiscussionRequest{},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if empty title",
			Request: &compassv1beta1.CreateDiscussionRequest{
				Body: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
				Type: discussion.TypeQAndA.String(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if no body",
			Request: &compassv1beta1.CreateDiscussionRequest{
				Title: "Lorem Ipsum",
				Type:  discussion.TypeQAndA.String(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if empty body",
			Request: &compassv1beta1.CreateDiscussionRequest{
				Title: "Lorem Ipsum",
				Body:  "",
				Type:  discussion.TypeQAndA.String(),
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if wrong type",
			Request: &compassv1beta1.CreateDiscussionRequest{
				Title: "Lorem Ipsum",
				Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
				Type:  "wrongtype",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  "should return internal server error if the discussion creation fails",
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Create(ctx, mock.AnythingOfType("*discussion.Discussion")).Return("", errors.New("some error"))
			},
		},
		{
			Description: "should return invalid argument if empty type",
			Request: &compassv1beta1.CreateDiscussionRequest{
				Title: "Lorem Ipsum",
				Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
				Type:  "",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  "should return OK and discussion ID if the discussion is successfully created",
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dsc := discussion.Discussion{
					Title: "Lorem Ipsum",
					Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
					Type:  discussion.TypeQAndA,
					State: discussion.StateOpen,
					Owner: user.User{ID: userID},
				}
				discussionWithID := dsc
				discussionWithID.ID = "12"
				dr.EXPECT().Create(ctx, &dsc).Run(func(ctx context.Context, dsc *discussion.Discussion) {
					dsc.ID = discussionWithID.ID
				}).Return(discussionWithID.ID, nil)
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

			_, err := handler.CreateDiscussion(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}

func TestGetDiscussion(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
	)

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetDiscussionRequest
		Setup        func(context.Context, *mocks.DiscussionRepository)
		ExpectStatus codes.Code
		PostCheck    func(resp *compassv1beta1.GetDiscussionResponse) error
	}

	var testCases = []TestCase{
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetDiscussionRequest{
				Id: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Get(ctx, discussionID).Return(discussion.Discussion{}, errors.New("unknown error"))
			},
		},
		{
			Description:  `should return invalid argument if discussion id not integer`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetDiscussionRequest{
				Id: "random",
			},
		},
		{
			Description:  `should return invalid argument if discussion id < 0`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetDiscussionRequest{
				Id: "-1",
			},
		},
		{
			Description:  `should return not found if discussion not found`,
			ExpectStatus: codes.NotFound,
			Request: &compassv1beta1.GetDiscussionRequest{
				Id: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Get(ctx, discussionID).Return(discussion.Discussion{}, discussion.NotFoundError{DiscussionID: discussionID})
			},
		},
		{
			Description:  "should return status OK along with discussions",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetDiscussionRequest{
				Id: discussionID,
			},
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Get(ctx, discussionID).Return(discussion.Discussion{ID: discussionID}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetDiscussionResponse) error {
				expected := &compassv1beta1.GetDiscussionResponse{
					Data: &compassv1beta1.Discussion{
						Id: discussionID,
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
			mockRepo := new(mocks.DiscussionRepository)

			if tc.Setup != nil {
				tc.Setup(ctx, mockRepo)
			}

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				DiscussionRepository: mockRepo,
			})
			got, err := handler.GetDiscussion(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", code.String(), tc.ExpectStatus.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					assert.Error(t, err)
					return
				}
			}
		})
	}
}

func TestPatchDiscussion(t *testing.T) {
	var (
		userID       = uuid.NewString()
		discussionID = "123"
	)

	var validRequest = &compassv1beta1.PatchDiscussionRequest{Id: discussionID, Title: "lorem ipsum"}

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.PatchDiscussionRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.DiscussionRepository)
	}

	testCases := []TestCase{
		{
			Description: "should return invalid argument if discussion id is not integer return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id: "random",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if discussion id is < 0 return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id: "-1",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if empty object return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id: discussionID,
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if invalid type return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id:   discussionID,
				Type: "random",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if invalid state return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id:    discussionID,
				State: "random",
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if assignees more than limit should return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id:        discussionID,
				Assignees: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if assets more than limit should return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id:     discussionID,
				Assets: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return invalid argument if labels more than limit should return invalid argument",
			Request: &compassv1beta1.PatchDiscussionRequest{
				Id:     discussionID,
				Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  "should return internal server error if the discussion patch fails",
			Request:      validRequest,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				expectedErr := errors.New("unknown error")
				dr.EXPECT().Patch(ctx, mock.AnythingOfType("*discussion.Discussion")).Return(expectedErr)
			},
		},
		{
			Description:  "should return Not Found if the discussion id is invalid",
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				expectedErr := discussion.NotFoundError{DiscussionID: discussionID}
				dr.EXPECT().Patch(ctx, mock.AnythingOfType("*discussion.Discussion")).Return(expectedErr)
			},
		},
		{
			Description:  "should return OK if the discussion is successfully patched",
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, dr *mocks.DiscussionRepository) {
				dr.EXPECT().Patch(ctx, mock.AnythingOfType("*discussion.Discussion")).Return(nil)
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

			_, err := handler.PatchDiscussion(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}
