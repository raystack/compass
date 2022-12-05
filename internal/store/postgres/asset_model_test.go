package postgres_test

import (
	"testing"

	"github.com/odpf/compass/internal/store/postgres"
	"github.com/stretchr/testify/assert"
)

func TestJSONMap(t *testing.T) {
	t.Run("return no error for valid type of value", func(t *testing.T) {
		value := []byte(`{"key1":"val1","key2":"val2"}`)
		m := postgres.JSONMap{}
		err := m.Scan(value)
		assert.NoError(t, err)
		s, err := m.Value()
		assert.Equal(t, string(value), s)
		assert.NoError(t, err)
		b, err := m.MarshalJSON()
		assert.NoError(t, err)
		err = m.UnmarshalJSON(b)
		assert.NoError(t, err)
	})
}
