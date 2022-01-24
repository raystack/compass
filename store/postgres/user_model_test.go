package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/columbus/user"
	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {

	t.Run("should return user domain entitiy", func(t *testing.T) {
		id := uuid.New()
		timestamp := time.Now().UTC()
		um := UserModel{
			ID:        id.String(),
			Email:     "user@odpf.io",
			Provider:  "columbus",
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		ud := um.toUser()

		assert.Equal(t, um.ID, ud.ID)
		assert.Equal(t, um.Email, ud.Email)
		assert.Equal(t, um.Provider, ud.Provider)
		assert.True(t, um.CreatedAt.Equal(ud.CreatedAt))
		assert.True(t, um.UpdatedAt.Equal(ud.UpdatedAt))
	})

	t.Run("should properly create user model from user", func(t *testing.T) {
		id := uuid.New()
		timestamp := time.Now().UTC()

		ud := &user.User{
			ID:        id.String(),
			Email:     "user@odpf.io",
			Provider:  "columbus",
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		um := newUserModel(ud)

		assert.Equal(t, um.ID, ud.ID)
		assert.Equal(t, um.Email, ud.Email)
		assert.Equal(t, um.Provider, ud.Provider)
		assert.True(t, um.CreatedAt.Equal(ud.CreatedAt))
		assert.True(t, um.UpdatedAt.Equal(ud.UpdatedAt))
	})
}
