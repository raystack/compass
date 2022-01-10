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
	"github.com/stretchr/testify/require"
)

func TestTypeRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateOrReplace", func(t *testing.T) {
		var testCases = []struct {
			Title      string
			Type       record.Type
			ShouldFail bool
			Validate   func(cli *elasticsearch.Client, recordType record.Type) error
		}{
			{
				Title:      "should successfully write the document to elasticsearch",
				Type:       daggerType,
				ShouldFail: false,
			},
			{
				Title:      "should create the index ${recordType.Name} in elasticsearch",
				Type:       daggerType,
				ShouldFail: false,
				Validate: func(cli *elasticsearch.Client, recordType record.Type) error {
					idxRequest := &esapi.IndicesExistsRequest{
						Index: []string{
							recordType.Name.String(),
						},
					}
					res, err := idxRequest.Do(context.Background(), cli)
					if err != nil {
						return fmt.Errorf("failed to query elasticsearch for index %q: %v", recordType.Name, err)
					}
					defer res.Body.Close()
					if res.IsError() {
						return fmt.Errorf("elasticsearch: error querying existence of %q index: %s", recordType.Name, res.Status())
					}
					return nil
				},
			},
			{
				Title: "should not accept any type that has the same name as the metadata index",
				Type: record.Type{
					Name: "meta", // defaultMetaIndex
				},
				ShouldFail: true,
			},
			{
				Title: "should not accept any type that has the same name as the search index",
				Type: record.Type{
					Name: "universe", // defaultSearchIndex
				},
				ShouldFail: true,
			},
			{
				Title: "should alias the type to the search index",
				Type:  daggerType,
				Validate: func(cli *elasticsearch.Client, recordType record.Type) error {
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
					if _, created := aliases[recordType.Name.String()]; !created {
						return fmt.Errorf("expected %q index to be aliased to %q, but it was not", recordType.Name, searchIndex)
					}
					return nil
				},
			},
			{
				Title: "type creation should be idempotent",
				Type:  daggerType,
				Validate: func(cli *elasticsearch.Client, recordType record.Type) error {
					// we'll try to save the type again, with the expectation
					// that it should succeed as normal
					repo := store.NewTypeRepository(cli)
					err := repo.CreateOrReplace(ctx, daggerType)
					if err != nil {
						return fmt.Errorf("repository returned unexpected error: %w", err)
					}
					return nil
				},
			},
			{
				Title: "created index should be able to correctly tokenize CamelCase text",
				Type:  daggerType,
				Validate: func(cli *elasticsearch.Client, recordType record.Type) error {
					textToAnalyze := "HelloWorld"
					analyzerPath := fmt.Sprintf("/%s/_analyze", recordType.Normalise().Name)
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
				repo := store.NewTypeRepository(cli)
				err := repo.CreateOrReplace(ctx, testCase.Type)
				if testCase.ShouldFail {
					assert.Error(t, err)
					return
				} else if err != nil {
					t.Errorf("repository returned unexpected error: %v", err)
					return
				}

				if testCase.Validate != nil {
					if err := testCase.Validate(cli, testCase.Type); err != nil {
						t.Error(err)
						return
					}
				}
			})
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		repo := store.NewTypeRepository(esTestServer.NewClient())
		err := repo.CreateOrReplace(ctx, daggerType)
		if err != nil {
			t.Errorf("error writing to elasticsearch: %v", err)
			return
		}

		typeFromRepo, err := repo.GetByName(ctx, daggerType.Name.String())
		if err != nil {
			t.Errorf("error getting type from repository: %v", err)
			return
		}
		if reflect.DeepEqual(daggerType, typeFromRepo) == false {
			t.Errorf("expected repository to return %#v, returned %#v instead", daggerType, typeFromRepo)
			return
		}
	})
	t.Run("GetAll", func(t *testing.T) {
		t.Run("should return empty list if no type is available", func(t *testing.T) {
			esClient := esTestServer.NewClient()
			repo := store.NewTypeRepository(esClient)
			_, err := esClient.Indices.Create("meta")
			if err != nil {
				t.Errorf("error creating meta index: %v", err)
				return
			}

			types, err := repo.GetAll(ctx)
			if err != nil {
				t.Errorf("error getting type from repository: %v", err)
				return
			}

			assert.Equal(t, []record.Type{}, types)
		})
		t.Run("should return types from elasticsearch", func(t *testing.T) {
			repo := store.NewTypeRepository(esTestServer.NewClient())
			err := repo.CreateOrReplace(ctx, daggerType)
			if err != nil {
				t.Errorf("error writing to elasticsearch: %v", err)
				return
			}

			types, err := repo.GetAll(ctx)
			if err != nil {
				t.Errorf("error getting type from repository: %v", err)
				return
			}
			var expect = []record.Type{daggerType}
			assert.Equal(t, expect, types)
		})
	})
	t.Run("GetRecordsCount", func(t *testing.T) {
		t.Run("should return empty map if no type is available", func(t *testing.T) {
			esClient := esTestServer.NewClient()
			repo := store.NewTypeRepository(esTestServer.NewClient())
			_, err := esClient.Indices.Create("meta")
			require.NoError(t, err)

			counts, err := repo.GetRecordsCount(ctx)
			require.NoError(t, err)

			assert.Equal(t, map[string]int{}, counts)
		})
		t.Run("should return map with 0 count if type has not been populated yet", func(t *testing.T) {
			typ := record.Type{
				Name: "test",
			}
			repo := store.NewTypeRepository(esTestServer.NewClient())
			err := repo.CreateOrReplace(ctx, typ)
			require.NoError(t, err)

			counts, err := repo.GetRecordsCount(ctx)
			require.NoError(t, err)

			expected := map[string]int{
				"test": 0,
			}
			assert.Equal(t, expected, counts)
		})
		t.Run("should return maps of record count with type as its key", func(t *testing.T) {
			typ := record.Type{
				Name: "test2",
			}
			records := []record.Record{
				{Urn: "record-1", Name: "record-1"},
				{Urn: "record-2", Name: "record-2"},
				{Urn: "record-3", Name: "record-3"},
			}

			esClient := esTestServer.NewClient()
			repo := store.NewTypeRepository(esClient)
			err := repo.CreateOrReplace(ctx, typ)
			require.NoError(t, err)

			rrf := store.NewRecordRepositoryFactory(esClient)
			rr, err := rrf.For(typ.Name.String())
			require.NoError(t, err)
			err = rr.CreateOrReplaceMany(ctx, records)
			require.NoError(t, err)

			counts, err := repo.GetRecordsCount(ctx)
			require.NoError(t, err)

			expected := map[string]int{
				"test2": len(records),
			}
			assert.Equal(t, expected, counts)
		})
	})
	t.Run("Delete", func(t *testing.T) {
		typeName := "delete-type"
		esClient := esTestServer.NewClient()
		repo := store.NewTypeRepository(esClient)

		t.Run("should return error if type name is reserved key", func(t *testing.T) {
			var err error

			err = repo.Delete(ctx, "meta")
			assert.NotNil(t, err)
			assert.IsType(t, record.ErrReservedTypeName{}, err)

			err = repo.Delete(ctx, "universe")
			assert.NotNil(t, err)
			assert.IsType(t, record.ErrReservedTypeName{}, err)
		})

		t.Run("should delete type by its name", func(t *testing.T) {
			err := repo.CreateOrReplace(ctx, record.Type{
				Name: record.TypeName(typeName),
			})
			if err != nil {
				t.Errorf("error writing to elasticsearch: %v", err)
				return
			}
			err = repo.Delete(ctx, typeName)
			if err != nil {
				t.Errorf("error deleting type: %v", err)
				return
			}

			_, err = repo.GetByName(ctx, typeName)
			assert.NotNil(t, err)
			assert.IsType(t, record.ErrNoSuchType{}, err)
		})

		t.Run("should delete the type's elasticsearch index", func(t *testing.T) {
			_, err := esClient.Indices.Create(typeName)
			if err != nil {
				t.Errorf("error creating index: %v", err)
				return
			}
			err = repo.Delete(ctx, typeName)
			if err != nil {
				t.Errorf("error deleting type: %v", err)
				return
			}

			response, err := esClient.Indices.Get([]string{typeName})
			if err != nil {
				t.Errorf("error getting indices type: %v", err)
				return
			}
			var indices map[string]interface{}
			if err := json.NewDecoder(response.Body).Decode(&indices); err != nil {
				t.Error(err)
				return
			}
			assert.Nil(t, indices[typeName])
		})
	})
}
