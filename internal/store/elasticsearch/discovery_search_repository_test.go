package elasticsearch_test

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/goto/compass/core/asset"
	store "github.com/goto/compass/internal/store/elasticsearch"
	"github.com/goto/salt/log"
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

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop())
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

		err = loadTestFixture(cli, esClient, "./testdata/search-test-fixture.json")
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop())

		type expectedRow struct {
			Type    string
			AssetID string
			Data    map[string]interface{}
			Service string
		}
		type searchTest struct {
			Description    string
			Config         asset.SearchConfig
			Expected       []expectedRow
			MatchTotalRows bool
		}

		tests := []searchTest{
			{
				Description: "should fetch assets with fields mentioned in included fields",
				Config: asset.SearchConfig{
					Text:          "topic",
					IncludeFields: []string{"id", "data.company", "type"},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic", Data: map[string]interface{}{"company": "gotocompany"}},
					{Type: "topic", AssetID: "order-topic", Data: map[string]interface{}{"company": "gotocompany"}},
					{Type: "topic", AssetID: "purchase-topic", Data: map[string]interface{}{"company": "microsoft"}},
					{Type: "topic", AssetID: "consumer-mq-2", Data: map[string]interface{}{"company": "gotocompany"}},
					{Type: "topic", AssetID: "transaction", Data: map[string]interface{}{"company": "gotocompany"}},
				},
			},
			{
				Description: "should fetch assets with default fields if included fields is empty",
				Config: asset.SearchConfig{
					Text: "topic",
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic", Service: "rabbitmq", Data: map[string]interface{}{"company": "gotocompany", "country": "id", "description": "Update on every rabbitmq customer creation/update", "environment": "production", "partition": float64(50), "topic_name": "consumer-topic"}},
					{Type: "topic", AssetID: "order-topic", Service: "kafka", Data: map[string]interface{}{"company": "gotocompany", "country": "us", "description": "Topic for each submitted order", "environment": "integration", "malformed": "2023-05-23T04:35:42Z", "partition": float64(250), "topic_name": "order-topic"}},
					{Type: "topic", AssetID: "purchase-topic", Service: "kafka", Data: map[string]interface{}{"company": "microsoft", "country": "id", "description": "Topic for each submitted purchase", "environment": "integration", "malformed": "", "partition": float64(100), "topic_name": "purchase-topic"}},
					{Type: "topic", AssetID: "consumer-mq-2", Service: "rabbitmq", Data: map[string]interface{}{"company": "gotocompany", "country": "id", "description": "Another rabbitmq topic", "environment": "production", "partition": float64(50), "topic_name": "consumer-mq-2"}},
					{Type: "topic", AssetID: "transaction", Service: "rabbitmq", Data: map[string]interface{}{"company": "gotocompany", "description": "This publishes all the invoices from each of invoice storage where the invoice will be filtered and checked using invoice filterer and invoice checker", "environment": "production", "partition": float64(1), "topic_name": "transaction"}},
				},
			},
			{
				Description: "should fetch assets which has text in any of its fields",
				Config: asset.SearchConfig{
					Text:          "topic",
					IncludeFields: []string{"type", "id"},
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
					Text:          "tpic",
					IncludeFields: []string{"type", "id"},
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
				Description: "should disable fuzzy search",
				Config: asset.SearchConfig{
					Text:          "tpic",
					Flags:         asset.SearchFlags{DisableFuzzy: true},
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{},
			},
			{
				Description: "should put more weight on id fields",
				Config: asset.SearchConfig{
					Text:          "invoice",
					IncludeFields: []string{"type", "id"},
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
					IncludeFields: []string{"type", "id"},
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
						"data.company": {"gotocompany"},
					},
					IncludeFields: []string{"type", "id"},
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
						"data.company":     {"gotocompany"},
					},
					IncludeFields: []string{"type", "id"},
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
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
				},
			},
			{
				Description: "should return a descendingly sorted based on usage count in search results if rank by usage in the config",
				Config: asset.SearchConfig{
					Text:          "bigquery",
					RankBy:        "data.profile.usage_count",
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-common"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-mid"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/test"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-1"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/abc-tablename-mid"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-abc-common-test"},
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
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{
					{Type: "topic", AssetID: "consumer-topic"},
				},
			},
			{
				Description: "should return 5 records with offset of 0",
				Config: asset.SearchConfig{
					Text:          "topic",
					Offset:        0,
					MaxResults:    5,
					IncludeFields: []string{"type", "id"},
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
				Description: "should return 4 records with offset of 1",
				Config: asset.SearchConfig{
					Text:          "topic",
					Offset:        1,
					MaxResults:    5,
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{
					// {Type: "topic", AssetID: "consumer-topic"},
					{Type: "topic", AssetID: "order-topic"},
					{Type: "topic", AssetID: "purchase-topic"},
					{Type: "topic", AssetID: "consumer-mq-2"},
					{Type: "topic", AssetID: "transaction"},
				},
			},
			{
				Description: "should return 'bigquery::gcpproject/dataset/tablename-common' resource on top if search by query table column name field with text 'tablename-common-column1'",
				Config: asset.SearchConfig{
					Text: "tablename",
					Queries: map[string]string{
						"data.schema.columns.name": "common",
					},
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-common"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-abc-common-test"},
				},
			},
			{
				Description: "should return 'bigquery::gcpproject/dataset/tablename-abc-common-test' resource on top if searched for text 'tablename-abc-common-test'",
				Config: asset.SearchConfig{
					Text:          "tablename-abc-common-test",
					RankBy:        "data.profile.usage_count",
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-abc-common-test"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-common"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/test"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/abc-tablename-mid"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-mid"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-1"},
				},
			},
			{
				Description: "should return highlighted text in resource if searched highlight text is enabled.",
				Config: asset.SearchConfig{
					Text:   "order",
					RankBy: "data.profile.usage_count",
					Flags: asset.SearchFlags{
						EnableHighlight: true,
					},
					IncludeFields: []string{"type", "id"},
				},

				Expected: []expectedRow{
					{
						Type:    "topic",
						AssetID: "order-topic",
						Data: map[string]interface{}{
							"_highlight": map[string]interface{}{
								"urn":              []interface{}{"<em>order</em>-topic"},
								"data.topic_name":  []interface{}{"<em>order</em>-topic"},
								"name":             []interface{}{"<em>order</em>-topic"},
								"description":      []interface{}{"Topic for each submitted <em>order</em>"},
								"id":               []interface{}{"<em>order</em>-topic"},
								"data.description": []interface{}{"Topic for each submitted <em>order</em>"},
							},
						},
					},
				},
			},
			{
				Description: "should return 'bigquery::gcpproject/dataset/tablename-abc-common-test' resource on top if " +
					"searched for text 'abc-test' as it has both the keywords searched",
				Config: asset.SearchConfig{
					Text:          "abc-test",
					RankBy:        "data.profile.usage_count",
					IncludeFields: []string{"type", "id"},
				},
				Expected: []expectedRow{
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/tablename-abc-common-test"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/test"},
					{Type: "table", AssetID: "bigquery::gcpproject/dataset/abc-tablename-mid"},
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
					assert.Equal(t, res.Data, results[i].Data)
					assert.Equal(t, res.Service, results[i].Service)
					if test.Config.Flags.EnableHighlight {
						assert.Equal(t, res.Data["_highlight"], results[i].Data["_highlight"])
					}
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

	err = loadTestFixture(cli, esClient, "./testdata/suggest-test-fixture.json")
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient, log.NewNoop())

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

func loadTestFixture(cli *elasticsearch.Client, esClient *store.Client, filePath string) error {
	testFixtureJSON, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var data []searchTestData
	err = json.Unmarshal(testFixtureJSON, &data)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	for _, testdata := range data {
		repo := store.NewDiscoveryRepository(esClient, log.NewNoop())
		for _, ast := range testdata.Assets {
			if err := repo.Upsert(ctx, ast); err != nil {
				return err
			}
		}
	}

	_, err = cli.Indices.Refresh(
		cli.Indices.Refresh.WithIgnoreUnavailable(true),
		cli.Indices.Refresh.WithIndex("universe"),
	)

	return err
}

func TestGroupAssets(t *testing.T) {
	ctx := context.TODO()
	t.Run("should return an error if group string array is empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop())
		_, err = repo.GroupAssets(ctx, asset.GroupConfig{
			GroupBy: []string{""},
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

		err = loadTestFixture(cli, esClient, "./testdata/search-test-fixture.json")
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop())

		type groupTest struct {
			Description string
			Config      asset.GroupConfig
			Expected    []asset.GroupResult
		}
		tests := []groupTest{
			{
				Description: "should group assets which match group by multiple fields",

				Config: asset.GroupConfig{
					GroupBy: []string{"type", "name"},
					Size:    15,
				},
				Expected: []asset.GroupResult{
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "abc-tablename-mid"},
						},
						Assets: []asset.Asset{{Name: "abc-tablename-mid"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "apple-invoice"},
						},
						Assets: []asset.Asset{{Name: "apple-invoice"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "microsoft-invoice"},
						},
						Assets: []asset.Asset{{Name: "microsoft-invoice"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "tablename-1"},
						},
						Assets: []asset.Asset{{Name: "tablename-1"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "tablename-abc-common-test"},
						},
						Assets: []asset.Asset{{Name: "tablename-abc-common-test"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "tablename-common"},
						},
						Assets: []asset.Asset{{Name: "tablename-common"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "tablename-mid"},
						},
						Assets: []asset.Asset{{Name: "tablename-mid"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
							{Name: "name", Value: "test"},
						},
						Assets: []asset.Asset{{Name: "test"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "topic"},
							{Name: "name", Value: "consumer-mq-2"},
						},
						Assets: []asset.Asset{{Name: "consumer-mq-2"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "name", Value: "consumer-topic"},
							{Name: "type", Value: "topic"},
						},
						Assets: []asset.Asset{{Name: "consumer-topic"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "topic"},
							{Name: "name", Value: "order-topic"},
						},
						Assets: []asset.Asset{{Name: "order-topic"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "topic"},
							{Name: "name", Value: "purchase-topic"},
						},
						Assets: []asset.Asset{{Name: "purchase-topic"}},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "topic"},
							{Name: "name", Value: "transaction"},
						},
						Assets: []asset.Asset{{Name: "transaction"}},
					},
				},
			},
			{
				Description: "should group assets which match group by fields",

				Config: asset.GroupConfig{
					GroupBy:        []string{"type"},
					IncludedFields: []string{"name"},
				},
				Expected: []asset.GroupResult{
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "table"},
						},
						Assets: []asset.Asset{
							{Name: "tablename-1"},
							{Name: "tablename-common"},
							{Name: "tablename-abc-common-test"},
						},
					},
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "topic"},
						},
						Assets: []asset.Asset{
							{Name: "order-topic"},
							{Name: "purchase-topic"},
							{Name: "consumer-topic"},
						},
					},
				},
			},
			{
				Description: "should not return assets without fields specified in filters",
				Config: asset.GroupConfig{
					GroupBy: []string{"type"},
					Filters: map[string][]string{
						"data.country":     {"id"},
						"data.environment": {"production"},
						"data.company":     {"gotocompany"},
					},
					IncludedFields: []string{"name"},
				},
				Expected: []asset.GroupResult{
					{
						Fields: []asset.GroupField{
							{Name: "type", Value: "topic"},
						},
						Assets: []asset.Asset{
							{Name: "consumer-topic"},
							{Name: "consumer-mq-2"},
						},
					},
				},
			},
		}
		for _, test := range tests {
			t.Run(test.Description, func(t *testing.T) {
				results, err := repo.GroupAssets(ctx, test.Config)
				assert.NoError(t, err)
				assert.Equal(t, len(test.Expected), len(results))

				for i, res := range test.Expected {
					assert.Equal(t, len(res.Fields), len(results[i].Fields))
					assert.Equal(t, len(res.Assets), len(results[i].Assets))
					sort.SliceStable(res.Fields, func(i, j int) bool {
						return res.Fields[i].Name > res.Fields[j].Name
					})

					sort.SliceStable(results[i].Fields, func(j, k int) bool {
						return results[i].Fields[j].Name > results[i].Fields[k].Name
					})
					assert.Equal(t, res.Fields, results[i].Fields)
					for j, assetRes := range res.Assets {
						assert.Equal(t, assetRes.Name, results[i].Assets[j].Name)
					}
				}
			})
		}
	})
}
