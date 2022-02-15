package user

//go:generate mockery --name Repository --outpkg mocks --output ../lib/mocks/ --structname UserRepository --filename user_repository.go
import (
	"context"
	"time"
)

// User is a basic entity of a user
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Provider  string    `json:"provider" db:"provider"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate validates a user is valid or not
func (u *User) Validate() error {
	if u == nil {
		return ErrNoUserInformation
	}

	if u.Email == "" || u.Provider == "" {
		return InvalidError{Email: u.Email, Provider: u.Provider}
	}

	return nil
}

// Repository contains interface of supported methods
type Repository interface {
	Create(ctx context.Context, u *User) (string, error)
	GetID(ctx context.Context, email string) (string, error)
}
