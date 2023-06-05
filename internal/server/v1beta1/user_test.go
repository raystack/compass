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
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/discussion"
	"github.com/goto/compass/core/star"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetUserStarredAssets(t *testing.T) {
	var (
		userUUID = uuid.NewString()
		userID   = uuid.NewString()
		offset   = 2
		size     = 10
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.StarService)
		PostCheck    func(resp *compassv1beta1.GetUserStarredAssetsResponse) error
	}

	testCases := []testCase{
		{
			Description:  "should return internal server error if failed to fetch starred",
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return(nil, errors.New("failed to fetch starred"))
			},
		},
		{
			Description:  "should return invalid argument if star repository return invalid error",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return(nil, star.InvalidError{})
			},
		},
		{
			Description:  "should return not found if starred not found",
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return starred assets of a user if no error",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return([]asset.Asset{
					{ID: "1", URN: "asset-urn-1", Type: "asset-type"},
					{ID: "2", URN: "asset-urn-2", Type: "asset-type"},
					{ID: "3", URN: "asset-urn-3", Type: "asset-type"},
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetUserStarredAssetsResponse) error {
				expected := &compassv1beta1.GetUserStarredAssetsResponse{
					Data: []*compassv1beta1.Asset{
						{
							Id:   "1",
							Urn:  "asset-urn-1",
							Type: "asset-type",
						},
						{
							Id:   "2",
							Urn:  "asset-urn-2",
							Type: "asset-type",
						},
						{
							Id:   "3",
							Urn:  "asset-urn-3",
							Type: "asset-type",
						},
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
			mockStarSvc := new(mocks.StarService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStarSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockStarSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, mockStarSvc, nil, nil, nil, mockUserSvc)

			got, err := handler.GetUserStarredAssets(ctx, &compassv1beta1.GetUserStarredAssetsRequest{
				UserId: userID,
				Offset: uint32(offset),
				Size:   uint32(size),
			})
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

func TestGetMyStarredAssets(t *testing.T) {
	var (
		userUUID = uuid.NewString()
		userID   = uuid.NewString()
		offset   = 2
		size     = 10
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.StarService)
		PostCheck    func(resp *compassv1beta1.GetMyStarredAssetsResponse) error
	}

	testCases := []testCase{
		{
			Description:  "should return internal server error if failed to fetch starred",
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return(nil, errors.New("failed to fetch starred"))
			},
		},
		{
			Description:  "should return invalid argument if star repository return invalid error",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return(nil, star.InvalidError{})
			},
		},
		{
			Description:  "should return not found if starred not found",
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return starred assets of a user if no error",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetsByUserID(ctx, star.Filter{Offset: offset, Size: size}, userID).Return([]asset.Asset{
					{ID: "1", URN: "asset-urn-1", Type: "asset-type"},
					{ID: "2", URN: "asset-urn-2", Type: "asset-type"},
					{ID: "3", URN: "asset-urn-3", Type: "asset-type"},
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetMyStarredAssetsResponse) error {
				expected := &compassv1beta1.GetMyStarredAssetsResponse{
					Data: []*compassv1beta1.Asset{
						{
							Id:   "1",
							Urn:  "asset-urn-1",
							Type: "asset-type",
						},
						{
							Id:   "2",
							Urn:  "asset-urn-2",
							Type: "asset-type",
						},
						{
							Id:   "3",
							Urn:  "asset-urn-3",
							Type: "asset-type",
						},
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
			mockStarSvc := new(mocks.StarService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStarSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockStarSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, mockStarSvc, nil, nil, nil, mockUserSvc)

			got, err := handler.GetMyStarredAssets(ctx, &compassv1beta1.GetMyStarredAssetsRequest{
				Offset: uint32(offset),
				Size:   uint32(size),
			})
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

func TestGetMyStarredAsset(t *testing.T) {
	var (
		userUUID  = uuid.NewString()
		userID    = uuid.NewString()
		assetID   = uuid.NewString()
		assetType = "an-asset-type"
		assetURN  = "dummy-asset-urn"
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.StarService)
		PostCheck    func(resp *compassv1beta1.GetMyStarredAssetResponse) error
	}

	testCases := []testCase{
		{
			Description:  "should return invalid argument if asset id is empty",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetByUserID(ctx, userID, assetID).Return(asset.Asset{}, star.ErrEmptyAssetID)
			},
		},
		{
			Description:  "should return invalid argument if repository return invalid error",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetByUserID(ctx, userID, assetID).Return(asset.Asset{}, star.InvalidError{})
			},
		},
		{
			Description:  "should return not found if star not found",
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetByUserID(ctx, userID, assetID).Return(asset.Asset{}, star.NotFoundError{})
			},
		},
		{
			Description:  "should return internal server error if failed to fetch a starred asset",
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetByUserID(ctx, userID, assetID).Return(asset.Asset{}, errors.New("failed to fetch starred"))
			},
		},
		{
			Description:  "should return a starred assets of a user if no error",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStarredAssetByUserID(ctx, userID, assetID).Return(asset.Asset{Type: asset.Type(assetType), URN: assetURN}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetMyStarredAssetResponse) error {
				expected := &compassv1beta1.GetMyStarredAssetResponse{
					Data: &compassv1beta1.Asset{
						Urn:  assetURN,
						Type: assetType,
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
			mockStarSvc := new(mocks.StarService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStarSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockStarSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, mockStarSvc, nil, nil, nil, mockUserSvc)

			got, err := handler.GetMyStarredAsset(ctx, &compassv1beta1.GetMyStarredAssetRequest{
				AssetId: assetID,
			})
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

func TestStarAsset(t *testing.T) {
	var (
		userID   = uuid.NewString()
		assetID  = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.StarService)
	}

	testCases := []testCase{
		{
			Description:  "should return invalid argument if asset id in param is invalid",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Stars(ctx, userID, assetID).Return("", star.ErrEmptyAssetID)
			},
		},
		{
			Description:  "should return invalid argument if star repository return invalid error",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Stars(ctx, userID, assetID).Return("", star.InvalidError{})
			},
		},
		{
			Description:  "should return invalid argument if user not found",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Stars(ctx, userID, assetID).Return("", star.UserNotFoundError{UserID: userID})
			},
		},
		{
			Description:  "should return internal server error if failed to star an asset",
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Stars(ctx, userID, assetID).Return("", errors.New("failed to star an asset"))
			},
		},
		{
			Description:  "should return ok if starring success",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Stars(ctx, userID, assetID).Return("1234", nil)
			},
		},
		{
			Description:  "should return ok if asset is already starred",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Stars(ctx, userID, assetID).Return("", star.DuplicateRecordError{})
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()

			mockUserSvc := new(mocks.UserService)
			mockStarSvc := new(mocks.StarService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStarSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockStarSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, mockStarSvc, nil, nil, nil, mockUserSvc)

			_, err := handler.StarAsset(ctx, &compassv1beta1.StarAssetRequest{
				AssetId: assetID,
			})
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}

func TestUnstarAsset(t *testing.T) {
	var (
		userID   = uuid.NewString()
		assetID  = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.StarService)
	}

	testCases := []testCase{
		{
			Description:  "should return invalid argument if asset id in param is empty",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Unstars(ctx, userID, assetID).Return(star.ErrEmptyAssetID)
			},
		},
		{
			Description:  "should return invalid argument if star repository return invalid error",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Unstars(ctx, userID, assetID).Return(star.InvalidError{})
			},
		},
		{
			Description:  "should return internal server error if failed to unstar an asset",
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Unstars(ctx, userID, assetID).Return(errors.New("failed to star an asset"))
			},
		},
		{
			Description:  "should return ok if unstarring success",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().Unstars(ctx, userID, assetID).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()

			mockUserSvc := new(mocks.UserService)
			mockStarSvc := new(mocks.StarService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStarSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockStarSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, mockStarSvc, nil, nil, nil, mockUserSvc)

			_, err := handler.UnstarAsset(ctx, &compassv1beta1.UnstarAssetRequest{
				AssetId: assetID,
			})
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}

func TestGetMyDiscussions(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Request      *compassv1beta1.GetMyDiscussionsRequest
		Setup        func(context.Context, *mocks.DiscussionService)
		PostCheck    func(resp *compassv1beta1.GetMyDiscussionsResponse) error
	}

	testCases := []testCase{
		{
			Description:  `should return internal server error if fetching fails`,
			Request:      &compassv1beta1.GetMyDiscussionsRequest{},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
					Type:                  "all",
					State:                 discussion.StateOpen.String(),
					Assignees:             []string{userID},
					SortBy:                "created_at",
					SortDirection:         "desc",
					DisjointAssigneeOwner: false,
				}).Return([]discussion.Discussion{}, errors.New("unknown error"))
			},
		},
		{
			Description: `should parse querystring to get filter`,
			Request: &compassv1beta1.GetMyDiscussionsRequest{
				Type:      "issues",
				State:     "closed",
				Labels:    "label1,label2,label4",
				Asset:     "e5d81dcd-3046-4d33-b1ac-efdd221e621d",
				Sort:      "updated_at",
				Direction: "asc",
				Size:      30,
				Offset:    50,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
					Type:                  "issues",
					State:                 "closed",
					Assignees:             []string{userID},
					Assets:                []string{"e5d81dcd-3046-4d33-b1ac-efdd221e621d"},
					Labels:                []string{"label1", "label2", "label4"},
					SortBy:                "updated_at",
					SortDirection:         "asc",
					Size:                  30,
					Offset:                50,
					DisjointAssigneeOwner: false,
				}).Return([]discussion.Discussion{}, nil)
			},
		},
		{
			Description: `should search by assigned or created if filter is all`,
			Request: &compassv1beta1.GetMyDiscussionsRequest{
				Filter: "all",
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
					Type:                  "all",
					State:                 "open",
					Assignees:             []string{userID},
					Owner:                 userID,
					SortBy:                "created_at",
					SortDirection:         "desc",
					DisjointAssigneeOwner: true,
				}).Return([]discussion.Discussion{}, nil)
			},
		},
		{
			Description:  `should set filter to default if empty`,
			ExpectStatus: codes.OK,
			Request:      &compassv1beta1.GetMyDiscussionsRequest{},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
					Type:                  "all",
					State:                 "open",
					Assignees:             []string{userID},
					SortBy:                "created_at",
					SortDirection:         "desc",
					Size:                  0,
					Offset:                0,
					DisjointAssigneeOwner: false,
				}).Return([]discussion.Discussion{}, nil)
			},
		},
		{
			Description:  "should return ok along with list of discussions",
			ExpectStatus: codes.OK,
			Request:      &compassv1beta1.GetMyDiscussionsRequest{},
			Setup: func(ctx context.Context, ds *mocks.DiscussionService) {
				ds.EXPECT().GetDiscussions(ctx, discussion.Filter{
					Type:                  "all",
					State:                 discussion.StateOpen.String(),
					Assignees:             []string{userID},
					SortBy:                "created_at",
					SortDirection:         "desc",
					DisjointAssigneeOwner: false,
				}).Return([]discussion.Discussion{
					{ID: "1122"},
					{ID: "2233"},
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetMyDiscussionsResponse) error {
				expected := &compassv1beta1.GetMyDiscussionsResponse{
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
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()

			mockUserSvc := new(mocks.UserService)
			mockDiscussionSvc := new(mocks.DiscussionService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscussionSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockDiscussionSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, nil, mockDiscussionSvc, nil, nil, mockUserSvc)

			got, err := handler.GetMyDiscussions(ctx, tc.Request)
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

func TestUserToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		User        user.User
		ExpectProto *compassv1beta1.User
	}

	testCases := []testCase{
		{
			Title:       "should return nil if UUID is empty",
			User:        user.User{},
			ExpectProto: nil,
		},
		{
			Title:       "should return fields without timestamp",
			User:        user.User{UUID: "uuid1", Email: "email@email.com", Provider: "provider", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.User{Uuid: "uuid1", Email: "email@email.com"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := userToProto(tc.User)
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestUserToFullProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title       string
		User        user.User
		ExpectProto *compassv1beta1.User
	}

	testCases := []testCase{
		{
			Title:       "should return nil if UUID is empty",
			User:        user.User{},
			ExpectProto: nil,
		},
		{
			Title:       "should return without timestamp pb if timestamp is zero",
			User:        user.User{UUID: "uuid1", Provider: "provider"},
			ExpectProto: &compassv1beta1.User{Uuid: "uuid1", Provider: "provider"},
		},
		{
			Title:       "should return with timestamp pb if timestamp is not zero",
			User:        user.User{UUID: "uuid1", Provider: "provider", CreatedAt: timeDummy, UpdatedAt: timeDummy},
			ExpectProto: &compassv1beta1.User{Uuid: "uuid1", Provider: "provider", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := userToFullProto(tc.User)
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestUserFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	type testCase struct {
		Title      string
		UserPB     *compassv1beta1.User
		ExpectUser user.User
	}

	testCases := []testCase{
		{
			Title:      "should return non empty time.Time if timestamp pb is not zero",
			UserPB:     &compassv1beta1.User{Uuid: "uuid1", Provider: "provider", CreatedAt: timestamppb.New(timeDummy), UpdatedAt: timestamppb.New(timeDummy)},
			ExpectUser: user.User{UUID: "uuid1", Provider: "provider", CreatedAt: timeDummy, UpdatedAt: timeDummy},
		},
		{
			Title:      "should return empty time.Time if timestamp pb is zero",
			UserPB:     &compassv1beta1.User{Uuid: "uuid1", Provider: "provider"},
			ExpectUser: user.User{UUID: "uuid1", Provider: "provider"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := userFromProto(tc.UserPB)
			if reflect.DeepEqual(got, tc.ExpectUser) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.ExpectUser, got)
			}
		})
	}
}
