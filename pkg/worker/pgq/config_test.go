package pgq_test

import (
	"testing"
	"time"

	"github.com/goto/compass/pkg/worker/pgq"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := pgq.Config{
		Host:            "dbhost",
		Port:            9876,
		Username:        "user",
		Password:        "pass",
		Name:            "database",
		MaxOpenConns:    3,
		MaxIdleConns:    10,
		ConnMaxIdleTime: 30 * time.Second,
		ConnMaxLifetime: 5 * time.Minute,
	}

	assert.Equal(t, "postgres://user:pass@dbhost:9876/database?sslmode=disable", cfg.ConnectionURL())
	assert.Equal(t, "dbname=database user=user password='pass' host=dbhost port=9876 sslmode=disable", cfg.ConnectionString())
	assert.Equal(t, 5*time.Minute, cfg.ConnMaxLifetimeWithJitter())

	cfg.ConnMaxLifetimeJitter = time.Minute
	maxLifetime := cfg.ConnMaxLifetimeWithJitter()
	assert.GreaterOrEqual(t, maxLifetime, 5*time.Minute)
	assert.LessOrEqual(t, maxLifetime, 6*time.Minute)
}
