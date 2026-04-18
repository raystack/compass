package principal_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/raystack/compass/core/principal"
)

func TestContext(t *testing.T) {
	t.Run("should return passed principal if exist in context", func(t *testing.T) {
		passed := principal.Principal{Subject: "sub-123", Type: "user", Name: "Alice"}
		ctx := principal.NewContext(context.Background(), passed)
		actual := principal.FromContext(ctx)
		if !cmp.Equal(passed, actual) {
			t.Fatalf("actual is \"%+v\" but expected was \"%+v\"", actual, passed)
		}
	})

	t.Run("should return empty principal if not exist in context", func(t *testing.T) {
		actual := principal.FromContext(context.Background())
		expected := principal.Principal{}
		if !cmp.Equal(actual, expected) {
			t.Fatalf("actual is \"%+v\" but expected was \"%+v\"", actual, expected)
		}
	})
}
