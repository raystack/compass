//go:build go1.16
// +build go1.16

package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	// Register database postgres
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Register golang migrate source
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var fs embed.FS

type Client struct {
	db *sqlx.DB
}

func (c *Client) RunWithinTx(ctx context.Context, f func(tx *sqlx.Tx) error) error {
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	if err := f(tx); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("rollback transaction error: %v (original error: %w)", txErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (c *Client) Migrate(cfg Config) (err error) {
	m, err := initMigration(cfg)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

// ExecQueries is used for executing list of db query
func (c *Client) ExecQueries(ctx context.Context, queries []string) error {
	for _, query := range queries {
		_, err := c.db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

// NewClient initializes database connection
func NewClient(cfg Config) (*Client, error) {
	db, err := sqlx.Connect("pgx", cfg.ConnectionURL().String())
	if err != nil {
		return nil, fmt.Errorf("error creating and connecting DB: %w", err)
	}
	if db == nil {
		return nil, errNilDBClient
	}
	return &Client{db}, nil
}

func initMigration(cfg Config) (*migrate.Migrate, error) {
	iofsDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", iofsDriver, cfg.ConnectionURL().String())
	if err != nil {
		log.Fatal(err)
	}
	return m, nil
}

func checkPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%w [%s]", errDuplicateKey, pgErr.Detail)
		case pgerrcode.CheckViolation:
			return fmt.Errorf("%w [%s]", errCheckViolation, pgErr.Detail)
		}
	}
	return err
}
