package asset_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/user"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			err := json.Unmarshal(tc.patchDataJSON, &patchData)
			require.NoError(t, err)

			tc.asset.Patch(patchData)
			assert.Equal(t, tc.expected, tc.asset)
		})
	}
}

func TestToProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	dataPB, err := structpb.NewStruct(map[string]interface{}{
		"data1": "datavalue1",
	})
	if err != nil {
		t.Fatal(err)
	}

	labelPB, err := structpb.NewStruct(map[string]interface{}{
		"label1": "labelvalue1",
	})
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		Title       string
		Asset       asset.Asset
		ExpectProto *compassv1beta1.Asset
	}

	var testCases = []testCase{
		{
			Title:       "should return nil data pb, label pb, empty owners pb, nil changelog pb, no timestamp pb if data is empty",
			Asset:       asset.Asset{ID: "id1", URN: "urn1"},
			ExpectProto: &compassv1beta1.Asset{Id: "id1", Urn: "urn1"},
		},
		{
			Title: "should return full pb if all fileds are not zero",
			Asset: asset.Asset{
				ID:  "id1",
				URN: "urn1",
				Data: map[string]interface{}{
					"data1": "datavalue1",
				},
				Labels: map[string]string{
					"label1": "labelvalue1",
				},
				Changelog: diff.Changelog{
					diff.Change{
						From: "1",
						To:   "2",
						Path: []string{"path1/path2"},
					},
				},
				CreatedAt: timeDummy,
				UpdatedAt: timeDummy,
			},
			ExpectProto: &compassv1beta1.Asset{
				Id:     "id1",
				Urn:    "urn1",
				Data:   dataPB,
				Labels: labelPB,
				Changelog: &compassv1beta1.Changelog{
					Changes: []*compassv1beta1.Change{
						{

							From: structpb.NewStringValue("1"),
							To:   structpb.NewStringValue("2"),
							Path: []string{"path1/path2"},
						},
					},
				},
				CreatedAt: timestamppb.New(timeDummy),
				UpdatedAt: timestamppb.New(timeDummy),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got, err := tc.Asset.ToProto()
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(got, tc.ExpectProto, protocmp.Transform()); diff != "" {
				t.Errorf("expected response to be %+v, was %+v", tc.ExpectProto, got)
			}
		})
	}
}

func TestNewFromProto(t *testing.T) {
	timeDummy := time.Date(2000, time.January, 7, 0, 0, 0, 0, time.UTC)
	dataPB, err := structpb.NewStruct(map[string]interface{}{
		"data1": "datavalue1",
	})
	if err != nil {
		t.Fatal(err)
	}

	labelPB, err := structpb.NewStruct(map[string]interface{}{
		"label1": "labelvalue1",
	})
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		Title       string
		AssetPB     *compassv1beta1.Asset
		ExpectAsset asset.Asset
	}

	var testCases = []testCase{
		{
			Title:       "should return empty labels, data, and owners if all pb empty",
			AssetPB:     &compassv1beta1.Asset{Id: "id1"},
			ExpectAsset: asset.Asset{ID: "id1"},
		},
		{
			Title: "should return non empty labels, data, and owners if all pb is not empty",
			AssetPB: &compassv1beta1.Asset{
				Id:     "id1",
				Urn:    "urn1",
				Name:   "name1",
				Data:   dataPB,
				Labels: labelPB,
				Owners: []*compassv1beta1.User{
					{
						Id: "uid1",
					},
					{
						Id: "uid2",
					},
				},
				Changelog: &compassv1beta1.Changelog{
					Changes: []*compassv1beta1.Change{
						{

							From: structpb.NewStringValue("1"),
							To:   structpb.NewStringValue("2"),
							Path: []string{"path1/path2"},
						},
					},
				},
				CreatedAt: timestamppb.New(timeDummy),
				UpdatedAt: timestamppb.New(timeDummy),
			},
			ExpectAsset: asset.Asset{
				ID:   "id1",
				URN:  "urn1",
				Name: "name1",
				Data: map[string]interface{}{
					"data1": "datavalue1",
				},
				Labels: map[string]string{
					"label1": "labelvalue1",
				},
				Owners: []user.User{
					{
						ID: "uid1",
					},
					{
						ID: "uid2",
					},
				},
				Changelog: diff.Changelog{
					diff.Change{
						From: "1",
						To:   "2",
						Path: []string{"path1/path2"},
					},
				},
				CreatedAt: timeDummy,
				UpdatedAt: timeDummy,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {

			got := asset.NewFromProto(tc.AssetPB)
			if reflect.DeepEqual(got, tc.ExpectAsset) == false {
				t.Errorf("expected returned asset to be to be %+v, was %+v", tc.ExpectAsset, got)
			}
		})
	}
}
