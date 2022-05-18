package handlersv1beta1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetTypesTypes(t *testing.T) {
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(tc *testCase, as *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAllTypesResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "should return internal server error if failing to fetch types",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, as *mocks.AssetService) {
				as.EXPECT().GetTypes(context.Background()).Return(map[asset.Type]int{}, errors.New("failed to fetch type"))
			},
		},
		{
			Description:  "should return internal server error if failing to fetch counts",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, as *mocks.AssetService) {
				as.EXPECT().GetTypes(context.Background()).Return(map[asset.Type]int{}, errors.New("failed to fetch assets count"))
			},
		},
		{
			Description:  "should return all valid types with its asset count",
			ExpectStatus: codes.OK,
			Setup: func(tc *testCase, as *mocks.AssetService) {
				as.EXPECT().GetTypes(context.Background()).Return(map[asset.Type]int{
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
			mockSvc := new(mocks.AssetService)
			logger := log.NewNoop()
			defer mockSvc.AssertExpectations(t)
			tc.Setup(&tc, mockSvc)

			handler := NewAPIServer(logger, mockSvc, nil, nil, nil, nil, nil)

			got, err := handler.GetAllTypes(context.TODO(), &compassv1beta1.GetAllTypesRequest{})
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
