package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/columbus/api/httpapi/handlers"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	dummyRoute        = "/v1beta1/dummy"
	identityHeaderKey = "Columbus-User-ID"
)

var userCfg = user.Config{IdentityProviderDefaultName: "shield"}

func TestValidateUser(t *testing.T) {

	t.Run("should return HTTP 400 when identity header not present", func(t *testing.T) {
		userSvc := user.NewService(nil, userCfg)

		r := runtime.NewServeMux()
		err := r.HandlePath(http.MethodGet, dummyRoute,
			ValidateUser(identityHeaderKey, userSvc, nil))
		if err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest("GET", dummyRoute, nil)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		response := &handlers.ErrorResponse{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "identity header is empty", response.Reason)
	})

	t.Run("should return HTTP 500 when something error with user service", func(t *testing.T) {
		customError := errors.New("some error")
		mockUserRepository := &mocks.UserRepository{}
		mockUserRepository.On("GetID", mock.Anything, mock.Anything).Return("", customError)
		mockUserRepository.On("Create", mock.Anything, mock.Anything).Return("", customError)

		userSvc := user.NewService(mockUserRepository, userCfg)

		r := runtime.NewServeMux()
		err := r.HandlePath(http.MethodGet, dummyRoute,
			ValidateUser(identityHeaderKey, userSvc, nil))
		if err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest("GET", dummyRoute, nil)
		req.Header.Set(identityHeaderKey, "some-email")
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		response := &handlers.ErrorResponse{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, customError.Error(), response.Reason)
	})

	t.Run("should return HTTP 200 with propagated user ID when user validation success", func(t *testing.T) {
		userID := "user-id"
		userEmail := "some-email"
		mockUserRepository := &mocks.UserRepository{}
		mockUserRepository.On("GetID", mock.Anything, mock.Anything).Return(userID, nil)
		mockUserRepository.On("Create", mock.Anything, mock.Anything).Return(userID, nil)

		userSvc := user.NewService(mockUserRepository, userCfg)

		r := runtime.NewServeMux()
		if err := r.HandlePath(http.MethodGet, dummyRoute,
			ValidateUser(identityHeaderKey, userSvc, runtime.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				propagatedUserID := user.FromContext(r.Context())
				_, err := rw.Write([]byte(propagatedUserID))
				if err != nil {
					t.Fatal(err)
				}
				rw.WriteHeader(http.StatusOK)
			}))); err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest("GET", dummyRoute, nil)
		req.Header.Set(identityHeaderKey, userEmail)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, userID, rr.Body.String())
	})
}
