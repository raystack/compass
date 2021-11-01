package lineage_test

import (
	"context"
	"testing"

	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/models"
	"github.com/stretchr/testify/assert"
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
		ctx     = context.Background()
	)
	for _, dataset := range datasets {
		typ := dataset.Type.Normalise()
		tr.On("GetByName", typ.Name).Return(typ, nil)
		recordIterator := new(mock.RecordIterator)
		recordIterator.On("Scan").Return(true).Once()
		recordIterator.On("Scan").Return(false).Once()
		recordIterator.On("Next").Return(dataset.Records)
		recordIterator.On("Close").Return(nil)
		recordRepo := new(mock.RecordRepository)
		recordRepo.On("GetAllIterator", ctx).Return(recordIterator, nil)
		rrf.On("For", typ).Return(recordRepo, nil)
		typList = append(typList, typ)
	}
	tr.On("GetAll", ctx).Return(typList, nil)
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
						},
						Records: []models.Record{
							{
								Urn: "1",
							},
							{
								Urn: "2",
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
						},
						Records: []models.Record{
							{
								Urn: "1",
								Upstreams: []models.LineageRecord{
									{
										Urn:  "A",
										Type: "related-resource-us",
									},
									{
										Urn:  "B",
										Type: "related-resource-us",
									},
								},
								Downstreams: []models.LineageRecord{
									{
										Urn:  "C",
										Type: "related-resource-ds",
									},
								},
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
						},
						Records: []models.Record{
							{
								Urn: "data-booking",
							},
						},
					},
					{
						Type: models.Type{
							Name:           "consumer",
							Classification: models.TypeClassificationResource,
						},
						Records: []models.Record{
							{
								Urn: "booking-aggregator",
								Upstreams: []models.LineageRecord{
									{
										Urn:  "data-booking",
										Type: "producer",
									},
								},
							},
							{
								Urn: "booking-fraud-detector",
								Upstreams: []models.LineageRecord{
									{
										Urn:  "data-booking",
										Type: "producer",
									},
								},
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
				graph, err := lineage.DefaultBuilder.Build(context.Background(), er, rrf)
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

				assert.Equal(t, tc.Result, result)
			})
		}
	})
}
