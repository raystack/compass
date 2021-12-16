package postgres_test

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestClient initializes database connection
func newTestClient(dsn string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}
