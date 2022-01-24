package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/odpf/salt/log"
	"net/http"
	"time"
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

func internalServerError(w http.ResponseWriter, logger log.Logger, msg string) {
	ref := time.Now().Unix()

	logger.Error("time format", "ref", ref, "msg", msg)
	response := &ErrorResponse{
		Reason: fmt.Sprintf(
			"%s - ref (%d)",
			http.StatusText(http.StatusInternalServerError),
			ref,
		),
	}

	writeJSON(w, http.StatusInternalServerError, response)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	response := &ErrorResponse{
		Reason: msg,
	}

	writeJSON(w, status, response)
}
