package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	dummyRoute             = "/v1beta1/dummy"
	identityUUIDHeaderKey  = "Columbus-User-ID"
	identityEmailHeaderKey = "Columbus-User-Email"
	userUUID               = "user-uuid"
	userID                 = "user-id"
	userEmail              = "some-email"
)

func TestValidateUser(t *testing.T) {

	type testCase struct {
		Description  string
		Setup        func(ctx context.Context, userRepo *mocks.UserRepository, req *http.Request)
		Handler      runtime.HandlerFunc
		ExpectStatus int
	}

	var testCases = []testCase{
		{
			Description:  "should return HTTP 400 when identity header not present",
			ExpectStatus: http.StatusBadRequest,
		},
		{
			Description: "should return HTTP 500 when something error with user service",
			Setup: func(ctx context.Context, userRepo *mocks.UserRepository, req *http.Request) {
				req.Header.Set(identityUUIDHeaderKey, userUUID)
				req.Header.Set(identityEmailHeaderKey, userEmail)

				customError := errors.New("some error")
				userRepo.EXPECT().GetByUUID(mock.Anything, mock.Anything).Return(user.User{}, customError)
				userRepo.EXPECT().UpsertByEmail(mock.Anything, mock.Anything).Return("", customError)
			},
			ExpectStatus: http.StatusInternalServerError,
		},
		{
			Description: "should return HTTP 200 with propagated user ID when user validation success",
			Handler: func(rw http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				propagatedUserID := user.FromContext(r.Context())
				_, err := rw.Write([]byte(propagatedUserID))
				if err != nil {
					t.Fatal(err)
				}
				rw.WriteHeader(http.StatusOK)
			},
			Setup: func(ctx context.Context, userRepo *mocks.UserRepository, req *http.Request) {
				req.Header.Set(identityUUIDHeaderKey, userUUID)
				req.Header.Set(identityEmailHeaderKey, userEmail)

				userRepo.EXPECT().GetByUUID(mock.Anything, mock.Anything).Return(user.User{
					ID:    userID,
					UUID:  userUUID,
					Email: userEmail,
				}, nil)
			},
			ExpectStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()
			logger := log.NewNoop()
			userRepo := new(mocks.UserRepository)
			userSvc := user.NewService(logger, userRepo)

			r := runtime.NewServeMux()
			err := r.HandlePath(http.MethodGet, dummyRoute,
				ValidateUser(identityUUIDHeaderKey, identityEmailHeaderKey, userSvc, tc.Handler))
			if err != nil {
				t.Fatal(err)
			}

			req, _ := http.NewRequest("GET", dummyRoute, nil)
			rr := httptest.NewRecorder()

			if tc.Setup != nil {
				tc.Setup(ctx, userRepo, req)
			}

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.ExpectStatus, rr.Code)
		})
	}
}
