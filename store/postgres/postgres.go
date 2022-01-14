package postgres

import (
	"github.com/golang-migrate/migrate/v4"
	postgres_migrate "github.com/golang-migrate/migrate/v4/database/postgres"

	// Register database postgres
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Register golang migrate source file
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const migrationsFilePath = "file://migrations"

type Client struct {
	Conn   *sqlx.DB
	Config Config
	logger logrus.FieldLogger
}

// NewClient initializes database connection
func NewClient(logger logrus.FieldLogger, cfg Config) (*Client, error) {
	dbConn, err := sqlx.Open("pgx", cfg.ConnectionURL().String())
	if err != nil {
		return nil, errors.Wrap(err, "error creating and connecting DB")
	}
	if dbConn == nil {
		return nil, errors.Wrap(err, "DB connection is nil in the creation")
	}
	return &Client{
		Conn:   dbConn,
		Config: cfg,
		logger: logger,
	}, nil
}

func (c *Client) Migrate() (err error) {
	m, err := c.initMigration()
	if err != nil {
		return errors.Wrap(err, "migration failed")
	}

	c.logger.Info("Migrating DB...")
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			c.logger.Infof("migration - no changes")
			return nil
		}
		return errors.Wrap(err, "migration failed")
	}
	return nil
}

func (c *Client) initMigration() (*migrate.Migrate, error) {
	if c.Conn != nil {
		driver, err := postgres_migrate.WithInstance(c.Conn.DB, &postgres_migrate.Config{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to initiate driver with db connection")
		}
		m, err := migrate.NewWithDatabaseInstance(migrationsFilePath, c.Config.Name, driver)
		return m, err
	}

	m, err := migrate.New(migrationsFilePath, c.Config.ConnectionURL().String())
	return m, err
}
