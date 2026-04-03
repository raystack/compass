package chunking

import (
	"strings"
	"testing"

	"github.com/raystack/compass/core/entity"
)

func TestSerializeEntity_Basic(t *testing.T) {
	ent := entity.Entity{
		URN:         "urn:bigquery:project.dataset.orders",
		Type:        entity.TypeTable,
		Name:        "orders",
		Source:      "bigquery",
		Description: "Customer purchase orders since 2020",
	}

	chunks := SerializeEntity(ent)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	c := chunks[0]
	if !strings.Contains(c.Content, "Name: orders") {
		t.Error("expected content to contain 'Name: orders'")
	}
	if !strings.Contains(c.Content, "Type: table") {
		t.Error("expected content to contain 'Type: table'")
	}
	if !strings.Contains(c.Content, "Source: bigquery") {
		t.Error("expected content to contain 'Source: bigquery'")
	}
	if !strings.Contains(c.Content, "Customer purchase orders") {
		t.Error("expected content to contain description")
	}
	if c.Context != "table: orders (urn:bigquery:project.dataset.orders)" {
		t.Errorf("unexpected context: %q", c.Context)
	}
}

func TestSerializeEntity_WithProperties(t *testing.T) {
	ent := entity.Entity{
		URN:  "urn:table:test",
		Type: entity.TypeTable,
		Name: "test",
		Properties: map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "id", "type": "INT"},
				map[string]interface{}{"name": "email", "type": "VARCHAR"},
			},
			"tags": []interface{}{"pii", "production"},
		},
	}

	chunks := SerializeEntity(ent)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	content := chunks[0].Content
	if !strings.Contains(content, "id (INT)") {
		t.Error("expected serialized column 'id (INT)'")
	}
	if !strings.Contains(content, "pii, production") {
		t.Error("expected serialized tags")
	}
}

func TestSerializeEntity_Empty(t *testing.T) {
	ent := entity.Entity{
		URN:  "urn:empty",
		Type: "service",
		Name: "empty-service",
	}

	chunks := SerializeEntity(ent)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if !strings.Contains(chunks[0].Content, "Name: empty-service") {
		t.Error("expected basic fields in content")
	}
}
