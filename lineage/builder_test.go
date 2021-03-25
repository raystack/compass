package lineage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/models"
)

type dataset struct {
	Type    models.Type
	Records []models.Record
}

func initialiseRepos(datasets []dataset) (models.TypeRepository, models.RecordRepositoryFactory) {
	var (
		tr      = new(mock.TypeRepository)
		rrf     = new(mock.RecordRepositoryFactory)
		typList = []models.Type{}
	)
	for _, dataset := range datasets {
		typ := dataset.Type.Normalise()
		tr.On("GetByName", typ.Name).Return(typ, nil)
		recordRepo := new(mock.RecordRepository)
		recordRepo.On("GetAll", models.RecordFilter{}).Return(dataset.Records, nil)
		rrf.On("For", typ).Return(recordRepo, nil)
		typList = append(typList, typ)
	}
	tr.On("GetAll").Return(typList, nil)
	return tr, rrf
}

func adjEntryWithTypeAndURN(typ, urn string) lineage.AdjacencyEntry {
	return lineage.AdjacencyEntry{
		Type:        typ,
		URN:         urn,
		Downstreams: set.NewStringSet(),
		Upstreams:   set.NewStringSet(),
	}
}

func TestDefaultBuilder(t *testing.T) {
	t.Run("graph construction algorithm", func(t *testing.T) {
		type testCase struct {
			Description string
			Datasets    []dataset
			Result      lineage.AdjacencyMap
			QueryCfg    lineage.QueryCfg
			BuildErr    error
			QueryErr    error
		}
		var testCases = []testCase{
			{
				Description: "smoke test",
				Datasets: []dataset{
					{
						Type: models.Type{
							Name:           "test",
							Classification: models.TypeClassificationResource,
							Fields: models.TypeFields{
								ID: "id",
							},
						},
						Records: []models.Record{
							{
								"id": "1",
							},
							{
								"id": "2",
							},
						},
					},
				},
				Result: lineage.AdjacencyMap{
					"test/1": adjEntryWithTypeAndURN("test", "1"),
					"test/2": adjEntryWithTypeAndURN("test", "2"),
				},
			},
			{
				// tests that the builder is able to use type.lineage to populate
				// related records
				Description: "internal ref test (simple)",
				Datasets: []dataset{
					{
						Type: models.Type{
							Name:           "internal-ref",
							Classification: models.TypeClassificationResource,
							Fields: models.TypeFields{
								ID: "id",
							},
							Lineage: []models.LineageDescriptor{
								{
									Type:  "related-resource-ds",
									Query: "$.downstream",
									Dir:   models.DataflowDirDownstream,
								},
								{
									Type:  "related-resource-us",
									Query: "$.upstreams",
									Dir:   models.DataflowDirUpstream,
								},
							},
						},
						Records: []models.Record{
							{
								"id":         "1",
								"upstreams":  []string{"A", "B"},
								"downstream": "C",
							},
						},
					},
				},
				Result: lineage.AdjacencyMap{
					"internal-ref/1": lineage.AdjacencyEntry{
						Type:        "internal-ref",
						URN:         "1",
						Downstreams: set.NewStringSet("related-resource-ds/C"),
						Upstreams:   set.NewStringSet("related-resource-us/A", "related-resource-us/B"),
					},
				},
			},
			{
				Description: "external ref test",
				Datasets: []dataset{
					{
						Type: models.Type{
							Name:           "producer",
							Classification: models.TypeClassificationResource,
							Fields: models.TypeFields{
								ID: "id",
							},
						},
						Records: []models.Record{
							{
								"id": "data-booking",
							},
						},
					},
					{
						Type: models.Type{
							Name:           "consumer",
							Classification: models.TypeClassificationResource,
							Fields: models.TypeFields{
								ID: "id",
							},
							Lineage: []models.LineageDescriptor{
								{
									Type:  "producer",
									Query: "$.src",
									Dir:   models.DataflowDirUpstream,
								},
							},
						},
						Records: []models.Record{
							{
								"id":  "booking-aggregator",
								"src": "data-booking",
							},
							{
								"id":  "booking-fraud-detector",
								"src": "data-booking",
							},
						},
					},
				},
				Result: lineage.AdjacencyMap{
					"producer/data-booking": lineage.AdjacencyEntry{
						Type:        "producer",
						URN:         "data-booking",
						Upstreams:   set.NewStringSet(),
						Downstreams: set.NewStringSet("consumer/booking-aggregator", "consumer/booking-fraud-detector"),
					},
					"consumer/booking-aggregator": lineage.AdjacencyEntry{
						Type:        "consumer",
						URN:         "booking-aggregator",
						Upstreams:   set.NewStringSet("producer/data-booking"),
						Downstreams: set.NewStringSet(),
					},
					"consumer/booking-fraud-detector": lineage.AdjacencyEntry{
						Type:        "consumer",
						URN:         "booking-fraud-detector",
						Upstreams:   set.NewStringSet("producer/data-booking"),
						Downstreams: set.NewStringSet(),
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				er, rrf := initialiseRepos(tc.Datasets)
				graph, err := lineage.DefaultBuilder.Build(er, rrf)
				if err != nil {
					if err != tc.BuildErr {
						t.Errorf("unexpected error when building graph: %v", err)
					}
					// the error was expected, terminate test case
					return
				}

				result, err := graph.Query(tc.QueryCfg)
				if err != nil {
					if err != tc.QueryErr {
						t.Errorf("unexpected error querying the graph: %v", err)
					}
					return
				}

				if reflect.DeepEqual(result, tc.Result) == false {
					var (
						msg = new(bytes.Buffer)
						enc = json.NewEncoder(msg)
					)
					enc.SetIndent("", "  ")
					fmt.Fprint(msg, "expected: ")
					enc.Encode(tc.Result)
					fmt.Fprint(msg, "got: ")
					enc.Encode(result)
					t.Error(msg.String())
					return
				}
			})
		}
	})
}
