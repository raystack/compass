package asset_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/goto/compass/core/asset"
)

func TestToAsset(t *testing.T) {
	type testCase struct {
		Title        string
		SearchResult asset.SearchResult
		Expect       asset.Asset
	}

	var testCases = []testCase{
		{
			Title: "should return correct asset",
			SearchResult: asset.SearchResult{
				ID:          "an-id",
				URN:         "an-urn",
				Title:       "a-title",
				Type:        "table",
				Service:     "a-service",
				Description: "a-description",
				Labels: map[string]string{
					"label1": "value1",
				},
			},
			Expect: asset.Asset{
				ID:          "an-id",
				URN:         "an-urn",
				Name:        "a-title",
				Type:        asset.TypeTable,
				Service:     "a-service",
				Description: "a-description",
				Labels: map[string]string{
					"label1": "value1",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			got := tc.SearchResult.ToAsset()
			if diff := cmp.Diff(got, tc.Expect); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.Expect, got)
			}
		})
	}
}
