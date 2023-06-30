package elasticsearch_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/namespace"
	"os"
	"testing"

	"github.com/raystack/compass/core/asset"
	store "github.com/raystack/compass/internal/store/elasticsearch"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchTestData struct {
	Assets []asset.Asset `json:"assets"`
}

func TestSearcherSearch(t *testing.T) {
	ctx := context.TODO()
	ns := &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "umbrella",
		State:    namespace.DedicatedState,
		Metadata: nil,
	}

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
		err = repo.CreateNamespace(ctx, ns)
		assert.NoError(t, err)

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

		err = loadTestFixture(esClient, ns, "./testdata/search-test-fixture.json")
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient)

		type searchTest struct {
			Description    string
			Config         asset.SearchConfig
			Expected       []asset.SearchResult
			MatchTotalRows bool
		}
		tests := []searchTest{
			{
				Description: "should fetch assets which has text in any of its fields",
				Config: asset.SearchConfig{
					Text:      "topic",
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "topic", ID: "consumer-topic"},
					{Type: "topic", ID: "order-topic"},
					{Type: "topic", ID: "purchase-topic"},
					{Type: "topic", ID: "consumer-mq-2"},
					{Type: "topic", ID: "transaction"},
				},
			},
			{
				Description: "should enable fuzzy search",
				Config: asset.SearchConfig{
					Text:      "tpic",
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "topic", ID: "consumer-topic"},
					{Type: "topic", ID: "order-topic"},
					{Type: "topic", ID: "purchase-topic"},
					{Type: "topic", ID: "consumer-mq-2"},
					{Type: "topic", ID: "transaction"},
				},
			},
			{
				Description: "should put more weight on id fields",
				Config: asset.SearchConfig{
					Text:      "invoice",
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "table", ID: "us1-apple-invoice"},
					{Type: "table", ID: "au2-microsoft-invoice"},
					{Type: "topic", ID: "transaction"},
				},
			},
			{
				Description: "should filter by service if given",
				Config: asset.SearchConfig{
					Text: "invoice",
					Filters: map[string][]string{
						"service": {"rabbitmq", "postgres"},
					},
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "table", ID: "au2-microsoft-invoice"},
					{Type: "topic", ID: "transaction"},
				},
			},
			{
				Description: "should match documents based on filter criteria",
				Config: asset.SearchConfig{
					Text: "topic",
					Filters: map[string][]string{
						"data.company": {"raystack"},
					},
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "topic", ID: "consumer-topic"},
					{Type: "topic", ID: "order-topic"},
					{Type: "topic", ID: "consumer-mq-2"},
					{Type: "topic", ID: "transaction"},
				},
			},
			{
				Description: "should not return assets without fields specified in filters",
				Config: asset.SearchConfig{
					Text: "invoice topic",
					Filters: map[string][]string{
						"data.country":     {"id"},
						"data.environment": {"production"},
						"data.company":     {"raystack"},
					},
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "topic", ID: "consumer-topic"},
					{Type: "topic", ID: "consumer-mq-2"},
				},
			},
			{
				Description: "should return 'consumer-topic' if filter owner email with 'john.doe@email.com'",
				Config: asset.SearchConfig{
					Text: "topic",
					Filters: map[string][]string{
						"owners.email": {"john.doe@email.com"},
					},
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "topic", ID: "consumer-topic"},
				},
			},
			{
				Description: "should return a descendingly sorted based on usage count in search results if rank by usage in the config",
				Config: asset.SearchConfig{
					Text:      "bigquery",
					RankBy:    "data.profile.usage_count",
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "table", ID: "bigquery::gcpproject/dataset/tablename-common"},
					{Type: "table", ID: "bigquery::gcpproject/dataset/tablename-mid"},
					{Type: "table", ID: "bigquery::gcpproject/dataset/tablename-1"},
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
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "topic", ID: "consumer-topic"},
				},
			},
			{
				Description: "should return 'bigquery::gcpproject/dataset/tablename-common' resource on top if search by query table column name field with text 'tablename-common-column1'",
				Config: asset.SearchConfig{
					Text: "tablename",
					Queries: map[string]string{
						"data.schema.columns.name": "common",
					},
					Namespace: ns,
				},
				Expected: []asset.SearchResult{
					{Type: "table", ID: "bigquery::gcpproject/dataset/tablename-common"},
				},
			},
		}
		for _, test := range tests {
			t.Run(test.Description, func(t *testing.T) {
				results, err := repo.Search(ctx, test.Config)
				require.NoError(t, err)

				require.Equal(t, len(test.Expected), len(results))
				for _, res := range test.Expected {
					assert.True(t, isContain(results, res))
				}
			})
		}
	})
}

func isContain(bag []asset.SearchResult, item asset.SearchResult) bool {
	for _, current := range bag {
		if current.Type == item.Type && current.ID == item.ID {
			return true
		}
	}
	return false
}

func TestSearcherSuggest(t *testing.T) {
	ctx := context.TODO()
	ns := &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "umbrella",
		State:    namespace.DedicatedState,
		Metadata: nil,
	}
	cli, err := esTestServer.NewClient()
	require.NoError(t, err)
	esClient, err := store.NewClient(
		log.NewNoop(),
		store.Config{},
		store.WithClient(cli),
	)
	require.NoError(t, err)

	err = loadTestFixture(esClient, ns, "./testdata/suggest-test-fixture.json")
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
			config := asset.SearchConfig{Text: tc.term, Namespace: ns}
			actual, err := repo.Suggest(ctx, config)
			assert.NoError(t, err)

			assert.Equal(t, tc.expected, actual, "suggestions are not as expected for term: %s and index: %d", tc.term, i)
		}
	})
}

func loadTestFixture(esClient *store.Client, ns *namespace.Namespace, filePath string) (err error) {
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
		repo := store.NewDiscoveryRepository(esClient, store.WithInstantRefresh())
		if err = repo.CreateNamespace(ctx, ns); err != nil {
			return err
		}
		for _, ast := range testdata.Assets {
			if err := repo.Upsert(ctx, ns, &ast); err != nil {
				return err
			}
		}
	}

	return err
}
