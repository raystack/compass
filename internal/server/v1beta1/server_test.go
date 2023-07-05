package handlersv1beta1_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/goto/compass/core/user"
	handlersv1beta1 "github.com/goto/compass/internal/server/v1beta1"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"github.com/goto/salt/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateUserInCtx(t *testing.T) {
	var (
		userUUID = uuid.NewString()
		userID   = uuid.NewString()
	)
	type testCase struct {
		Description  string
		UserID       string
		UserUUID     string
		ExpectStatus codes.Code
		Setup        func(context.Context, *mocks.UserService)
		PostCheck    func(resp *compassv1beta1.GetUserStarredAssetsResponse) error
	}

	testCases := []testCase{
		{
			Description:  "should return invalid argument error if ValidateUser empty uuid is passed",
			UserID:       "",
			UserUUID:     "",
			ExpectStatus: codes.InvalidArgument,
			Setup: func(ctx context.Context, us *mocks.UserService) {
				us.EXPECT().ValidateUser(ctx, "", "").Return("", user.ErrNoUserInformation)
			},
		},
		{
			Description:  "should return internal error if ValidateUser returns some error",
			UserID:       userID,
			UserUUID:     userUUID,
			ExpectStatus: codes.Internal,
			Setup: func(ctx context.Context, us *mocks.UserService) {
				us.EXPECT().ValidateUser(ctx, userUUID, "").Return("", errors.New("internal error"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := user.NewContext(context.Background(), user.User{UUID: tc.UserUUID})

			logger := log.NewNoop()

			mockUserSvc := mocks.NewUserService(t)
			mockStarSvc := mocks.NewStarService(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockUserSvc)
			}

			handler := handlersv1beta1.NewAPIServer(handlersv1beta1.APIServerDeps{StarSvc: mockStarSvc, UserSvc: mockUserSvc, Logger: logger})

			_, err := handler.ValidateUserInCtx(ctx)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %sinstead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}
