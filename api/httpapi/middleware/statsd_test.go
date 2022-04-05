package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/metrics"
	"github.com/stretchr/testify/require"
)

func TestStatsD(t *testing.T) {
	var (
		statsdPrefix     = "columbusApi"
		metricsSeparator = "."
	)
	t.Run("MonitorRouter", func(t *testing.T) {
		statsdClient := new(mocks.StatsdClient)
		statsdClient.EXPECT().Increment("columbusApi.responseStatusCode,statusCode=200,method=POST,url=/").Once()
		statsdClient.EXPECT().Timing("columbusApi.responseTime,method=POST,url=/", int64(0)).Once()
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
}
