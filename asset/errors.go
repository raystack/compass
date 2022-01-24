package asset

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyID = errors.New("asset does not have ID")
)

type NotFoundError struct {
	AssetID string
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf("no such record: %q", err.AssetID)
}
