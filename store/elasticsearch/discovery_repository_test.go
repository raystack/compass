package elasticsearch_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/odpf/compass/asset"
	store "github.com/odpf/compass/store/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryRepositoryUpsert(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error if id empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)
		err = repo.Upsert(ctx, asset.Asset{
			ID:   "",
			Type: asset.TypeTable,
		})
		assert.ErrorIs(t, err, asset.ErrEmptyID)
	})

	t.Run("should return error if type is not known", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)
		err = repo.Upsert(ctx, asset.Asset{
			ID:   "sample-id",
			Type: asset.Type("unknown-type"),
		})
		assert.ErrorIs(t, err, asset.ErrUnknownType)
	})

	t.Run("should insert asset to the correct index by its type", func(t *testing.T) {
		ast := asset.Asset{
			ID:          "sample-id",
			URN:         "sample-urn",
			Type:        asset.TypeTable,
			Service:     "bigquery",
			Name:        "sample-name",
			Description: "sample-description",
			Data: map[string]interface{}{
				"foo": map[string]interface{}{
					"company": "odpf",
				},
			},
			Labels: map[string]string{
				"bar": "foo",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)
		err = repo.Upsert(ctx, ast)
		assert.NoError(t, err)

		res, err := cli.API.Get("table", ast.ID)
		require.NoError(t, err)
		require.False(t, res.IsError())

		var payload struct {
			Source asset.Asset `json:"_source"`
		}
		err = json.NewDecoder(res.Body).Decode(&payload)
		require.NoError(t, err)

		assert.Equal(t, ast.ID, payload.Source.ID)
		assert.Equal(t, ast.URN, payload.Source.URN)
		assert.Equal(t, ast.Type, payload.Source.Type)
		assert.Equal(t, ast.Service, payload.Source.Service)
		assert.Equal(t, ast.Name, payload.Source.Name)
		assert.Equal(t, ast.Description, payload.Source.Description)
		assert.Equal(t, ast.Data, payload.Source.Data)
		assert.Equal(t, ast.Labels, payload.Source.Labels)
		assert.WithinDuration(t, ast.CreatedAt, payload.Source.CreatedAt, 0)
		assert.WithinDuration(t, ast.UpdatedAt, payload.Source.UpdatedAt, 0)
	})

	t.Run("should update existing asset if ID exists", func(t *testing.T) {
		existingAsset := asset.Asset{
			ID:          "existing-id",
			URN:         "existing-urn",
			Type:        asset.TypeTable,
			Service:     "bigquery",
			Name:        "existing-name",
			Description: "existing-description",
		}
		newAsset := existingAsset
		newAsset.URN = "new-urn"
		newAsset.Name = "new-name"
		newAsset.Description = "new-description"

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)

		err = repo.Upsert(ctx, existingAsset)
		assert.NoError(t, err)
		err = repo.Upsert(ctx, newAsset)
		assert.NoError(t, err)

		res, err := cli.API.Get("table", existingAsset.ID)
		require.NoError(t, err)
		require.False(t, res.IsError())

		var payload struct {
			Source asset.Asset `json:"_source"`
		}
		err = json.NewDecoder(res.Body).Decode(&payload)
		require.NoError(t, err)

		assert.Equal(t, existingAsset.ID, payload.Source.ID)
		assert.Equal(t, newAsset.URN, payload.Source.URN)
		assert.Equal(t, newAsset.Name, payload.Source.Name)
		assert.Equal(t, newAsset.Description, payload.Source.Description)
	})
}

func TestDiscoveryRepositoryDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error if id empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)
		err = repo.Delete(ctx, "")
		assert.ErrorIs(t, err, asset.ErrEmptyID)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := asset.Asset{
			ID:   "delete-id",
			Type: asset.TypeTable,
			URN:  "some-urn",
		}

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)

		err = repo.Upsert(ctx, ast)
		require.NoError(t, err)

		err = repo.Delete(ctx, ast.ID)
		assert.NoError(t, err)
	})
}

func TestDiscoveryRepositoryGetTypes(t *testing.T) {
	ctx := context.Background()

	t.Run("should return empty map if no type is available", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)
		counts, err := repo.GetTypes(ctx)
		require.NoError(t, err)

		assert.Equal(t, map[asset.Type]int{}, counts)
	})

	t.Run("should return map with 0 count if type has not been populated yet", func(t *testing.T) {
		typ := asset.TypeTable
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		err = store.Migrate(ctx, cli, typ)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(cli)
		counts, err := repo.GetTypes(ctx)
		require.NoError(t, err)

		expected := map[asset.Type]int{
			asset.TypeTable: 0,
		}
		assert.Equal(t, expected, counts)
	})

	t.Run("should return maps of asset count with valid type as its key", func(t *testing.T) {
		typ := asset.TypeDashboard
		assets := []asset.Asset{
			{ID: "id-asset-1", URN: "asset-1", Name: "asset-1", Type: typ},
			{ID: "id-asset-2", URN: "asset-2", Name: "asset-2", Type: typ},
			{ID: "id-asset-3", URN: "asset-3", Name: "asset-3", Type: typ},
		}

		esClient, err := esTestServer.NewClient()
		require.NoError(t, err)

		err = store.Migrate(ctx, esClient, asset.TypeDashboard)
		require.NoError(t, err)

		invalidType := "invalid-type"
		err = store.Migrate(ctx, esClient, asset.Type(invalidType))
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		_, err = repo.GetTypes(ctx)
		require.NoError(t, err)

		for _, ast := range assets {
			err = repo.Upsert(ctx, ast)
			require.NoError(t, err)
		}

		counts, err := repo.GetTypes(ctx)
		require.NoError(t, err)

		expected := map[asset.Type]int{
			asset.TypeDashboard: len(assets),
		}
		assert.Equal(t, expected, counts)
	})
}
