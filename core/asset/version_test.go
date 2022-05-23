package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersionSemver(t *testing.T) {
	t.Run("parse invalid version will return non nil error", func(t *testing.T) {
		v := "xx"
		sv, err := ParseVersion(v)
		assert.Error(t, err)
		assert.Nil(t, sv)
	})

	t.Run("parse valid version will return nil error", func(t *testing.T) {
		v := "1.0"
		sv, err := ParseVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, sv.Major(), uint64(1))
		assert.Equal(t, sv.Minor(), uint64(0))
	})

	t.Run("parse valid version with prefix 'v' will return nil error", func(t *testing.T) {
		v := "v1.0"
		sv, err := ParseVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, sv.Major(), uint64(1))
		assert.Equal(t, sv.Minor(), uint64(0))
	})
}

func TestIncreaseMinorVersion(t *testing.T) {
	t.Run("increase minor version of invalid version will return non nil error", func(t *testing.T) {
		v := "xx"
		sv, err := IncreaseMinorVersion(v)
		assert.Error(t, err)
		assert.Empty(t, sv)
	})

	t.Run("increase minor version of valid version will return nil error", func(t *testing.T) {
		v := "1.0"
		sv, err := IncreaseMinorVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, "1.1", sv)
	})

	t.Run("increase minor version of valid version with prefix 'v' will return nil error", func(t *testing.T) {
		v := "v1.0"
		sv, err := IncreaseMinorVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, "1.1", sv)
	})
}
