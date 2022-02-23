package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/columbus/lib/mocks"
	"github.com/odpf/columbus/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var userCfg = user.Config{IdentityProviderDefaultName: "shield"}

func TestValidateWithHeader(t *testing.T) {
	ctx := context.TODO()
	t.Run("should return no user error when param is empty", func(t *testing.T) {
		userSvc := user.NewService(nil, userCfg)

		id, err := userSvc.ValidateUser(ctx, "")

		assert.ErrorIs(t, err, user.ErrNoUserInformation)
		assert.Empty(t, id)
	})

	t.Run("should return error when user id from DB is empty", func(t *testing.T) {
		mockUserRepository := &mocks.UserRepository{}
		mockUserRepository.On("GetID", mock.Anything, mock.Anything).Return("", nil)

		userSvc := user.NewService(mockUserRepository, userCfg)

		id, err := userSvc.ValidateUser(ctx, "an-email")

		assert.ErrorIs(t, err, user.ErrNoUserInformation)
		assert.Empty(t, id)
	})

	t.Run("should return user ID when successfully found user ID from DB", func(t *testing.T) {
		userID := "user-id"
		mockUserRepository := &mocks.UserRepository{}
		mockUserRepository.On("GetID", mock.Anything, mock.Anything).Return(userID, nil)

		userSvc := user.NewService(mockUserRepository, userCfg)

		id, err := userSvc.ValidateUser(ctx, "an-email")

		assert.NoError(t, err)
		assert.Equal(t, id, userID)
	})

	t.Run("should return user ID when not found user ID from DB but can create the new one", func(t *testing.T) {
		userID := "user-id"
		mockUserRepository := &mocks.UserRepository{}
		mockUserRepository.On("GetID", mock.Anything, mock.Anything).Return("", nil)
		mockUserRepository.On("Create", mock.Anything, mock.Anything).Return(userID, nil)

		userSvc := user.NewService(mockUserRepository, userCfg)

		id, err := userSvc.ValidateUser(ctx, "an-email")

		assert.ErrorIs(t, err, user.ErrNoUserInformation)
		assert.Empty(t, id)
	})

	t.Run("should return error when not found user ID from DB but can't create the new one", func(t *testing.T) {
		mockErr := errors.New("error adding user")
		mockUserRepository := &mocks.UserRepository{}
		mockUserRepository.On("GetID", mock.Anything, mock.Anything).Return("", mockErr)
		mockUserRepository.On("Create", mock.Anything, mock.Anything).Return("", mockErr)

		userSvc := user.NewService(mockUserRepository, userCfg)

		id, err := userSvc.ValidateUser(ctx, "an-email")

		assert.ErrorIs(t, err, mockErr)
		assert.Empty(t, id)
	})
}
