package lineage

import "context"

type MetricsMonitor interface {
	Duration(string, int)
}

type dummyMetricMonitor struct{}

func (c dummyMetricMonitor) Duration(string, int) {}

type PerformanceMonitor interface {
	StartTransaction(ctx context.Context, operation string) (context.Context, func())
}

type dummyPerformanceMonitor struct{}

func (c dummyPerformanceMonitor) StartTransaction(ctx context.Context, operation string) (context.Context, func()) {
	return ctx, func() {}
}
