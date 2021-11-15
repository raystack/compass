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
		Type:    record.TypeTable,
	}

	t.Run("should return error if factory returns error", func(t *testing.T) {
		records := []record.Record{sampleRecord}

		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", record.TypeTable).Return(new(mock.RecordRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, record.TypeTable, records)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		records := []record.Record{sampleRecord}

		rr := new(mock.RecordRepository)
		rr.On("CreateOrReplaceMany", ctx, records).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", record.TypeTable).Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, record.TypeTable, records)
		assert.Error(t, err)
	})

	t.Run("should return no error on success", func(t *testing.T) {
		records := []record.Record{sampleRecord}

		rr := new(mock.RecordRepository)
		rr.On("CreateOrReplaceMany", ctx, records).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", record.TypeTable).Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.Upsert(ctx, record.TypeTable, records)
		assert.NoError(t, err)
	})
}

func TestServiceDeleteRecord(t *testing.T) {
	ctx := context.TODO()
	recordURN := "sample-urn"

	t.Run("should return error if factory returns error", func(t *testing.T) {
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", record.TypeTable).Return(new(mock.RecordRepository), errors.New("error"))
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, record.TypeTable, recordURN)
		assert.Error(t, err)
	})

	t.Run("should return error if repo returns error", func(t *testing.T) {
		rr := new(mock.RecordRepository)
		rr.On("Delete", ctx, recordURN).Return(errors.New("error"))
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", record.TypeTable).Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, record.TypeTable, recordURN)
		assert.Error(t, err)
	})

	t.Run("should delete record", func(t *testing.T) {
		rr := new(mock.RecordRepository)
		rr.On("Delete", ctx, recordURN).Return(nil)
		defer rr.AssertExpectations(t)
		rrf := new(mock.RecordRepositoryFactory)
		rrf.On("For", record.TypeTable).Return(rr, nil)
		defer rrf.AssertExpectations(t)

		service := discovery.NewService(rrf, nil)
		err := service.DeleteRecord(ctx, record.TypeTable, recordURN)
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
		searcher.On("Search", ctx, cfg).Return([]record.Record{}, errors.New("error"))
		defer searcher.AssertExpectations(t)

		service := discovery.NewService(nil, searcher)
		_, err := service.Search(ctx, cfg)

		assert.Error(t, err)
	})

	t.Run("should return records from searcher", func(t *testing.T) {
		expected := []record.Record{
			{Urn: "record-1"},
			{Urn: "record-2"},
			{Urn: "record-3"},
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
