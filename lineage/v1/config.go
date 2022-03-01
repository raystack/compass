package lineage

type Config struct {
	RefreshInterval    string
	MetricsMonitor     MetricsMonitor
	PerformanceMonitor PerformanceMonitor
	Builder            Builder
	TimeSource         TimeSource
}
