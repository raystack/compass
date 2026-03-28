package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/handler/mocks"
	"github.com/raystack/compass/internal/middleware"
	compassv1beta1 "github.com/raystack/compass/proto/gen/raystack/compass/v1beta1"
	log "github.com/raystack/salt/observability/logger"
	
	
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetTypes(t *testing.T) {
	var (
		userUUID = uuid.NewString()
		ns       = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "tenant",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)
	type testCase struct {
		Description  string
		ExpectStatus connect.Code
		Setup        func(tc *testCase, ctx context.Context, as *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAllTypesResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "should return internal server error if failing to fetch types",
			ExpectStatus: connect.CodeInternal,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{}, errors.New("failed to fetch type"))
			},
		},
		{
			Description:  "should return internal server error if failing to fetch counts",
			ExpectStatus: connect.CodeInternal,
			Setup: func(tc *testCase, ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{}, errors.New("failed to fetch assets count"))
			},
		},
		{
			Description:  "should return all valid types with its asset count",
			ExpectStatus: 0,
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
						{
							Name:  "query",
							Count: 0,
						},
						{
							Name:  "metric",
							Count: 0,
						},
						{
							Name:  "experiment",
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
			ctx = middleware.BuildContextWithNamespace(ctx, ns)

			mockSvc := new(mocks.AssetService)
			logger := log.NewNoop()
			defer mockSvc.AssertExpectations(t)
			tc.Setup(&tc, ctx, mockSvc)

			defer mockSvc.AssertExpectations(t)

			handler := NewAPIServer(logger, nil, mockSvc, nil, nil, nil, nil, nil)

			got, err := handler.GetAllTypes(ctx, connect.NewRequest(&compassv1beta1.GetAllTypesRequest{}))
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
