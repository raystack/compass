package lineage

type Config struct {
	RefreshInterval string
	MetricsMonitor  MetricsMonitor
	Builder         Builder
	TimeSource      TimeSource
}
