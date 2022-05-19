package asset

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

const BaseVersion = "0.1"

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
