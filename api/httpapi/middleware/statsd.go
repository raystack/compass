package middleware

import (
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/compass/metrics"
)

func StatsD(mm *metrics.StatsdMonitor, h runtime.HandlerFunc) runtime.HandlerFunc {
	return runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		if mm == nil {
			h(w, r, pathParams)
			return
		}
		start := time.Now()
		rw := responseWriter(w)
		h(w, r, pathParams)
		mm.ResponseTime(r.Method, r.URL.Path, int64(time.Since(start)/time.Millisecond))
		mm.ResponseStatus(r.Method, r.URL.Path, rw.statusCode)
	})
}

func responseWriter(w http.ResponseWriter) *interceptedResponseWriter {
	return &interceptedResponseWriter{w, http.StatusOK}
}

type interceptedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *interceptedResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
