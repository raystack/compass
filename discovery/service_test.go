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
	sampleRecord := asset.Asset{
		URN:     "sample-urn",
		Service: "bigquery",
	}

	t.Run("should return error if factory returns error", func(t *testing.T) {
		assets := []asset.Asset{sampleRecord}

		rrf := new(mocks.RecordRepositoryFactory)
		rrf.On("For", "table").Return(new(mocks.RecordRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", assets)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		assets := []asset.Asset{sampleRecord}

		rr := new(mocks.RecordRepository)
		rr.On("CreateOrReplaceMany", ctx, assets).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mocks.RecordRepositoryFactory)
		rrf.On("For", "table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", assets)
		assert.Error(t, err)
	})

	t.Run("should return no error on success", func(t *testing.T) {
		assets := []asset.Asset{sampleRecord}

		rr := new(mocks.RecordRepository)
		rr.On("CreateOrReplaceMany", ctx, assets).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mocks.RecordRepositoryFactory)
		rrf.On("For", "table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", assets)
		assert.NoError(t, err)
	})
}

func TestServiceDeleteRecord(t *testing.T) {
	ctx := context.TODO()
	recordURN := "sample-urn"

	t.Run("should return error if factory returns error", func(t *testing.T) {
		rrf := new(mocks.RecordRepositoryFactory)
		rrf.On("For", "table").Return(new(mocks.RecordRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, "table", recordURN)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		rr := new(mocks.RecordRepository)
		rr.On("Delete", ctx, recordURN).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mocks.RecordRepositoryFactory)
		rrf.On("For", "table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, "table", recordURN)
		assert.Error(t, err)
	})

	t.Run("should delete record", func(t *testing.T) {
		rr := new(mocks.RecordRepository)
		rr.On("Delete", ctx, recordURN).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mocks.RecordRepositoryFactory)
		rrf.On("For", "table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, "table", recordURN)
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
		searcher := new(mocks.RecordSearcher)
		searcher.On("Search", ctx, cfg).Return([]discovery.SearchResult{}, errors.New("error"))
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		_, err := service.Search(ctx, cfg)

		assert.Error(t, err)
	})

	t.Run("should return assets from searcher", func(t *testing.T) {
		expected := []discovery.SearchResult{
			{ID: "record-1"},
			{ID: "record-2"},
			{ID: "record-3"},
		}
		searcher := new(mocks.RecordSearcher)
		searcher.On("Search", ctx, cfg).Return(expected, nil)
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
		searcher := new(mocks.RecordSearcher)
		searcher.On("Suggest", ctx, cfg).Return([]string{}, errors.New("error"))
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		_, err := service.Suggest(ctx, cfg)

		assert.Error(t, err)
	})

	t.Run("should return assets from searcher", func(t *testing.T) {
		expected := []string{
			"record-1",
			"record-2",
			"record-3",
		}
		searcher := new(mocks.RecordSearcher)
		searcher.On("Suggest", ctx, cfg).Return(expected, nil)
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		actual, err := service.Suggest(ctx, cfg)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
