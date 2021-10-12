package lineage

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/odpf/columbus/models"
)

const (
	semUnlocked uint32 = iota
	semLocked
)

type TimeSource interface {
	Now() time.Time
}

type TimeSourceFunc func() time.Time

func (tsf TimeSourceFunc) Now() time.Time {
	return tsf()
}

// Service represents a high-level interface to the
// lineage package. Allows the client to configure
// an interval at which it'll construct the graph, while
// serving an old copy in between ticks.
type Service struct {
	typeRepo           models.TypeRepository
	recordRepoFactory  models.RecordRepositoryFactory
	metricsMonitor     MetricsMonitor
	performanceMonitor PerformanceMonitor
	builder            Builder
	timeSource         TimeSource

	refreshInterval time.Duration
	lastBuilt       time.Time
	sem             uint32 // semaphore
	mu              sync.RWMutex
	graph           Graph
	err             error
}

func (srv *Service) ForceBuild() {
	srv.build()
}

func (srv *Service) build() {
	ctx, endTxn := srv.performanceMonitor.StartTransaction(context.Background(), "lineage:Service/build")
	defer endTxn()

	startTime := srv.timeSource.Now()
	graph, err := srv.builder.Build(ctx, srv.typeRepo, srv.recordRepoFactory)
	now := srv.timeSource.Now()
	srv.metricsMonitor.Duration("lineageBuildTime", int(now.Sub(startTime)/time.Millisecond))
	srv.mu.Lock()
	defer srv.mu.Unlock()

	srv.graph = graph
	srv.err = err
	srv.lastBuilt = now
}

func (srv *Service) Graph() (Graph, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()
	srv.refreshIfNeeded()
	return srv.graph, srv.err
}

func (srv *Service) refreshIfNeeded() {
	threshold := srv.lastBuilt.Add(srv.refreshInterval)
	if srv.timeSource.Now().After(threshold) {
		srv.requestRefresh()
	}
}

func (srv *Service) requestRefresh() {
	// only one requestRefresh() call will be honored, and any and all following requests will be will be discarded until
	// the goroutine spawned by the former requestRefresh() is not finished.
	// WARN: do not touch this block of code unless you're _absolutely_ sure about what you're doing.
	if atomic.CompareAndSwapUint32(&srv.sem, semUnlocked, semLocked) {
		go func() {
			defer atomic.CompareAndSwapUint32(&srv.sem, semLocked, semUnlocked)
			srv.build()
		}()
	}
}

func NewService(er models.TypeRepository, rrf models.RecordRepositoryFactory, config Config) (*Service, error) {
	srv := &Service{
		builder:            DefaultBuilder,
		typeRepo:           er,
		recordRepoFactory:  rrf,
		refreshInterval:    time.Minute,
		timeSource:         TimeSourceFunc(time.Now),
		metricsMonitor:     dummyMetricMonitor{},
		performanceMonitor: &dummyPerformanceMonitor{},
		graph:              NewInMemoryGraph(AdjacencyMap{}),
	}

	err := applyConfig(srv, config)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func applyConfig(service *Service, config Config) error {
	refreshInterval := config.RefreshInterval
	if refreshInterval == "" {
		refreshInterval = "5m"
	}
	lineageRefreshInterval, err := time.ParseDuration(refreshInterval)
	if err != nil {
		return fmt.Errorf("error parsing lineage refresh interval: %v", err)
	}
	service.refreshInterval = lineageRefreshInterval

	if !isNilMonitor(config.MetricsMonitor) {
		service.metricsMonitor = config.MetricsMonitor
	}

	if !isNilMonitor(config.PerformanceMonitor) {
		service.performanceMonitor = config.PerformanceMonitor
	}

	if config.Builder != nil {
		service.builder = config.Builder
	}

	if config.TimeSource != nil {
		service.timeSource = config.TimeSource
	}

	return nil
}

func isNilMonitor(monitor interface{}) bool {
	v := reflect.ValueOf(monitor)
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
