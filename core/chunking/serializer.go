package chunking

import (
	"fmt"
	"sort"
	"strings"

	"github.com/raystack/compass/core/entity"
)

// SerializeEntity converts an entity's metadata into readable text for embedding.
// Most entities produce a single chunk.
func SerializeEntity(ent entity.Entity) []Chunk {
	var b strings.Builder

	fmt.Fprintf(&b, "Name: %s\n", ent.Name)
	fmt.Fprintf(&b, "Type: %s\n", ent.Type)
	fmt.Fprintf(&b, "URN: %s\n", ent.URN)
	if ent.Source != "" {
		fmt.Fprintf(&b, "Source: %s\n", ent.Source)
	}
	if ent.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", ent.Description)
	}

	if len(ent.Properties) > 0 {
		b.WriteString("Properties:\n")
		serializeProperties(&b, ent.Properties, "  ")
	}

	context := fmt.Sprintf("%s: %s (%s)", ent.Type, ent.Name, ent.URN)

	return []Chunk{{
		Content:  b.String(),
		Context:  context,
		Heading:  string(ent.Type) + ": " + ent.Name,
		Position: 0,
	}}
}

func serializeProperties(b *strings.Builder, props map[string]interface{}, indent string) {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := props[k]
		switch val := v.(type) {
		case map[string]interface{}:
			fmt.Fprintf(b, "%s%s:\n", indent, k)
			serializeProperties(b, val, indent+"  ")
		case []interface{}:
			serializeSlice(b, k, val, indent)
		default:
			fmt.Fprintf(b, "%s%s: %v\n", indent, k, v)
		}
	}
}

func serializeSlice(b *strings.Builder, key string, items []interface{}, indent string) {
	if len(items) == 0 {
		return
	}

	// Check if items are simple strings/numbers
	allSimple := true
	for _, item := range items {
		switch item.(type) {
		case string, float64, int, bool:
		default:
			allSimple = false
		}
	}

	if allSimple {
		vals := make([]string, len(items))
		for i, item := range items {
			vals[i] = fmt.Sprintf("%v", item)
		}
		fmt.Fprintf(b, "%s%s: %s\n", indent, key, strings.Join(vals, ", "))
		return
	}

	// Complex items (e.g., columns)
	fmt.Fprintf(b, "%s%s:\n", indent, key)
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			// Try to format as "name (type)" for column-like structures
			name, _ := m["name"].(string)
			typ, _ := m["type"].(string)
			if name != "" && typ != "" {
				fmt.Fprintf(b, "%s  - %s (%s)\n", indent, name, typ)
				continue
			}
		}
		fmt.Fprintf(b, "%s  - %v\n", indent, item)
	}
}
