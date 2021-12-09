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

func TestSearch(t *testing.T) {
	ctx := context.Background()
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

		testFixture, err := loadTestFixture()
		if err != nil {
			t.Error(err)
		}
		err = populateSearchData(esClient, testFixture)
		if err != nil {
			t.Error(err)
		}
		searcher, err := store.NewSearcher(store.SearcherConfig{
			Client: esClient,
		})
		if err != nil {
			t.Error(err)
		}

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

func TestSearchWithUsageBoosting(t *testing.T) {
	ctx := context.Background()
	t.Run("should return a descendingly sorted based on usage count in search results if sortby usage in the config", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		testFixture, err := loadTestFixture()
		if err != nil {
			t.Error(err)
		}
		err = populateSearchData(esClient, testFixture)
		if err != nil {
			t.Error(err)
		}
		searcher, err := store.NewSearcher(store.SearcherConfig{
			Client: esClient,
		})
		if err != nil {
			t.Error(err)
			return
		}
		searchResults, err := searcher.Search(ctx, discovery.SearchConfig{
			Text:   "bigquery",
			SortBy: "data.profile.usage_count",
		})
		expectedOrder := []string{"bigquery::gcpproject/dataset/tablename-common", "bigquery::gcpproject/dataset/tablename-mid", "bigquery::gcpproject/dataset/tablename-1"}

		resultsOrder := []string{}
		for _, r := range searchResults {
			resultsOrder = append(resultsOrder, r.ID)
		}

		assert.Nil(t, err)
		assert.EqualValues(t, expectedOrder, resultsOrder[:3])
	})
}

func loadTestFixture() (testFixture []searchTestData, err error) {
	testFixtureJSON, err := ioutil.ReadFile("./testdata/search-test-fixture.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(testFixtureJSON, &testFixture)
	if err != nil {
		return
	}

	return testFixture, err
}

func populateSearchData(esClient *elasticsearch.Client, data []searchTestData) error {
	for _, sample := range data {
		recordRepo, _ := store.NewRecordRepositoryFactory(esClient).For(sample.Type)
		if err := recordRepo.CreateOrReplaceMany(context.Background(), sample.Records); err != nil {
			return err
		}
	}

	return nil
}
