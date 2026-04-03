package chunking

import (
	"strings"
	"testing"
)

func TestSplitDocument_SingleSection(t *testing.T) {
	body := "This is a short document about the orders table."
	chunks := SplitDocument("Orders", body, Options{MaxTokens: 512})

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Context != "Document: Orders" {
		t.Errorf("unexpected context: %q", chunks[0].Context)
	}
	if chunks[0].Heading != "Introduction" {
		t.Errorf("unexpected heading: %q", chunks[0].Heading)
	}
}

func TestSplitDocument_MultipleSections(t *testing.T) {
	body := `## Overview
This table tracks customer orders.

## Schema
The table has columns: id, customer_id, amount, created_at.

## Usage
Query the orders_mart view instead of this table directly.`

	chunks := SplitDocument("Orders Table", body, Options{MaxTokens: 512})

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	if chunks[0].Heading != "Overview" {
		t.Errorf("expected heading 'Overview', got %q", chunks[0].Heading)
	}
	if chunks[0].Context != "Document: Orders Table > Section: Overview" {
		t.Errorf("unexpected context: %q", chunks[0].Context)
	}
	if chunks[1].Heading != "Schema" {
		t.Errorf("expected heading 'Schema', got %q", chunks[1].Heading)
	}
	if chunks[2].Heading != "Usage" {
		t.Errorf("expected heading 'Usage', got %q", chunks[2].Heading)
	}
}

func TestSplitDocument_LargeSection(t *testing.T) {
	// Create a large section that exceeds 20 tokens
	var b strings.Builder
	b.WriteString("## Procedures\n")
	for i := 0; i < 20; i++ {
		b.WriteString("This is paragraph number one with enough words to count.\n\n")
	}

	chunks := SplitDocument("Runbook", b.String(), Options{MaxTokens: 30})

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for large section, got %d", len(chunks))
	}
	for _, c := range chunks {
		if c.Heading != "Procedures" {
			t.Errorf("expected heading 'Procedures', got %q", c.Heading)
		}
	}
}

func TestSplitDocument_Empty(t *testing.T) {
	chunks := SplitDocument("Empty", "", Options{})
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty doc, got %d", len(chunks))
	}
}

func TestSplitDocument_HeadingLevels(t *testing.T) {
	body := `# Top Level
Content under h1.

### Sub Level
Content under h3.

#### Deep Level
Content under h4.`

	chunks := SplitDocument("Doc", body, Options{MaxTokens: 512})
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0].Heading != "Top Level" {
		t.Errorf("expected 'Top Level', got %q", chunks[0].Heading)
	}
	if chunks[1].Heading != "Sub Level" {
		t.Errorf("expected 'Sub Level', got %q", chunks[1].Heading)
	}
	if chunks[2].Heading != "Deep Level" {
		t.Errorf("expected 'Deep Level', got %q", chunks[2].Heading)
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"one two three four", 5}, // 4 words * 4/3 ≈ 5
	}

	for _, tt := range tests {
		got := EstimateTokens(tt.text)
		if got != tt.expected {
			t.Errorf("EstimateTokens(%q) = %d, want %d", tt.text, got, tt.expected)
		}
	}
}
