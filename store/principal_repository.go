package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/principal"
)

// PrincipalRepository is a type that manages principal operations to the primary database.
type PrincipalRepository struct {
	client *Client
}

// Upsert inserts or updates a principal by subject.
func (r *PrincipalRepository) Upsert(ctx context.Context, ns *namespace.Namespace, p *principal.Principal) (string, error) {
	var id string
	if err := p.Validate(); err != nil {
		return "", err
	}

	pm := newPrincipalModel(p)

	err := r.client.QueryFn(ctx, func(conn *sqlx.Conn) error {
		return conn.QueryRowxContext(ctx, `
				INSERT INTO principals (uuid, subject, name, type, metadata, namespace_id)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (subject, namespace_id) WHERE subject IS NOT NULL AND subject != ''
				DO UPDATE SET name = EXCLUDED.name, type = EXCLUDED.type, metadata = EXCLUDED.metadata
				RETURNING id
		`, pm.UUID, pm.Subject, pm.Name, pm.Type, pm.Metadata, ns.ID).Scan(&id)
	})
	if err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return "", principal.DuplicateRecordError{Subject: p.Subject}
		}
		return "", err
	}

	if id == "" {
		return "", fmt.Errorf("error principal ID is empty from DB")
	}
	return id, nil
}

// Create inserts a principal to the database.
func (r *PrincipalRepository) Create(ctx context.Context, ns *namespace.Namespace, p *principal.Principal) (string, error) {
	var id string
	if p == nil {
		return "", principal.ErrNoPrincipalInformation
	}
	if p.Subject == "" {
		return "", principal.ErrNoPrincipalInformation
	}
	pm := newPrincipalModel(p)

	err := r.client.QueryFn(ctx, func(conn *sqlx.Conn) error {
		return conn.QueryRowxContext(ctx, `
					INSERT INTO
					principals
						(uuid, subject, name, type, metadata, namespace_id)
					VALUES
						($1, $2, $3, $4, $5, $6)
					RETURNING id
					`, pm.UUID, pm.Subject, pm.Name, pm.Type, pm.Metadata, ns.ID).Scan(&id)
	})
	if err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", principal.DuplicateRecordError{Subject: p.Subject}
		}
		return "", err
	}

	if id == "" {
		return "", fmt.Errorf("error principal ID is empty from DB")
	}
	return id, nil
}

// GetBySubject retrieves a principal given the subject.
func (r *PrincipalRepository) GetBySubject(ctx context.Context, subject string) (principal.Principal, error) {
	var pm PrincipalModel
	if err := r.client.GetContext(ctx, &pm, `
		SELECT * FROM principals WHERE subject = $1
	`, subject); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Fall back to uuid column for backward compat
			if err2 := r.client.GetContext(ctx, &pm, `
				SELECT * FROM principals WHERE uuid = $1
			`, subject); err2 != nil {
				if errors.Is(err2, sql.ErrNoRows) {
					return principal.Principal{}, principal.NotFoundError{Subject: subject}
				}
				return principal.Principal{}, err2
			}
			return pm.toPrincipal(), nil
		}
		return principal.Principal{}, err
	}
	return pm.toPrincipal(), nil
}

// NewPrincipalRepository initializes principal repository clients.
func NewPrincipalRepository(c *Client) (*PrincipalRepository, error) {
	if c == nil {
		return nil, errNilPostgresClient
	}
	return &PrincipalRepository{
		client: c,
	}, nil
}
