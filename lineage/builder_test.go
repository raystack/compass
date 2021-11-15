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

func initialiseRepos(datasets map[record.Type][]record.Record) discovery.RecordRepositoryFactory {
	var (
		rrf = new(mock.RecordRepositoryFactory)
		ctx = context.Background()
	)

	for _, t := range record.TypeList {
		var records []record.Record
		records, ok := datasets[t]
		recordIterator := new(mock.RecordIterator)
		if ok {
			recordIterator.On("Scan").Return(true).Once()
			recordIterator.On("Scan").Return(false).Once()
			recordIterator.On("Next").Return(records)
		} else {
			recordIterator.On("Scan").Return(false).Once()
		}

		recordIterator.On("Close").Return(nil)
		recordRepo := new(mock.RecordRepository)
		recordRepo.On("GetAllIterator", ctx).Return(recordIterator, nil)
		rrf.On("For", t).Return(recordRepo, nil)
	}

	return rrf
}

func adjEntryWithTypeAndURN(typ record.Type, urn, service string) lineage.AdjacencyEntry {
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
			Datasets    map[record.Type][]record.Record
			Result      lineage.AdjacencyMap
			QueryCfg    lineage.QueryCfg
			BuildErr    error
			QueryErr    error
		}
		var testCases = []testCase{
			{
				Description: "smoke test",
				Datasets: map[record.Type][]record.Record{
					record.TypeTable: {
						{
							Urn:     "1",
							Service: "service-A",
							Type:    record.TypeTable,
						},
						{
							Urn:     "2",
							Type:    record.TypeTable,
							Service: "service-A",
						},
					},
				},
				Result: lineage.AdjacencyMap{
					"table/1": adjEntryWithTypeAndURN(record.TypeTable, "1", "service-A"),
					"table/2": adjEntryWithTypeAndURN(record.TypeTable, "2", "service-A"),
				},
			},
			{
				// tests that the builder is able to use type.lineage to populate
				// related records
				Description: "internal ref test (simple)",
				Datasets: map[record.Type][]record.Record{
					record.TypeJob: {
						{
							Urn:     "1",
							Service: "service-A",
							Type:    record.TypeJob,
							Upstreams: []record.LineageRecord{
								{
									Urn:  "A",
									Type: record.TypeDashboard,
								},
								{
									Urn:  "B",
									Type: record.TypeDashboard,
								},
							},
							Downstreams: []record.LineageRecord{
								{
									Urn:  "C",
									Type: record.TypeTopic,
								},
							},
						},
					},
				},
				Result: lineage.AdjacencyMap{
					"job/1": lineage.AdjacencyEntry{
						Type:        record.TypeJob,
						URN:         "1",
						Service:     "service-A",
						Downstreams: set.NewStringSet("topic/C"),
						Upstreams:   set.NewStringSet("dashboard/A", "dashboard/B"),
					},
				},
			},
			{
				Description: "external ref test",
				Datasets: map[record.Type][]record.Record{
					record.TypeTopic: {
						{
							Urn:  "data-booking",
							Type: record.TypeTopic,
						},
					},
					record.TypeTable: {
						{
							Urn:  "booking-aggregator",
							Type: record.TypeTable,
							Upstreams: []record.LineageRecord{
								{
									Urn:  "data-booking",
									Type: record.TypeTopic,
								},
							},
						},
						{
							Urn:  "booking-fraud-detector",
							Type: record.TypeTable,
							Upstreams: []record.LineageRecord{
								{
									Urn:  "data-booking",
									Type: record.TypeTopic,
								},
							},
						},
					},
				},
				Result: lineage.AdjacencyMap{
					"topic/data-booking": lineage.AdjacencyEntry{
						Type:        record.TypeTopic,
						URN:         "data-booking",
						Upstreams:   set.NewStringSet(),
						Downstreams: set.NewStringSet("table/booking-aggregator", "table/booking-fraud-detector"),
					},
					"table/booking-aggregator": lineage.AdjacencyEntry{
						Type:        record.TypeTable,
						URN:         "booking-aggregator",
						Upstreams:   set.NewStringSet("topic/data-booking"),
						Downstreams: set.NewStringSet(),
					},
					"table/booking-fraud-detector": lineage.AdjacencyEntry{
						Type:        record.TypeTable,
						URN:         "booking-fraud-detector",
						Upstreams:   set.NewStringSet("topic/data-booking"),
						Downstreams: set.NewStringSet(),
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Description, func(t *testing.T) {
				rrf := initialiseRepos(tc.Datasets)
				graph, err := lineage.DefaultBuilder.Build(context.Background(), rrf)
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
