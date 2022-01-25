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
	"github.com/odpf/columbus/record"
	store "github.com/odpf/columbus/store/elasticsearch"
	"github.com/stretchr/testify/assert"
)

func TestElasticsearch(t *testing.T) {
	ctx := context.Background()
	t.Run("CreateOrReplace", func(t *testing.T) {
		var testCases = []struct {
			Title      string
			TypeName   record.TypeName
			ShouldFail bool
			Validate   func(cli *elasticsearch.Client, recordTypeName record.TypeName) error
		}{
			{
				Title:      "should successfully write the document to elasticsearch",
				TypeName:   daggerType,
				ShouldFail: false,
			},
			{
				Title:      "should create the index ${recordType.Name} in elasticsearch",
				TypeName:   daggerType,
				ShouldFail: false,
				Validate: func(cli *elasticsearch.Client, recordTypeName record.TypeName) error {
					idxRequest := &esapi.IndicesExistsRequest{
						Index: []string{
							recordTypeName.String(),
						},
					}
					res, err := idxRequest.Do(context.Background(), cli)
					if err != nil {
						return fmt.Errorf("failed to query elasticsearch for index %q: %w", recordTypeName, err)
					}
					defer res.Body.Close()
					if res.IsError() {
						return fmt.Errorf("elasticsearch: error querying existence of %q index: %s", recordTypeName, res.Status())
					}
					return nil
				},
			},
			{
				Title:      "should not accept any type that has the same name as the search index",
				TypeName:   record.TypeName("universe"), // defaultSearchIndex
				ShouldFail: true,
			},
			{
				Title:    "should alias the type to the search index",
				TypeName: daggerType,
				Validate: func(cli *elasticsearch.Client, recordTypeName record.TypeName) error {
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
					if _, created := aliases[recordTypeName.String()]; !created {
						return fmt.Errorf("expected %q index to be aliased to %q, but it was not", recordTypeName, searchIndex)
					}
					return nil
				},
			},
			{
				Title:    "type creation should be idempotent",
				TypeName: daggerType,
				Validate: func(cli *elasticsearch.Client, recordTypeName record.TypeName) error {
					// we'll try to save the type again, with the expectation
					// that it should succeed as normal
					err := store.Migrate(ctx, cli, daggerType)
					if err != nil {
						return fmt.Errorf("repository returned unexpected error: %w", err)
					}
					return nil
				},
			},
			{
				Title:    "created index should be able to correctly tokenize CamelCase text",
				TypeName: daggerType,
				Validate: func(cli *elasticsearch.Client, recordTypeName record.TypeName) error {
					textToAnalyze := "HelloWorld"
					analyzerPath := fmt.Sprintf("/%s/_analyze", recordTypeName)
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
				cli := esTestServer.NewClient()
				err := store.Migrate(ctx, cli, testCase.TypeName)
				if testCase.ShouldFail {
					assert.Error(t, err)
					return
				} else if err != nil {
					t.Errorf("repository returned unexpected error: %v", err)
					return
				}

				if testCase.Validate != nil {
					if err := testCase.Validate(cli, testCase.TypeName); err != nil {
						t.Error(err)
						return
					}
				}
			})
		}
	})
}
