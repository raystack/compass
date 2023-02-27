package asset_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/asset/mocks"
	"github.com/stretchr/testify/assert"
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

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}

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

			mockAssetRepo := mocks.NewAssetRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

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
	sampleNodes1 := []string{"1-urn-1", "1-urn-2"}
	sampleNodes2 := []string{"2-urn-1", "2-urn-2"}
	type testCase struct {
		Description string
		Asset       *asset.Asset
		Upstreams   []string
		Downstreams []string
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
				lr.EXPECT().Upsert(ctx, sampleAsset.URN, sampleNodes1, sampleNodes2).Return(errors.New("unknown error"))
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
				lr.EXPECT().Upsert(ctx, sampleAsset.URN, sampleNodes1, sampleNodes2).Return(nil)
			},
			Err:        nil,
			ReturnedID: sampleAsset.ID,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			rid, err := svc.UpsertAsset(ctx, tc.Asset, tc.Upstreams, tc.Downstreams)
			if tc.Err != nil {
				assert.EqualError(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.ReturnedID, rid)
		})
	}
}

func TestService_UpsertAssetWithoutLineage(t *testing.T) {
	sampleAsset := &asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.TypeDashboard, Service: "some-service"}
	var testCases = []struct {
		Description string
		Asset       *asset.Asset
		Err         error
		ReturnedID  string
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository)
	}{
		{
			Description: `should return error if asset repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return("", errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return error if discovery repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset.ID, nil)
				dr.EXPECT().Upsert(ctx, *sampleAsset).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return no error if all repositories upsert return no error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset.ID, nil)
				dr.EXPECT().Upsert(ctx, *sampleAsset).Return(nil)
			},
			ReturnedID: sampleAsset.ID,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo)
			}

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mocks.NewLineageRepository(t))
			rid, err := svc.UpsertAssetWithoutLineage(ctx, tc.Asset)
			if tc.Err != nil {
				assert.EqualError(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.ReturnedID, rid)
		})
	}
}

func TestService_DeleteAsset(t *testing.T) {
	assetID := "d9351e2e-a6b2-4c5d-af68-b95432e30203"
	urn := "my-test-urn"
	type testCase struct {
		Description string
		ID          string
		Err         error
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	var testCases = []testCase{
		{
			Description: `with ID, should return error if asset repository delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByID(ctx, assetID).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with ID, should return error if discovery repository delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByID(ctx, assetID).Return(nil)
				dr.EXPECT().DeleteByID(ctx, assetID).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with URN, should return error if asset repository delete return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with URN, should return error if discovery repository delete return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should call DeleteByID on repositories when given a UUID`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByID(ctx, assetID).Return(nil)
				dr.EXPECT().DeleteByID(ctx, assetID).Return(nil)
			},
			Err: nil,
		},
		{
			Description: `should call DeleteByURN on repositories when not given a UUID`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(nil)
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			err := svc.DeleteAsset(ctx, tc.ID)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
		})
	}
}

func TestService_GetAssetByID(t *testing.T) {
	assetID := "f742aa61-1100-445c-8d72-355a42e2fb59"
	urn := "my-test-urn"
	now := time.Now().UTC()
	type testCase struct {
		Description string
		ID          string
		Expected    *asset.Asset
		ExpectedErr error
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	ast := asset.Asset{
		ID: assetID,
	}

	var testCases = []testCase{
		{
			Description: `should return error if the repository return error without id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{})
			},
			ExpectedErr: asset.NotFoundError{},
		},
		{
			Description: `should return error if the repository return error, with id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: ast.ID})
			},
			ExpectedErr: asset.NotFoundError{AssetID: ast.ID},
		},
		{
			Description: `should return error if the repository return error, with invalid id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.InvalidError{AssetID: ast.ID})
			},
			ExpectedErr: asset.InvalidError{AssetID: ast.ID},
		},
		{
			Description: `with URN, should return error from repository`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByURN(ctx, urn).Return(asset.Asset{}, errors.New("the world exploded"))
			},
			ExpectedErr: errors.New("the world exploded"),
		},
		{
			Description: `with ID, should return no error if asset is found`,
			ID:          assetID,
			Expected: &asset.Asset{
				ID:        assetID,
				URN:       urn,
				CreatedAt: now,
				Probes: []asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				},
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{
					ID:        assetID,
					URN:       urn,
					CreatedAt: now,
				}, nil)
				ar.EXPECT().GetProbes(ctx, urn).Return([]asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				}, nil)
			},
			ExpectedErr: nil,
		},
		{
			Description: `with URN, should return no error if asset is found`,
			ID:          urn,
			Expected: &asset.Asset{
				ID:        assetID,
				URN:       urn,
				CreatedAt: now,
				Probes: []asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				},
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByURN(ctx, urn).Return(asset.Asset{
					ID:        assetID,
					URN:       urn,
					CreatedAt: now,
				}, nil)
				ar.EXPECT().GetProbes(ctx, urn).Return([]asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				}, nil)
			},
			ExpectedErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc := asset.NewService(mockAssetRepo, mocks.NewDiscoveryRepository(t), mocks.NewLineageRepository(t))
			actual, err := svc.GetAssetByID(ctx, tc.ID)
			if tc.Expected != nil {
				assert.Equal(t, *tc.Expected, actual)
			}
			if tc.ExpectedErr != nil {
				assert.ErrorContains(t, err, tc.ExpectedErr.Error())
				assert.ErrorAs(t, err, &tc.ExpectedErr)
			}
		})
	}
}

func TestService_GetAssetByVersion(t *testing.T) {
	assetID := "f742aa61-1100-445c-8d72-355a42e2fb59"
	urn := "my-test-urn"
	type testCase struct {
		Description string
		ID          string
		ExpectedErr error
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	var testCases = []testCase{
		{
			Description: `should return error if the GetByVersionWithID function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithID(ctx, assetID, "v0.0.2").
					Return(asset.Asset{}, errors.New("error fetching asset"))
			},
			ExpectedErr: errors.New("error fetching asset"),
		},
		{
			Description: `should return error if the GetByVersionWithURN function return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithURN(ctx, urn, "v0.0.2").
					Return(asset.Asset{}, errors.New("error fetching asset"))
			},
			ExpectedErr: errors.New("error fetching asset"),
		},
		{
			Description: `should return no error if asset is found with ID`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithID(ctx, assetID, "v0.0.2").Return(asset.Asset{}, nil)
			},
			ExpectedErr: nil,
		},
		{
			Description: `should return no error if asset is found with URN`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithURN(ctx, urn, "v0.0.2").Return(asset.Asset{}, nil)
			},
			ExpectedErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc := asset.NewService(mockAssetRepo, mocks.NewDiscoveryRepository(t), mocks.NewLineageRepository(t))
			_, err := svc.GetAssetByVersion(ctx, tc.ID, "v0.0.2")
			if tc.ExpectedErr != nil {
				assert.EqualError(t, err, tc.ExpectedErr.Error())
			}
		})
	}
}

func TestService_GetAssetVersionHistory(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		Err         error
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	ast := []asset.Asset{}
	var testCases = []testCase{
		{
			Description: `should return error if the GetVersionHistory function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{}, assetID).Return(ast, errors.New("error fetching asset"))
			},
			Err: errors.New("error fetching asset"),
		},
		{
			Description: `should return no error if asset is found by the version`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{}, assetID).Return(ast, nil)
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			_, err := svc.GetAssetVersionHistory(ctx, asset.Filter{}, tc.ID)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
		})
	}
}

func TestService_GetLineage(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
		Expected    asset.Lineage
		Err         error
	}

	var testCases = []testCase{
		{
			Description: `should return error if the GetGraph function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{}).
					Return(asset.LineageGraph{}, errors.New("error fetching graph"))
			},
			Expected: asset.Lineage{},
			Err:      errors.New("error fetching graph"),
		},
		{
			Description: `should return no error if graph with 0 edges are returned`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{}).
					Return(asset.LineageGraph{}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1"},
					MaxRows:   1,
				}).Return(nil, nil)
			},
			Expected: asset.Lineage{Edges: []asset.LineageEdge{}, NodeAttrs: map[string]asset.NodeAttributes{}},
			Err:      nil,
		},
		{
			Description: `should return an error if GetProbesWithFilter function returns error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{}).Return(asset.LineageGraph{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1", "urn-target-1", "urn-target-2", "urn-target-3"},
					MaxRows:   1,
				}).Return(nil, errors.New("error fetching probes"))
			},
			Expected: asset.Lineage{},
			Err:      errors.New("error fetching probes"),
		},
		{
			Description: `should return no error if GetProbesWithFilter function returns 0 probes`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{}).Return(asset.LineageGraph{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1", "urn-target-1", "urn-target-2", "urn-target-3"},
					MaxRows:   1,
				}).Return(nil, nil)
			},
			Expected: asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				},
				NodeAttrs: map[string]asset.NodeAttributes{},
			},
			Err: nil,
		},
		{
			Description: `should return lineage with edges and node attributes`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{}).Return(asset.LineageGraph{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1", "urn-target-1", "urn-target-2", "urn-target-3"},
					MaxRows:   1,
				}).Return(
					map[string][]asset.Probe{
						"urn-source-1": {
							asset.Probe{Status: "SUCCESS"},
						},
						"urn-target-2": {},
						"urn-target-3": {
							asset.Probe{Status: "FAILED"},
						},
					},
					nil,
				)
			},
			Expected: asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				},
				NodeAttrs: map[string]asset.NodeAttributes{
					"urn-source-1": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "SUCCESS"},
						},
						Attributes: map[string]interface{}{},
					},
					"urn-target-2": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{},
						},
						Attributes: map[string]interface{}{},
					},
					"urn-target-3": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "FAILED"},
						},
						Attributes: map[string]interface{}{},
					},
				},
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			actual, err := svc.GetLineage(ctx, "urn-source-1", asset.LineageQuery{})
			if tc.Err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.Err.Error())
			}
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestService_SearchSuggestAssets(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		ErrSearch   error
		ErrSuggest  error
		Setup       func(context.Context, *mocks.DiscoveryRepository)
	}

	DisErr := asset.DiscoveryError{Err: errors.New("could not find")}

	searchResults := []asset.SearchResult{}
	var testCases = []testCase{
		{
			Description: `should return error if the GetGraph function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, dr *mocks.DiscoveryRepository) {
				dr.EXPECT().Search(ctx, asset.SearchConfig{}).Return(searchResults, DisErr)
				dr.EXPECT().Suggest(ctx, asset.SearchConfig{}).Return([]string{}, DisErr)
			},
			ErrSearch:  DisErr,
			ErrSuggest: DisErr,
		},
		{
			Description: `should return no error if search and suggest function work`,
			ID:          assetID,
			Setup: func(ctx context.Context, dr *mocks.DiscoveryRepository) {
				dr.EXPECT().Search(ctx, asset.SearchConfig{}).Return(searchResults, nil)
				dr.EXPECT().Suggest(ctx, asset.SearchConfig{}).Return([]string{}, nil)
			},
			ErrSearch:  nil,
			ErrSuggest: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscoveryRepo)
			}

			svc := asset.NewService(mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			_, err := svc.SearchAssets(ctx, asset.SearchConfig{})
			if err != nil && !assert.Equal(t, tc.ErrSearch, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.ErrSearch)
			}
			_, err = svc.SuggestAssets(ctx, asset.SearchConfig{})
			if err != nil && !assert.Equal(t, tc.ErrSuggest.Error(), err.Error()) {
				t.Fatalf("got error %v, expected error was %v", err, tc.ErrSuggest)
			}
		})
	}
}

func TestService_CreateAssetProbe(t *testing.T) {
	var (
		ctx      = context.Background()
		assetURN = "sample-urn"
		probe    = asset.Probe{
			Status: "RUNNING",
		}
	)

	t.Run("should return no error on success", func(t *testing.T) {
		mockAssetRepo := mocks.NewAssetRepository(t)
		mockAssetRepo.EXPECT().AddProbe(ctx, assetURN, &probe).Return(nil)

		svc := asset.NewService(mockAssetRepo, nil, nil)
		err := svc.AddProbe(ctx, assetURN, &probe)
		assert.NoError(t, err)
	})

	t.Run("should return error on failed", func(t *testing.T) {
		expectedErr := errors.New("test error")

		mockAssetRepo := mocks.NewAssetRepository(t)
		mockAssetRepo.EXPECT().AddProbe(ctx, assetURN, &probe).Return(expectedErr)

		svc := asset.NewService(mockAssetRepo, nil, nil)
		err := svc.AddProbe(ctx, assetURN, &probe)
		assert.Equal(t, expectedErr, err)
	})
}
