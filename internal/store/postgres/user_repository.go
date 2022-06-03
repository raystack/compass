package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/compass/core/user"
)

// UserRepository is a type that manages user operation to the primary database
type UserRepository struct {
	client *Client
}

// UpsertByEmail updates a row if email match and uuid is empty
// if email not found, insert a new row
func (r *UserRepository) UpsertByEmail(ctx context.Context, ud *user.User) (string, error) {
	var userID string

	if err := ud.Validate(); err != nil {
		return "", err
	}

	um := newUserModel(ud)

	if err := r.client.db.QueryRowxContext(ctx, `
				INSERT INTO users (uuid, email, provider) VALUES ($1, $2, $3) ON CONFLICT (email)
				DO UPDATE SET uuid = $1, email = $2 WHERE users.uuid IS NULL
				RETURNING id
		`, um.UUID, um.Email, um.Provider).Scan(&userID); err != nil {
		err := checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return "", user.DuplicateRecordError{UUID: ud.UUID, Email: ud.Email}
		}
		return "", err
	}
	if userID == "" {
		return "", fmt.Errorf("error User UUID is empty from DB")
	}
	return userID, nil
}

// Create insert a user to the database
// a new data is still inserted if either uuid or email is empty
// but returns error if both uuid and email are empty
func (r *UserRepository) Create(ctx context.Context, ud *user.User) (string, error) {
	return r.create(ctx, r.client.db, ud)
}

// Create insert a user to the database using given transaction as client
func (r *UserRepository) CreateWithTx(ctx context.Context, tx *sqlx.Tx, ud *user.User) (string, error) {
	return r.create(ctx, tx, ud)
}

func (r *UserRepository) create(ctx context.Context, querier sqlx.QueryerContext, ud *user.User) (string, error) {
	var userID string

	if ud == nil {
		return "", user.ErrNoUserInformation
	}

	if ud.UUID == "" && ud.Email == "" {
		return "", user.ErrNoUserInformation
	}

	um := newUserModel(ud)

	if err := querier.QueryRowxContext(ctx, `
					INSERT INTO
					users
						(uuid, email, provider)
					VALUES
						($1, $2, $3)
					RETURNING id
					`, um.UUID, um.Email, um.Provider).Scan(&userID); err != nil {
		err := checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", user.DuplicateRecordError{UUID: ud.UUID, Email: ud.Email}
		}
		return "", err
	}
	if userID == "" {
		return "", fmt.Errorf("error User UUID is empty from DB")
	}
	return userID, nil
}

// GetUUID retrieves user UUID given the email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	var um UserModel
	if err := r.client.db.GetContext(ctx, &um, `
		SELECT * FROM users WHERE email = $1
	`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.NotFoundError{Email: email}
		}
		return user.User{}, err
	}
	return um.toUser(), nil
}

// GetbyUUID retrieves user given the uuid
func (r *UserRepository) GetByUUID(ctx context.Context, uuid string) (user.User, error) {
	var um UserModel
	if err := r.client.db.GetContext(ctx, &um, `
		SELECT * FROM users WHERE uuid = $1
	`, uuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.NotFoundError{UUID: uuid}
		}
		return user.User{}, err
	}
	return um.toUser(), nil
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
