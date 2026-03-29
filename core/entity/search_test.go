package entity

import (
	"testing"
)

func TestReciprocalRankFusion(t *testing.T) {
	list1 := []SearchResult{
		{URN: "a", Name: "Alpha"},
		{URN: "b", Name: "Beta"},
		{URN: "c", Name: "Charlie"},
	}
	list2 := []SearchResult{
		{URN: "b", Name: "Beta"},
		{URN: "d", Name: "Delta"},
		{URN: "a", Name: "Alpha"},
	}

	fused := reciprocalRankFusion(list1, list2)

	if len(fused) != 4 {
		t.Fatalf("expected 4 results, got %d", len(fused))
	}

	// "a" and "b" appear in both lists, should be ranked higher
	// "b" is rank 2 in list1 and rank 1 in list2 → strong
	// "a" is rank 1 in list1 and rank 3 in list2 → strong
	topTwo := map[string]bool{fused[0].URN: true, fused[1].URN: true}
	if !topTwo["a"] || !topTwo["b"] {
		t.Errorf("expected a and b in top 2, got %s and %s", fused[0].URN, fused[1].URN)
	}
}

func TestReciprocalRankFusion_EmptyLists(t *testing.T) {
	fused := reciprocalRankFusion(nil, nil)
	if len(fused) != 0 {
		t.Fatalf("expected 0 results, got %d", len(fused))
	}
}

func TestReciprocalRankFusion_SingleList(t *testing.T) {
	list := []SearchResult{
		{URN: "a"},
		{URN: "b"},
	}
	fused := reciprocalRankFusion(list)
	if len(fused) != 2 {
		t.Fatalf("expected 2 results, got %d", len(fused))
	}
	if fused[0].URN != "a" {
		t.Errorf("expected first result to be 'a', got %q", fused[0].URN)
	}
}
