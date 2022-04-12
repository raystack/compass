package metrics_test

import (
	"testing"

	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/metrics"
)

func TestNewStatsdMonitor(t *testing.T) {
	var (
		statsdPrefix     = "columbusApi"
		metricsSeparator = "."
	)

	t.Run("MonitorLineage", func(t *testing.T) {
		operationName := "build"
		duration := 100
		statsdClient := new(mocks.StatsdClient)
		statsdClient.EXPECT().Timing("columbusApi.duration,operation=build", int64(duration)).Once()

		monitor := metrics.NewStatsdMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.Duration(operationName, duration)

		statsdClient.AssertExpectations(t)
	})
}
