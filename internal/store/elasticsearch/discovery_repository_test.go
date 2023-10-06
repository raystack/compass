package elasticsearch_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/goto/compass/core/asset"
	store "github.com/goto/compass/internal/store/elasticsearch"
	"github.com/goto/salt/log"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryRepositoryUpsert(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
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

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "",
			Type:    asset.TypeTable,
			Service: bigqueryService,
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

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "sample-id",
			Type:    asset.Type("unknown-type"),
			Service: bigqueryService,
		})
		assert.ErrorIs(t, err, asset.ErrUnknownType)
	})

	t.Run("should return error if response.body.Errors is true", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)

		// upsert with create_time as a object
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "sample-id",
			Type:    asset.TypeTable,
			Service: bigqueryService,
			Data: map[string]interface{}{
				"create_time": map[string]interface{}{
					"seconds": 1618103237,
					"nanos":   897000000,
				},
			},
		})
		require.NoError(t, err)

		// upsert with create_time as a string
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "sample-id",
			Type:    asset.TypeTable,
			Service: bigqueryService,
			Data: map[string]interface{}{
				"create_time": "2023-04-10T22:33:57.897Z",
			},
		})
		assert.EqualError(t, err, "discovery error: IndexDoc: doc ID 'sample-id': "+
			"index 'bigquery-test': "+
			"elasticsearch code 'mapper_parsing_exception': "+
			"object mapping for [data.create_time] tried to parse field [create_time] as object, but found a concrete value")
	})

	t.Run("should insert asset to the correct index by its service", func(t *testing.T) {
		ast := asset.Asset{
			ID:          "sample-id",
			URN:         "sample-urn",
			Type:        asset.TypeTable,
			Service:     bigqueryService,
			Name:        "sample-name",
			Description: "sample-description",
			Data: map[string]interface{}{
				"foo": map[string]interface{}{
					"company": "gotocompany",
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

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)
		err = repo.Upsert(ctx, ast)
		assert.NoError(t, err)

		res, err := cli.API.Get(bigqueryService, ast.ID)
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
			Service:     bigqueryService,
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

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)

		err = repo.Upsert(ctx, existingAsset)
		assert.NoError(t, err)
		err = repo.Upsert(ctx, newAsset)
		assert.NoError(t, err)

		res, err := cli.API.Get(bigqueryService, existingAsset.ID)
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
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		kafkaService    = "kafka-test"
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

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)
		err = repo.DeleteByID(ctx, "")
		assert.ErrorIs(t, err, asset.ErrEmptyID)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := asset.Asset{
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

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)

		err = repo.Upsert(ctx, ast)
		require.NoError(t, err)

		err = repo.DeleteByID(ctx, ast.ID)
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

	t.Run("should ignore unavailable indices", func(t *testing.T) {
		ast1 := asset.Asset{
			ID:      "id1",
			Type:    asset.TypeTable,
			Service: bigqueryService,
			URN:     "urn1",
		}
		ast2 := asset.Asset{
			ID:      "id2",
			Type:    asset.TypeTopic,
			Service: kafkaService,
			URN:     "urn2",
		}
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)

		err = repo.Upsert(ctx, ast1)
		require.NoError(t, err)

		err = repo.Upsert(ctx, ast2)
		require.NoError(t, err)

		_, err = cli.Indices.Close([]string{kafkaService})
		require.NoError(t, err)

		err = repo.DeleteByID(ctx, ast1.ID)
		assert.NoError(t, err)
	})
}

func TestDiscoveryRepositoryDeleteByURN(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		kafkaService    = "kafka-test"
	)

	cli, err := esTestServer.NewClient()
	require.NoError(t, err)

	esClient, err := store.NewClient(
		log.NewNoop(), store.Config{}, store.WithClient(cli),
	)
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)

	t.Run("should return error if the given urn is empty", func(t *testing.T) {
		err = repo.DeleteByURN(ctx, "")
		assert.ErrorIs(t, err, asset.ErrEmptyURN)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := asset.Asset{
			ID:      "delete-id",
			Type:    asset.TypeTable,
			Service: bigqueryService,
			URN:     "some-urn",
		}

		err = repo.Upsert(ctx, ast)
		require.NoError(t, err)

		err = repo.DeleteByURN(ctx, ast.URN)
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

	t.Run("should ignore unavailable indices", func(t *testing.T) {
		ast1 := asset.Asset{
			ID:      "id1",
			Type:    asset.TypeTable,
			Service: bigqueryService,
			URN:     "urn1",
		}
		ast2 := asset.Asset{
			ID:      "id2",
			Type:    asset.TypeTopic,
			Service: kafkaService,
			URN:     "urn2",
		}
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10)

		err = repo.Upsert(ctx, ast1)
		require.NoError(t, err)

		err = repo.Upsert(ctx, ast2)
		require.NoError(t, err)

		_, err = cli.Indices.Close([]string{kafkaService})
		require.NoError(t, err)

		err = repo.DeleteByURN(ctx, ast1.URN)
		assert.NoError(t, err)
	})
}
