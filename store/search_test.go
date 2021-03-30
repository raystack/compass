package store_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/odpf/columbus/models"
	"github.com/odpf/columbus/store"
	"github.com/stretchr/testify/assert"
)

type dataset struct {
	Type    models.Type
	Records []models.Record
}

func TestSearcher(t *testing.T) {
	type testCase struct {
		Title         string
		Datasets      []dataset
		SearchConfig  models.SearchConfig
		ShouldFail    bool
		TypeWhiteList []string
		Check         func(tc *testCase, results []models.SearchResult) error
	}

	type resultFilter func(record models.Record, recordType models.Type) bool

	// stdResults is a helper for generating the search response
	// generates a results slice, depending on an optional filter condition
	var stdResults = func(datasets []dataset, filter resultFilter) (results []models.SearchResult) {
		for _, dataset := range datasets {
			for _, record := range dataset.Records {
				if filter != nil && filter(record, dataset.Type) == false {
					continue
				}
				results = append(results, models.SearchResult{
					TypeName: dataset.Type.Name,
					Record:   record,
				})
			}
		}
		return
	}

	daggerTypeClone := daggerType
	daggerTypeClone.Name = "dagger2"

	var daggerTestRecord = models.Record{
		"urn":       "dagger-test-1",
		"landscape": "id",
		"title":     "dagger test",
	}

	var testCases = []testCase{
		{
			Title: "should return matched documents",
			Datasets: []dataset{
				{
					Type:    daggerType,
					Records: []models.Record{daggerTestRecord},
				},
			},
			SearchConfig: models.SearchConfig{
				Text: "test",
			},
			Check: func(tc *testCase, results []models.SearchResult) error {
				expectResults := stdResults(tc.Datasets, nil)
				if reflect.DeepEqual(expectResults, results) == false {
					return incorrectResultsError(expectResults, results)
				}
				return nil
			},
		},
		{
			Title: "should handle search on CamelCase'd field as well",
			Datasets: []dataset{
				{
					Type: daggerType,
					Records: []models.Record{
						{
							"urn":       "BookingLogDagger",
							"landscape": "id",
							"title":     "booking log aggregator",
						},
					},
				},
			},
			SearchConfig: models.SearchConfig{
				Text: "booking",
			},
			Check: func(tc *testCase, results []models.SearchResult) error {
				expectResults := stdResults(tc.Datasets, nil)
				if reflect.DeepEqual(expectResults, results) == false {
					return incorrectResultsError(expectResults, results)
				}
				return nil
			},
		},
		{
			Title: "should return an error if search string is empty",
			Datasets: []dataset{
				{
					Type:    daggerType,
					Records: []models.Record{daggerTestRecord},
				},
			},
			ShouldFail: true,
		},
		{
			Title: "should match documents based on filter criteria",
			Datasets: []dataset{
				{
					Type: daggerType,
					Records: []models.Record{
						{
							"urn":       "test-dagger-1",
							"landscape": "id",
							"title":     "test dagger 1",
						},
						{
							"urn":       "test-dagger-2",
							"landscape": "vn",
							"title":     "test dagger 2",
						},
						{
							"urn":       "test-dagger-3",
							"landscape": "vn",
							"title":     "test dagger 3",
						},
					},
				},
			},
			SearchConfig: models.SearchConfig{
				Text: "dagger",
				Filters: map[string][]string{
					"landscape": {"vn"},
				},
			},
			Check: func(tc *testCase, results []models.SearchResult) error {
				expectResults := stdResults(tc.Datasets, func(record models.Record, ent models.Type) bool {
					return record["landscape"].(string) == "vn"
				})
				if reflect.DeepEqual(results, expectResults) == false {
					return incorrectResultsError(expectResults, results)
				}
				return nil
			},
		},
		{
			Title: "should match documents that don't contain the filter key",
			Datasets: []dataset{
				{
					Type: daggerType,
					Records: []models.Record{
						{
							"urn":       "test-dagger-1",
							"landscape": "id",
							"title":     "test dagger 1",
						},
						{
							"urn":   "test-dagger-2",
							"title": "test dagger 2",
						},
					},
				},
			},
			SearchConfig: models.SearchConfig{
				Text: "dagger",
				Filters: map[string][]string{
					"landscape": {"id"},
				},
			},
			Check: func(tc *testCase, results []models.SearchResult) error {
				expectResults := stdResults(tc.Datasets, nil)
				if reflect.DeepEqual(results, expectResults) == false {
					return incorrectResultsError(expectResults, results)
				}
				return nil
			},
		},
		{
			Title: "should restrict search to globally white listed type types",
			SearchConfig: models.SearchConfig{
				Text: "dagger",
			},
			Datasets: []dataset{
				{
					Type:    daggerType,
					Records: []models.Record{daggerTestRecord},
				},
				{
					Type:    daggerTypeClone,
					Records: []models.Record{daggerTestRecord},
				},
			},
			TypeWhiteList: []string{daggerType.Name},
			Check: func(tc *testCase, results []models.SearchResult) error {
				expectResults := stdResults(tc.Datasets, func(record models.Record, recordType models.Type) bool {
					for _, name := range tc.TypeWhiteList {
						if name == recordType.Name {
							return true
						}
					}
					return false
				})
				if reflect.DeepEqual(expectResults, results) == false {
					return incorrectResultsError(expectResults, results)
				}
				return nil
			},
		},
		{
			Title: "should restrict search to locally white listed type types",
			SearchConfig: models.SearchConfig{
				Text:          "dagger",
				TypeWhiteList: []string{"dagger"},
			},
			Datasets: []dataset{
				{
					Type:    daggerType,
					Records: []models.Record{daggerTestRecord},
				},
				{
					Type:    daggerTypeClone,
					Records: []models.Record{daggerTestRecord},
				},
			},
			Check: func(tc *testCase, results []models.SearchResult) error {
				expectResults := stdResults(tc.Datasets, func(record models.Record, recordType models.Type) bool {
					for _, name := range tc.SearchConfig.TypeWhiteList {
						if name == recordType.Name {
							return true
						}
					}
					return false
				})
				if reflect.DeepEqual(expectResults, results) == false {
					return incorrectResultsError(expectResults, results)
				}
				return nil
			},
		},
		{
			Title: "should restrict search to the common subset of global and local type types",
			SearchConfig: models.SearchConfig{
				Text:          "dagger",
				TypeWhiteList: []string{"dagger", "dagger2", "firehose"},
			},
			TypeWhiteList: []string{"sakaar", "dagger"},
			Datasets: []dataset{
				{
					Type:    daggerType,
					Records: []models.Record{daggerTestRecord},
				},
				{
					Type:    daggerTypeClone,
					Records: []models.Record{daggerTestRecord},
				},
			},
			Check: func(tc *testCase, results []models.SearchResult) error {
				expectResults := stdResults(tc.Datasets, func(record models.Record, recordType models.Type) bool {
					if recordType.Name == "dagger" {
						return true
					}
					return false
				})
				if reflect.DeepEqual(expectResults, results) == false {
					return incorrectResultsError(expectResults, results)
				}
				return nil
			},
		},
	}

	var setupError = func(err error) string {
		return fmt.Sprintf("error setting up testcase: %v", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			var (
				cli               = esTestServer.NewClient()
				typeRepo          = store.NewTypeRepository(cli)
				recordRepoFactory = store.NewRecordRepositoryFactory(cli)
				searcher, err     = store.NewSearcher(cli, testCase.TypeWhiteList)
			)

			assert.Nil(t, err)

			for _, ds := range testCase.Datasets {
				err := typeRepo.CreateOrReplace(ds.Type)
				if err != nil {
					t.Fatal(setupError(err))
				}
				recordRepo, err := recordRepoFactory.For(ds.Type)
				if err != nil {
					t.Fatal(setupError(err))
				}
				err = recordRepo.CreateOrReplaceMany(ds.Records)
				if err != nil {
					t.Fatal(setupError(err))
				}

			}

			results, err := searcher.Search(testCase.SearchConfig)
			if testCase.ShouldFail {
				assert.Error(t, err)
				return
			}

			if err != nil {
				t.Errorf("Search: %v", err)
				return
			}
			if err = testCase.Check(&testCase, results); err != nil {
				t.Errorf("check failed: %v", err)
			}
		})
	}
}

func TestNewSearcher(t *testing.T) {
	t.Run("should return an error if an internal index is specified in TypeWhiteList", func(t *testing.T) {
		reservedIndices := []string{
			"meta",
			"universe",
		}
		for _, ri := range reservedIndices {
			t.Run(ri, func(t *testing.T) {
				_, err := store.NewSearcher(esTestServer.NewClient(), []string{ri})
				assert.Error(t, err)
			})
		}
	})
}
