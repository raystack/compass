package elasticsearch_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/record"
	store "github.com/odpf/columbus/store/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchTestData struct {
	Type    string          `json:"type"`
	Records []record.Record `json:"records"`
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
			Type     string `json:"type"`
			RecordID string `json:"record_id"`
		}
		type searchTest struct {
			Description    string
			Config         discovery.SearchConfig
			Expected       []expectedRow
			MatchTotalRows bool
		}
		tests := []searchTest{
			{
				Description: "should fetch records which has text in any of its fields",
				Config: discovery.SearchConfig{
					Text: "topic",
				},
				Expected: []expectedRow{
					{Type: "topic", RecordID: "order-topic"},
					{Type: "topic", RecordID: "purchase-topic"},
					{Type: "topic", RecordID: "consumer-topic"},
					{Type: "topic", RecordID: "consumer-mq-2"},
				},
			},
			{
				Description: "should enable fuzzy search",
				Config: discovery.SearchConfig{
					Text: "tpic",
				},
				Expected: []expectedRow{
					{Type: "topic", RecordID: "order-topic"},
					{Type: "topic", RecordID: "purchase-topic"},
					{Type: "topic", RecordID: "consumer-topic"},
					{Type: "topic", RecordID: "consumer-mq-2"},
				},
			},
			{
				Description: "should put more weight on id fields",
				Config: discovery.SearchConfig{
					Text: "invoice",
				},
				Expected: []expectedRow{
					{Type: "table", RecordID: "au2-microsoft-invoice"},
					{Type: "table", RecordID: "us1-apple-invoice"},
					{Type: "topic", RecordID: "transaction"},
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
					{Type: "table", RecordID: "au2-microsoft-invoice"},
					{Type: "topic", RecordID: "transaction"},
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
					{Type: "topic", RecordID: "order-topic"},
					{Type: "topic", RecordID: "consumer-topic"},
					{Type: "topic", RecordID: "consumer-mq-2"},
				},
			},
			{
				Description: "should not return records without fields specified in filters",
				Config: discovery.SearchConfig{
					Text: "invoice topic",
					Filters: map[string][]string{
						"data.country":     {"id"},
						"data.environment": {"production"},
						"data.company":     {"odpf"},
					},
				},
				Expected: []expectedRow{
					{Type: "topic", RecordID: "consumer-topic"},
					{Type: "topic", RecordID: "consumer-mq-2"},
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
					{Type: "topic", RecordID: "consumer-topic"},
				},
			},
			{
				Description: "should return a descendingly sorted based on usage count in search results if rank by usage in the config",
				Config: discovery.SearchConfig{
					Text:   "bigquery",
					RankBy: "data.profile.usage_count",
				},
				Expected: []expectedRow{
					{Type: "table", RecordID: "bigquery::gcpproject/dataset/tablename-common"},
					{Type: "table", RecordID: "bigquery::gcpproject/dataset/tablename-mid"},
					{Type: "table", RecordID: "bigquery::gcpproject/dataset/tablename-1"},
				},
			},
			{
				Description: "should return consumer-topic if search by query description field with text 'rabbitmq' and owners name 'johndoe'",
				Config: discovery.SearchConfig{
					Text: "consumer",
					Queries: map[string]string{
						"description": "rabbitmq",
						"owners.name": "john doe",
					},
				},
				Expected: []expectedRow{
					{Type: "topic", RecordID: "consumer-topic"},
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
					{Type: "table", RecordID: "bigquery::gcpproject/dataset/tablename-common"},
				},
			},
		}
		for _, test := range tests {
			t.Run(test.Description, func(t *testing.T) {
				results, err := searcher.Search(ctx, test.Config)
				require.NoError(t, err)

				assert.Equal(t, len(test.Expected), len(results))
				for i, res := range test.Expected {
					assert.Equal(t, res.Type, results[i].Type)
					assert.Equal(t, res.RecordID, results[i].ID)
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

	typeRepo := store.NewTypeRepository(esClient)

	ctx := context.TODO()
	for _, testdata := range data {
		if err := typeRepo.CreateOrReplace(ctx, record.Type{Name: testdata.Type}); err != nil {
			return err
		}
		recordRepo, _ := store.NewRecordRepositoryFactory(esClient).For(testdata.Type)
		if err := recordRepo.CreateOrReplaceMany(ctx, testdata.Records); err != nil {
			return err
		}
	}

	return err
}
