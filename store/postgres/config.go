package postgres

import "fmt"

type Config struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// ConnectionURL
func (c *Config) ConnectionURL(driver string) string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
		driver,
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}

func NewConfig(
	host string,
	port int,
	name string,
	user string,
	password string,
	sslMode string) Config {
	return Config{
		Host:     host,
		Port:     port,
		Name:     name,
		User:     user,
		Password: password,
		SSLMode:  sslMode,
	}
}
