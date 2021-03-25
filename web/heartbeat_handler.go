package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// HeartbeatHandler responds to any GET requests an HTTP 200 OK
// This is used an an indicator to determine whether the service is up or not
type HeartbeatHandler struct {
	mux *mux.Router
}

func (handler *HeartbeatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mux.ServeHTTP(w, r)
}

func (handler *HeartbeatHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func NewHeartbeatHandler() *HeartbeatHandler {
	handler := &HeartbeatHandler{
		mux: mux.NewRouter(),
	}
	handler.mux.HandleFunc("/ping", handler.Heartbeat)
	return handler
}
