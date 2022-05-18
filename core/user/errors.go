package user

import (
	"errors"
	"fmt"
)

var (
	ErrNoUserInformation = errors.New("no user information")
)

type NotFoundError struct {
	UUID  string
	Email string
}

func (e NotFoundError) Error() string {
	cause := "could not find user"
	if e.UUID != "" {
		cause += fmt.Sprintf(" with uuid \"%s\"", e.UUID)
	}
	if e.Email != "" {
		cause += fmt.Sprintf(" with email \"%s\"", e.Email)
	}
	return cause
}

type DuplicateRecordError struct {
	UUID  string
	Email string
}

func (e DuplicateRecordError) Error() string {
	cause := "duplicate user"
	if e.UUID != "" {
		cause += fmt.Sprintf(" with uuid \"%s\"", e.UUID)
	}
	if e.Email != "" {
		cause += fmt.Sprintf(" with email \"%s\"", e.Email)
	}
	return cause
}

type InvalidError struct {
	UUID string
}

func (e InvalidError) Error() string {
	return fmt.Sprintf("empty field with uuid \"%s\"", e.UUID)
}
