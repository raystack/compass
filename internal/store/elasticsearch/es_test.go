package elasticsearch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/compass/core/asset"
	store "github.com/odpf/compass/internal/store/elasticsearch"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElasticsearch(t *testing.T) {
	ctx := context.Background()
	t.Run("Migrate", func(t *testing.T) {
		var testCases = []struct {
			Title      string
			Type       asset.Type
			ShouldFail bool
			Validate   func(esClient *store.Client, cli *elasticsearch.Client, assetType asset.Type) error
		}{
			{
				Title:      "should successfully write the document to elasticsearch",
				Type:       daggerType,
				ShouldFail: false,
			},
			{
				Title:      "should create the index ${assetType} in elasticsearch",
				Type:       daggerType,
				ShouldFail: false,
				Validate: func(esClient *store.Client, cli *elasticsearch.Client, assetType asset.Type) error {
					idxRequest := &esapi.IndicesExistsRequest{
						Index: []string{
							assetType.String(),
						},
					}
					res, err := idxRequest.Do(context.Background(), cli)
					if err != nil {
						return fmt.Errorf("failed to query elasticsearch for index %q: %w", assetType, err)
					}
					defer res.Body.Close()
					if res.IsError() {
						return fmt.Errorf("elasticsearch: error querying existence of %q index: %s", assetType, res.Status())
					}
					return nil
				},
			},
			{
				Title:      "should not accept any type that has the same name as the search index",
				Type:       asset.Type("universe"), // defaultSearchIndex
				ShouldFail: true,
			},
			{
				Title: "should alias the type to the search index",
				Type:  daggerType,
				Validate: func(esClient *store.Client, cli *elasticsearch.Client, assetType asset.Type) error {
					searchIndex := "universe"
					req, err := http.NewRequest("GET", "/_alias/"+searchIndex, nil)
					if err != nil {
						return fmt.Errorf("error creating request: %w", err)
					}
					res, err := cli.Perform(req)
					if err != nil {
						return fmt.Errorf("error calling elasticsearch alias API: %w", err)
					}
					defer res.Body.Close()

					aliases := make(map[string]interface{})
					err = json.NewDecoder(res.Body).Decode(&aliases)
					if err != nil {
						return fmt.Errorf("error decoding elasticsearch response: %w", err)
					}
					if _, created := aliases[assetType.String()]; !created {
						return fmt.Errorf("expected %q index to be aliased to %q, but it was not", assetType, searchIndex)
					}
					return nil
				},
			},
			{
				Title: "type creation should be idempotent",
				Type:  daggerType,
				Validate: func(esClient *store.Client, cli *elasticsearch.Client, assetType asset.Type) error {
					// we'll try to save the type again, with the expectation
					// that it should succeed as normal
					err := esClient.Migrate(ctx, daggerType)
					if err != nil {
						return fmt.Errorf("repository returned unexpected error: %w", err)
					}
					return nil
				},
			},
			{
				Title: "created index should be able to correctly tokenize CamelCase text",
				Type:  daggerType,
				Validate: func(esClient *store.Client, cli *elasticsearch.Client, assetType asset.Type) error {
					textToAnalyze := "HelloWorld"
					analyzerPath := fmt.Sprintf("/%s/_analyze", assetType)
					analyzerPayload := fmt.Sprintf(`{"analyzer": "my_analyzer", "text": %q}`, textToAnalyze)

					req, err := http.NewRequest("POST", analyzerPath, strings.NewReader(analyzerPayload))
					if err != nil {
						return fmt.Errorf("error creating analyzer request: %w", err)
					}
					req.Header.Add("content-type", "application/json")

					res, err := cli.Perform(req)
					if err != nil {
						return fmt.Errorf("error invoking analyzer: %v", err)
					}
					defer res.Body.Close()
					if res.StatusCode != http.StatusOK {
						return fmt.Errorf("elasticsearch returned non-200 response: %d", res.StatusCode)
					}
					var response struct {
						Tokens []struct {
							Token string `json:"token"`
						} `json:"tokens"`
					}
					err = json.NewDecoder(res.Body).Decode(&response)
					if err != nil {
						return fmt.Errorf("error decoding response: %w", err)
					}
					expectTokens := []string{"hello", "world"}
					analyzedTokens := []string{}
					for _, tok := range response.Tokens {
						analyzedTokens = append(analyzedTokens, tok.Token)
					}

					if reflect.DeepEqual(expectTokens, analyzedTokens) == false {
						return fmt.Errorf("expected analyzer to tokenize %q as %v, was %v", textToAnalyze, expectTokens, analyzedTokens)
					}
					return nil
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Title, func(t *testing.T) {
				cli, err := esTestServer.NewClient()
				require.NoError(t, err)
				esClient, err := store.NewClient(
					log.NewNoop(),
					store.Config{},
					store.WithClient(cli),
				)
				require.NoError(t, err)

				err = esClient.Migrate(ctx, testCase.Type)
				if testCase.ShouldFail {
					assert.Error(t, err)
					return
				} else if err != nil {
					t.Errorf("repository returned unexpected error: %v", err)
					return
				}

				if testCase.Validate != nil {
					if err := testCase.Validate(esClient, cli, testCase.Type); err != nil {
						t.Error(err)
						return
					}
				}
			})
		}
	})
}
