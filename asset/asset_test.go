package asset_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/odpf/columbus/asset"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffTopLevel(t *testing.T) {
	cases := []struct {
		Name           string
		Source, Target string
		Changelog      diff.Changelog
		Error          error
	}{
		{
			"ignored field won't be compared",
			`{
				"id": "1234",
				"urn": "urn1234",
				"type": "dashboard",
				"service": "service1234"	
			}`,
			`{
				"id": "5678",
				"urn": "urn5678",
				"type": "job",
				"service": "service5678"	
			}`,
			nil,
			nil,
		},
		{
			"updated top level field should be reflected",
			`{
				"name":	"old-name"
			}`,
			`{
				"name":	"updated-name",
				"description":	"updated-decsription"
			}`,
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"name"}, From: "old-name", To: "updated-name"},
				diff.Change{Type: diff.UPDATE, Path: []string{"description"}, From: "", To: "updated-decsription"},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			var sourceAsset asset.Asset
			err := json.Unmarshal([]byte(tc.Source), &sourceAsset)
			if err != nil {
				t.Fatal(err)
			}
			var targetAsset asset.Asset
			err = json.Unmarshal([]byte(tc.Target), &targetAsset)
			if err != nil {
				t.Fatal(err)
			}

			cl, err := sourceAsset.Diff(&targetAsset)

			assert.Equal(t, tc.Error, err)
			require.Equal(t, len(tc.Changelog), len(cl))

			fmt.Printf("%+v\n", cl)
			for i, c := range cl {
				assert.Equal(t, tc.Changelog[i].Type, c.Type)
				assert.Equal(t, tc.Changelog[i].Path, c.Path)
				assert.Equal(t, tc.Changelog[i].From, c.From)
				assert.Equal(t, tc.Changelog[i].To, c.To)
			}
		})
	}
}

func TestDiffData(t *testing.T) {
	cases := []struct {
		Name           string
		Source, Target string
		Changelog      diff.Changelog
		Error          error
	}{
		{
			"updated data value string should be reflected",
			`{
				"name": "jane-kafka-1a",
				"description": "",
				"data": {
				  "title": "jane-kafka-1a",
				  "entity": "odpf",
				  "country": "vn"
				}
			  }`,
			`{
				"name": "jane-kafka-1a",
				"service": "kafka",
				"description": "",
				"data": {
				  "title": "jane-kafka-1a",
				  "description": "a new description inside",
				  "entity": "odpf",
				  "country": "id"
				}
			  }`,
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"data", "country"}, From: "vn", To: "id"},
				diff.Change{Type: diff.CREATE, Path: []string{"data", "description"}, To: "a new description inside"},
			},
			nil,
		},
		{
			"updated data value array should be reflected",
			`{
				"name": "jane-kafka-1a",
				"data": {
				  "some_array": [
					  {
						  "id": "element1id"
					  }
				  ],
				  "entity": "odpf",
				  "country": "vn"
				}
			  }`,
			`{
				"name": "jane-kafka-1a",
				"data": {
					"some_array": [
						{
							"id": "element2id"
						}
					],
				  "entity": "odpf",
				  "country": "vn"
				}
			  }`,
			diff.Changelog{
				diff.Change{Type: diff.UPDATE, Path: []string{"data", "some_array", "0", "id"}, From: "element1id", To: "element2id"},
			},
			nil,
		},
		{
			"created data value array should be reflected",
			`{
				"name": "jane-kafka-1a",
				"data": {
				  "some_array": [
					  {
						  "id": "element1id"
					  }
				  ],
				  "entity": "odpf",
				  "country": "vn"
				}
			  }`,
			`{
				"name": "jane-kafka-1a",
				"data": {
					"some_array": [
						{
							"id": "element1id"
						},
						{
							"id": "element2id"
						}
					],
				  "entity": "odpf",
				  "country": "vn"
				}
			  }`,
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"data", "some_array", "1"}, To: map[string]interface{}(map[string]interface{}{"id": "element2id"})},
			},
			nil,
		},
		{
			"deleted data value array should be reflected",
			`{
				"name": "jane-kafka-1a",
				"data": {
					"some_array": [
						{
							"id": "element1id"
						},
						{
							"id": "element2id"
						}
					],
					"entity": "odpf",
					"country": "vn"
				}
			  }`,
			`{
				"name": "jane-kafka-1a",
				"data": {
					"some_array": [
						{
							"id": "element1id"
						}
					],
					"entity": "odpf",
					"country": "vn"
				}
			  }`,
			diff.Changelog{
				diff.Change{Type: diff.DELETE, Path: []string{"data", "some_array", "1"}, From: map[string]interface{}(map[string]interface{}{"id": "element2id"})},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {

			var sourceAsset asset.Asset
			err := json.Unmarshal([]byte(tc.Source), &sourceAsset)
			if err != nil {
				t.Fatal(err)
			}
			var targetAsset asset.Asset
			err = json.Unmarshal([]byte(tc.Target), &targetAsset)
			if err != nil {
				t.Fatal(err)
			}

			cl, err := sourceAsset.Diff(&targetAsset)

			assert.Equal(t, tc.Error, err)
			require.Equal(t, len(tc.Changelog), len(cl))

			for i, c := range cl {
				assert.Equal(t, tc.Changelog[i].Type, c.Type)
				assert.Equal(t, tc.Changelog[i].Path, c.Path)
				assert.Equal(t, tc.Changelog[i].From, c.From)
				assert.Equal(t, tc.Changelog[i].To, c.To)
			}
		})
	}
}
