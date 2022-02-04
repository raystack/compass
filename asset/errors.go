package asset

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyID     = errors.New("asset does not have ID")
	ErrUnknownType = errors.New("unknown type")
)

type NotFoundError struct {
	AssetID string
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf("no such record: %q", err.AssetID)
}

type InvalidError struct {
	AssetID string
}

func (err InvalidError) Error() string {
	return fmt.Sprintf("invalid asset id: %q", err.AssetID)
}
