package elasticsearch_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/namespace"
	"strings"
	"testing"
	"time"

	"github.com/raystack/compass/core/asset"
	store "github.com/raystack/compass/internal/store/elasticsearch"
	"github.com/raystack/salt/log"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryRepositoryUpsert(t *testing.T) {
	var (
		ctx = context.Background()
		ns  = &namespace.Namespace{
			ID:       uuid.New(),
			Name:     "umbrella",
			State:    namespace.DedicatedState,
			Metadata: nil,
		}
		indexAlias = store.BuildAliasNameFromNamespace(ns)
	)

	t.Run("should return error if id empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		err = repo.CreateNamespace(ctx, ns)
		assert.NoError(t, err)

		err = repo.Upsert(ctx, ns, &asset.Asset{
			ID:      "",
			Type:    asset.TypeTable,
			Service: indexAlias,
		})
		assert.ErrorIs(t, err, asset.ErrEmptyID)
	})

	t.Run("should return error if type is not known", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		err = repo.CreateNamespace(ctx, ns)
		assert.NoError(t, err)

		err = repo.Upsert(ctx, ns, &asset.Asset{
			ID:      "sample-id",
			Type:    asset.Type("unknown-type"),
			Service: indexAlias,
		})
		assert.ErrorIs(t, err, asset.ErrUnknownType)
	})

	t.Run("should insert asset to the correct index by its service", func(t *testing.T) {
		ast := &asset.Asset{
			ID:          "sample-id",
			URN:         "sample-urn",
			Type:        asset.TypeTable,
			Service:     indexAlias,
			Name:        "sample-name",
			Description: "sample-description",
			Data: map[string]interface{}{
				"foo": map[string]interface{}{
					"company": "raystack",
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
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		err = repo.CreateNamespace(ctx, ns)
		assert.NoError(t, err)
		err = repo.Upsert(ctx, ns, ast)
		assert.NoError(t, err)

		res, err := cli.API.Get(indexAlias, ast.ID)
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
		existingAsset := &asset.Asset{
			ID:          "existing-id",
			URN:         "existing-urn",
			Type:        asset.TypeTable,
			Service:     indexAlias,
			Name:        "existing-name",
			Description: "existing-description",
		}
		newAsset := existingAsset
		newAsset.URN = "new-urn"
		newAsset.Name = "new-name"
		newAsset.Description = "new-description"

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		err = repo.CreateNamespace(ctx, ns)
		assert.NoError(t, err)

		err = repo.Upsert(ctx, ns, existingAsset)
		assert.NoError(t, err)
		err = repo.Upsert(ctx, ns, newAsset)
		assert.NoError(t, err)

		res, err := cli.API.Get(indexAlias, existingAsset.ID)
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

func TestDiscoveryRepositoryDeleteByID(t *testing.T) {
	nsID, _ := uuid.NewUUID()
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		ns              = &namespace.Namespace{
			ID:       nsID,
			Name:     "umbrella",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)

	t.Run("should return error if id empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		err = repo.CreateNamespace(ctx, ns)
		assert.NoError(t, err)

		err = repo.DeleteByID(ctx, ns, "")
		assert.ErrorIs(t, err, asset.ErrEmptyID)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := &asset.Asset{
			ID:      "delete-id",
			Type:    asset.TypeTable,
			Service: bigqueryService,
			URN:     "some-urn",
		}

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		err = repo.CreateNamespace(ctx, ns)
		assert.NoError(t, err)

		err = repo.Upsert(ctx, ns, ast)
		require.NoError(t, err)

		err = repo.DeleteByID(ctx, ns, ast.ID)
		assert.NoError(t, err)

		res, err := cli.Search(
			cli.Search.WithBody(strings.NewReader(`{"query":{"term":{"_id": "delete-id"}}}`)),
			cli.Search.WithIndex("_all"),
		)
		require.NoError(t, err)
		assert.False(t, res.IsError())

		var body struct {
			Hits struct {
				Total elastic.TotalHits `json:"total"`
			} `json:"hits"`
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

		assert.Equal(t, int64(0), body.Hits.Total.Value)
	})
}

func TestDiscoveryRepositoryDeleteByURN(t *testing.T) {
	nsID, _ := uuid.NewUUID()
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		ns              = &namespace.Namespace{
			ID:       nsID,
			Name:     "umbrella",
			State:    namespace.SharedState,
			Metadata: nil,
		}
	)

	cli, err := esTestServer.NewClient()
	require.NoError(t, err)

	esClient, err := store.NewClient(
		log.NewNoop(), store.Config{}, store.WithClient(cli),
	)
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient)
	err = repo.CreateNamespace(ctx, ns)
	assert.NoError(t, err)

	t.Run("should return error if the given urn is empty", func(t *testing.T) {
		err = repo.DeleteByURN(ctx, ns, "")
		assert.ErrorIs(t, err, asset.ErrEmptyURN)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := &asset.Asset{
			ID:      "delete-id",
			Type:    asset.TypeTable,
			Service: bigqueryService,
			URN:     "some-urn",
		}

		err = repo.Upsert(ctx, ns, ast)
		require.NoError(t, err)

		err = repo.DeleteByURN(ctx, ns, ast.URN)
		assert.NoError(t, err)

		res, err := cli.Search(
			cli.Search.WithBody(strings.NewReader(`{"query":{"term":{"urn.keyword": "some-urn"}}}`)),
			cli.Search.WithIndex("_all"),
		)
		require.NoError(t, err)
		assert.False(t, res.IsError())

		var body struct {
			Hits struct {
				Total elastic.TotalHits `json:"total"`
			} `json:"hits"`
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

		assert.Equal(t, int64(0), body.Hits.Total.Value)
	})
}
