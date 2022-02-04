package postgres

import (
	"time"

	"github.com/odpf/columbus/user"
)

type UserModel struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Provider  string    `db:"provider"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u *UserModel) toUser() *user.User {
	return &user.User{
		ID:        u.ID,
		Email:     u.Email,
		Provider:  u.Provider,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func newUserModel(u *user.User) *UserModel {
	return &UserModel{
		ID:        u.ID,
		Email:     u.Email,
		Provider:  u.Provider,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type UserModels []UserModel

func (us UserModels) toUsers() []user.User {
	users := []user.User{}
	for _, u := range us {
		users = append(users, user.User{
			ID:        u.ID,
			Email:     u.Email,
			Provider:  u.Provider,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}
	return users
}
