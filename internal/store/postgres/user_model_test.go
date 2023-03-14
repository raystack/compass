package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/goto/compass/core/user"
	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {

	t.Run("should return user domain entitiy", func(t *testing.T) {
		someUUID := uuid.NewString()
		timestamp := time.Now().UTC()
		um := UserModel{
			ID:        sql.NullString{String: "12", Valid: true},
			UUID:      sql.NullString{String: someUUID, Valid: true},
			Email:     sql.NullString{String: "user@gotocompany.com", Valid: true},
			Provider:  sql.NullString{String: "compass", Valid: true},
			CreatedAt: sql.NullTime{Time: timestamp, Valid: true},
			UpdatedAt: sql.NullTime{Time: timestamp, Valid: true},
		}

		ud := um.toUser()

		assert.Equal(t, um.ID.String, ud.ID)
		assert.Equal(t, um.UUID.String, ud.UUID)
		assert.Equal(t, um.Email.String, ud.Email)
		assert.Equal(t, um.Provider.String, ud.Provider)
		assert.True(t, um.CreatedAt.Time.Equal(ud.CreatedAt))
		assert.True(t, um.UpdatedAt.Time.Equal(ud.UpdatedAt))
	})

	t.Run("should properly create user model from user", func(t *testing.T) {
		someUUID := uuid.NewString()
		timestamp := time.Now().UTC()

		ud := &user.User{
			ID:        "12",
			UUID:      someUUID,
			Email:     "user@gotocompany.com",
			Provider:  "compass",
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		um := newUserModel(ud)

		assert.Equal(t, um.ID.String, ud.ID)
		assert.Equal(t, um.UUID.String, ud.UUID)
		assert.Equal(t, um.Email.String, ud.Email)
		assert.Equal(t, um.Provider.String, ud.Provider)
		assert.True(t, um.CreatedAt.Time.Equal(ud.CreatedAt))
		assert.True(t, um.UpdatedAt.Time.Equal(ud.UpdatedAt))
	})
}
