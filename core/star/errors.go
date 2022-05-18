package star

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyUserID  = errors.New("star is not related to any user")
	ErrEmptyAssetID = errors.New("star is not related to any asset")
)

type NotFoundError struct {
	AssetID string
	UserID  string
}

func (e NotFoundError) Error() string {
	fields := []string{"could not find starred asset"}
	if e.AssetID != "" {
		fields = append(fields, fmt.Sprintf("with asset id \"%s\"", e.AssetID))
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
	AssetID string
}

func (e DuplicateRecordError) Error() string {
	return fmt.Sprintf("duplicate starred asset id \"%s\" with user id \"%s\"", e.AssetID, e.UserID)
}

type InvalidError struct {
	UserID  string
	AssetID string
}

func (e InvalidError) Error() string {
	fields := []string{"invalid"}
	if e.AssetID != "" {
		fields = append(fields, fmt.Sprintf("asset id \"%s\"", e.AssetID))
	}
	if e.UserID != "" {
		fields = append(fields, fmt.Sprintf("user id \"%s\"", e.UserID))
	}
	return strings.Join(fields, " ")
}
