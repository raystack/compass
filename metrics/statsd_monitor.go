package metrics

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	statsd "github.com/etsy/statsd/examples/go"
	"github.com/gorilla/mux"
)

type statsdClient interface {
	Timing(string, int64)
	Increment(string)
}

func NewStatsdClient(statsdAddress string) *statsd.StatsdClient {
	statsdHost, statsdPortStr, _ := net.SplitHostPort(statsdAddress)
	statsdPort, _ := strconv.Atoi(statsdPortStr)
	return statsd.New(statsdHost, statsdPort)
}

type StatsdMonitor struct {
	client    statsdClient
	prefix    string
	separator string
}

func NewStatsdMonitor(client statsdClient, prefix string, separator string) *StatsdMonitor {
	return &StatsdMonitor{
		client:    client,
		prefix:    prefix,
		separator: separator,
	}
}

func (mm *StatsdMonitor) MonitorRouter(router *mux.Router) {
	router.Use(mm.routerMiddleware)
}

func (mm *StatsdMonitor) Duration(operation string, duration int) {
	statName := fmt.Sprintf("%s%s%s,operation=%s", mm.prefix, mm.separator, "duration", operation)
	mm.client.Timing(statName, int64(duration))
}

func (mm *StatsdMonitor) routerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := responseWriter(w)
		h.ServeHTTP(rw, r)

		mm.responseTime(r.Method, r.URL.Path, int64(time.Since(start)/time.Millisecond))
		mm.responseStatus(r.Method, r.URL.Path, rw.statusCode)
	})
}

func (mm *StatsdMonitor) responseTime(requestMethod string, requestUrl string, responseTime int64) {
	statName := fmt.Sprintf("%s%s%s,%s",
		mm.prefix,
		mm.separator,
		"responseTime",
		Tags{requestMethod, requestUrl})
	mm.client.Timing(statName, responseTime)
}

func (mm *StatsdMonitor) responseStatus(requestMethod string, requestUrl string, responseCode int) {
	statName := fmt.Sprintf("%s%s%s,statusCode=%d,%s",
		mm.prefix,
		mm.separator,
		"responseStatusCode",
		responseCode,
		Tags{requestMethod, requestUrl})
	mm.client.Increment(statName)
}

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
