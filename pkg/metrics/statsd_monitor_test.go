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

	t.Run("Duration", func(t *testing.T) {
		operationName := "build"
		duration := 100
		statsdClient := new(mocks.StatsDClient)
		statsdClient.EXPECT().Timing("compassApi.duration,operation=build", int64(duration)).Once()

		monitor := metrics.NewStatsDMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.Duration(operationName, duration)

		statsdClient.AssertExpectations(t)
	})

	t.Run("ResponseTime", func(t *testing.T) {
		requestMethod := "somemethod"
		responseTime := int64(100)
		requestURL := "requesturl"
		statsdClient := new(mocks.StatsDClient)
		statsdClient.EXPECT().Timing("compassApi.responseTime,method=somemethod,url=requesturl", responseTime).Once()

		monitor := metrics.NewStatsDMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.ResponseTime(requestMethod, requestURL, responseTime)

		statsdClient.AssertExpectations(t)
	})

	t.Run("ResponseStatus", func(t *testing.T) {
		requestMethod := "somemethod"
		responseStatusCode := 200
		requestURL := "requesturl"
		statsdClient := new(mocks.StatsDClient)
		statsdClient.EXPECT().Increment("compassApi.responseStatusCode,statusCode=200,method=somemethod,url=requesturl").Once()

		monitor := metrics.NewStatsDMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.ResponseStatus(requestMethod, requestURL, responseStatusCode)

		statsdClient.AssertExpectations(t)
	})

	t.Run("ResponseTimeGRPC", func(t *testing.T) {
		fullMethod := "somemethod/grpc"
		responseTime := int64(100)
		statsdClient := new(mocks.StatsDClient)
		statsdClient.EXPECT().Timing("compassApi.responseTime,method=somemethod/grpc", responseTime).Once()

		monitor := metrics.NewStatsDMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.ResponseTimeGRPC(fullMethod, responseTime)

		statsdClient.AssertExpectations(t)
	})

	t.Run("ResponseStatusGRPC", func(t *testing.T) {
		fullMethod := "somemethod/grpc"
		responseStatusCode := "OK"
		statsdClient := new(mocks.StatsDClient)
		statsdClient.EXPECT().Increment("compassApi.responseStatusCode,statusCode=OK,method=somemethod/grpc").Once()

		monitor := metrics.NewStatsDMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.ResponseStatusGRPC(fullMethod, responseStatusCode)

		statsdClient.AssertExpectations(t)
	})
}
