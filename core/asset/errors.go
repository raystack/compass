package asset

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyID     = errors.New("asset does not have ID")
	ErrEmptyURN    = errors.New("asset does not have URN")
	ErrUnknownType = errors.New("unknown type")
	ErrNilAsset    = errors.New("nil asset")
)

type NotFoundError struct {
	AssetID string
	URN     string
	Type    Type
	Service string
}

func (err NotFoundError) Error() string {
	if err.AssetID != "" {
		return fmt.Sprintf("no such record: %q", err.AssetID)
	}

	return fmt.Sprintf("could not find asset with urn = %s, type = %s, service = %s", err.URN, err.Type, err.Service)
}

type InvalidError struct {
	AssetID string
}

func (err InvalidError) Error() string {
	return fmt.Sprintf("invalid asset id: %q", err.AssetID)
}

type DiscoveryError struct {
	Err error
}

func (err DiscoveryError) Error() string {
	return fmt.Sprintf("discovery error: %s", err.Err)
}
