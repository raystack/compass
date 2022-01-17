package postgres

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	postgres_migrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/sirupsen/logrus"

	// Register golang migrate source file
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Client struct {
	db *sqlx.DB
}

func (c *Client) RunWithinTx(ctx context.Context, f func(tx *sqlx.Tx) error) error {
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}

	if err := f(tx); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("rollback transaction error: %v (original error: %w)", txErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

func (c *Client) Migrate(cfg Config, migrationFilePath string) (err error) {
	m, err := initMigration(c.db, cfg, migrationFilePath)
	if err != nil {
		return errors.Wrap(err, "migration failed")
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return errors.Wrap(err, "migration failed")
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
func NewClient(logger logrus.FieldLogger, cfg Config) (*Client, error) {
	db, err := sqlx.Connect("pgx", cfg.ConnectionURL().String())
	if err != nil {
		return nil, errors.Wrap(err, "error creating and connecting DB")
	}
	if db == nil {
		return nil, errNilDBClient
	}
	return &Client{db}, nil
}

func initMigration(db *sqlx.DB, cfg Config, migrationsFilePath string) (*migrate.Migrate, error) {
	if db != nil {
		driver, err := postgres_migrate.WithInstance(db.DB, &postgres_migrate.Config{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to initiate driver with db connection")
		}
		m, err := migrate.NewWithDatabaseInstance(migrationsFilePath, cfg.Name, driver)
		return m, err
	}

	m, err := migrate.New(migrationsFilePath, cfg.ConnectionURL().String())
	return m, err
}
