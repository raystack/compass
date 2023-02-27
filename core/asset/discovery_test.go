package asset_test

import (
	"github.com/odpf/compass/core/namespace"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/odpf/compass/core/asset"
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

func TestSearchConfig_Validate(t *testing.T) {
	type fields struct {
		Text       string
		Filters    asset.SearchFilter
		MaxResults int
		RankBy     string
		Queries    map[string]string
		Namespace  *namespace.Namespace
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "fail validation if namespace is empty",
			fields: fields{
				Namespace: nil,
			},
			wantErr: true,
		},
		{
			name: "fail validation if text is empty",
			fields: fields{
				Text:      "",
				Namespace: &namespace.Namespace{},
			},
			wantErr: true,
		},
		{
			name: "should not fail validation if all required fields are non empty",
			fields: fields{
				Text:      "query",
				Namespace: &namespace.Namespace{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := asset.SearchConfig{
				Text:       tt.fields.Text,
				Filters:    tt.fields.Filters,
				MaxResults: tt.fields.MaxResults,
				RankBy:     tt.fields.RankBy,
				Queries:    tt.fields.Queries,
				Namespace:  tt.fields.Namespace,
			}
			if !tt.wantErr {
				assert.Nil(t, s.Validate())
			} else {
				assert.NotNil(t, s.Validate())
			}
		})
	}
}
