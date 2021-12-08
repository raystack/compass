package metrics_test

import (
	"testing"

	"github.com/odpf/columbus/metrics"
)

func TestMonitoringHandler(t *testing.T) {

	tag := metrics.Tags{"POST", "/ping"}

	expectedResponse := "method=POST,url=/ping"
	actualResponse := tag.String()
	if actualResponse != expectedResponse {
		t.Errorf("expected tag to be %q, was %q instead", expectedResponse, actualResponse)
		return
	}
}
