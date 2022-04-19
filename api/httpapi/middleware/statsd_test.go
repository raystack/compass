package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/compass/lib/mocks"
	"github.com/odpf/compass/metrics"
	"github.com/stretchr/testify/require"
)

func TestStatsD(t *testing.T) {
	var (
		statsdPrefix     = "compassApi"
		metricsSeparator = "."
	)

	t.Run("StatsD should be called if not nil", func(t *testing.T) {
		statsdClient := new(mocks.StatsdClient)
		statsdClient.EXPECT().Increment("compassApi.responseStatusCode,statusCode=200,method=POST,url=/").Once()
		statsdClient.EXPECT().Timing("compassApi.responseTime,method=POST,url=/", int64(0)).Once()
		monitor := metrics.NewStatsdMonitor(statsdClient, statsdPrefix, metricsSeparator)
		router := runtime.NewServeMux()
		handler := runtime.HandlerFunc(func(res http.ResponseWriter, req *http.Request, pathParams map[string]string) {
			_, err := res.Write([]byte(""))
			require.NoError(t, err)
		})
		err := router.HandlePath(http.MethodPost, "/", StatsD(monitor, handler))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRequest("POST", "/", nil)
		rw := httptest.NewRecorder()
		router.ServeHTTP(rw, rr)

		statsdClient.AssertExpectations(t)
	})

	t.Run("Handlers should still be called if StatsD is nil", func(t *testing.T) {
		statsdClient := new(mocks.StatsdClient)
		router := runtime.NewServeMux()
		handler := runtime.HandlerFunc(func(res http.ResponseWriter, req *http.Request, pathParams map[string]string) {
			_, err := res.Write([]byte(""))
			require.NoError(t, err)
		})
		err := router.HandlePath(http.MethodPost, "/", StatsD(nil, handler))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRequest("POST", "/", nil)
		rw := httptest.NewRecorder()
		router.ServeHTTP(rw, rr)

		statsdClient.AssertExpectations(t)
	})
}
