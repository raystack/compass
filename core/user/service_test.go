package user_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/namespace"
	"testing"

	"github.com/raystack/compass/core/user"
	"github.com/raystack/compass/core/user/mocks"
	"github.com/raystack/compass/pkg/statsd"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/assert"
)

func TestValidateUser(t *testing.T) {
	ns := &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "tenant",
		State:    namespace.SharedState,
		Metadata: nil,
	}
	type testCase struct {
		Description string
		UUID        string
		Email       string
		Setup       func(ctx context.Context, inputUUID, inputEmail string, userRepo *mocks.UserRepository)
		ExpectErr   error
	}

	var testCases = []testCase{
		{
			Description: "should return no user error when uuid is empty and email is optional",
			UUID:        "",
			ExpectErr:   user.ErrNoUserInformation,
		},
		{
			Description: "should return user UUID when successfully found user UUID from DB",
			UUID:        "a-uuid",
			Setup: func(ctx context.Context, inputUUID, inputEmail string, userRepo *mocks.UserRepository) {
				userRepo.EXPECT().GetByUUID(ctx, inputUUID).Return(user.User{ID: "user-id", UUID: inputUUID}, nil)
			},
			ExpectErr: nil,
		},
		{
			Description: "should return user error if GetByUUID return nil error and empty ID",
			UUID:        "a-uuid",
			Setup: func(ctx context.Context, inputUUID, inputEmail string, userRepo *mocks.UserRepository) {
				userRepo.EXPECT().GetByUUID(ctx, inputUUID).Return(user.User{}, nil)
			},
			ExpectErr: errors.New("fetched user uuid from DB is empty"),
		},
		{
			Description: "should return user UUID when not found user UUID from DB but can create the new one without email",
			UUID:        "an-email-success-create",
			Setup: func(ctx context.Context, inputUUID, inputEmail string, userRepo *mocks.UserRepository) {
				userRepo.EXPECT().GetByUUID(ctx, inputUUID).Return(user.User{}, user.NotFoundError{UUID: inputUUID})
				userRepo.EXPECT().UpsertByEmail(ctx, ns, &user.User{UUID: inputUUID}).Return("some-id", nil)
			},
			ExpectErr: nil,
		},
		{
			Description: "should return user UUID when not found user UUID from DB but can create the new one wit email",
			UUID:        "an-uuid-error",
			Email:       "an-email-success-create",
			Setup: func(ctx context.Context, inputUUID, inputEmail string, userRepo *mocks.UserRepository) {
				userRepo.EXPECT().GetByUUID(ctx, inputUUID).Return(user.User{}, user.NotFoundError{UUID: inputUUID})
				userRepo.EXPECT().UpsertByEmail(ctx, ns, &user.User{UUID: inputUUID, Email: inputEmail}).Return("some-id", nil)
			},
			ExpectErr: nil,
		},
		{
			Description: "should return error when not found user ID from DB but can't create the new one",
			UUID:        "an-uuid-error",
			Email:       "an-email",
			Setup: func(ctx context.Context, inputUUID, inputEmail string, userRepo *mocks.UserRepository) {
				mockErr := errors.New("error upserting user")
				userRepo.EXPECT().GetByUUID(ctx, inputUUID).Return(user.User{}, mockErr)
				userRepo.EXPECT().UpsertByEmail(ctx, ns, &user.User{UUID: inputUUID, Email: inputEmail}).Return("", mockErr)
			},
			ExpectErr: errors.New("error upserting user"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.TODO()
			logger := log.NewNoop()
			mockUserRepo := new(mocks.UserRepository)

			if tc.Setup != nil {
				tc.Setup(ctx, tc.UUID, tc.Email, mockUserRepo)
			}

			userSvc := user.NewService(logger, mockUserRepo)

			_, err := userSvc.ValidateUser(ctx, ns, tc.UUID, tc.Email)

			assert.Equal(t, tc.ExpectErr, err)
		})
	}
}

func TestServiceWithStatsDResporter(t *testing.T) {
	t.Run("should create statsDReport for a service", func(t *testing.T) {
		usr := &user.Service{}
		s := user.ServiceWithStatsDReporter(&statsd.Reporter{})
		s(usr)
	})
}
