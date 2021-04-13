package lineage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/odpf/columbus/models"
	"github.com/stretchr/testify/assert"
)

type roundRobinTimeSource struct {
	Index  int
	Mu     sync.Mutex
	Values []time.Time
}

func (ts *roundRobinTimeSource) Now() time.Time {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()
	index := ts.Index
	ts.Index = (ts.Index + 1) % len(ts.Values)
	return ts.Values[index]
}

type roundRobinBuilder struct {
	Graphs []Graph
	Index  int
}

func (builder *roundRobinBuilder) Build(ctx context.Context, er models.TypeRepository, rrf models.RecordRepositoryFactory) (Graph, error) {
	index := builder.Index
	builder.Index = (builder.Index + 1) % len(builder.Graphs)
	return builder.Graphs[index], nil
}

func TestService(t *testing.T) {
	t.Run("test caching", func(t *testing.T) {
		now := time.Now()
		builder := &roundRobinBuilder{
			Graphs: []Graph{
				new(InMemoryGraph),
				new(InMemoryGraph),
			},
		}
		ts := &roundRobinTimeSource{
			Values: []time.Time{
				now,                                // first build()
				now,                                // end timestamp
				now,                                // first test
				now.Add(time.Second * 30),          // second test
				now.Add(time.Minute + time.Second), // third test
				now.Add(time.Minute + time.Second), // async build() complete
				now.Add(time.Minute + time.Second), // + end ts
				now.Add(time.Minute + time.Second), // fourth test
			},
		}
		srv := &Service{
			timeSource:         ts,
			builder:            builder,
			refreshInterval:    time.Minute,
			lastBuilt:          now,
			metricsMonitor:     dummyMetricMonitor{},
			performanceMonitor: dummyPerformanceMonitor{},
		}
		srv.build()

		// first test, lastBuilt == now
		graph, err := srv.Graph()
		assert.Nil(t, err)
		assert.Equal(t, builder.Graphs[0], graph)

		// second test, lastBuilt = old now, now = old now + 30s
		graph, err = srv.Graph()
		assert.Nil(t, err)
		assert.Equal(t, builder.Graphs[0], graph)

		// third test, lastBuilt = old now, now = old now + 1m1s, triggers refresh
		graph, err = srv.Graph()
		assert.Nil(t, err)
		assert.Equal(t, builder.Graphs[0], graph)

		// wait for build() to finish
		ts.Mu.Lock()
		timeout := time.Now().Add(time.Second * 10)
		for ts.Index != 5 {
			if time.Now().After(timeout) {
				panic("internal error: timed out waiting for test condition")
			}
			time.Sleep(time.Second / 30)
		}
		ts.Mu.Unlock()

		// fourth test, lastBuilt = now, now = old now + 1m1s, update completed
		graph, err = srv.Graph()
		assert.Nil(t, err)
		assert.Equal(t, builder.Graphs[1], graph)
	})
}
