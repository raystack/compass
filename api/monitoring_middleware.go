package api

import (
	"net/http"
	"time"

	"github.com/odpf/columbus/metrics"
)

type interceptedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func responseWriter(w http.ResponseWriter) *interceptedResponseWriter {
	return &interceptedResponseWriter{w, http.StatusOK}
}

func (lrw *interceptedResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// MonitoringHandler - middleware that intercepts the response and pushes
// data like response_time, status code etc to statsd
func MonitoringHandler(h http.Handler, metricsMonitor metrics.Monitor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := responseWriter(w)
		h.ServeHTTP(rw, r)

		metricsMonitor.ResponseTime(r.Method, r.URL.Path, int64(time.Since(start)/time.Millisecond))
		metricsMonitor.ResponseStatus(r.Method, r.URL.Path, rw.statusCode)
	})
}
