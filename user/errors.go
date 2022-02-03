package user

import (
	"errors"
	"fmt"
)

var (
	ErrNoUserInformation = errors.New("no user information")
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

type InvalidError struct {
	Email    string
	Provider string
}

func (e InvalidError) Error() string {
	return fmt.Sprintf("empty field with email \"%s\" and provider \"%s\"", e.Email, e.Provider)
}
