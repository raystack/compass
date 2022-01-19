package api

import (
	"github.com/odpf/salt/log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func decodeURLMiddleware(log log.Logger) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			newVars := map[string]string{}
			for key, val := range mux.Vars(r) {
				decodedVal, err := url.QueryUnescape(val)
				if err != nil {
					log.Warn("error decoding url", "value", val)
					decodedVal = val
				}

				newVars[key] = decodedVal
			}
			r = mux.SetURLVars(r, newVars)
			h.ServeHTTP(rw, r)
		})
	}
}
