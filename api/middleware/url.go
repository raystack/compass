package middleware

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func DecodeURL(cfg Config) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			newVars := map[string]string{}
			for key, val := range mux.Vars(r) {
				decodedVal, err := url.QueryUnescape(val)
				if err != nil {
					cfg.Logger.Warn("error decoding url", "value", val)
					decodedVal = val
				}

				newVars[key] = decodedVal
			}
			r = mux.SetURLVars(r, newVars)
			h.ServeHTTP(rw, r)
		})
	}
}
