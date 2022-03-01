package lineage_test

import (
	"context"
	"testing"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/lineage/v1"
	"github.com/stretchr/testify/assert"
)

type dataset struct {
	TypeName asset.Type
	Records  []asset.Asset
}

func initialiseRepos(datasets []dataset) (discovery.TypeRepository, discovery.RecordRepositoryFactory) {
	var (
		tr          = new(mocks.TypeRepository)
		rrf         = new(mocks.RecordRepositoryFactory)
		typNameList = map[asset.Type]int{}
		ctx         = context.Background()
	)
	for _, dataset := range datasets {
		tr.On("GetByName", dataset.TypeName).Return(dataset.TypeName, nil)
		recordIterator := new(mocks.RecordIterator)
		recordIterator.On("Scan").Return(true).Once()
		recordIterator.On("Scan").Return(false).Once()
		recordIterator.On("Next").Return(dataset.Records)
		recordIterator.On("Close").Return(nil)
		recordRepo := new(mocks.RecordRepository)
		recordRepo.On("GetAllIterator", ctx).Return(recordIterator, nil)
		rrf.On("For", dataset.TypeName.String()).Return(recordRepo, nil)
		typNameList[dataset.TypeName] = 1
	}
	tr.On("GetAll", ctx).Return(typNameList, nil)
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
						TypeName: asset.Type("test"),
						Records: []asset.Asset{
							{
								URN:     "1",
								Service: "service-A",
							},
							{
								URN:     "2",
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
						TypeName: asset.Type("internal-ref"),
						Records: []asset.Asset{
							{
								URN:     "1",
								Service: "service-A",
								Upstreams: []asset.LineageRecord{
									{
										URN:  "A",
										Type: "related-resource-us",
									},
									{
										URN:  "B",
										Type: "related-resource-us",
									},
								},
								Downstreams: []asset.LineageRecord{
									{
										URN:  "C",
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
						TypeName: asset.Type("producer"),
						Records: []asset.Asset{
							{
								URN: "data-booking",
							},
						},
					},
					{
						TypeName: asset.Type("consumer"),
						Records: []asset.Asset{
							{
								URN: "booking-aggregator",
								Upstreams: []asset.LineageRecord{
									{
										URN:  "data-booking",
										Type: "producer",
									},
								},
							},
							{
								URN: "booking-fraud-detector",
								Upstreams: []asset.LineageRecord{
									{
										URN:  "data-booking",
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
