package middleware

import (
	"errors"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/columbus/api/httpapi/handlers"
	"github.com/odpf/columbus/user"
)

// ValidateUser middleware will propagate a valid user ID as string
// within request context
// use `user.FromContext` function to get the user ID string
func ValidateUser(identityUUIDHeaderKey, identityEmailHeaderKey string, userSvc *user.Service, h runtime.HandlerFunc) runtime.HandlerFunc {
	return runtime.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		userUUID := r.Header.Get(identityUUIDHeaderKey)
		if userUUID == "" {
			handlers.WriteJSONError(rw, http.StatusBadRequest, "identity header uuid is empty")
			return
		}
		userEmail := r.Header.Get(identityEmailHeaderKey)
		userID, err := userSvc.ValidateUser(r.Context(), userUUID, userEmail)
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
		h(rw, r, pathParams)
	})
}
