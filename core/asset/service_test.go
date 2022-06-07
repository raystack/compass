package asset_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/asset/mocks"
)

func TestService_GetAllAssets(t *testing.T) {
	type testCase struct {
		Description string
		Filter      asset.Filter
		WithTotal   bool
		Err         error
		ResultLen   int
		TotalCnt    uint32
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	var testCases = []testCase{
		{
			Description: `should return error if asset repository get all return error and with total false`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{}, errors.New("unknown error"))
			},
			Err:       errors.New("unknown error"),
			ResultLen: 0,
			TotalCnt:  0,
		},
		{
			Description: `should return assets if asset repository get all return no error and with total false`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{
					{
						ID: "some-id",
					},
				}, nil)
			},
			Err:       errors.New("unknown error"),
			ResultLen: 1,
			TotalCnt:  0,
		},
		{
			Description: `should return error if asset repository get count return error and with total true`,
			WithTotal:   true,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{
					{
						ID: "some-id",
					},
				}, nil)
				ar.EXPECT().GetCount(ctx, asset.Filter{}).Return(0, errors.New("unknown error"))
			},
			Err:       errors.New("unknown error"),
			ResultLen: 0,
			TotalCnt:  0,
		},
		{
			Description: `should return no error if asset repository get count return no error and with total true`,
			WithTotal:   true,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{
					{
						ID: "some-id",
					},
				}, nil)
				ar.EXPECT().GetCount(ctx, asset.Filter{}).Return(1, nil)
			},
			Err:       nil,
			ResultLen: 1,
			TotalCnt:  1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			mockLineageRepo := new(mocks.LineageRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)
			defer mockLineageRepo.AssertExpectations(t)

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			got, cnt, err := svc.GetAllAssets(ctx, tc.Filter, tc.WithTotal)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
			if tc.ResultLen != len(got) {
				t.Fatalf("got result len %v, expected result len was %v", len(got), tc.ResultLen)
			}
			if tc.TotalCnt != cnt {
				t.Fatalf("got total count %v, expected total count was %v", cnt, tc.TotalCnt)
			}
		})
	}
}

func TestService_GetTypes(t *testing.T) {
	type testCase struct {
		Description string
		Filter      asset.Filter
		Err         error
		Result      map[asset.Type]int
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	var testCases = []testCase{
		{
			Description: `should return error if asset repository get types return error`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetTypes(ctx, asset.Filter{}).Return(nil, errors.New("unknown error"))
			},
			Result: nil,
			Err:    errors.New("unknown error"),
		},
		{
			Description: `should return map types if asset repository get types return no error`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{
					asset.TypeJob:   1,
					asset.TypeTable: 1,
					asset.TypeTopic: 1,
				}, nil)
			},
			Result: map[asset.Type]int{
				asset.TypeJob:   1,
				asset.TypeTable: 1,
				asset.TypeTopic: 1,
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := new(mocks.AssetRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)

			svc := asset.NewService(mockAssetRepo, nil, nil)
			got, err := svc.GetTypes(ctx, tc.Filter)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
			if !cmp.Equal(tc.Result, got) {
				t.Fatalf("got result %+v, expected result was %+v", got, tc.Result)
			}
		})
	}
}

func TestService_UpsertAsset(t *testing.T) {
	sampleAsset := &asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.TypeDashboard, Service: "some-service"}
	sampleNodes1 := []asset.LineageNode{
		{
			URN:     "1-urn-1",
			Type:    asset.TypeJob,
			Service: "service-1",
		},
		{
			URN:     "1-urn-2",
			Type:    asset.TypeJob,
			Service: "service-1",
		},
	}
	sampleNodes2 := []asset.LineageNode{
		{
			URN:     "2-urn-1",
			Type:    asset.TypeTopic,
			Service: "service-2",
		},
		{
			URN:     "2-urn-2",
			Type:    asset.TypeJob,
			Service: "service-2",
		},
	}
	type testCase struct {
		Description string
		Asset       *asset.Asset
		Upstreams   []asset.LineageNode
		Downstreams []asset.LineageNode
		Err         error
		ReturnedID  string
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	var testCases = []testCase{
		{
			Description: `should return error if asset repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return("", errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: "",
		},
		{
			Description: `should return error if discovery repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset.ID, nil)
				dr.EXPECT().Upsert(ctx, *sampleAsset).Return(errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: sampleAsset.ID,
		},
		{
			Description: `should return error if lineage repository upsert return error`,
			Asset:       sampleAsset,
			Upstreams:   sampleNodes1,
			Downstreams: sampleNodes2,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset.ID, nil)
				dr.EXPECT().Upsert(ctx, *sampleAsset).Return(nil)
				lr.EXPECT().Upsert(ctx, asset.LineageNode{
					URN:     sampleAsset.URN,
					Type:    sampleAsset.Type,
					Service: sampleAsset.Service,
				}, sampleNodes1, sampleNodes2).Return(errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: sampleAsset.ID,
		},
		{
			Description: `should return no error if all repositories upsert return no error`,
			Asset:       sampleAsset,
			Upstreams:   sampleNodes1,
			Downstreams: sampleNodes2,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset.ID, nil)
				dr.EXPECT().Upsert(ctx, *sampleAsset).Return(nil)
				lr.EXPECT().Upsert(ctx, asset.LineageNode{
					URN:     sampleAsset.URN,
					Type:    sampleAsset.Type,
					Service: sampleAsset.Service,
				}, sampleNodes1, sampleNodes2).Return(nil)
			},
			Err:        nil,
			ReturnedID: sampleAsset.ID,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			mockLineageRepo := new(mocks.LineageRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)
			defer mockLineageRepo.AssertExpectations(t)

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			rid, err := svc.UpsertPatchAsset(ctx, tc.Asset, tc.Upstreams, tc.Downstreams)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
			if tc.ReturnedID != rid {
				t.Fatalf("got returned id %v, expected returned id was %v", rid, tc.ReturnedID)
			}
		})
	}
}

func TestService_DeleteAsset(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		Err         error
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	var testCases = []testCase{
		{
			Description: `should return error if asset repository delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Delete(ctx, assetID).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return error if discovery repository delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Delete(ctx, assetID).Return(nil)
				dr.EXPECT().Delete(ctx, assetID).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return no error if all repositories return no error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Delete(ctx, assetID).Return(nil)
				dr.EXPECT().Delete(ctx, assetID).Return(nil)
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			mockLineageRepo := new(mocks.LineageRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)
			defer mockLineageRepo.AssertExpectations(t)

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			err := svc.DeleteAsset(ctx, tc.ID)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
		})
	}
}
