package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, http.StatusNotFound, mux.ErrNotFound.Error())
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	WriteJSONError(w, http.StatusMethodNotAllowed, mux.ErrMethodMismatch.Error())
}
