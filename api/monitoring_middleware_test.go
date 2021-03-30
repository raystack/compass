package api_test

import (
	"net/http/httptest"
	"testing"

	"github.com/odpf/columbus/api"
	"github.com/odpf/columbus/metrics"
)

func TestMonitoringHandler(t *testing.T) {
	statsdPrefix := "a_prefix"
	metricsSeparator := "."
	statsdClient := metrics.NewStatsdClient("127.0.0.1:8125")
	metricsMonitor := metrics.NewMonitor(statsdClient, statsdPrefix, metricsSeparator)

	handler := api.MonitoringHandler(api.NewHeartbeatHandler(), metricsMonitor)
	rr := httptest.NewRequest("GET", "/ping", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, rr)

	if rw.Code != 200 {
		t.Errorf("expected handler to respond with HTTP 200, got HTTP %d instead", rw.Code)
		return
	}
	expectedResponse := "pong"
	actualResponse := rw.Body.String()
	if actualResponse != expectedResponse {
		t.Errorf("expected handler response to be %q, was %q instead", expectedResponse, actualResponse)
		return
	}
}
