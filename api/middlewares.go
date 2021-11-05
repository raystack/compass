package api

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func decodeURLMiddleware(logger logrus.FieldLogger) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			newVars := map[string]string{}
			for key, val := range mux.Vars(r) {
				decodedVal, err := url.QueryUnescape(val)
				if err != nil {
					logger.Warnf("error decoding url %s", val)
					decodedVal = val
				}

				newVars[key] = decodedVal
			}

			r = mux.SetURLVars(r, newVars)
			h.ServeHTTP(rw, r)
		})
	}
}
