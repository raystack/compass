package v1beta1_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/odpf/compass/api"
	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/lib/mocks"
	"github.com/odpf/compass/lineage"
	"github.com/odpf/compass/star"
	"github.com/odpf/compass/user"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGetAllAssets(t *testing.T) {
	var userID = uuid.NewString()
	type testCase struct {
		Description  string
		Request      *compassv1beta1.GetAllAssetsRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetRepository)
		PostCheck    func(resp *compassv1beta1.GetAllAssetsResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request:      &compassv1beta1.GetAllAssetsRequest{},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Filter{}).Return([]asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description: `should return internal server error if fetching total fails`,
			Request: &compassv1beta1.GetAllAssetsRequest{
				WithTotal: true,
			},
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Filter{}).Return([]asset.Asset{}, nil, nil)
				ar.On("GetCount", ctx, asset.Filter{}).Return(0, errors.New("unknown error"))
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
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
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
				ar.On("GetAll", ctx, cfg).Return([]asset.Asset{}, nil, nil)
			},
		},
		{
			Description:  "should return status OK along with list of assets",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Filter{}).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
				}, nil, nil)
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
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetAll", ctx, asset.Filter{
					Types:    []asset.Type{"job"},
					Services: []string{"kafka"},
					Size:     10,
					Offset:   5,
				}).Return([]asset.Asset{
					{ID: "testid-1"},
					{ID: "testid-2"},
					{ID: "testid-3"},
				}, nil, nil)
				ar.On("GetCount", ctx, asset.Filter{
					Types:    []asset.Type{"job"},
					Services: []string{"kafka"},
					Size:     10,
					Offset:   5,
				}).Return(150, nil, nil)
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
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockAssetRepo := new(mocks.AssetRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				AssetRepository: mockAssetRepo,
			})

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
		userID  = uuid.NewString()
		assetID = uuid.NewString()
		ast     = asset.Asset{
			ID: assetID,
		}
	)

	type testCase struct {
		Description  string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetRepository)
		PostCheck    func(resp *compassv1beta1.GetAssetByIDResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return invalid argument if asset id is not uuid`,
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return not found if asset doesn't exist`,
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: assetID})
			},
		},
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  "should return http 200 status along with the asset, if found",
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.On("GetByID", ctx, assetID).Return(ast, nil, nil)
			},
			PostCheck: func(resp *compassv1beta1.GetAssetByIDResponse) error {
				expected := &compassv1beta1.GetAssetByIDResponse{
					Data: &compassv1beta1.Asset{
						Id: assetID,
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
			mockAssetRepo := new(mocks.AssetRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				AssetRepository: mockAssetRepo,
			})

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

func TestUpsertAsset(t *testing.T) {
	var (
		userID       = uuid.NewString()
		assetID      = uuid.NewString()
		validPayload = &compassv1beta1.UpsertAssetRequest{
			Asset: &compassv1beta1.UpsertAssetRequest_BaseAsset{
				Urn:     "test dagger",
				Type:    "table",
				Name:    "de-dagger-test",
				Service: "kafka",
				Data:    &structpb.Struct{},
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
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.UpsertAssetRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
		PostCheck    func(resp *compassv1beta1.UpsertAssetResponse) error
	}

	var testCases = []testCase{
		{
			Description:  "empty object asset will return invalid argument",
			Request:      &compassv1beta1.UpsertAssetRequest{},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty urn will return invalid argument",
			Request: &compassv1beta1.UpsertAssetRequest{
				Asset: &compassv1beta1.UpsertAssetRequest_BaseAsset{
					Urn:     "",
					Name:    "some-name",
					Data:    &structpb.Struct{},
					Service: "some-service",
					Type:    "table",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty name will return invalid argument",
			Request: &compassv1beta1.UpsertAssetRequest{
				Asset: &compassv1beta1.UpsertAssetRequest_BaseAsset{
					Urn:     "some-urn",
					Name:    "",
					Data:    &structpb.Struct{},
					Service: "some-service",
					Type:    "table",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "nil data will return invalid argument",
			Request: &compassv1beta1.UpsertAssetRequest{
				Asset: &compassv1beta1.UpsertAssetRequest_BaseAsset{
					Urn:     "some-urn",
					Name:    "some-name",
					Service: "some-service",
					Type:    "table",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty service will return invalid argument",
			Request: &compassv1beta1.UpsertAssetRequest{
				Asset: &compassv1beta1.UpsertAssetRequest_BaseAsset{
					Urn:     "some-urn",
					Name:    "some-name",
					Data:    &structpb.Struct{},
					Service: "",
					Type:    "table",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty type will return invalid argument",
			Request: &compassv1beta1.UpsertAssetRequest{
				Asset: &compassv1beta1.UpsertAssetRequest_BaseAsset{
					Urn:     "some-urn",
					Name:    "some-name",
					Data:    &structpb.Struct{},
					Service: "some-service",
					Type:    "",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "invalid type will return invalid argument",
			Request: &compassv1beta1.UpsertAssetRequest{
				Asset: &compassv1beta1.UpsertAssetRequest_BaseAsset{
					Urn:     "some-urn",
					Name:    "some-name",
					Data:    &structpb.Struct{},
					Service: "some-service",
					Type:    "invalid type",
				},
			},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "should return internal server error when the asset repository fails",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				expectedErr := errors.New("unknown error")
				ar.On("Upsert", ctx, mock.AnythingOfType("*asset.Asset")).Return("1234-5678", expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return internal server error when the discovery repository fails",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				expectedErr := errors.New("unknown error")
				ar.On("Upsert", ctx, mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil, nil)
				dr.On("Upsert", ctx, mock.AnythingOfType("asset.Asset")).Return(expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return internal server error when the lineage repository fails",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				expectedErr := errors.New("unknown error")
				ar.EXPECT().Upsert(ctx, mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
				lr.EXPECT().Upsert(ctx,
					mock.AnythingOfType("lineage.Node"),
					mock.AnythingOfType("[]lineage.Node"),
					mock.AnythingOfType("[]lineage.Node"),
				).Return(expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return OK and asset's ID if the asset is successfully created/updated",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ast := asset.Asset{
					URN:       "test dagger",
					Type:      asset.TypeTable,
					Name:      "de-dagger-test",
					Service:   "kafka",
					UpdatedBy: user.User{ID: userID},
					Data:      map[string]interface{}{},
				}
				upstreams := []lineage.Node{
					{URN: "upstream-1", Type: asset.TypeJob, Service: "optimus"},
				}
				downstreams := []lineage.Node{
					{URN: "downstream-1", Type: asset.TypeDashboard, Service: "metabase"},
					{URN: "downstream-2", Type: asset.TypeDashboard, Service: "tableau"},
				}
				assetWithID := ast
				assetWithID.ID = assetID

				ar.EXPECT().Upsert(ctx, &ast).Return(assetWithID.ID, nil).Run(func(ctx context.Context, ast *asset.Asset) {
					ast.ID = assetWithID.ID
				})
				dr.EXPECT().Upsert(ctx, assetWithID).Return(nil)
				lr.EXPECT().Upsert(ctx,
					lineage.Node{
						URN:     ast.URN,
						Type:    ast.Type,
						Service: ast.Service,
					},
					upstreams,
					downstreams,
				).Return(nil)
			},
			Request:      validPayload,
			ExpectStatus: codes.OK,
			PostCheck: func(resp *compassv1beta1.UpsertAssetResponse) error {
				expected := &compassv1beta1.UpsertAssetResponse{
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
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			mockLineageRepo := new(mocks.LineageRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)
			defer mockLineageRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				AssetRepository:     mockAssetRepo,
				DiscoveryRepository: mockDiscoveryRepo,
				LineageRepository:   mockLineageRepo,
			})

			got, err := handler.UpsertAsset(ctx, tc.Request)
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
		assetID      = uuid.NewString()
		validPayload = &compassv1beta1.UpsertPatchAssetRequest{
			Asset: &compassv1beta1.UpsertPatchAssetRequest_BaseAsset{
				Urn:     "test dagger",
				Type:    "table",
				Name:    wrapperspb.String("new-name"),
				Service: "kafka",
				Data:    &structpb.Struct{},
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
		}
	)
	type testCase struct {
		Description  string
		Request      *compassv1beta1.UpsertPatchAssetRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
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
			Request:      &compassv1beta1.UpsertPatchAssetRequest{Asset: &compassv1beta1.UpsertPatchAssetRequest_BaseAsset{}},
			ExpectStatus: codes.InvalidArgument,
		},
		{
			Description: "empty urn will return invalid argument",
			Request: &compassv1beta1.UpsertPatchAssetRequest{
				Asset: &compassv1beta1.UpsertPatchAssetRequest_BaseAsset{
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
				Asset: &compassv1beta1.UpsertPatchAssetRequest_BaseAsset{
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
				Asset: &compassv1beta1.UpsertPatchAssetRequest_BaseAsset{
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
				Asset: &compassv1beta1.UpsertPatchAssetRequest_BaseAsset{
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
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				expectedErr := errors.New("unknown error")
				ar.EXPECT().Find(ctx, "test dagger", asset.TypeTable, "kafka").Return(currentAsset, expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return internal server error when upserting asset repository failed",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				expectedErr := errors.New("unknown error")
				ar.EXPECT().Find(ctx, "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
				ar.EXPECT().Upsert(ctx, mock.AnythingOfType("*asset.Asset")).Return("1234-5678", expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return internal server error when the discovery repository fails",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				expectedErr := errors.New("unknown error")
				ar.EXPECT().Find(ctx, "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
				ar.EXPECT().Upsert(ctx, mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return internal server error when the lineage repository fails",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				expectedErr := errors.New("unknown error")
				ar.EXPECT().Find(ctx, "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
				ar.EXPECT().Upsert(ctx, mock.AnythingOfType("*asset.Asset")).Return("1234-5678", nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
				lr.EXPECT().Upsert(ctx,
					mock.AnythingOfType("lineage.Node"),
					mock.AnythingOfType("[]lineage.Node"),
					mock.AnythingOfType("[]lineage.Node"),
				).Return(expectedErr)
			},
			Request:      validPayload,
			ExpectStatus: codes.Internal,
		},
		{
			Description: "should return OK and asset's ID if the asset is successfully created/patched",
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				patchedAsset := asset.Asset{
					URN:       "test dagger",
					Type:      asset.TypeTable,
					Name:      "new-name",
					Service:   "kafka",
					UpdatedBy: user.User{ID: userID},
					Data:      map[string]interface{}{},
				}
				upstreams := []lineage.Node{
					{URN: "upstream-1", Type: asset.TypeJob, Service: "optimus"},
				}
				downstreams := []lineage.Node{
					{URN: "downstream-1", Type: asset.TypeDashboard, Service: "metabase"},
					{URN: "downstream-2", Type: asset.TypeDashboard, Service: "tableau"},
				}

				assetWithID := patchedAsset
				assetWithID.ID = assetID

				ar.EXPECT().Find(ctx, "test dagger", asset.TypeTable, "kafka").Return(currentAsset, nil)
				ar.EXPECT().Upsert(ctx, &patchedAsset).Return(assetWithID.ID, nil).Run(func(ctx context.Context, ast *asset.Asset) {
					patchedAsset.ID = assetWithID.ID
				})
				dr.EXPECT().Upsert(ctx, assetWithID).Return(nil)
				lr.EXPECT().Upsert(ctx,
					lineage.Node{
						URN:     patchedAsset.URN,
						Type:    patchedAsset.Type,
						Service: patchedAsset.Service,
					},
					upstreams,
					downstreams,
				).Return(nil)
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
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			mockLineageRepo := new(mocks.LineageRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)
			defer mockLineageRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				AssetRepository:     mockAssetRepo,
				DiscoveryRepository: mockDiscoveryRepo,
				LineageRepository:   mockLineageRepo,
			})

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
		userID = uuid.NewString()
	)

	type TestCase struct {
		Description  string
		AssetID      string
		ExpectStatus codes.Code
		Setup        func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, astID string)
	}

	var testCases = []TestCase{
		{
			Description:  "should return invalid argument when asset id is not uuid",
			AssetID:      "not-uuid",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, astID string) {
				ar.EXPECT().Delete(ctx, astID).Return(asset.InvalidError{AssetID: astID})
			},
		},
		{
			Description:  "should return not found when asset cannot be found",
			AssetID:      uuid.NewString(),
			ExpectStatus: codes.NotFound,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, astID string) {
				ar.On("Delete", ctx, astID).Return(asset.NotFoundError{AssetID: astID})
			},
		},
		{
			Description:  "should return 500 on error deleting asset",
			AssetID:      uuid.NewString(),
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, astID string) {
				ar.On("Delete", ctx, astID).Return(errors.New("error deleting asset"))
			},
		},
		{
			Description:  "should return internal server error on error deleting asset from discovery",
			AssetID:      uuid.NewString(),
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, astID string) {
				ar.On("Delete", ctx, astID).Return(nil)
				dr.On("Delete", ctx, astID).Return(asset.NotFoundError{AssetID: astID})
			},
		},
		{
			Description:  "should return OK on success",
			AssetID:      uuid.NewString(),
			ExpectStatus: codes.OK,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, astID string) {
				ar.On("Delete", ctx, astID).Return(nil)
				dr.On("Delete", ctx, astID).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, tc.AssetID)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				AssetRepository:     mockAssetRepo,
				DiscoveryRepository: mockDiscoveryRepo,
			})

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
	)

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetAssetStargazersRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.StarRepository)
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
			Setup: func(ctx context.Context, sr *mocks.StarRepository) {
				sr.EXPECT().GetStargazers(ctx, defaultStarCfg, assetID).Return(nil, errors.New("some error"))
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
			Setup: func(ctx context.Context, sr *mocks.StarRepository) {
				sr.EXPECT().GetStargazers(ctx, defaultStarCfg, assetID).Return(nil, star.NotFoundError{})
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
			Setup: func(ctx context.Context, sr *mocks.StarRepository) {
				sr.EXPECT().GetStargazers(ctx, defaultStarCfg, assetID).Return([]user.User{{ID: "1"}, {ID: "2"}, {ID: "3"}}, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), userID)

			logger := log.NewNoop()
			mockStarRepo := new(mocks.StarRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockStarRepo)
			}
			defer mockStarRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				StarRepository: mockStarRepo,
			})

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

	var assetID = uuid.NewString()

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetAssetVersionHistoryRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetRepository)
		PostCheck    func(resp *compassv1beta1.GetAssetVersionHistoryResponse) error
	}

	var testCases = []TestCase{
		{
			Description:  `should return invalid argument if asset id is not uuid`,
			ExpectStatus: codes.InvalidArgument,
			Request: &compassv1beta1.GetAssetVersionHistoryRequest{
				Id: assetID,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{}, assetID).Return([]asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetAssetVersionHistoryRequest{
				Id: assetID,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{}, assetID).Return([]asset.Asset{}, errors.New("unknown error"))
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
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{
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
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{}, assetID).Return([]asset.Asset{
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
			ctx := context.Background()
			logger := log.NewNoop()
			mockAssetRepo := new(mocks.AssetRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				AssetRepository: mockAssetRepo,
			})

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
		assetID = uuid.NewString()
		version = "0.2"
		ast     = asset.Asset{
			ID:      assetID,
			Version: version,
		}
	)

	type TestCase struct {
		Description  string
		Request      *compassv1beta1.GetAssetByVersionRequest
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.AssetRepository)
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
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersion(ctx, assetID, version).Return(asset.Asset{}, asset.InvalidError{AssetID: assetID})
			},
		},
		{
			Description:  `should return not found if asset doesn't exist`,
			ExpectStatus: codes.NotFound,
			Request: &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: version,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersion(ctx, assetID, version).Return(asset.Asset{}, asset.NotFoundError{AssetID: assetID})
			},
		},
		{
			Description:  `should return internal server error if fetching fails`,
			ExpectStatus: codes.Internal,
			Request: &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: version,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersion(ctx, assetID, version).Return(asset.Asset{}, errors.New("unknown error"))
			},
		},
		{
			Description:  "should return status OK along with the asset if found",
			ExpectStatus: codes.OK,
			Request: &compassv1beta1.GetAssetByVersionRequest{
				Id:      assetID,
				Version: version,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersion(ctx, assetID, version).Return(ast, nil)
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
			ctx := context.Background()
			logger := log.NewNoop()
			mockAssetRepo := new(mocks.AssetRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)

			handler := api.NewGRPCHandler(logger, &api.Dependencies{
				AssetRepository: mockAssetRepo,
			})

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
