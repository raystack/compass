package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/odpf/columbus/api/handlers"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	dummyRoute             = "/v1beta1/dummy"
	identityHeader         = "Columbus-User-Email"
	identityProviderHeader = "Columbus-User-Provider"
)

func TestValidateUser(t *testing.T) {
	middlewareCfg := Config{
		Logger:         log.NewNoop(),
		IdentityHeader: identityHeader,
	}

	t.Run("should return HTTP 400 when identity header not present", func(t *testing.T) {
		userSvc := user.NewService(nil)
		r := mux.NewRouter()
		r.Use(ValidateUser(middlewareCfg, userSvc))
		r.Path(dummyRoute).Methods(http.MethodGet)

		req, _ := http.NewRequest("GET", dummyRoute, nil)

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		response := &handlers.ErrorResponse{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
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

		userSvc := user.NewService(mockUserRepository)
		r := mux.NewRouter()
		r.Use(ValidateUser(middlewareCfg, userSvc))
		r.Path(dummyRoute).Methods(http.MethodGet)

		req, _ := http.NewRequest("GET", dummyRoute, nil)
		req.Header.Set(identityHeader, "some-email")
		req.Header.Set(identityProviderHeader, "some-provider")

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		response := &handlers.ErrorResponse{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, customError.Error(), response.Reason)
	})

	t.Run("should return HTTP 200 with propagated user ID and email when user validation success", func(t *testing.T) {
		userID := "user-id"
		userEmail := "some-email"
		mockUserRepository := &mocks.UserRepository{}
		mockUserRepository.On("GetID", mock.Anything, mock.Anything).Return(userID, nil)
		mockUserRepository.On("Create", mock.Anything, mock.Anything).Return(userID, nil)

		userSvc := user.NewService(mockUserRepository)
		r := mux.NewRouter()
		r.Use(ValidateUser(middlewareCfg, userSvc))
		r.Path(dummyRoute).Methods(http.MethodGet).HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			propagatedUserID := user.IDFromContext(r.Context())
			propagatedUserEmail := user.EmailFromContext(r.Context())
			_, err := rw.Write([]byte(propagatedUserID + propagatedUserEmail))
			if err != nil {
				t.Fatal(err)
			}
			rw.WriteHeader(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", dummyRoute, nil)
		req.Header.Set(identityHeader, userEmail)
		req.Header.Set(identityProviderHeader, "some-provider")

		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, userID+userEmail, rr.Body.String())
	})
}
