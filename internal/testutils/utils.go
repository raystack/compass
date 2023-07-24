package testutils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func AssertEqualProto(t *testing.T, expected, actual proto.Message) {
	t.Helper()

	if diff := cmp.Diff(actual, expected, protocmp.Transform()); diff != "" {
		msg := fmt.Sprintf(
			"Not equal:\n"+
				"expected:\n\t'%s'\n"+
				"actual:\n\t'%s'\n"+
				"diff (-expected +actual):\n%s",
			expected, actual, diff,
		)
		assert.Fail(t, msg)
	}
}

func Marshal(t *testing.T, v interface{}) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return data
}
