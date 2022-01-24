package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/odpf/columbus/user"
)

// UserRepository is a type that manages user operation to the primary database
type UserRepository struct {
	client *Client
}

// Create insert a user to the database
func (r *UserRepository) Create(ctx context.Context, ud *user.User) (string, error) {
	var userID string
	if ud == nil {
		return "", user.ErrNilUser
	}

	// either success inserting a row or return error
	// no need to check rows affected
	if err := r.client.db.QueryRowxContext(ctx, `
					INSERT INTO 
					users 
						(email, provider)
					VALUES 
						($1, $2)
					RETURNING id
					`, ud.Email, ud.Provider).Scan(&userID); err != nil {
		err := checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", user.DuplicateRecordError{ID: ud.ID, Email: ud.Email}
		}
	}
	if userID == "" {
		return "", fmt.Errorf("error User ID is empty from DB")
	}
	return userID, nil
}

// GetID  retrieves user UUID given the email
func (r *UserRepository) GetID(ctx context.Context, email string) (string, error) {
	var userID string
	if err := r.client.db.GetContext(ctx, &userID, `
		SELECT 
			id
		FROM
			users
		WHERE
			email = $1
	`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", user.NotFoundError{Email: email}
		}
		return "", err
	}
	return userID, nil
}

// NewUserRepository initializes user repository clients
func NewUserRepository(c *Client) (*UserRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	return &UserRepository{
		client: c,
	}, nil
}
