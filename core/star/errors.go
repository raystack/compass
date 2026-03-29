package star

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyUserID  = errors.New("star is not related to any user")
	ErrEmptyEntityID = errors.New("star is not related to any entity")
)

type NotFoundError struct {
	EntityID string
	UserID  string
}

func (e NotFoundError) Error() string {
	fields := []string{"could not find starred entity"}
	if e.EntityID != "" {
		fields = append(fields, fmt.Sprintf("with entity id \"%s\"", e.EntityID))
	}
	if e.UserID != "" {
		fields = append(fields, fmt.Sprintf("by user id \"%s\"", e.UserID))
	}
	return strings.Join(fields, ", ")
}

type UserNotFoundError struct {
	UserID string
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("could not find user with id \"%s\"", e.UserID)
}

type DuplicateRecordError struct {
	UserID  string
	EntityID string
}

func (e DuplicateRecordError) Error() string {
	return fmt.Sprintf("duplicate starred entity id \"%s\" with user id \"%s\"", e.EntityID, e.UserID)
}

type InvalidError struct {
	UserID  string
	EntityID string
}

func (e InvalidError) Error() string {
	fields := []string{"invalid"}
	if e.EntityID != "" {
		fields = append(fields, fmt.Sprintf("asset id \"%s\"", e.EntityID))
	}
	if e.UserID != "" {
		fields = append(fields, fmt.Sprintf("user id \"%s\"", e.UserID))
	}
	return strings.Join(fields, " ")
}
