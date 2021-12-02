package lineage_test

import (
	"context"
	"testing"

	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lib/mock"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/record"
	"github.com/stretchr/testify/assert"
)

type dataset struct {
	Type    record.Type
	Records []record.Record
}

func initialiseRepos(datasets []dataset) (record.TypeRepository, discovery.RecordRepositoryFactory) {
	var (
		tr      = new(mock.TypeRepository)
		rrf     = new(mock.RecordRepositoryFactory)
		typList = []record.Type{}
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
		rrf.On("For", typ.Name).Return(recordRepo, nil)
		typList = append(typList, typ)
	}
	tr.On("GetAll", ctx).Return(typList, nil)
	return tr, rrf
}

func adjEntryWithTypeAndURN(typ, urn, service string) lineage.AdjacencyEntry {
	return lineage.AdjacencyEntry{
		Type:        typ,
		URN:         urn,
		Service:     service,
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
						Type: record.Type{
							Name:           "test",
							Classification: record.TypeClassificationResource,
						},
						Records: []record.Record{
							{
								Urn:     "1",
								Service: "service-A",
							},
							{
								Urn:     "2",
								Service: "service-A",
							},
						},
					},
				},
				Result: lineage.AdjacencyMap{
					"test/1": adjEntryWithTypeAndURN("test", "1", "service-A"),
					"test/2": adjEntryWithTypeAndURN("test", "2", "service-A"),
				},
			},
			{
				// tests that the builder is able to use type.lineage to populate
				// related records
				Description: "internal ref test (simple)",
				Datasets: []dataset{
					{
						Type: record.Type{
							Name:           "internal-ref",
							Classification: record.TypeClassificationResource,
						},
						Records: []record.Record{
							{
								Urn:     "1",
								Service: "service-A",
								Upstreams: []record.LineageRecord{
									{
										Urn:  "A",
										Type: "related-resource-us",
									},
									{
										Urn:  "B",
										Type: "related-resource-us",
									},
								},
								Downstreams: []record.LineageRecord{
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
						Service:     "service-A",
						Downstreams: set.NewStringSet("related-resource-ds/C"),
						Upstreams:   set.NewStringSet("related-resource-us/A", "related-resource-us/B"),
					},
				},
			},
			{
				Description: "external ref test",
				Datasets: []dataset{
					{
						Type: record.Type{
							Name:           "producer",
							Classification: record.TypeClassificationResource,
						},
						Records: []record.Record{
							{
								Urn: "data-booking",
							},
						},
					},
					{
						Type: record.Type{
							Name:           "consumer",
							Classification: record.TypeClassificationResource,
						},
						Records: []record.Record{
							{
								Urn: "booking-aggregator",
								Upstreams: []record.LineageRecord{
									{
										Urn:  "data-booking",
										Type: "producer",
									},
								},
							},
							{
								Urn: "booking-fraud-detector",
								Upstreams: []record.LineageRecord{
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
