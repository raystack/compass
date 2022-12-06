package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetTypes(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(tc *testCase, ctx context.Context, as *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAllTypesResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "should return internal server error if failing to fetch types",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{}, errors.New("failed to fetch type"))
			},
		},
		{
			Description:  "should return internal server error if failing to fetch counts",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{}, errors.New("failed to fetch assets count"))
			},
		},
		{
			Description:  "should return all valid types with its asset count",
			ExpectStatus: codes.OK,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{
					asset.Type("table"): 10,
					asset.Type("topic"): 30,
					asset.Type("job"):   15,
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllTypesResponse) error {
				expected := &compassv1beta1.GetAllTypesResponse{
					Data: []*compassv1beta1.Type{
						{
							Name:  "table",
							Count: 10,
						},
						{
							Name:  "job",
							Count: 15,
						},
						{
							Name:  "dashboard",
							Count: 0,
						},
						{
							Name:  "topic",
							Count: 30,
						},
						{
							Name:  "feature_table",
							Count: 0,
						},
						{
							Name:  "application",
							Count: 0,
						},
						{
							Name:  "model",
							Count: 0,
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

			mockUserSvc := new(mocks.UserService)
			mockSvc := new(mocks.AssetService)
			logger := log.NewNoop()
			defer mockSvc.AssertExpectations(t)
			tc.Setup(&tc, ctx, mockSvc)

			defer mockUserSvc.AssertExpectations(t)
			defer mockSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetAllTypes(ctx, &compassv1beta1.GetAllTypesRequest{})
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
