package elasticsearch_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/discovery"
	store "github.com/odpf/columbus/discovery/elasticsearch"
	"github.com/odpf/columbus/record"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchTestData struct {
	Type    record.Type     `json:"type"`
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
			Type     record.Type `json:"type"`
			RecordID string      `json:"record_id"`
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
					{Type: record.TypeTopic, RecordID: "order-topic"},
					{Type: record.TypeTopic, RecordID: "purchase-topic"},
					{Type: record.TypeTopic, RecordID: "consumer-topic"},
				},
			},
			{
				Description: "should enable fuzzy search",
				Config: discovery.SearchConfig{
					Text: "tpic",
				},
				Expected: []expectedRow{
					{Type: record.TypeTopic, RecordID: "order-topic"},
					{Type: record.TypeTopic, RecordID: "purchase-topic"},
					{Type: record.TypeTopic, RecordID: "consumer-topic"},
				},
			},
			{
				Description: "should put more weight on id fields",
				Config: discovery.SearchConfig{
					Text: "invoice",
				},
				Expected: []expectedRow{
					{Type: record.TypeTable, RecordID: "au2-microsoft-invoice"},
					{Type: record.TypeTable, RecordID: "us1-apple-invoice"},
					{Type: record.TypeTopic, RecordID: "transaction"},
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
					{Type: record.TypeTable, RecordID: "au2-microsoft-invoice"},
					{Type: record.TypeTopic, RecordID: "transaction"},
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
					{Type: record.TypeTopic, RecordID: "order-topic"},
					{Type: record.TypeTopic, RecordID: "consumer-topic"},
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
					{Type: record.TypeTopic, RecordID: "consumer-topic"},
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
					assert.Equal(t, res.RecordID, results[i].Urn)
				}
			})
		}
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
