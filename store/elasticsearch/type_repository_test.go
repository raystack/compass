package elasticsearch_test

import (
	"context"
	"testing"

	"github.com/odpf/columbus/record"
	store "github.com/odpf/columbus/store/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("GetAll", func(t *testing.T) {
		t.Run("should return empty map if no type is available", func(t *testing.T) {
			repo := store.NewTypeRepository(esTestServer.NewClient())
			counts, err := repo.GetAll(ctx)
			require.NoError(t, err)

			assert.Equal(t, map[record.TypeName]int{}, counts)
		})

		t.Run("should return map with 0 count if type has not been populated yet", func(t *testing.T) {
			typ := record.TypeNameTable
			cli := esTestServer.NewClient()

			err := store.Migrate(ctx, cli, typ)
			require.NoError(t, err)

			repo := store.NewTypeRepository(cli)
			counts, err := repo.GetAll(ctx)
			require.NoError(t, err)

			expected := map[record.TypeName]int{
				record.TypeNameTable: 0,
			}
			assert.Equal(t, expected, counts)
		})

		t.Run("should return maps of record count with valid type as its key", func(t *testing.T) {
			typName := record.TypeNameDashboard
			records := []record.Record{
				{Urn: "record-1", Name: "record-1"},
				{Urn: "record-2", Name: "record-2"},
				{Urn: "record-3", Name: "record-3"},
			}

			esClient := esTestServer.NewClient()
			err := store.Migrate(ctx, esClient, record.TypeNameDashboard)
			require.NoError(t, err)

			invalidTypeName := "invalid-type"
			err = store.Migrate(ctx, esClient, record.TypeName(invalidTypeName))
			require.NoError(t, err)

			repo := store.NewTypeRepository(esClient)
			_, err = repo.GetAll(ctx)
			require.NoError(t, err)

			rrf := store.NewRecordRepositoryFactory(esClient)
			rr, err := rrf.For(typName.String())
			require.NoError(t, err)
			err = rr.CreateOrReplaceMany(ctx, records)
			require.NoError(t, err)

			rr, err = rrf.For(invalidTypeName)
			require.NoError(t, err)
			err = rr.CreateOrReplaceMany(ctx, records)
			require.NoError(t, err)

			counts, err := repo.GetAll(ctx)
			require.NoError(t, err)

			expected := map[record.TypeName]int{
				record.TypeNameDashboard: len(records),
			}
			assert.Equal(t, expected, counts)
		})
	})
}
