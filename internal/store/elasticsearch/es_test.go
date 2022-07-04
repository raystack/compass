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
	store "github.com/odpf/compass/internal/store/elasticsearch"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElasticsearch(t *testing.T) {
	ctx := context.Background()
	t.Run("Create", func(t *testing.T) {
		var testCases = []struct {
			Title      string
			Service    string
			ShouldFail bool
			Validate   func(esClient *store.Client, cli *elasticsearch.Client, indexName string) error
		}{
			{
				Title:      "should successfully write the document to elasticsearch",
				Service:    daggerService,
				ShouldFail: false,
			},
			{
				Title:      "should create the index ${service} in elasticsearch",
				Service:    daggerService,
				ShouldFail: false,
				Validate: func(esClient *store.Client, cli *elasticsearch.Client, indexName string) error {
					idxRequest := &esapi.IndicesExistsRequest{
						Index: []string{
							indexName,
						},
					}
					res, err := idxRequest.Do(context.Background(), cli)
					if err != nil {
						return fmt.Errorf("failed to query elasticsearch for index %q: %w", indexName, err)
					}
					defer res.Body.Close()
					if res.IsError() {
						return fmt.Errorf("elasticsearch: error querying existence of %q index: %s", indexName, res.Status())
					}
					return nil
				},
			},
			{
				Title:      "should not accept any service that has the same name as the search index",
				Service:    "universe", // defaultSearchIndex
				ShouldFail: true,
			},
			{
				Title:   "should alias the type to the search index",
				Service: daggerService,
				Validate: func(esClient *store.Client, cli *elasticsearch.Client, indexName string) error {
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
					if _, created := aliases[indexName]; !created {
						return fmt.Errorf("expected %q index to be aliased to %q, but it was not", indexName, searchIndex)
					}
					return nil
				},
			},
			{
				Title:   "created index should be able to correctly tokenize CamelCase text",
				Service: daggerService,
				Validate: func(esClient *store.Client, cli *elasticsearch.Client, indexName string) error {
					textToAnalyze := "HelloWorld"
					analyzerPath := fmt.Sprintf("/%s/_analyze", indexName)
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
				_, err = esClient.Init()
				assert.NoError(t, err)
				err = esClient.CreateIdx(ctx, testCase.Service)
				if testCase.ShouldFail {
					assert.Error(t, err)
					return
				} else if err != nil {
					t.Errorf("repository returned unexpected error: %v", err)
					return
				}

				if testCase.Validate != nil {
					if err := testCase.Validate(esClient, cli, testCase.Service); err != nil {
						t.Error(err)
						return
					}
				}
			})
		}
	})
}

// func TestInit(t *testing.T){
// 	s, err := elasticsearch.NewClient()
// }
