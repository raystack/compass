package v1beta1_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/columbus/api"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetAllTypes(t *testing.T) {
	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(tc *testCase, er *mocks.TypeRepository)
		PostCheck    func(resp *compassv1beta1.GetAllTypesResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "should return internal server error if failing to fetch types",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, er *mocks.TypeRepository) {
				er.On("GetAll", context.Background()).Return(map[asset.Type]int{}, errors.New("failed to fetch type"))
			},
		},
		{
			Description:  "should return internal server error if failing to fetch counts",
			ExpectStatus: codes.Internal,
			Setup: func(tc *testCase, er *mocks.TypeRepository) {
				er.On("GetAll", context.Background()).Return(map[asset.Type]int{}, errors.New("failed to fetch assets count"))
			},
		},
		{
			Description:  "should return all valid types with its record count",
			ExpectStatus: codes.OK,
			Setup: func(tc *testCase, er *mocks.TypeRepository) {
				er.On("GetAll", context.Background()).Return(map[asset.Type]int{
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
			er := new(mocks.TypeRepository)
			logger := log.NewNoop()
			defer er.AssertExpectations(t)
			tc.Setup(&tc, er)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				TypeRepository: er,
			})

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
