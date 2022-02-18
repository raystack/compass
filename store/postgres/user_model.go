package postgres

import (
	"database/sql"

	"github.com/odpf/columbus/user"
)

type UserModel struct {
	ID        sql.NullString `db:"id"`
	Email     sql.NullString `db:"email"`
	Provider  sql.NullString `db:"provider"`
	CreatedAt sql.NullTime   `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
}

func (u *UserModel) toUser() user.User {
	return user.User{
		ID:        u.ID.String,
		Email:     u.Email.String,
		Provider:  u.Provider.String,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func newUserModel(u *user.User) UserModel {
	return UserModel{
		ID:        sql.NullString{String: u.ID, Valid: true},
		Email:     sql.NullString{String: u.Email, Valid: true},
		Provider:  sql.NullString{String: u.Provider, Valid: true},
		CreatedAt: sql.NullTime{Time: u.CreatedAt, Valid: true},
		UpdatedAt: sql.NullTime{Time: u.UpdatedAt, Valid: true},
	}
}

type UserModels []UserModel

func (us UserModels) toUsers() []user.User {
	users := []user.User{}
	for _, u := range us {
		users = append(users, u.toUser())
	}
	return users
}
