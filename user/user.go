package user

import (
	"context"
	"time"
)

type User struct {
	ID        string
	Email     string
	Provider  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	GetID(ctx context.Context, email string) (string, error)
	Create(ctx context.Context, u *User) error
}
