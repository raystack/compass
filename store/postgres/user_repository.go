package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/columbus/user"
)

// UserRepository is a type that manages user operation to the primary database
type UserRepository struct {
	client *Client
}

// Create insert a user to the database
func (r *UserRepository) Create(ctx context.Context, ud *user.User) (string, error) {
	return r.create(ctx, r.client.db, ud)
}

// Create insert a user to the database using given transaction as client
func (r *UserRepository) CreateWithTx(ctx context.Context, tx *sqlx.Tx, ud *user.User) (string, error) {
	return r.create(ctx, tx, ud)
}

func (r *UserRepository) create(ctx context.Context, querier sqlx.QueryerContext, ud *user.User) (string, error) {
	var userID string

	if err := ud.Validate(); err != nil {
		return "", err
	}

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
		return "", err
	}
	if userID == "" {
		return "", fmt.Errorf("error User ID is empty from DB")
	}
	return userID, nil
}

// GetID retrieves user UUID given the email
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
		return nil, errNilPostgresClient
	}
	return &UserRepository{
		client: c,
	}, nil
}
