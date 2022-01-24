package postgres

import (
	"time"

	"github.com/odpf/columbus/user"
)

type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Provider  string    `db:"provider"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u *User) toUser() *user.User {
	return &user.User{
		ID:        u.ID,
		Email:     u.Email,
		Provider:  u.Provider,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func newUserModel(u *user.User) *User {
	return &User{
		ID:        u.ID,
		Email:     u.Email,
		Provider:  u.Provider,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
