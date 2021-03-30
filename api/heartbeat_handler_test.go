package api_test

import (
	"net/http/httptest"
	"testing"

	"github.com/odpf/columbus/api"
)

func TestHeartbeatHandler(t *testing.T) {
	handler := api.NewHeartbeatHandler()
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
