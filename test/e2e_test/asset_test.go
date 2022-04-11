//go:build e2e
// +build e2e

package endtoend_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/odpf/columbus/asset"
	"github.com/stretchr/testify/suite"
)

type AssetEndToEndTestSuite struct {
	suite.Suite
	ctx    context.Context
	client *Client
}

func (r *AssetEndToEndTestSuite) SetupSuite() {
	r.client = NewClient()
}

func (r *AssetEndToEndTestSuite) TestAllNormalFlow() {
	// create 5 assets, get all, get 1, get 1 asset version, patch 1 asset 2 times, get 1 asset version, get asset version v0.3
	assetIDs := []string{}
	for i := 0; i < 5; i++ {
		uniqueAssetURN := strings.ReplaceAll(uuid.NewString()+r.T().Name(), "/", "-")
		uniqueName := strings.ReplaceAll(r.T().Name()+" "+fmt.Sprintf("%d", (i+1)), "/", "-")
		ast := generateAsset(uniqueAssetURN, uniqueName)
		id, err := r.client.PatchAsset(ast)
		if err != nil {
			r.T().Fatal(err)
		}
		assetIDs = append(assetIDs, id)
	}

	// Get all assets
	retreivedAssets, err := r.client.GetAllAssets()
	if err != nil {
		r.T().Fatal(err)
	}
	r.Len(retreivedAssets, 5)

	// GetAsset
	sampleAsset, err := r.client.GetAnAsset(assetIDs[0])
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.Version, "0.1")
	// PatchAsset
	descriptionV2 := "new description v0.2"
	descriptionV3 := "new description v0.3"
	sampleAsset.Description = descriptionV2
	id, err := r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(id, sampleAsset.ID)

	sampleAsset.Description = descriptionV3
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(id, sampleAsset.ID)

	// Get All Versions
	assetVersions, err := r.client.GetAssetVersions(sampleAsset.ID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Len(assetVersions, 3)

	// Get Latest Version
	sampleAsset, err = r.client.GetAnAsset(sampleAsset.ID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.Version, "0.3")

	// Get a specific Version
	sampleAsset, err = r.client.GetAssetWithVersion(sampleAsset.ID, "0.3")
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.Description, descriptionV3)
	r.Equal(sampleAsset.Version, "0.3")

}

func (r *AssetEndToEndTestSuite) TestPatchAssetsAllFields() {
	uniqueAssetURN := strings.ReplaceAll(uuid.NewString()+r.T().Name(), "/", "-")
	uniqueName := strings.ReplaceAll(r.T().Name(), "/", "-")
	ast := generateAsset(uniqueAssetURN, uniqueName)
	assetID, err := r.client.PatchAsset(ast)
	if err != nil {
		r.T().Fatal(err)
	}

	// GetAsset
	sampleAsset, err := r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.1", sampleAsset.Version)

	// v0.2 PatchAsset field name
	sampleAsset.Name = "new name"
	id, err := r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err := r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.2", retrievedAsset.Version)
	r.Equal("new name", retrievedAsset.Name)

	// v0.3 PatchAsset field data update data type
	sampleAsset.Data["key1"] = 987
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.3", retrievedAsset.Version)
	r.Equal(float64(987), retrievedAsset.Data["key1"])

	// v0.4 PatchAsset field data update nested data type
	sampleAsset.Data["key3"].(map[string]interface{})["key31"] = 987
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(id, sampleAsset.ID)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.4", retrievedAsset.Version)
	r.Equal(float64(987), retrievedAsset.Data["key3"].(map[string]interface{})["key31"])

	// v0.5 PatchAsset field data add new bool entry
	sampleAsset.Data["key4"] = true
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.5", retrievedAsset.Version)
	r.Equal(true, retrievedAsset.Data["key4"])

	// v0.6 PatchAsset field data remove entry
	sampleAsset.Data["key2"] = nil
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.6", retrievedAsset.Version)
	r.Equal(nil, retrievedAsset.Data["key2"])

	// v0.7 PatchAsset field data add new nested bool entry
	sampleAsset.Data["key3"].(map[string]interface{})["key34"] = true
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.7", retrievedAsset.Version)
	r.Equal(true, retrievedAsset.Data["key3"].(map[string]interface{})["key34"])

	// v0.8 PatchAsset field data remove entry
	sampleAsset.Data["key3"].(map[string]interface{})["key32"] = nil
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.8", retrievedAsset.Version)
	r.Equal(nil, sampleAsset.Data["key3"].(map[string]interface{})["key32"])

	// v0.9 PatchAsset field label update value
	sampleAsset.Labels["label1"] = "new label 1"
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.9", retrievedAsset.Version)
	r.Equal("new label 1", retrievedAsset.Labels["label1"])

	// v0.10 PatchAsset field label add new entry
	sampleAsset.Labels["label4"] = "new label 4"
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(sampleAsset.ID, id)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.10", retrievedAsset.Version)
	r.Equal("new label 4", retrievedAsset.Labels["label4"])

	// v0.11 PatchAsset field label remove entry
	delete(sampleAsset.Labels, "label2")
	id, err = r.client.PatchAsset(sampleAsset)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal(id, sampleAsset.ID)

	retrievedAsset, err = r.client.GetAnAsset(assetID)
	if err != nil {
		r.T().Fatal(err)
	}
	r.Equal("0.11", retrievedAsset.Version)
	_, ok := retrievedAsset.Labels["label2"]
	r.False(ok)

}

func generateAsset(urn, name string) asset.Asset {
	return asset.Asset{
		URN:         urn,
		Type:        "table",
		Service:     "postgres",
		Name:        name,
		Description: "description about " + name,
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": map[string]interface{}{
				"key31": "value31",
				"key32": 123,
			},
		},
		Labels: map[string]string{
			"label1": "valuelabel1",
			"label2": "valuelabel2",
			"label3": "valuelabel3",
		},
	}
}

func TestAssetEndToEnd(t *testing.T) {
	suite.Run(t, &AssetEndToEndTestSuite{})
}
