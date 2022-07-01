package asset_test

import (
	"encoding/json"
	"testing"

	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/user"
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
		{
			"created owners should be reflected",
			`{
				"name":	"old-name"
			}`,
			`{
				"name":	"old-name",
				"owners": [
					{
						"email": "email@odpf.io"
					}
				]
			}`,
			diff.Changelog{
				diff.Change{Type: diff.CREATE, Path: []string{"owners", "0", "email"}, To: "email@odpf.io"},
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

func TestAssetPatch(t *testing.T) {
	testcases := []struct {
		description   string
		asset         asset.Asset
		patchDataJSON json.RawMessage
		patchData     map[string]interface{}
		expected      asset.Asset
	}{
		{
			description: "should patch all allowed fields",
			asset: asset.Asset{
				URN:         "some-urn",
				Type:        asset.TypeJob,
				Service:     "optimus",
				Description: "sample-description",
				Name:        "old-name",
				Labels: map[string]string{
					"foo": "bar",
				},
				Owners: []user.User{
					{Email: "old@example.com"},
				},
			},
			patchDataJSON: []byte(`{
				"urn":         "new-urn",
				"type":        "table",
				"service":     "firehose",
				"description": "new-description",
				"name":        "new-name",
				"labels": {
					"bar":  "foo",
					"bar2": "foo2"
				},
				"owners": [
					{"email": "new@example.com"},
					{"email": "new2@example.com"}
				]
			}`),
			expected: asset.Asset{
				URN:         "new-urn",
				Type:        asset.TypeTable,
				Service:     "firehose",
				Description: "new-description",
				Name:        "new-name",
				Labels: map[string]string{
					"bar":  "foo",
					"bar2": "foo2",
				},
				Owners: []user.User{
					{Email: "new@example.com"},
					{Email: "new2@example.com"},
				},
			},
		},
		{
			description: "should patch all allowed fields without JSON",
			asset: asset.Asset{
				URN:         "some-urn",
				Type:        asset.TypeJob,
				Service:     "optimus",
				Description: "sample-description",
				Name:        "old-name",
				Labels: map[string]string{
					"foo": "bar",
				},
				Owners: []user.User{
					{Email: "old@example.com"},
				},
			},
			patchData: map[string]interface{}{
				"urn":         "new-urn",
				"type":        "table",
				"service":     "firehose",
				"description": "new-description",
				"name":        "new-name",
				"labels": map[string]string{
					"bar":  "foo",
					"bar2": "foo2",
				},
				"owners": []map[string]interface{}{
					{"email": "new@example.com"},
					{"email": "new2@example.com"},
				},
			},
			expected: asset.Asset{
				URN:         "new-urn",
				Type:        asset.TypeTable,
				Service:     "firehose",
				Description: "new-description",
				Name:        "new-name",
				Labels: map[string]string{
					"bar":  "foo",
					"bar2": "foo2",
				},
				Owners: []user.User{
					{Email: "new@example.com"},
					{Email: "new2@example.com"},
				},
			},
		}, {
			description: "should patch all allowed fields without labels and owners",
			asset: asset.Asset{
				URN:         "some-urn",
				Type:        asset.TypeJob,
				Service:     "optimus",
				Description: "sample-description",
				Name:        "old-name",
				Labels: map[string]string{
					"foo": "bar",
				},
				Owners: []user.User{
					{Email: "old@example.com"},
				},
			},
			patchData: map[string]interface{}{
				"urn":         "new-urn",
				"type":        "table",
				"service":     "firehose",
				"description": "new-description",
				"name":        "new-name",
				"labels":      "",
				"owners":      "",
			},
			expected: asset.Asset{
				URN:         "new-urn",
				Type:        asset.TypeTable,
				Service:     "firehose",
				Description: "new-description",
				Name:        "new-name",
			},
		},
		{
			description: "should patch data field",
			asset: asset.Asset{
				Data: map[string]interface{}{
					"user": map[string]interface{}{
						"name":  "sample-name",
						"email": "sample@test.com",
					},
					"properties": map[string]interface{}{
						"attributes": map[string]interface{}{
							"entity":      "odpf",
							"environment": "staging",
						},
					},
				},
			},
			patchDataJSON: []byte(`{
				"data": {
					"user": {
						"email": "new-email@test.com",
						"description": "user description"
					},
					"schemas": [
						"schema1",
						"schema2"
					],
					"properties": {
						"attributes": {
							"environment": "production",
							"type": "some-type"
						}
					}
				}
			}`),
			expected: asset.Asset{
				Data: map[string]interface{}{
					"user": map[string]interface{}{
						"name":        "sample-name",
						"email":       "new-email@test.com",
						"description": "user description",
					},
					"properties": map[string]interface{}{
						"attributes": map[string]interface{}{
							"entity":      "odpf",
							"environment": "production",
							"type":        "some-type",
						},
					},
					"schemas": []interface{}{
						"schema1",
						"schema2",
					},
				},
			},
		},
		{
			description: "should not panic if current asset's data is nil",
			asset: asset.Asset{
				Data: nil,
			},
			patchDataJSON: []byte(`{
				"data": {
					"foo": "bar"
				}
			}`),
			expected: asset.Asset{
				Data: map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			var patchData map[string]interface{}
			if tc.patchDataJSON != nil {
				err := json.Unmarshal(tc.patchDataJSON, &patchData)
				assert.NoError(t, err)
			} else {
				patchData = tc.patchData
			}
			tc.asset.Patch(patchData)
			assert.Equal(t, tc.expected, tc.asset)
		})
	}
}
