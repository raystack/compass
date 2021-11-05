package api

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func decodeURLMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		newVars := map[string]string{}
		for key, val := range mux.Vars(r) {
			decodedValue, _ := url.QueryUnescape(val)
			newVars[key] = decodedValue
		}

		r = mux.SetURLVars(r, newVars)
		h.ServeHTTP(rw, r)
	})
}
