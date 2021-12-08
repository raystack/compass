package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(status)
		code, err := w.Write([]byte("error encoding response to json"))
		if err != nil {
			log.Print(fmt.Sprintf("error writing response with code: %d", code))
		}
	}
}

func internalServerError(w http.ResponseWriter, logger logrus.FieldLogger, msg string) {
	ref := time.Now().Unix()

	logger.Errorf("ref (%d): %s", ref, msg)
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
