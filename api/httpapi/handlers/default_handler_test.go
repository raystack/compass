package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/odpf/columbus/api/httpapi/handlers"
)

func TestNotFoundHandler(t *testing.T) {
	handler := http.HandlerFunc(handlers.NotFound)
	rr := httptest.NewRequest("GET", "/xxx", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, rr)
	if rw.Code != 404 {
		t.Errorf("expected handler to respond with HTTP 404, got HTTP %d instead", rw.Code)
		return
	}
	expectedResponse := "{\"reason\":\"no matching route was found\"}\n"
	actualResponse := rw.Body.String()
	if actualResponse != expectedResponse {
		t.Errorf("expected handler response to be %q, was %q instead", expectedResponse, actualResponse)
		return
	}
}

func TestMethodNotAllowedHandler(t *testing.T) {
	handler := http.HandlerFunc(handlers.MethodNotAllowed)
	rr := httptest.NewRequest("POST", "/ping", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, rr)
	if rw.Code != 405 {
		t.Errorf("expected handler to respond with HTTP 405, got HTTP %d instead", rw.Code)
		return
	}
	expectedResponse := "{\"reason\":\"method is not allowed\"}\n"
	actualResponse := rw.Body.String()
	if actualResponse != expectedResponse {
		t.Errorf("expected handler response to be %q, was %q instead", expectedResponse, actualResponse)
		return
	}
}
