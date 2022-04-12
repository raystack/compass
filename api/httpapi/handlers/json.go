package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/odpf/salt/log"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	logger := log.NewLogrus()
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(status)
		code, err := w.Write([]byte("error encoding response to json"))
		if err != nil {
			logger.Info(fmt.Sprintf("error writing response with code: %d", code))
		}
	}
}

func WriteJSONError(w http.ResponseWriter, status int, msg string) {
	response := &ErrorResponse{
		Reason: msg,
	}

	writeJSON(w, status, response)
}
