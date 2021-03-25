package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	var payload = map[string]string{
		"foo": "bar",
	}
	rw := httptest.NewRecorder()
	writeJSON(rw, http.StatusOK, payload)

	expectContentType := "application/json"
	actualContentType := rw.Result().Header.Get("content-type")

	if expectContentType != actualContentType {
		t.Errorf("expected 'content-type' of response to be %q, was %q", expectContentType, actualContentType)
		return
	}

	expectPayload := `{"foo":"bar"}`
	actualPayload := strings.TrimSpace(rw.Body.String())

	if expectPayload != actualPayload {
		t.Errorf("expected response to be %q, was %q instead", expectPayload, actualPayload)
		return
	}
}
