package store_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/models"
	"github.com/odpf/columbus/store"
	"github.com/stretchr/testify/assert"
)

type searchTestData struct {
	Type    models.Type     `json:"type"`
	Records []models.Record `json:"records"`
}

type searchTest struct {
	Description string `json:"description"`
	Config      models.SearchConfig
	Expected    []struct {
		Type     string `json:"type"`
		RecordID string `json:"record_id"`
	}
	MatchTotalRows bool `json:"match_total_rows"`
}

type searchTestFixture struct {
	Data  []searchTestData `json:"data"`
	Tests []searchTest     `json:"tests"`
}

func TestSearch(t *testing.T) {
	t.Run("should return an error if search string is empty", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		defer esTestServer.Close()

		searcher, err := store.NewSearcher(esClient, nil)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = searcher.Search(models.SearchConfig{
			Text: "",
		})

		assert.NotNil(t, err)
	})

	t.Run("should restrict search to globally white listed type types", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		defer esTestServer.Close()

		whitelistedType := "whitelisted_type"
		queryText := "sample"
		testData := []searchTestData{
			buildSampleSearchData(whitelistedType),
			buildSampleSearchData("random_type"),
		}

		_, err := populateSearchData(esClient, testData)
		if err != nil {
			t.Error(err)
			return
		}
		searcher, err := store.NewSearcher(esClient, []string{whitelistedType})
		if err != nil {
			t.Error(err)
			return
		}
		results, err := searcher.Search(models.SearchConfig{Text: queryText})
		if err != nil {
			t.Errorf("Search: %v", err)
			return
		}

		assert.Equal(t, 1, len(results))
		assert.Equal(t, whitelistedType, results[0].TypeName)
	})

	t.Run("should restrict search to locally white listed type types", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		defer esTestServer.Close()

		whitelistedType := "whitelisted_type"
		queryText := "sample"
		testData := []searchTestData{
			buildSampleSearchData(whitelistedType),
			buildSampleSearchData("random_type"),
		}

		_, err := populateSearchData(esClient, testData)
		if err != nil {
			t.Error(err)
			return
		}
		searcher, err := store.NewSearcher(esClient, []string{})
		if err != nil {
			t.Error(err)
			return
		}
		results, err := searcher.Search(models.SearchConfig{
			Text:          queryText,
			TypeWhiteList: []string{whitelistedType},
		})
		if err != nil {
			t.Errorf("Search: %v", err)
			return
		}

		assert.Equal(t, 1, len(results))
		assert.Equal(t, whitelistedType, results[0].TypeName)
	})

	t.Run("should restrict search to the common subset of global and local type types", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		defer esTestServer.Close()

		subsetType := "type_c"
		localWhitelist := []string{"type_a", "type_b", subsetType}
		globalWhitelist := []string{subsetType, "type_d", "type_e"}
		queryText := "sample"
		testData := []searchTestData{
			buildSampleSearchData("type_a"),
			buildSampleSearchData("type_b"),
			buildSampleSearchData("type_c"),
			buildSampleSearchData("type_d"),
			buildSampleSearchData("type_e"),
		}

		_, err := populateSearchData(esClient, testData)
		if err != nil {
			t.Error(err)
			return
		}
		searcher, err := store.NewSearcher(esClient, globalWhitelist)
		if err != nil {
			t.Error(err)
			return
		}
		results, err := searcher.Search(models.SearchConfig{
			Text:          queryText,
			TypeWhiteList: localWhitelist,
		})
		if err != nil {
			t.Errorf("Search: %v", err)
			return
		}

		assert.Equal(t, 1, len(results))
		assert.Equal(t, subsetType, results[0].TypeName)
	})

	t.Run("fixtures", func(t *testing.T) {
		esClient := esTestServer.NewClient()
		defer esTestServer.Close()

		testFixture, err := loadTestFixture()
		if err != nil {
			t.Error(err)
		}
		types, err := populateSearchData(esClient, testFixture.Data)
		if err != nil {
			t.Error(err)
		}

		typesMap := mapTypesToTypesMap(types)
		searcher, err := store.NewSearcher(esClient, mapTypesToTypeNames(types))
		if err != nil {
			t.Error(err)
		}

		for _, test := range testFixture.Tests {
			t.Run(test.Description, func(t *testing.T) {
				results, err := searcher.Search(test.Config)
				if err != nil {
					t.Error(err)
					return
				}

				if test.MatchTotalRows {
					assert.Equal(t, len(test.Expected), len(results))
				}

				for i, res := range test.Expected {
					recordIDKey := typesMap[res.Type].Fields.ID
					assert.Equal(t, res.Type, results[i].TypeName)
					assert.Equal(t, res.RecordID, results[i].Record[recordIDKey])
				}
			})
		}
	})
}

func buildSampleSearchData(typeName string) searchTestData {
	return searchTestData{
		Type: models.Type{Name: typeName, Fields: models.TypeFields{ID: "urn"}},
		Records: []models.Record{{
			"urn":       "sample-test-1",
			"landscape": "id",
			"title":     "sample test",
		}},
	}
}

func loadTestFixture() (testFixture searchTestFixture, err error) {
	testFixtureJSON, err := ioutil.ReadFile("./testdata/search-test-fixture.json")
	err = json.Unmarshal(testFixtureJSON, &testFixture)
	if err != nil {
		return testFixture, err
	}

	return testFixture, err
}

func populateSearchData(esClient *elasticsearch.Client, data []searchTestData) (types []models.Type, err error) {
	typeRepo := store.NewTypeRepository(esClient)
	for _, sample := range data {
		types = append(types, sample.Type)
		if err := typeRepo.CreateOrReplace(sample.Type); err != nil {
			return types, err
		}

		recordRepo, _ := store.NewRecordRepositoryFactory(esClient).For(sample.Type)
		if err := recordRepo.CreateOrReplaceMany(sample.Records); err != nil {
			return types, err
		}
	}

	return types, nil
}

func mapTypesToTypeNames(types []models.Type) []string {
	var result []string
	for _, typ := range types {
		result = append(result, typ.Name)
	}

	return result
}

func mapTypesToTypesMap(types []models.Type) map[string]models.Type {
	result := map[string]models.Type{}
	for _, typ := range types {
		result[typ.Name] = typ
	}

	return result
}
