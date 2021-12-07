package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	response := &ErrorResponse{
		Reason: msg,
	}

	writeJSON(w, status, response)
}
