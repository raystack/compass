package asset

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/r3labs/diff/v2"
)

const BaseVersion = "0.1"

// AssetVersion is the changes summary of asset versions
type AssetVersion struct {
	ID        string         `json:"id" db:"id"`
	URN       string         `json:"urn" db:"urn"`
	Type      string         `json:"type" db:"type"`
	Service   string         `json:"service" db:"service"`
	Version   string         `json:"version" db:"version"`
	Changelog diff.Changelog `json:"changelog" db:"changelog"`
	UpdatedBy string         `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// ValidateVersion returns error if version string is not in MAJOR.MINOR format
func ValidateVersion(v string) error {
	_, err := semver.NewVersion(v)
	if err != nil {
		return fmt.Errorf("invalid version \"%s\"", v)
	}
	return nil
}

// IncreaseMinorVersion bumps up the minor version +0.1
func IncreaseMinorVersion(v string) (string, error) {
	if err := ValidateVersion(v); err != nil {
		return "", err
	}
	s, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return "", err
	}
	s = s + 0.1
	return strconv.FormatFloat(s, 'f', -1, 32), nil
}
