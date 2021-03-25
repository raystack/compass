package lineage

type dummyMetricMonitor struct{}

func (c dummyMetricMonitor) Duration(string, int64) {}
