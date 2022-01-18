package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotFound, mux.ErrNotFound.Error())
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusMethodNotAllowed, mux.ErrMethodMismatch.Error())
}
