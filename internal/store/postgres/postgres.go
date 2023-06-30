//go:build go1.16
// +build go1.16

package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/raystack/compass/pkg/grpc_interceptor"
	"log"
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

const (
	columnNameCreatedAt     = "created_at"
	columnNameUpdatedAt     = "updated_at"
	sortDirectionAscending  = "ASC"
	sortDirectionDescending = "DESC"
	DefaultMaxResultSize    = 100

	namespaceRLSSetQuery   = "SET app.current_tenant = '%s'"
	namespaceRLSResetQuery = "RESET app.current_tenant"
)

// Client is a wrapper over sqlx for strict multi-tenancy using postgres RLS
type Client struct {
	_db *sqlx.DB
}

func (c *Client) ExecContext(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	conn, err := c._db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// inject context
	if _, err = conn.ExecContext(ctx, fmt.Sprintf(namespaceRLSSetQuery, namespaceFromContext(ctx))); err != nil {
		return
	}

	// execute main query
	result, err = conn.ExecContext(ctx, query, args...)

	// reset context, do it even if main query failed with err
	if _, _err := conn.ExecContext(ctx, namespaceRLSResetQuery); _err != nil {
		// this should not happen, risk of namespace context spills
		panic(_err)
	}
	return result, err
}

func (c *Client) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	conn, err := c._db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// inject context
	if _, err = conn.ExecContext(ctx, fmt.Sprintf(namespaceRLSSetQuery, namespaceFromContext(ctx))); err != nil {
		return err
	}

	// execute main query
	qerr := conn.GetContext(ctx, dest, query, args...)

	// reset context, do it even if main query failed with err
	if _, _err := conn.ExecContext(ctx, namespaceRLSResetQuery); _err != nil {
		// this should not happen, risk of namespace context spills
		panic(_err)
	}
	return qerr
}

func (c *Client) QueryFn(ctx context.Context, f func(*sqlx.Conn) error) error {
	conn, err := c._db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// inject context
	if _, err = conn.ExecContext(ctx, fmt.Sprintf(namespaceRLSSetQuery, namespaceFromContext(ctx))); err != nil {
		return err
	}

	// execute main query
	ferr := f(conn)

	// reset context, do it even if main query failed with err
	if _, _err := conn.ExecContext(ctx, namespaceRLSResetQuery); _err != nil {
		// this should not happen, risk of namespace context spills
		panic(_err)
	}
	return ferr
}

func (c *Client) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	conn, err := c._db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// inject context
	if _, err = conn.ExecContext(ctx, fmt.Sprintf(namespaceRLSSetQuery, namespaceFromContext(ctx))); err != nil {
		return err
	}

	// execute main query
	qerr := conn.SelectContext(ctx, dest, query, args...)

	// reset context, do it even if main query failed with err
	if _, _err := conn.ExecContext(ctx, namespaceRLSResetQuery); _err != nil {
		// this should not happen, risk of namespace context spills
		panic(_err)
	}
	return qerr
}

func (c *Client) RunWithinTx(ctx context.Context, f func(tx *sqlx.Tx) error) error {
	conn, err := c._db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// inject context
	if _, err = conn.ExecContext(ctx, fmt.Sprintf(namespaceRLSSetQuery, namespaceFromContext(ctx))); err != nil {
		return err
	}

	tx, err := conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	if err = f(tx); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			err = fmt.Errorf("rollback transaction error: %v (original error: %w)", txErr, err)
		}

		// reset context
		if _, _err := conn.ExecContext(ctx, namespaceRLSResetQuery); _err != nil {
			// this should not happen, risk of namespace context spills
			panic(_err)
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	// reset context, do it even if main query failed with err
	if _, _err := conn.ExecContext(ctx, namespaceRLSResetQuery); _err != nil {
		// this should not happen, risk of namespace context spills
		panic(_err)
	}
	return nil
}

func (c *Client) Migrate(cfg Config) (ver uint, err error) {
	m, err := initMigration(cfg)
	if err != nil {
		return 0, fmt.Errorf("migration failed: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return 0, fmt.Errorf("migration failed: %w", err)
	}
	if ver, _, err = m.Version(); err != nil {
		return ver, err
	}
	return ver, nil
}

func (c *Client) MigrateDown(cfg Config) (ver uint, err error) {
	m, err := initMigration(cfg)
	if err != nil {
		return 0, fmt.Errorf("migration failed: %w", err)
	}
	// down one step
	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return 0, fmt.Errorf("migration failed: %w", err)
	}
	if ver, _, err = m.Version(); err != nil {
		return ver, err
	}
	return ver, nil
}

// ExecQueries is used for executing list of _db query
func (c *Client) ExecQueries(ctx context.Context, queries []string) error {
	for _, query := range queries {
		_, err := c._db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Close() error {
	return c._db.Close()
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
		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%w [%s]", errForeignKeyViolation, pgErr.Detail)
		}
	}
	return err
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func namespaceFromContext(ctx context.Context) uuid.UUID {
	return grpc_interceptor.FetchNamespaceFromContext(ctx).ID
}
