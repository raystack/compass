package asset_test

import (
	"context"
	"errors"
	"testing"

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

func TestService_GetAsset(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		Err         error
		ErrID       error
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	ast := asset.Asset{
		ID: assetID,
	}

	var testCases = []testCase{
		{
			Description: `should return error if the GetAsset functions return error without id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{})
				ar.EXPECT().GetByVersion(ctx, assetID, "v0.0.2").Return(asset.Asset{}, errors.New("error fetching asset"))
				ar.EXPECT().Find(ctx, "some-urn", ast.Type, assetID).Return(ast, errors.New("error fetching asset"))
			},
			ErrID: asset.NotFoundError{},
			Err:   errors.New("error fetching asset"),
		},
		{
			Description: `should return error if the GetAsset functions return error, with id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: ast.ID})
				ar.EXPECT().GetByVersion(ctx, assetID, "v0.0.2").Return(asset.Asset{}, errors.New("error fetching asset"))
				ar.EXPECT().Find(ctx, "some-urn", ast.Type, assetID).Return(ast, errors.New("error fetching asset"))
			},
			ErrID: asset.NotFoundError{AssetID: ast.ID},
			Err:   errors.New("error fetching asset"),
		},
		{
			Description: `should return error if the GetAsset functions return error, with invalid id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.InvalidError{AssetID: ast.ID})
				ar.EXPECT().GetByVersion(ctx, assetID, "v0.0.2").Return(asset.Asset{}, errors.New("error fetching asset"))
				ar.EXPECT().Find(ctx, "some-urn", ast.Type, assetID).Return(ast, errors.New("error fetching asset"))
			},
			ErrID: asset.InvalidError{AssetID: ast.ID},
			Err:   errors.New("error fetching asset"),
		},
		{
			Description: `should return no error if asset is found`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, nil)
				ar.EXPECT().Find(ctx, "some-urn", ast.Type, assetID).Return(ast, nil)
				ar.EXPECT().GetByVersion(ctx, assetID, "v0.0.2").Return(asset.Asset{}, nil)
			},
			ErrID: nil,
			Err:   nil,
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
			_, err := svc.GetAssetByID(ctx, tc.ID)
			if err != nil && !assert.Equal(t, tc.ErrID.Error(), err.Error()) {
				t.Fatalf("got error %v, expected error was %v", err, tc.ErrID)
			}
			_, err = svc.GetAssetByURN(ctx, "some-urn", ast.Type, tc.ID)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
			_, err = svc.GetAssetByVersion(ctx, tc.ID, "v0.0.2")
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
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

			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			mockLineageRepo := new(mocks.LineageRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)
			defer mockLineageRepo.AssertExpectations(t)

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
		Err         error
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	var testCases = []testCase{
		{
			Description: `should return error if the GetGraph function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, asset.LineageNode{}).Return(asset.LineageGraph{}, errors.New("error fetching graph"))
			},
			Err: errors.New("error fetching graph"),
		},
		{
			Description: `should return no error if graph nodes are returned`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, asset.LineageNode{}).Return(asset.LineageGraph{}, nil)
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
			_, err := svc.GetLineage(ctx, asset.LineageNode{})
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
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

			mockAssetRepo := new(mocks.AssetRepository)
			mockDiscoveryRepo := new(mocks.DiscoveryRepository)
			mockLineageRepo := new(mocks.LineageRepository)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscoveryRepo)
			}
			defer mockAssetRepo.AssertExpectations(t)
			defer mockDiscoveryRepo.AssertExpectations(t)
			defer mockLineageRepo.AssertExpectations(t)

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
