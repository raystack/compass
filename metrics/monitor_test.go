package metrics_test

import (
	"testing"

	"github.com/odpf/columbus/metrics"
	"github.com/stretchr/testify/mock"
)

type MockMetricsClient struct {
	mock.Mock
}

func (c *MockMetricsClient) Timing(key string, val int64) {
	c.Called(key, val)
}

func (c *MockMetricsClient) Increment(key string) {
	c.Called(key)
}

func TestNewMonitor(t *testing.T) {
	var (
		statsdPrefix     = "columbusApi"
		metricsSeparator = "."
		statsdClient     = &MockMetricsClient{}
	)

	statsdClient.On("Increment", "columbusApi.responseStatusCode,statusCode=200,method=POST,url=/").Once()
	statsdClient.On("Timing", "columbusApi.responseTime,method=POST,url=/", int64(100)).Once()
	statsdClient.On("Timing", "columbusApi.duration,operation=build", int64(100)).Once()

	metricsMonitor := metrics.NewMonitor(statsdClient, statsdPrefix, metricsSeparator)
	metricsMonitor.ResponseStatus("POST", "/", 200)
	metricsMonitor.ResponseTime("POST", "/", int64(100))
	metricsMonitor.Duration("build", int64(100))

	statsdClient.AssertExpectations(t)
}
