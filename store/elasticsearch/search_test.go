package elasticsearch_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	store "github.com/odpf/columbus/store/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchTestData struct {
	Type   asset.Type    `json:"type"`
	Assets []asset.Asset `json:"assets"`
}

func TestSearcherSearch(t *testing.T) {
	ctx := context.TODO()
	t.Run("should return an error if search string is empty", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		searcher, err := store.NewSearcher(store.SearcherConfig{
			Client: esClient,
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err = searcher.Search(ctx, discovery.SearchConfig{
			Text: "",
		})

		assert.Error(t, err)
	})

	t.Run("fixtures", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		err := loadTestFixture(esClient, "./testdata/search-test-fixture.json")
		require.NoError(t, err)

		searcher, err := store.NewSearcher(store.SearcherConfig{
			Client: esClient,
		})
		require.NoError(t, err)

		type expectedRow struct {
			Type    string
			AssetID string
		}
		type searchTest struct {
			Description    string
			Config         discovery.SearchConfig
			Expected       []expectedRow
			MatchTotalRows bool
		}
		tests := []searchTest{
			{
				Description: "should fetch assets which has text in any of its fields",
				Config: discovery.SearchConfig{
					Text: "topic",
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "order-topic"},
					{Type: "topic", AssetID: "purchase-topic"},
					{Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
				},
			},
			{
				Description: "should enable fuzzy search",
				Config: discovery.SearchConfig{
					Text: "tpic",
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "order-topic"},
					{Type: "topic", AssetID: "purchase-topic"},
					{Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
				},
			},
			{
				Description: "should put more weight on id fields",
				Config: discovery.SearchConfig{
					Text: "invoice",
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "au2-microsoft-invoice"},
					{Type: "table", AssetID: "us1-apple-invoice"},
					{Type: "topic", AssetID: "transaction"},
				},
			},
			{
				Description: "should filter by service if given",
				Config: discovery.SearchConfig{
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
				Config: discovery.SearchConfig{
					Text: "topic",
					Filters: map[string][]string{
						"data.company": {"odpf"},
					},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "order-topic"},
					{Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
				},
			},
			{
				Description: "should not return assets without fields specified in filters",
				Config: discovery.SearchConfig{
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
				Config: discovery.SearchConfig{
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
				Config: discovery.SearchConfig{
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
				Config: discovery.SearchConfig{
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
				Config: discovery.SearchConfig{
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
				results, err := searcher.Search(ctx, test.Config)
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
	esClient := esTestServer.NewClient()
	err := loadTestFixture(esClient, "./testdata/suggest-test-fixture.json")
	require.NoError(t, err)

	searcher, err := store.NewSearcher(store.SearcherConfig{
		Client: esClient,
	})
	require.NoError(t, err)

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
			config := discovery.SearchConfig{Text: tc.term}
			actual, err := searcher.Suggest(ctx, config)
			assert.NoError(t, err)

			assert.Equal(t, tc.expected, actual, "suggestions are not as expected for term: %s and index: %d", tc.term, i)
		}
	})
}

func loadTestFixture(esClient *elasticsearch.Client, filePath string) (err error) {
	testFixtureJSON, err := ioutil.ReadFile(filePath)
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
		if err := store.Migrate(ctx, esClient, testdata.Type); err != nil {
			return err
		}
		recordRepo, _ := store.NewRecordRepositoryFactory(esClient).For(testdata.Type.String())
		if err := recordRepo.CreateOrReplaceMany(ctx, testdata.Assets); err != nil {
			return err
		}
	}

	return err
}
