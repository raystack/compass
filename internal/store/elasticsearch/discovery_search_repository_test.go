package elasticsearch_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/odpf/compass/core/asset"
	store "github.com/odpf/compass/internal/store/elasticsearch"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchTestData struct {
	Assets []asset.Asset `json:"assets"`
}

func TestSearcherSearch(t *testing.T) {
	ctx := context.TODO()
	t.Run("should return an error if search string is empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)
		_, err = repo.Search(ctx, asset.SearchConfig{
			Text: "",
		})

		assert.Error(t, err)
	})

	t.Run("fixtures", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		err = loadTestFixture(esClient, "./testdata/search-test-fixture.json")
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)

		type expectedRow struct {
			Type    string
			AssetID string
		}
		type searchTest struct {
			Description    string
			Config         asset.SearchConfig
			Expected       []expectedRow
			MatchTotalRows bool
		}
		tests := []searchTest{
			{
				Description: "should fetch assets which has text in any of its fields",
				Config: asset.SearchConfig{
					Text: "topic",
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "order-topic"},
					{Type: "topic", AssetID: "purchase-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
					{Type: "topic", AssetID: "transaction"},
				},
			},
			{
				Description: "should enable fuzzy search",
				Config: asset.SearchConfig{
					Text: "tpic",
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "order-topic"},
					{Type: "topic", AssetID: "purchase-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
					{Type: "topic", AssetID: "transaction"},
				},
			},
			{
				Description: "should put more weight on id fields",
				Config: asset.SearchConfig{
					Text: "invoice",
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "us1-apple-invoice"},
					{Type: "table", AssetID: "au2-microsoft-invoice"},
					{Type: "topic", AssetID: "transaction"},
				},
			},
			{
				Description: "should filter by service if given",
				Config: asset.SearchConfig{
					Text: "invoice",
					Filters: map[string][]string{
						"service": {"rabbitmq", "postgres"},
					},
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "au2-microsoft-invoice"},
					{Type: "topic", AssetID: "transaction"},
				},
			},
			{
				Description: "should match documents based on filter criteria",
				Config: asset.SearchConfig{
					Text: "topic",
					Filters: map[string][]string{
						"data.company": {"odpf"},
					},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "order-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
					{Type: "topic", AssetID: "transaction"},
				},
			},
			{
				Description: "should not return assets without fields specified in filters",
				Config: asset.SearchConfig{
					Text: "invoice topic",
					Filters: map[string][]string{
						"data.country":     {"id"},
						"data.environment": {"production"},
						"data.company":     {"odpf"},
					},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
				},
			},
			{
				Description: "should return 'consumer-topic' if filter owner email with 'john.doe@email.com'",
				Config: asset.SearchConfig{
					Text: "topic",
					Filters: map[string][]string{
						"owners.email": {"john.doe@email.com"},
					},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
				},
			},
			{
				Description: "should return a descendingly sorted based on usage count in search results if rank by usage in the config",
				Config: asset.SearchConfig{
					Text:   "bigquery",
					RankBy: "data.profile.usage_count",
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-common"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-mid"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-1"},
				},
			},
			{
				Description: "should return consumer-topic if search by query description field with text 'rabbitmq' and owners name 'johndoe'",
				Config: asset.SearchConfig{
					Text: "consumer",
					Queries: map[string]string{
						"description":  "rabbitmq",
						"owners.email": "john.doe",
					},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
				},
			},
			{
				Description: "should return 'bigquery::gcpproject/dataset/tablename-common' resource on top if search by query table column name field with text 'tablename-common-column1'",
				Config: asset.SearchConfig{
					Text: "tablename",
					Queries: map[string]string{
						"data.schema.columns.name": "common",
					},
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-common"},
				},
			},
		}
		for _, test := range tests {
			t.Run(test.Description, func(t *testing.T) {
				results, err := repo.Search(ctx, test.Config)
				require.NoError(t, err)

				require.Equal(t, len(test.Expected), len(results))
				for i, res := range test.Expected {
					assert.Equal(t, res.Type, results[i].Type)
					assert.Equal(t, res.AssetID, results[i].ID)
				}
			})
		}
	})
}

func TestSearcherSuggest(t *testing.T) {
	ctx := context.TODO()
	cli, err := esTestServer.NewClient()
	require.NoError(t, err)
	esClient, err := store.NewClient(
		log.NewNoop(),
		store.Config{},
		store.WithClient(cli),
	)
	require.NoError(t, err)

	err = loadTestFixture(esClient, "./testdata/suggest-test-fixture.json")
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient)

	t.Run("fixtures", func(t *testing.T) {
		testCases := []struct {
			term     string
			expected []string
		}{
			{"wallet", []string{"wallet-usage", "wallet/event", "wallet_usage"}},
			{"wallet_usa", []string{"wallet-usage", "wallet_usage"}},
			{"test_t", []string{"test_table"}},
			{"te", []string{"test_table"}},
		}

		for i, tc := range testCases {
			config := asset.SearchConfig{Text: tc.term}
			actual, err := repo.Suggest(ctx, config)
			assert.NoError(t, err)

			assert.Equal(t, tc.expected, actual, "suggestions are not as expected for term: %s and index: %d", tc.term, i)
		}
	})
}

func loadTestFixture(esClient *store.Client, filePath string) (err error) {
	testFixtureJSON, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var data []searchTestData
	err = json.Unmarshal(testFixtureJSON, &data)
	if err != nil {
		return
	}

	ctx := context.TODO()
	for _, testdata := range data {
		repo := store.NewDiscoveryRepository(esClient)
		for _, ast := range testdata.Assets {
			if err := repo.Upsert(ctx, ast); err != nil {
				return err
			}
		}
	}

	return err
}
