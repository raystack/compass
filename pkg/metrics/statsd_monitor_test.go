package metrics_test

import (
	"testing"

	"github.com/odpf/compass/pkg/metrics"
	"github.com/odpf/compass/pkg/metrics/mocks"
)

func TestNewStatsDMonitor(t *testing.T) {
	var (
		statsdPrefix     = "compassApi"
		metricsSeparator = "."
	)

	t.Run("MonitorLineage", func(t *testing.T) {
		operationName := "build"
		duration := 100
		statsdClient := new(mocks.StatsDClient)
		statsdClient.EXPECT().Timing("compassApi.duration,operation=build", int64(duration)).Once()

		monitor := metrics.NewStatsDMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.Duration(operationName, duration)

		statsdClient.AssertExpectations(t)
	})
}
