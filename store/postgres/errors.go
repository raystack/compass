package postgres

import "errors"

var (
	errNilDBClient         = errors.New("db client is nil")
	errNilPostgresClient   = errors.New("postgres client is nil")
	errDuplicateKey        = errors.New("duplicate key")
	errCheckViolation      = errors.New("check constraint violation")
	errForeignKeyViolation = errors.New("foreign key violation")
)
