package user

import (
	"errors"
	"fmt"
)

var (
	ErrNilUser = errors.New("user is nil")
)

type NotFoundError struct {
	Email string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("could not find user \"%s\"", e.Email)
}

type DuplicateRecordError struct {
	ID    string
	Email string
}

func (e DuplicateRecordError) Error() string {
	return fmt.Sprintf("duplicate user \"%s\" with user id \"%s\"", e.Email, e.ID)
}
