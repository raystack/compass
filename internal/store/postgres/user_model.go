package postgres

import (
	"database/sql"

	"github.com/goto/compass/core/user"
)

type UserModel struct {
	ID        sql.NullString `db:"id"`
	UUID      sql.NullString `db:"uuid"`
	Email     sql.NullString `db:"email"`
	Provider  sql.NullString `db:"provider"`
	CreatedAt sql.NullTime   `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
}

func (u *UserModel) toUser() user.User {
	return user.User{
		ID:        u.ID.String,
		UUID:      u.UUID.String,
		Email:     u.Email.String,
		Provider:  u.Provider.String,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func newUserModel(u *user.User) UserModel {
	um := UserModel{}
	if u.ID != "" {
		um.ID = sql.NullString{String: u.ID, Valid: true}
	}
	if u.UUID != "" {
		um.UUID = sql.NullString{String: u.UUID, Valid: true}
	}
	if u.Email != "" {
		um.Email = sql.NullString{String: u.Email, Valid: true}
	}
	if u.Provider != "" {
		um.Provider = sql.NullString{String: u.Provider, Valid: true}
	}
	um.CreatedAt = sql.NullTime{Time: u.CreatedAt, Valid: true}
	um.UpdatedAt = sql.NullTime{Time: u.UpdatedAt, Valid: true}

	return um
}

type UserModels []UserModel

func (us UserModels) toUsers() []user.User {
	if len(us) == 0 {
		return nil
	}
	users := []user.User{}
	for _, u := range us {
		users = append(users, u.toUser())
	}
	return users
}
