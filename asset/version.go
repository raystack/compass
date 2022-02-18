package asset

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/odpf/columbus/user"
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
	UpdatedBy user.User      `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// ParseVersion returns error if version string is not in MAJOR.MINOR format
func ParseVersion(v string) (*semver.Version, error) {
	semverVersion, err := semver.NewVersion(v)
	if err != nil {
		return nil, fmt.Errorf("invalid version \"%s\"", v)
	}
	return semverVersion, nil
}

// IncreaseMinorVersion bumps up the minor version +0.1
func IncreaseMinorVersion(v string) (string, error) {
	oldVersion, err := ParseVersion(v)
	if err != nil {
		return "", err
	}
	newVersion := oldVersion.IncMinor()
	return fmt.Sprintf("%d.%d", newVersion.Major(), newVersion.Minor()), nil
}
