package sqlstore

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// NewPostgreSQLClient initializes database connection
func NewPostgreSQLClient(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		strings.TrimSpace(cfg.Host),
		cfg.Port,
		strings.TrimSpace(cfg.Name),
		strings.TrimSpace(cfg.User),
		strings.TrimSpace(cfg.Password),
		strings.TrimSpace(cfg.SSLMode),
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// NewSQLiteClient initializes database connection
func NewSQLiteClient(dsn string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}
