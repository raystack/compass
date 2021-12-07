package discovery_test

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/record"
	"github.com/stretchr/testify/assert"
)

func TestServiceUpsert(t *testing.T) {
	ctx := context.TODO()
	sampleRecord := record.Record{
		Urn:     "sample-urn",
		Service: "bigquery",
	}

	t.Run("should return error if factory returns error", func(t *testing.T) {
		records := []record.Record{sampleRecord}

		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", "table").Return(new(mock.RecordRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", records)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		records := []record.Record{sampleRecord}

		rr := new(mock.RecordRepository)
		rr.On("CreateOrReplaceMany", ctx, records).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", "table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", records)
		assert.Error(t, err)
	})

	t.Run("should return no error on success", func(t *testing.T) {
		records := []record.Record{sampleRecord}

		rr := new(mock.RecordRepository)
		rr.On("CreateOrReplaceMany", ctx, records).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", "table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, "table", records)
		assert.NoError(t, err)
	})
}

func TestServiceDeleteRecord(t *testing.T) {
	ctx := context.TODO()
	recordURN := "sample-urn"

	t.Run("should return error if factory returns error", func(t *testing.T) {
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", "table").Return(new(mock.RecordRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, "table", recordURN)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		rr := new(mock.RecordRepository)
		rr.On("Delete", ctx, recordURN).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", "table").Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, "table", recordURN)
		assert.Error(t, err)
	})

	t.Run("should delete record", func(t *testing.T) {
		rr := new(mock.RecordRepository)
		rr.On("Delete", ctx, recordURN).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
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
		searcher := new(mock.RecordSearcher)
		searcher.On("Search", ctx, cfg).Return([]discovery.SearchResult{}, errors.New("error"))
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		_, err := service.Search(ctx, cfg)

		assert.Error(t, err)
	})

	t.Run("should return records from searcher", func(t *testing.T) {
		expected := []discovery.SearchResult{
			{ID: "record-1"},
			{ID: "record-2"},
			{ID: "record-3"},
		}
		searcher := new(mock.RecordSearcher)
		searcher.On("Search", ctx, cfg).Return(expected, nil)
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		actual, err := service.Search(ctx, cfg)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
