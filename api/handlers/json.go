package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func internalServerError(w http.ResponseWriter, logger logrus.FieldLogger, msg string) error {
	ref := time.Now().Unix()

	logger.Errorf("ref (%d): %s", ref, msg)
	response := &ErrorResponse{
		Reason: fmt.Sprintf(
			"%s - ref (%d)",
			http.StatusText(http.StatusInternalServerError),
			ref,
		),
	}

	return writeJSON(w, http.StatusInternalServerError, response)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) error {
	response := &ErrorResponse{
		Reason: msg,
	}
	return writeJSON(w, status, response)
}
