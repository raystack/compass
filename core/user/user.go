package user

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname UserRepository --filename user_repository.go --output=./mocks
import (
	"context"
	"time"
)

// User is a basic entity of a user
type User struct {
	ID        string    `json:"-" diff:"-" db:"id"`
	UUID      string    `json:"uuid,omitempty" diff:"-" db:"uuid"`
	Email     string    `json:"email" diff:"email" db:"email"`
	Provider  string    `json:"provider" diff:"-" db:"provider"`
	CreatedAt time.Time `json:"-" diff:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" diff:"-" db:"updated_at"`
}

// Validate validates a user is valid or not
func (u *User) Validate() error {
	if u == nil {
		return ErrNoUserInformation
	}

	if u.UUID == "" {
		return InvalidError{UUID: u.UUID}
	}

	return nil
}

// Repository contains interface of supported methods
type Repository interface {
	Create(ctx context.Context, u *User) (string, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByUUID(ctx context.Context, uuid string) (User, error)
	UpsertByEmail(ctx context.Context, u *User) (string, error)
}
