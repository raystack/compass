package metrics_test

import (
	"fmt"
	"testing"

	"github.com/odpf/columbus/metrics"
)

func TestMonitoringHandler(t *testing.T) {

	tag := metrics.Tags{"POST", "/ping"}

	expectedResponse := "method=POST,url=/ping"
	actualResponse := fmt.Sprintf("%s", tag)
	if actualResponse != expectedResponse {
		t.Errorf("expected tag to be %q, was %q instead", expectedResponse, actualResponse)
		return
	}
}
