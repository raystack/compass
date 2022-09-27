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
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/star"
	"github.com/odpf/compass/core/user"
	"github.com/odpf/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"github.com/odpf/salt/log"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGetAllAssets(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllAssetsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAllAssetsResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request:      &compassv1beta1.GetAllAssetsRequest{},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAllAssets(ctx, asset.Filter{}, false).Return([]asset.Asset{}, 0, errors.New("unknown error"))
			},
		},
		{
			Description: `should return internal server error if fetching total fails`,
			Request: &compassv1beta1.GetAllAssetsRequest{
				WithTotal: true,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAllAssets(ctx, asset.Filter{}, true).Return([]asset.Asset{}, 0, errors.New("unknown error"))
			},
		},
		{
			Description: `should successfully get config from request`,
			Request: &compassv1beta1.GetAllAssetsRequest{
				Types:     "table,topic",
				Services:  "bigquery,kafka",
				Sort:      "type",
				Direction: "asc",
				Data: map[string]string{
					"dataset": "booking",
					"project": "p-godata-id",
				},
				Q:         "internal",
				QFields:   "name,urn",
				Size:      30,
				Offset:    50,
				WithTotal: false,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				cfg := asset.Filter{
					Types:         []asset.Type{"table", "topic"},
					Services:      []string{"bigquery", "kafka"},
					Size:          30,
					Offset:        50,
					SortBy:        "type",
					SortDirection: "asc",
					QueryFields:   []string{"name", "urn"},
					Query:         "internal",
					Data: map[string]string{
						"dataset": "booking",
						"project": "p-godata-id",
					},
				}
				as.EXPECT().GetAllAssets(ctx, cfg, false).Return([]asset.Asset{}, 0, nil)
			},
		},
		{
			Description:  "should return status OK along with list of assets",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAllAssets(ctx, asset.Filter{}, false).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
				}, 0, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllAssetsResponse) error {
				expected := &compassv1beta1.GetAllAssetsResponse{
					Data: []*compassv1beta1.Asset{
						{Id: "testid-1"},
						{Id: "testid-2"},
					},
				}

				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("expected response to be %+v, was %+v", expected, resp)
				}
				return nil
			},
		},
		{
			Description:  "should return total in the payload if with_total flag is given",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetAllAssetsRequest{
				Types:     "job",
				Services:  "kafka",
				Size:      10,
				Offset:    5,
				WithTotal: true,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAllAssets(ctx, asset.Filter{
					Types:    []asset.Type{"job"},
					Services: []string{"kafka"},
					Size:     10,
					Offset:   5,
				}, true).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
					{ID: "testid-3"},
				}, 150, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAllAssetsResponse) error {
				expected := &compassv1beta1.GetAllAssetsResponse{
					Total: 150,
					Data: []*compassv1beta1.Asset{
						{Id: "testid-1"},
						{Id: "testid-2"},
						{Id: "testid-3"},
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
			mockAssetSvc := new(mocks.AssetService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockAssetSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockAssetSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetAllAssets(ctx, tc.Request)
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

func TestGetAssetByID(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
		assetID  = uuid.NewString()
		now      = time.Now()
		ast      = asset.Asset{
			ID: assetID,
			Probes: []asset.Probe{
				{
					ID:           uuid.NewString(),
					AssetURN:     assetID,
					Status:       "RUNNING",
					StatusReason: "reason-1",
					Metadata: map[string]interface{}{
						"foo": "bar",
					},
					Timestamp: now,
					CreatedAt: now.Add(-24 * time.Hour),
				},
				{
					ID:           uuid.NewString(),
					AssetURN:     assetID,
					Status:       "FAILED",
					StatusReason: "reason-2",
					Timestamp:    now.Add(2 * time.Hour),
					CreatedAt:    now.Add(-26 * time.Hour),
				},
			},
		}
	)

	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAssetByIDResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return invalid argument if asset id is not uuid`,
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByID(ctx, assetID).Return(asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return not found if asset doesn't exist`,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: assetID})
			},
		},
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByID(ctx, assetID).Return(asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  "should return http 200 status along with the asset, if found",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByID(ctx, assetID).Return(ast, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAssetByIDResponse) error {
				expected := &compassv1beta1.GetAssetByIDResponse{
					Data: &compassv1beta1.Asset{
						Id: assetID,
						Probes: []*compassv1beta1.Probe{
							{
								Id:           ast.Probes[0].ID,
								AssetUrn:     ast.Probes[0].AssetURN,
								Status:       ast.Probes[0].Status,
								StatusReason: ast.Probes[0].StatusReason,
								Metadata:     newStructpb(t, ast.Probes[0].Metadata),
								Timestamp:    timestamppb.New(ast.Probes[0].Timestamp),
								CreatedAt:    timestamppb.New(ast.Probes[0].CreatedAt),
							},
							{
								Id:           ast.Probes[1].ID,
								AssetUrn:     ast.Probes[1].AssetURN,
								Status:       ast.Probes[1].Status,
								StatusReason: ast.Probes[1].StatusReason,
								Timestamp:    timestamppb.New(ast.Probes[1].Timestamp),
								CreatedAt:    timestamppb.New(ast.Probes[1].CreatedAt),
							},
						},
					},
				}
				if diff := cmp.Diff(resp, expected, protocmp.Transform()); diff != "" {
					return fmt.Errorf("mismatch (-want +got):\n%s", diff)
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := mocks.NewUserService(t)
			mockAssetSvc := mocks.NewAssetService(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetSvc)
			}

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockAssetSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetAssetByID(ctx, &compassv1beta1.GetAssetByIDRequest{Id: assetID})
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

func TestUpsertPatchAsset(t *testing.T) {
	var (
		userID       = uuid.NewString()
		userUUID     = uuid.NewString()
		assetID      = uuid.NewString()
		validPayload = &compassv1beta1.UpsertPatchAssetRequest{
			Asset: &compassv1beta1.UpsertPatchAssetRequest_Asset{
				Urn:     "test dagger",
				Type:    "table",
				Name:    wrapperspb.String("new-name"),
				Service: "kafka",
				Data:    &structpb.Struct{},
				Owners:  []*compassv1beta1.User{{Id: "id", Uuid: "", Email: "email@email.com", Provider: "provider"}},
			},
			Upstreams: []*compassv1beta1.LineageNode{
				{
					Urn:     "upstream-1",
					Type:    "job",
					Service: "optimus",
				},
			},
			Downstreams: []*compassv1beta1.LineageNode{
				{
					Urn:     "downstream-1",
					Type:    "dashboard",
					Service: "metabase",
				},
				{
					Urn:     "downstream-2",
					Type:    "dashboard",
					Service: "tableau",
				},
			},
		}
		currentAsset = asset.Asset{
			URN:       "test dagger",
			Type:      asset.TypeTable,
			Name:      "old-name", // this value will be updated
			Service:   "kafka",
			UpdatedBy: user.User{ID: userID},
			Data:      map[string]interface{}{},
			Owners:    []user.User{{ID: "id", UUID: "", Email: "email@email.com", Provider: "provider"}},
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.UpsertPatchAssetRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.UpsertPatchAssetResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "empty payload will return invalid argument",
			Request:      &compassv1beta1.UpsertPatchAssetRequest{},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description:  "empty asset will return invalid argument",
			Request:      &compassv1beta1.UpsertPatchAssetRequest{Asset: &compassv1beta1.UpsertPatchAssetRequest_Asset{}},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty urn will return invalid argument",
			Request: &compassv1beta1.UpsertPatchAssetRequest{
				Asset: &compassv1beta1.UpsertPatchAssetRequest_Asset{
					Urn:     "",
					Name:    wrapperspb.String("some-name"),
					Data:    &structpb.Struct{},
					Service: "some-service",
					Type:    "table",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty service will return invalid argument",
			Request: &compassv1beta1.UpsertPatchAssetRequest{
				Asset: &compassv1beta1.UpsertPatchAssetRequest_Asset{
					Urn:     "some-urn",
					Name:    wrapperspb.String("some-name"),
					Data:    &structpb.Struct{},
					Service: "",
					Type:    "table",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty type will return invalid argument",
			Request: &compassv1beta1.UpsertPatchAssetRequest{
				Asset: &compassv1beta1.UpsertPatchAssetRequest_Asset{
					Urn:     "some-urn",
					Name:    wrapperspb.String("some-name"),
					Data:    &structpb.Struct{},
					Service: "some-service",
					Type:    "",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "invalid type will return invalid argument",
			Request: &compassv1beta1.UpsertPatchAssetRequest{
				Asset: &compassv1beta1.UpsertPatchAssetRequest_Asset{
					Urn:     "some-urn",
					Name:    wrapperspb.String("some-name"),
					Data:    &structpb.Struct{},
					Service: "some-service",
					Type:    "invalid type",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return internal server error when finding asset failed",
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				expectedErr := errors.New("unknown error")
				as.EXPECT().GetAssetByID(ctx, "test dagger").Return(currentAsset, expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return internal server error when upserting asset service failed",
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				expectedErr := errors.New("unknown error")
				as.EXPECT().GetAssetByID(ctx, "test dagger").Return(currentAsset, nil)
				as.EXPECT().UpsertPatchAsset(ctx, mock.AnythingOfType("*asset.Asset"), mock.AnythingOfType("[]asset.LineageNode"), mock.AnythingOfType("[]asset.LineageNode")).Return("1234-5678", expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return OK and asset's ID if the asset is successfully created/patched",
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				patchedAsset := asset.Asset{
					URN:       "test dagger",
					Type:      asset.TypeTable,
					Name:      "new-name",
					Service:   "kafka",
					UpdatedBy: user.User{ID: userID},
					Data:      map[string]interface{}{},
					Owners:    []user.User{{ID: "id", UUID: "", Email: "email@email.com", Provider: "provider"}},
				}
				upstreams := []asset.LineageNode{
					{URN: "upstream-1", Type: asset.TypeJob, Service: "optimus"},
				}
				downstreams := []asset.LineageNode{
					{URN: "downstream-1", Type: asset.TypeDashboard, Service: "metabase"},
					{URN: "downstream-2", Type: asset.TypeDashboard, Service: "tableau"},
				}

				assetWithID := patchedAsset
				assetWithID.ID = assetID

				as.EXPECT().GetAssetByID(ctx, "test dagger").Return(currentAsset, nil)
				as.EXPECT().UpsertPatchAsset(ctx, &patchedAsset, upstreams, downstreams).Return(assetWithID.ID, nil).Run(func(ctx context.Context, ast *asset.Asset, upstreams, downstreams []asset.LineageNode) {
					patchedAsset.ID = assetWithID.ID
				})
			},
			Request:      validPayload,
			ExpectStatus: codes.OK,
			PostCheck: func(resp *compassv1beta1.UpsertPatchAssetResponse) error {
				expected := &compassv1beta1.UpsertPatchAssetResponse{
					Id: assetID,
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
			mockAssetSvc := new(mocks.AssetService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockAssetSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockAssetSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.UpsertPatchAsset(ctx, tc.Request)
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

func TestDeleteAsset(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)
	type TestCase struct {
		Description  string
		AssetID      string
		ExpectStatus codes.Code
		Setup        func(ctx context.Context, as *mocks.AssetService, astID string)
	}

	var testCases = []TestCase{
		{
			Description:  "should return invalid argument when asset id is not uuid",
			AssetID:      "not-uuid",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, as *mocks.AssetService, astID string) {
				as.EXPECT().DeleteAsset(ctx, "not-uuid").Return(asset.InvalidError{AssetID: astID})
			},
		},
		{
			Description:  "should return not found when asset cannot be found",
			AssetID:      assetID,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, as *mocks.AssetService, astID string) {
				as.EXPECT().DeleteAsset(ctx, astID).Return(asset.NotFoundError{AssetID: astID})
			},
		},
		{
			Description:  "should return 500 on error deleting asset",
			AssetID:      assetID,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, as *mocks.AssetService, astID string) {
				as.EXPECT().DeleteAsset(ctx, astID).Return(errors.New("error deleting asset"))
			},
		},
		{
			Description:  "should return OK on success",
			AssetID:      assetID,
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, as *mocks.AssetService, astID string) {
				as.EXPECT().DeleteAsset(ctx, astID).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: userUUID})

			logger := log.NewNoop()
			mockUserSvc := new(mocks.UserService)
			mockAssetSvc := new(mocks.AssetService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetSvc, assetID)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockAssetSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockAssetSvc, nil, nil, nil, nil, mockUserSvc)

			_, err := handler.DeleteAsset(ctx, &compassv1beta1.DeleteAssetRequest{Id: tc.AssetID})
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}

func TestGetAssetStargazers(t *testing.T) {

	var (
		offset         = 10
		size           = 20
		defaultStarCfg = star.Filter{Offset: offset, Size: size}
		assetID        = uuid.NewString()
		userID         = uuid.NewString()
		userUUID       = uuid.NewString()
	)

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetAssetStargazersRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.StarService)
		PostCheck    func(resp *compassv1beta1.GetAssetStargazersResponse) error
	}

	var testCases = []TestCase{
		{
			Description:  "should return internal server error if failed to fetch star repository",
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetAssetStargazersRequest{
				Id:     assetID,
				Size:   uint32(size),
				Offset: uint32(offset),
			},
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStargazers(ctx, defaultStarCfg, assetID).Return(nil, errors.New("some error"))
			},
		},
		{
			Description:  "should return not found if star repository return not found error",
			ExpectStatus: codes.NotFound,
			Request: &compassv1beta1.GetAssetStargazersRequest{
				Id:     assetID,
				Size:   uint32(size),
				Offset: uint32(offset),
			},
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStargazers(ctx, defaultStarCfg, assetID).Return(nil, star.NotFoundError{})
			},
		},
		{
			Description:  "should return OK if star repository return nil error",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetAssetStargazersRequest{
				Id:     assetID,
				Size:   uint32(size),
				Offset: uint32(offset),
			},
			Setup: func(ctx context.Context, ss *mocks.StarService) {
				ss.EXPECT().GetStargazers(ctx, defaultStarCfg, assetID).Return([]user.User{{ID: "1"}, {ID: "2"}, {ID: "3"}}, nil)
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
			defer mockStarSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, nil, mockStarSvc, nil, nil, nil, mockUserSvc)

			got, err := handler.GetAssetStargazers(ctx, tc.Request)
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

func TestGetAssetVersionHistory(t *testing.T) {

	var (
		assetID  = uuid.NewString()
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
	)

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetAssetVersionHistoryRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAssetVersionHistoryResponse) error
	}

	var testCases = []TestCase{
		{
			Description:  `should return invalid argument if asset id is not uuid`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetAssetVersionHistoryRequest{
				Id: assetID,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetVersionHistory(ctx, asset.Filter{}, assetID).Return([]asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetAssetVersionHistoryRequest{
				Id: assetID,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetVersionHistory(ctx, asset.Filter{}, assetID).Return([]asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description: `should parse querystring to get config`,
			Request: &compassv1beta1.GetAssetVersionHistoryRequest{
				Id:     assetID,
				Size:   30,
				Offset: 50,
			},
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetVersionHistory(ctx, asset.Filter{
					Size:   30,
					Offset: 50,
				}, assetID).Return([]asset.Asset{}, nil)
			},
		},
		{
			Description:  "should return status OK along with list of asset versions",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetAssetVersionHistoryRequest{
				Id: assetID,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetVersionHistory(ctx, asset.Filter{}, assetID).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
				}, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAssetVersionHistoryResponse) error {
				expected := &compassv1beta1.GetAssetVersionHistoryResponse{
					Data: []*compassv1beta1.Asset{
						{
							Id: "testid-1",
						},
						{
							Id: "testid-2",
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
			mockAssetSvc := new(mocks.AssetService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockAssetSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockAssetSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetAssetVersionHistory(ctx, tc.Request)
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

func TestGetAssetByVersion(t *testing.T) {

	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
		assetID  = uuid.NewString()
		version  = "0.2"
		ast      = asset.Asset{
			ID:      assetID,
			Version: version,
		}
	)

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetAssetByVersionRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.GetAssetByVersionResponse) error
	}

	var testCases = []TestCase{
		{
			Description: `should return invalid argument if asset id is not uuid`,
			Request: &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: version,
			},
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByVersion(ctx, assetID, version).Return(asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return not found if asset doesn't exist`,
			ExpectStatus: codes.NotFound,
			Request: &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: version,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByVersion(ctx, assetID, version).Return(asset.Asset{}, asset.NotFoundError{AssetID: assetID})
			},
		},
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: version,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByVersion(ctx, assetID, version).Return(asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  "should return status OK along with the asset if found",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: version,
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().GetAssetByVersion(ctx, assetID, version).Return(ast, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAssetByVersionResponse) error {
				expected := &compassv1beta1.GetAssetByVersionResponse{
					Data: &compassv1beta1.Asset{
						Id:      assetID,
						Version: version,
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
			mockAssetSvc := new(mocks.AssetService)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetSvc)
			}
			defer mockUserSvc.AssertExpectations(t)
			defer mockAssetSvc.AssertExpectations(t)

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockAssetSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.GetAssetByVersion(ctx, tc.Request)
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

func TestCreateAssetProbe(t *testing.T) {
	var (
		userID   = uuid.NewString()
		userUUID = uuid.NewString()
		assetURN = "test-urn"
		now      = time.Now().UTC()
		probeID  = uuid.NewString()
	)

	type testCase struct {
		Description  string
		Request      *compassv1beta1.CreateAssetProbeRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetService)
		PostCheck    func(resp *compassv1beta1.CreateAssetProbeResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return error if payload is invalid`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.CreateAssetProbeRequest{
				AssetUrn: assetURN,
				Probe:    &compassv1beta1.CreateAssetProbeRequest_Probe{},
			},
		},
		{
			Description:  `should return not found if asset doesn't exist`,
			ExpectStatus: codes.NotFound,
			Request: &compassv1beta1.CreateAssetProbeRequest{
				AssetUrn: assetURN,
				Probe: &compassv1beta1.CreateAssetProbeRequest_Probe{
					Status: "RUNNING",
				},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().
					AddProbe(ctx, assetURN, mock.AnythingOfType("*asset.Probe")).
					Return(asset.NotFoundError{URN: assetURN})
			},
		},
		{
			Description:  `should return internal server error if adding probe fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.CreateAssetProbeRequest{
				AssetUrn: assetURN,
				Probe: &compassv1beta1.CreateAssetProbeRequest_Probe{
					Status: "RUNNING",
				},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				as.EXPECT().
					AddProbe(ctx, assetURN, mock.AnythingOfType("*asset.Probe")).
					Return(errors.New("unknown error"))
			},
		},
		{
			Description:  "should return probe on success",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.CreateAssetProbeRequest{
				AssetUrn: assetURN,
				Probe: &compassv1beta1.CreateAssetProbeRequest_Probe{
					Status:       "FINISHED",
					StatusReason: "test reason",
					Timestamp:    timestamppb.New(now),
					Metadata: newStructpb(t, map[string]interface{}{
						"foo1": "bar1",
						"foo2": "bar2",
					}),
				},
			},
			Setup: func(ctx context.Context, as *mocks.AssetService) {
				expectedProbe := &asset.Probe{
					Status:       "FINISHED",
					StatusReason: "test reason",
					Timestamp:    now,
					Metadata: map[string]interface{}{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				}
				as.EXPECT().AddProbe(ctx, assetURN, expectedProbe).Run(func(ctx context.Context, assetURN string, probe *asset.Probe) {
					probe.ID = probeID
				}).Return(nil)
			},
			PostCheck: func(resp *compassv1beta1.CreateAssetProbeResponse) error {
				expected := &compassv1beta1.CreateAssetProbeResponse{
					Id: probeID,
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
			mockUserSvc := mocks.NewUserService(t)
			mockAssetSvc := mocks.NewAssetService(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetSvc)
			}

			mockUserSvc.EXPECT().ValidateUser(ctx, userUUID, "").Return(userID, nil)

			handler := NewAPIServer(logger, mockAssetSvc, nil, nil, nil, nil, mockUserSvc)

			got, err := handler.CreateAssetProbe(ctx, tc.Request)
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

func TestAssetToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	dataPB, err := structpb.NewStruct(map[string]interface{}{
		"data1": "datavalue1",
	})
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		Title       string
		Asset       asset.Asset
		ExpectProto *compassv1beta1.Asset
	}

	var testCases = []testCase{
		{
			Title:       "should return nil data pb, label pb, empty owners pb, nil changelog pb, no timestamp pb if data is empty",
			Asset:       asset.Asset{ID: "id1", URN: "urn1"},
			ExpectProto: &compassv1beta1.Asset{Id: "id1", Urn: "urn1"},
		},
		{
			Title: "should return full pb if all fileds are not zero",
			Asset: asset.Asset{
				ID:  "id1",
				URN: "urn1",
				Data: map[string]interface{}{
					"data1": "datavalue1",
				},
				Labels: map[string]string{
					"label1": "labelvalue1",
				},
				Changelog: diff.Changelog{
					diff.Change{
						From: "1",
						To:   "2",
						Path: []string{"path1/path2"},
					},
				},
				CreatedAt: timeDummy,
				UpdatedAt: timeDummy,
			},
			ExpectProto: &compassv1beta1.Asset{
				Id:   "id1",
				Urn:  "urn1",
				Data: dataPB,
				Labels: map[string]string{
					"label1": "labelvalue1",
				},
				Changelog: []*compassv1beta1.Change{
					{

						From: structpb.NewStringValue("1"),
						To:   structpb.NewStringValue("2"),
						Path: []string{"path1/path2"},
					},
				},
				CreatedAt: timestamppb.New(timeDummy),
				UpdatedAt: timestamppb.New(timeDummy),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got, err := assetToProto(tc.Asset, true)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestAssetFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	dataPB, err := structpb.NewStruct(map[string]interface{}{
		"data1": "datavalue1",
	})
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		Title       string
		AssetPB     *compassv1beta1.Asset
		ExpectAsset asset.Asset
	}

	var testCases = []testCase{
		{
			Title:       "should return empty labels, data, and owners if all pb empty",
			AssetPB:     &compassv1beta1.Asset{Id: "id1"},
			ExpectAsset: asset.Asset{ID: "id1"},
		},
		{
			Title: "should return non empty labels, data, and owners if all pb is not empty",
			AssetPB: &compassv1beta1.Asset{
				Id:   "id1",
				Urn:  "urn1",
				Name: "name1",
				Data: dataPB,
				Labels: map[string]string{
					"label1": "labelvalue1",
				},
				Owners: []*compassv1beta1.User{
					{
						Id: "uid1",
					},
					{
						Id: "uid2",
					},
				},
				Changelog: []*compassv1beta1.Change{
					{

						From: structpb.NewStringValue("1"),
						To:   structpb.NewStringValue("2"),
						Path: []string{"path1/path2"},
					},
				},
				CreatedAt: timestamppb.New(timeDummy),
				UpdatedAt: timestamppb.New(timeDummy),
			},
			ExpectAsset: asset.Asset{
				ID:   "id1",
				URN:  "urn1",
				Name: "name1",
				Data: map[string]interface{}{
					"data1": "datavalue1",
				},
				Labels: map[string]string{
					"label1": "labelvalue1",
				},
				Owners: []user.User{
					{
						ID: "uid1",
					},
					{
						ID: "uid2",
					},
				},
				Changelog: diff.Changelog{
					diff.Change{
						From: "1",
						To:   "2",
						Path: []string{"path1/path2"},
					},
				},
				CreatedAt: timeDummy,
				UpdatedAt: timeDummy,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := assetFromProto(tc.AssetPB)
			if reflect.DeepEqual(got, tc.ExpectAsset) == false {
				t.Errorf("expected returned asset to be %+v, was %+v", tc.ExpectAsset, got)
			}
		})
	}
}

func newStructpb(t *testing.T, v map[string]interface{}) *structpb.Struct {
	res, err := structpb.NewStruct(v)
	require.NoError(t, err)

	return res
}
