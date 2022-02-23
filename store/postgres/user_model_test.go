package postgres

import (
	"database/sql"
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
			ID:        sql.NullString{String: id.String(), Valid: true},
			Email:     sql.NullString{String: "user@odpf.io", Valid: true},
			Provider:  sql.NullString{String: "columbus", Valid: true},
			CreatedAt: sql.NullTime{Time: timestamp, Valid: true},
			UpdatedAt: sql.NullTime{Time: timestamp, Valid: true},
		}

		ud := um.toUser()

		assert.Equal(t, um.ID.String, ud.ID)
		assert.Equal(t, um.Email.String, ud.Email)
		assert.Equal(t, um.Provider.String, ud.Provider)
		assert.True(t, um.CreatedAt.Time.Equal(ud.CreatedAt))
		assert.True(t, um.UpdatedAt.Time.Equal(ud.UpdatedAt))
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

		assert.Equal(t, um.ID.String, ud.ID)
		assert.Equal(t, um.Email.String, ud.Email)
		assert.Equal(t, um.Provider.String, ud.Provider)
		assert.True(t, um.CreatedAt.Time.Equal(ud.CreatedAt))
		assert.True(t, um.UpdatedAt.Time.Equal(ud.UpdatedAt))
	})
}
