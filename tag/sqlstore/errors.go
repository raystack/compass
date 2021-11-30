package sqlstore

import "errors"

var (
	errNilDBClient = errors.New("db client is nil")
)
