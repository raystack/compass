package discovery_test

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/discovery"
	"github.com/odpf/compass/lib/mocks"
	"github.com/stretchr/testify/assert"
)

func TestServiceUpsert(t *testing.T) {
	ctx := context.TODO()
	sampleAsset := asset.Asset{
		URN:     "sample-urn",
		Service: "bigquery",
	}

	t.Run("should return error if factory returns error", func(t *testing.T) {
		assets := []asset.Asset{sampleAsset}

		rrf := new(mocks.DiscoveryAssetRepositoryFactory)
		rrf.EXPECT().For("table").Return(new(mocks.DiscoveryAssetRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", assets)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		assets := []asset.Asset{sampleAsset}

		rr := new(mocks.DiscoveryAssetRepository)
		rr.EXPECT().CreateOrReplaceMany(ctx, assets).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mocks.DiscoveryAssetRepositoryFactory)
		rrf.EXPECT().For("table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", assets)
		assert.Error(t, err)
	})

	t.Run("should return no error on success", func(t *testing.T) {
		assets := []asset.Asset{sampleAsset}

		rr := new(mocks.DiscoveryAssetRepository)
		rr.EXPECT().CreateOrReplaceMany(ctx, assets).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mocks.DiscoveryAssetRepositoryFactory)
		rrf.EXPECT().For("table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", assets)
		assert.NoError(t, err)
	})
}

func TestServiceDeleteAsset(t *testing.T) {
	ctx := context.TODO()
	assetURN := "sample-urn"

	t.Run("should return error if factory returns error", func(t *testing.T) {
		rrf := new(mocks.DiscoveryAssetRepositoryFactory)
		rrf.EXPECT().For("table").Return(new(mocks.DiscoveryAssetRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteAsset(ctx, "table", assetURN)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		rr := new(mocks.AssetRepository)
		rr.EXPECT().Delete(ctx, assetURN).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mocks.DiscoveryAssetRepositoryFactory)
		rrf.EXPECT().For("table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteAsset(ctx, "table", assetURN)
		assert.Error(t, err)
	})

	t.Run("should delete asset", func(t *testing.T) {
		rr := new(mocks.AssetRepository)
		rr.EXPECT().Delete(ctx, assetURN).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mocks.DiscoveryAssetRepositoryFactory)
		rrf.EXPECT().For("table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteAsset(ctx, "table", assetURN)
		assert.NoError(t, err)
	})
}

func TestServiceSearch(t *testing.T) {
	ctx := context.TODO()
	cfg := discovery.SearchConfig{
		Text: "test",
		Filters: map[string][]string{
			"foo": {"bar"},
		},
	}
	t.Run("should return error if searcher fails", func(t *testing.T) {
		searcher := new(mocks.DiscoveryAssetSearcher)
		searcher.EXPECT().Search(ctx, cfg).Return([]discovery.SearchResult{}, errors.New("error"))
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		_, err := service.Search(ctx, cfg)

		assert.Error(t, err)
	})

	t.Run("should return assets from searcher", func(t *testing.T) {
		expected := []discovery.SearchResult{
			{ID: "asset-1"},
			{ID: "asset-2"},
			{ID: "asset-3"},
		}
		searcher := new(mocks.DiscoveryAssetSearcher)
		searcher.EXPECT().Search(ctx, cfg).Return(expected, nil)
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		actual, err := service.Search(ctx, cfg)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestServiceSuggest(t *testing.T) {
	ctx := context.TODO()
	cfg := discovery.SearchConfig{
		Text: "test",
		Filters: map[string][]string{
			"foo": {"bar"},
		},
	}
	t.Run("should return error if searcher fails", func(t *testing.T) {
		searcher := new(mocks.DiscoveryAssetSearcher)
		searcher.EXPECT().Suggest(ctx, cfg).Return([]string{}, errors.New("error"))
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		_, err := service.Suggest(ctx, cfg)

		assert.Error(t, err)
	})

	t.Run("should return assets from searcher", func(t *testing.T) {
		expected := []string{
			"asset-1",
			"asset-2",
			"asset-3",
		}
		searcher := new(mocks.DiscoveryAssetSearcher)
		searcher.EXPECT().Suggest(ctx, cfg).Return(expected, nil)
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		actual, err := service.Suggest(ctx, cfg)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
