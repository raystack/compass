package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeString(t *testing.T) {
	for typ, expected := range map[Type]string{
		TypeDashboard:    "dashboard",
		TypeJob:          "job",
		TypeTable:        "table",
		TypeTopic:        "topic",
		TypeFeatureTable: "feature_table",
		TypeApplication:  "application",
		TypeModel:        "model",
	} {
		t.Run((string)(typ), func(t *testing.T) {
			assert.Equal(t, expected, typ.String())
		})
	}
}

func TestTypeIsValid(t *testing.T) {
	for _, typ := range []Type{
		"dashboard", "job", "table", "topic", "feature_table", "application", "model",
	} {
		t.Run((string)(typ), func(t *testing.T) {
			assert.Truef(t, typ.IsValid(), "%s should be valid", typ)
		})
	}

	if typ := Type("random"); typ.IsValid() {
		t.Fatalf("type %s should not be valid", typ)
	}
}
