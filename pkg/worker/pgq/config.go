package pgq

import (
	"fmt"
	"math/rand"
	"time"
)

type Config struct {
	Host     string `mapstructure:"host" default:"localhost"`
	Port     int    `mapstructure:"port" default:"5432"`
	Name     string `mapstructure:"name" default:"postgres"`
	Username string `mapstructure:"username" default:"root"`
	Password string `mapstructure:"password" default:""`

	// Connection pool settings
	MaxOpenConns          int           `mapstructure:"max_open_conns" default:"10"`
	MaxIdleConns          int           `mapstructure:"max_idle_conns" default:"4"`
	ConnMaxIdleTime       time.Duration `mapstructure:"conn_max_idle_time" default:"5m"`
	ConnMaxLifetime       time.Duration `mapstructure:"conn_max_lifetime" default:"5m"`
	ConnMaxLifetimeJitter time.Duration `mapstructure:"conn_max_lifetime_jitter" default:"2m"`
}

func (c Config) ConnectionString() string {
	return fmt.Sprintf(
		"dbname=%s user=%s password='%s' host=%s port=%d sslmode=disable",
		c.Name, c.Username, c.Password, c.Host, c.Port,
	)
}

func (c Config) ConnectionURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.Username, c.Password, c.Host, c.Port, c.Name)
}

func (c Config) ConnMaxLifetimeWithJitter() time.Duration {
	var jitter time.Duration
	if c.ConnMaxLifetimeJitter > 0 {
		//nolint:gosec
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		jitter = time.Duration(r.Int63n(int64(c.ConnMaxLifetimeJitter)))
	}

	return c.ConnMaxLifetime + jitter
}
