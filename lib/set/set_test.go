package set_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/odpf/columbus/lib/set"
)

func TestStringSet(t *testing.T) {
	t.Run("json encode", func(t *testing.T) {
		s := set.NewStringSet("1337")

		var buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(s)
		if err != nil {
			t.Fatalf("error serialising string set: %v", err)
			return
		}

		got := strings.TrimSpace(buf.String())
		want := `["1337"]`

		if want != got {
			t.Errorf("expected %#v = %q, was %q", s, want, got)
			return
		}
	})
	t.Run("json decode", func(t *testing.T) {

		// we start with a non-empty set to test if it correctly
		// empties the set before decoding the JSON values
		got := set.NewStringSet("420")
		var raw = `["1337"]`
		err := json.NewDecoder(strings.NewReader(raw)).Decode(&got)
		if err != nil {
			t.Fatalf("error deserialising string set: %v", err)
			return
		}

		want := set.NewStringSet("1337")

		if reflect.DeepEqual(got, want) == false {
			t.Errorf("expected %q to decode to %#v, but was %#v", raw, want, got)
			return
		}
	})
}
