package postgres

import (
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	Host     string `mapstructure:"host" default:"localhost"`
	Port     int    `mapstructure:"port" default:"5432"`
	Name     string `mapstructure:"name" default:"postgres"`
	User     string `mapstructure:"user" default:"root"`
	Password string `mapstructure:"password" default:""`
	SSLMode  string `mapstructure:"sslmode" default:"disable"`
}

// ConnectionURL
func (c *Config) ConnectionURL() *url.URL {
	pgURL := &url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		User:   url.UserPassword(c.User, c.Password),
		Path:   c.Name,
	}
	q := pgURL.Query()
	q.Add("sslmode", "disable")
	pgURL.RawQuery = q.Encode()

	return pgURL
}
