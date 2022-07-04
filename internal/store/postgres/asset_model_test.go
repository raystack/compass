package postgres_test

import (
	"testing"

	"github.com/odpf/compass/internal/store/postgres"
	"github.com/stretchr/testify/assert"
)

func TestJSONMap(t *testing.T) {
	t.Run("return no error for valid type of value", func(t *testing.T) {
		m := postgres.JSONMap{}
		m.Scan("value")
		_, err := m.Value()
		assert.NoError(t, err)
		b, err := m.MarshalJSON()
		assert.NoError(t, err)
		err = m.UnmarshalJSON(b)
		assert.NoError(t, err)
	})
}
