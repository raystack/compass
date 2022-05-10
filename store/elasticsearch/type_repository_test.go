package elasticsearch_test

import (
	"context"
	"testing"

	"github.com/odpf/compass/asset"
	store "github.com/odpf/compass/store/elasticsearch"
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

			assert.Equal(t, map[asset.Type]int{}, counts)
		})

		t.Run("should return map with 0 count if type has not been populated yet", func(t *testing.T) {
			typ := asset.TypeTable
			cli := esTestServer.NewClient()

			err := store.Migrate(ctx, cli, typ)
			require.NoError(t, err)

			repo := store.NewTypeRepository(cli)
			counts, err := repo.GetAll(ctx)
			require.NoError(t, err)

			expected := map[asset.Type]int{
				asset.TypeTable: 0,
			}
			assert.Equal(t, expected, counts)
		})

		t.Run("should return maps of asset count with valid type as its key", func(t *testing.T) {
			typName := asset.TypeDashboard
			assets := []asset.Asset{
				{URN: "asset-1", Name: "asset-1"},
				{URN: "asset-2", Name: "asset-2"},
				{URN: "asset-3", Name: "asset-3"},
			}

			esClient := esTestServer.NewClient()
			err := store.Migrate(ctx, esClient, asset.TypeDashboard)
			require.NoError(t, err)

			invalidType := "invalid-type"
			err = store.Migrate(ctx, esClient, asset.Type(invalidType))
			require.NoError(t, err)

			repo := store.NewTypeRepository(esClient)
			_, err = repo.GetAll(ctx)
			require.NoError(t, err)

			rrf := store.NewAssetRepositoryFactory(esClient)
			rr, err := rrf.For(typName.String())
			require.NoError(t, err)
			err = rr.CreateOrReplaceMany(ctx, assets)
			require.NoError(t, err)

			rr, err = rrf.For(invalidType)
			require.NoError(t, err)
			err = rr.CreateOrReplaceMany(ctx, assets)
			require.NoError(t, err)

			counts, err := repo.GetAll(ctx)
			require.NoError(t, err)

			expected := map[asset.Type]int{
				asset.TypeDashboard: len(assets),
			}
			assert.Equal(t, expected, counts)
		})
	})
}
