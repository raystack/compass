package postgres

import (
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
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
