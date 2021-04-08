package lineage

type MetricsMonitor interface {
	Duration(string, int)
}

type dummyMetricMonitor struct{}

func (c dummyMetricMonitor) Duration(string, int) {}
