package middleware

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/httpapi/handlers"
	"github.com/odpf/columbus/user"
)

// ValidateUser middleware will propagate a valid user ID as string
// within request context
// use `user.FromContext` function to get the user ID string
func ValidateUser(identityHeaderKey string, userSvc *user.Service) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			userEmail := r.Header.Get(identityHeaderKey)
			if userEmail == "" {
				handlers.WriteJSONError(rw, http.StatusBadRequest, "identity header is empty")
				return
			}
			userID, err := userSvc.ValidateUser(r.Context(), userEmail)
			if err != nil {
				if errors.Is(err, user.ErrNoUserInformation) {
					handlers.WriteJSONError(rw, http.StatusBadRequest, err.Error())
					return
				}
				handlers.WriteJSONError(rw, http.StatusInternalServerError, err.Error())
				return
			}
			ctx := user.NewContext(r.Context(), userID)
			r = r.WithContext(ctx)
			h.ServeHTTP(rw, r)
		})
	}
}
