package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/goto/compass/core/discussion"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetAllDiscussions(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllDiscussionsRequest
		Setup        func(context.Context, *mocks.DiscussionService)
		ExpectStatus codes.Code
		PostCheck    func(resp *compassv1beta1.GetAllDiscussionsResponse) error
	}

	testCases := []testCase{
		{
			Description: `should return internal server error if fetching fails`,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
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
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
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
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
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
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{DiscussionSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

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
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	validRequest := &compassv1beta1.CreateDiscussionRequest{
		Title: "Lorem Ipsum",
		Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		Type:  discussion.TypeQAndA.String(),
	}

	type testCase struct {
		Description  string
		Request      *compassv1beta1.CreateDiscussionRequest
		Setup        func(context.Context, *mocks.DiscussionService)
		ExpectStatus codes.Code
	}

	testCases := []testCase{
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
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().CreateDiscussion(ctx, mock.AnythingOfType("*discussion.Discussion")).Return("", errors.New("some error"))
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
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				dsc := discussion.Discussion{
					Title: "Lorem Ipsum",
					Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
					Type:  discussion.TypeQAndA,
					State: discussion.StateOpen,
					Owner: user.User{ID: userID},
				}
				discussionWithID := dsc
				discussionWithID.ID = "12"
				ds.EXPECT().CreateDiscussion(ctx, &dsc).Run(func(ctx context.Context, dsc *discussion.Discussion) {
					dsc.ID = discussionWithID.ID
				}).Return(discussionWithID.ID, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{DiscussionSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

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
		userUUID     = uuid.NewString()
		discussionID = "123"
	)

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetDiscussionRequest
		Setup        func(context.Context, *mocks.DiscussionService)
		ExpectStatus codes.Code
		PostCheck    func(resp *compassv1beta1.GetDiscussionResponse) error
	}

	testCases := []TestCase{
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetDiscussionRequest{
				Id: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussion(ctx, discussionID).Return(discussion.Discussion{}, errors.New("unknown error"))
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
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussion(ctx, discussionID).Return(discussion.Discussion{}, discussion.NotFoundError{DiscussionID: discussionID})
			},
		},
		{
			Description:  "should return status OK along with discussions",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetDiscussionRequest{
				Id: discussionID,
			},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussion(ctx, discussionID).Return(discussion.Discussion{ID: discussionID}, nil)
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
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)
			handler := NewAPIServer(APIServerDeps{DiscussionSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

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
		userUUID     = uuid.NewString()
		discussionID = "123"
	)

	validRequest := &compassv1beta1.PatchDiscussionRequest{Id: discussionID, Title: "lorem ipsum"}

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.PatchDiscussionRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.DiscussionService)
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
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				expectedErr := errors.New("unknown error")
				ds.EXPECT().PatchDiscussion(ctx, mock.AnythingOfType("*discussion.Discussion")).Return(expectedErr)
			},
		},
		{
			Description:  "should return Not Found if the discussion id is invalid",
			Request:      validRequest,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				expectedErr := discussion.NotFoundError{DiscussionID: discussionID}
				ds.EXPECT().PatchDiscussion(ctx, mock.AnythingOfType("*discussion.Discussion")).Return(expectedErr)
			},
		},
		{
			Description:  "should return OK if the discussion is successfully patched",
			Request:      validRequest,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().PatchDiscussion(ctx, mock.AnythingOfType("*discussion.Discussion")).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(APIServerDeps{DiscussionSvc: mockSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.PatchDiscussion(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}

func TestDiscussionToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		Discussion  discussion.Discussion
		ExpectProto *compassv1beta1.Discussion
	}

	testCases := []testCase{
		{
			Title:       "should return no timestamp pb if timestamp is zero",
			Discussion:  discussion.Discussion{ID: "id1"},
			ExpectProto: &compassv1beta1.Discussion{Id: "id1"},
		},
		{
			Title:       "should return timestamp pb if timestamp is not zero",
			Discussion:  discussion.Discussion{ID: "id1", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.Discussion{Id: "id1", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := discussionToProto(tc.Discussion)
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestDiscussionFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title            string
		DiscussionPB     *compassv1beta1.Discussion
		ExpectDiscussion discussion.Discussion
	}

	testCases := []testCase{
		{
			Title:            "should return non empty time.Time and owner if pb is not empty or zero",
			DiscussionPB:     &compassv1beta1.Discussion{Id: "id1", Owner: &compassv1beta1.User{Id: "uid1"}, CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			ExpectDiscussion: discussion.Discussion{ID: "id1", Owner: user.User{ID: "uid1"}, Type: "openended", State: "open", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:            "should return empty time.Time and owner if pb is empty or zero",
			DiscussionPB:     &compassv1beta1.Discussion{Id: "id1"},
			ExpectDiscussion: discussion.Discussion{ID: "id1", Type: "openended", State: "open"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := discussionFromProto(tc.DiscussionPB)
			if reflect.DeepEqual(got, tc.ExpectDiscussion) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.ExpectDiscussion, got)
			}
		})
	}
}
