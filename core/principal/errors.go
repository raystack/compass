package principal

import (
	"errors"
	"fmt"
)

var (
	ErrNoPrincipalInformation = errors.New("no principal information")
)

type NotFoundError struct {
	Subject string
}

func (e NotFoundError) Error() string {
	cause := "could not find principal"
	if e.Subject != "" {
		cause += fmt.Sprintf(" with subject %q", e.Subject)
	}
	return cause
}

type DuplicateRecordError struct {
	Subject string
}

func (e DuplicateRecordError) Error() string {
	cause := "duplicate principal"
	if e.Subject != "" {
		cause += fmt.Sprintf(" with subject %q", e.Subject)
	}
	return cause
}

type InvalidError struct {
	Subject string
}

func (e InvalidError) Error() string {
	return fmt.Sprintf("invalid principal: empty subject %q", e.Subject)
}
