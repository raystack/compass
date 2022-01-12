package postgres

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewClient initializes database connection
func NewClient(cfg Config) (*gorm.DB, error) {
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
