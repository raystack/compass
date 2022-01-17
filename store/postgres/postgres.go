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

// NewClient initializes database connection
func NewClient(logger logrus.FieldLogger, cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", cfg.ConnectionURL().String())
	if err != nil {
		return nil, errors.Wrap(err, "error creating and connecting DB")
	}
	return db, nil
}

func Migrate(db *sqlx.DB, cfg Config) (err error) {
	m, err := initMigration(db, cfg)
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

func initMigration(db *sqlx.DB, cfg Config) (*migrate.Migrate, error) {
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
