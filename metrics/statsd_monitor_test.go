package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/metrics"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestNewStatsdMonitor(t *testing.T) {
	var (
		statsdPrefix     = "columbusApi"
		metricsSeparator = "."
	)
	t.Run("MonitorRouter", func(t *testing.T) {
		statsdClient := &MockMetricsClient{}
		statsdClient.On("Increment", "columbusApi.responseStatusCode,statusCode=200,method=POST,url=/").Once()
		statsdClient.On("Timing", "columbusApi.responseTime,method=POST,url=/", int64(0)).Once()
		monitor := metrics.NewStatsdMonitor(statsdClient, statsdPrefix, metricsSeparator)
		router := mux.NewRouter()
		monitor.MonitorRouter(router)
		handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			_, err := res.Write([]byte(""))
			require.NoError(t, err)
		})
		router.Path("/").HandlerFunc(handler)

		rr := httptest.NewRequest("POST", "/", nil)
		rw := httptest.NewRecorder()
		router.ServeHTTP(rw, rr)

		statsdClient.AssertExpectations(t)
	})

	t.Run("MonitorLineage", func(t *testing.T) {
		operationName := "build"
		duration := 100
		statsdClient := &MockMetricsClient{}
		statsdClient.On("Timing", "columbusApi.duration,operation=build", int64(duration)).Once()

		monitor := metrics.NewStatsdMonitor(statsdClient, statsdPrefix, metricsSeparator)
		monitor.Duration(operationName, duration)

		statsdClient.AssertExpectations(t)
	})
}
