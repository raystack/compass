package middleware

import "github.com/odpf/salt/log"

type Config struct {
	Logger         log.Logger
	IdentityHeader string
}
