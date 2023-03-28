package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/goto/compass/core/user"
	"github.com/jmoiron/sqlx"
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
	return r.GetByEmailWithTx(ctx, r.client.db, email)
}

func (r *UserRepository) GetByEmailWithTx(ctx context.Context, querier sqlx.QueryerContext, email string) (user.User, error) {
	u, err := getUserByPredicate(ctx, querier, sq.Eq{"email": email})
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.NotFoundError{Email: email}
	}
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

// GetbyUUID retrieves user given the uuid
func (r *UserRepository) GetByUUID(ctx context.Context, uuid string) (user.User, error) {
	return r.GetByUUIDWithTx(ctx, r.client.db, uuid)
}

func (r *UserRepository) GetByUUIDWithTx(ctx context.Context, querier sqlx.QueryerContext, uuid string) (user.User, error) {
	u, err := getUserByPredicate(ctx, querier, sq.Eq{"uuid": uuid})
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.NotFoundError{UUID: uuid}
	}
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func getUserByPredicate(ctx context.Context, querier sqlx.QueryerContext, pred sq.Eq) (user.User, error) {
	qry, args, err := sq.Select("id", "uuid", "email", "provider", "created_at", "updated_at").
		From("users").
		Where(pred).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return user.User{}, fmt.Errorf("build query to get user by predicate: %w", err)
	}
	var um UserModel
	if err := sqlx.GetContext(ctx, querier, &um, qry, args...); err != nil {
		return user.User{}, fmt.Errorf("get user by predicate: %w", err)
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
