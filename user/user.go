package user

// go:generate mockery --name Repository --outpkg mocks --output ../lib/mocks/ --structname UserRepository --filename user_repository.go

import (
	"context"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	if u == nil {
		return ErrNoUserInformation
	}

	if u.Email == "" || u.Provider == "" {
		return InvalidError{Email: u.Email, Provider: u.Provider}
	}

	return nil
}

type Repository interface {
	GetID(ctx context.Context, email string) (string, error)
	Create(ctx context.Context, u *User) (string, error)
}
