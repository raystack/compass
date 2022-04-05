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

// func (r *AssetRepositoryTestSuite) createUsers(userRepo user.Repository) []user.User {
// 	var err error
// 	users := []user.User{}

// 	user1 := user.User{Email: "user-test-1@odpf.io", Provider: defaultProviderName}
// 	user1.ID, err = userRepo.Create(r.ctx, &user1)
// 	r.Require().NoError(err)
// 	users = append(users, user1)

// 	user2 := user.User{Email: "user-test-2@odpf.io", Provider: defaultProviderName}
// 	user2.ID, err = userRepo.Create(r.ctx, &user2)
// 	r.Require().NoError(err)
// 	users = append(users, user2)

// 	user3 := user.User{Email: "user-test-3@odpf.io", Provider: defaultProviderName}
// 	user3.ID, err = userRepo.Create(r.ctx, &user3)
// 	r.Require().NoError(err)
// 	users = append(users, user3)

// 	user4 := user.User{Email: "user-test-4@odpf.io", Provider: defaultProviderName}
// 	user4.ID, err = userRepo.Create(r.ctx, &user4)
// 	r.Require().NoError(err)
// 	users = append(users, user4)

// 	return users
// }

// func (r *AssetRepositoryTestSuite) TearDownSuite() {
// 	// Clean tests
// 	err := r.client.Close()
// 	if err != nil {
// 		r.T().Fatal(err)
// 	}
// 	err = purgeDocker(r.pool, r.resource)
// 	if err != nil {
// 		r.T().Fatal(err)
// 	}
// }

// func (r *AssetRepositoryTestSuite) TestGetAll() {
// 	// populate assets
// 	total := 12
// 	assets := []asset.Asset{}
// 	for i := 0; i < total; i++ {
// 		var typ asset.Type
// 		var service string
// 		if (i % 2) == 0 { // if even
// 			typ = asset.TypeJob
// 			service = "postgres"
// 		} else {
// 			typ = asset.TypeDashboard
// 			service = "metabase"
// 		}

// 		ast := asset.Asset{
// 			URN:       fmt.Sprintf("urn-get-%d", i),
// 			Type:      typ,
// 			Service:   service,
// 			Version:   asset.BaseVersion,
// 			UpdatedBy: r.users[0],
// 		}
// 		id, err := r.repository.Upsert(r.ctx, &ast)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		ast.ID = id
// 		assets = append(assets, ast)
// 	}

// 	r.Run("should return all assets limited by default size", func() {
// 		results, err := r.repository.GetAll(r.ctx, asset.Config{})
// 		r.Require().NoError(err)
// 		r.Require().Len(results, defaultGetMaxSize)
// 		for i := 0; i < defaultGetMaxSize; i++ {
// 			r.assertAsset(&assets[i], &results[i])
// 		}
// 	})

// 	r.Run("should override default size using GetConfig.Size", func() {
// 		size := 8
// 		results, err := r.repository.GetAll(r.ctx, asset.Config{
// 			Size: size,
// 		})
// 		r.Require().NoError(err)
// 		r.Require().Len(results, size)
// 		for i := 0; i < size; i++ {
// 			r.assertAsset(&assets[i], &results[i])
// 		}
// 	})

// 	r.Run("should fetch assets by offset defined in GetConfig.Offset", func() {
// 		offset := 2
// 		results, err := r.repository.GetAll(r.ctx, asset.Config{
// 			Offset: offset,
// 		})
// 		r.Require().NoError(err)
// 		for i := offset; i < defaultGetMaxSize+offset; i++ {
// 			r.assertAsset(&assets[i], &results[i-offset])

// 		}
// 	})

// 	r.Run("should filter using type", func() {
// 		results, err := r.repository.GetAll(r.ctx, asset.Config{
// 			Type: asset.TypeDashboard,
// 			Size: total,
// 		})
// 		r.Require().NoError(err)
// 		r.Require().Len(results, total/2)
// 		for _, ast := range results {
// 			r.Equal(asset.TypeDashboard, ast.Type)
// 		}
// 	})

// 	r.Run("should filter using service", func() {
// 		results, err := r.repository.GetAll(r.ctx, asset.Config{
// 			Service: "postgres",
// 			Size:    total,
// 		})
// 		r.Require().NoError(err)
// 		r.Require().Len(results, total/2)
// 		for _, ast := range results {
// 			r.Equal("postgres", ast.Service)
// 		}
// 	})
// }

// func (r *AssetRepositoryTestSuite) TestGetCount() {
// 	// populate assets
// 	total := 12
// 	typ := asset.TypeJob
// 	service := "service-getcount"
// 	for i := 0; i < total; i++ {
// 		ast := asset.Asset{
// 			URN:       fmt.Sprintf("urn-getcount-%d", i),
// 			Type:      typ,
// 			Service:   service,
// 			UpdatedBy: r.users[0],
// 		}
// 		id, err := r.repository.Upsert(r.ctx, &ast)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		ast.ID = id
// 	}

// 	r.Run("should return total assets with filter", func() {
// 		actual, err := r.repository.GetCount(r.ctx, asset.Config{
// 			Type:    typ,
// 			Service: service,
// 		})
// 		r.Require().NoError(err)
// 		r.Equal(total, actual)
// 	})
// }

// func (r *AssetRepositoryTestSuite) TestGetByID() {
// 	r.Run("return error from client if asset not an uuid", func() {
// 		_, err := r.repository.GetByID(r.ctx, "invalid-uuid")
// 		r.Error(err)
// 		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
// 	})

// 	r.Run("return NotFoundError if asset does not exist", func() {
// 		uuid := "2aabb450-f986-44e2-a6db-7996861d5004"
// 		_, err := r.repository.GetByID(r.ctx, uuid)
// 		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuid})
// 	})

// 	r.Run("return correct asset from db", func() {
// 		asset1 := asset.Asset{
// 			URN:       "urn-gbi-1",
// 			Type:      "table",
// 			Service:   "bigquery",
// 			Version:   asset.BaseVersion,
// 			UpdatedBy: r.users[1],
// 		}
// 		asset2 := asset.Asset{
// 			URN:       "urn-gbi-2",
// 			Type:      "topic",
// 			Service:   "kafka",
// 			Version:   asset.BaseVersion,
// 			UpdatedBy: r.users[1],
// 		}

// 		var err error
// 		id, err := r.repository.Upsert(r.ctx, &asset1)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		asset1.ID = id

// 		id, err = r.repository.Upsert(r.ctx, &asset2)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		asset2.ID = id

// 		result, err := r.repository.GetByID(r.ctx, asset2.ID)
// 		r.NoError(err)
// 		asset2.UpdatedBy = r.users[1]
// 		r.assertAsset(&asset2, &result)
// 	})

// 	r.Run("return owners if any", func() {

// 		ast := asset.Asset{
// 			URN:     "urn-gbi-3",
// 			Type:    "table",
// 			Service: "bigquery",
// 			Owners: []user.User{
// 				r.users[1],
// 				r.users[2],
// 			},
// 			UpdatedBy: r.users[1],
// 		}

// 		id, err := r.repository.Upsert(r.ctx, &ast)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		ast.ID = id

// 		result, err := r.repository.GetByID(r.ctx, ast.ID)
// 		r.NoError(err)
// 		r.Len(result.Owners, len(ast.Owners))
// 		for i, owner := range result.Owners {
// 			r.Equal(ast.Owners[i].ID, owner.ID)
// 		}
// 	})
// }

// func (r *AssetRepositoryTestSuite) TestFind() {
// 	r.Run("return NotFoundError if asset does not exist", func() {
// 		urn := "some-urn"
// 		typ := asset.TypeDashboard
// 		service := "bigquery"
// 		_, err := r.repository.Find(r.ctx, urn, typ, service)
// 		r.ErrorAs(err, &asset.NotFoundError{URN: urn, Type: typ, Service: service})
// 	})

// 	r.Run("return correct asset from db", func() {
// 		asset1 := asset.Asset{
// 			URN:       "urn-find-1",
// 			Type:      "table",
// 			Service:   "bigquery",
// 			Version:   asset.BaseVersion,
// 			UpdatedBy: r.users[1],
// 		}
// 		asset2 := asset.Asset{
// 			URN:       "urn-find-2",
// 			Type:      "topic",
// 			Service:   "kafka",
// 			Version:   asset.BaseVersion,
// 			UpdatedBy: r.users[1],
// 		}

// 		var err error
// 		id, err := r.repository.Upsert(r.ctx, &asset1)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		asset1.ID = id

// 		id, err = r.repository.Upsert(r.ctx, &asset2)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		asset2.ID = id

// 		result, err := r.repository.Find(r.ctx, asset2.URN, asset2.Type, asset2.Service)
// 		r.NoError(err)
// 		asset2.UpdatedBy = r.users[1]
// 		r.assertAsset(&asset2, &result)

// 		// clean up
// 		err = r.repository.Delete(r.ctx, asset1.ID)
// 		r.Require().NoError(err)
// 		err = r.repository.Delete(r.ctx, asset2.ID)
// 		r.Require().NoError(err)
// 	})

// 	r.Run("return owners if any", func() {
// 		ast := asset.Asset{
// 			URN:     "urn-find-3",
// 			Type:    "table",
// 			Service: "bigquery",
// 			Owners: []user.User{
// 				r.users[1],
// 				r.users[2],
// 			},
// 			UpdatedBy: r.users[1],
// 		}

// 		id, err := r.repository.Upsert(r.ctx, &ast)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		ast.ID = id

// 		result, err := r.repository.Find(r.ctx, ast.URN, ast.Type, ast.Service)
// 		r.NoError(err)
// 		r.Len(result.Owners, len(ast.Owners))
// 		for i, owner := range result.Owners {
// 			r.Equal(ast.Owners[i].ID, owner.ID)
// 		}

// 		// clean up
// 		err = r.repository.Delete(r.ctx, ast.ID)
// 		r.Require().NoError(err)
// 	})
// }

// func (r *AssetRepositoryTestSuite) TestVersions() {
// 	assetURN := "urn-u-2-version"
// 	// v0.1
// 	astVersioning := asset.Asset{
// 		URN:       assetURN,
// 		Type:      "table",
// 		Service:   "bigquery",
// 		UpdatedBy: r.users[1],
// 	}

// 	id, err := r.repository.Upsert(r.ctx, &astVersioning)
// 	r.Require().NoError(err)
// 	r.Require().Equal(r.lengthOfString(id), 36)
// 	astVersioning.ID = id

// 	// v0.2
// 	astVersioning.Description = "new description in v0.2"
// 	id, err = r.repository.Upsert(r.ctx, &astVersioning)
// 	r.Require().NoError(err)
// 	r.Require().Equal(id, astVersioning.ID)

// 	// v0.3
// 	astVersioning.Owners = []user.User{
// 		{
// 			Email: "user@odpf.io",
// 		},
// 		{
// 			Email:    "meteor@odpf.io",
// 			Provider: "meteor",
// 		},
// 	}
// 	id, err = r.repository.Upsert(r.ctx, &astVersioning)
// 	r.Require().NoError(err)
// 	r.Require().Equal(id, astVersioning.ID)

// 	// v0.4
// 	astVersioning.Data = map[string]interface{}{
// 		"data1": float64(12345),
// 	}
// 	id, err = r.repository.Upsert(r.ctx, &astVersioning)
// 	r.Require().NoError(err)
// 	r.Require().Equal(id, astVersioning.ID)

// 	// v0.5
// 	astVersioning.Labels = map[string]string{
// 		"key1": "value1",
// 	}

// 	id, err = r.repository.Upsert(r.ctx, &astVersioning)
// 	r.Require().NoError(err)
// 	r.Require().Equal(id, astVersioning.ID)

// 	r.Run("should return 3 last versions of an assets if there are exist", func() {

// 		expectedAssetVersions := []asset.Asset{
// 			{
// 				ID:      astVersioning.ID,
// 				URN:     assetURN,
// 				Type:    "table",
// 				Service: "bigquery",
// 				Version: "0.4",
// 				Changelog: diff.Changelog{
// 					diff.Change{Type: "create", Path: []string{"labels", "key1"}, From: interface{}(nil), To: "value1"},
// 				},
// 				UpdatedBy: r.users[1],
// 			},
// 			{
// 				ID:      astVersioning.ID,
// 				URN:     assetURN,
// 				Type:    "table",
// 				Service: "bigquery",
// 				Version: "0.3",
// 				Changelog: diff.Changelog{
// 					diff.Change{Type: "create", Path: []string{"data", "data1"}, From: interface{}(nil), To: float64(12345)},
// 				},
// 				UpdatedBy: r.users[1],
// 			},
// 			{
// 				ID:      astVersioning.ID,
// 				URN:     assetURN,
// 				Type:    "table",
// 				Service: "bigquery",
// 				Version: "0.2",
// 				Changelog: diff.Changelog{
// 					diff.Change{Type: "create", Path: []string{"owners", "0", "email"}, From: interface{}(nil), To: "user@odpf.io"},
// 					diff.Change{Type: "create", Path: []string{"owners", "1", "email"}, From: interface{}(nil), To: "meteor@odpf.io"},
// 				},
// 				UpdatedBy: r.users[1],
// 			},
// 		}

// 		assetVersions, err := r.repository.GetVersionHistory(r.ctx, asset.Config{Size: 3}, astVersioning.ID)
// 		r.NoError(err)
// 		// making updatedby user time empty to make ast comparable
// 		for i := 0; i < len(assetVersions); i++ {
// 			assetVersions[i].UpdatedBy.CreatedAt = time.Time{}
// 			assetVersions[i].UpdatedBy.UpdatedAt = time.Time{}
// 			assetVersions[i].CreatedAt = time.Time{}
// 			assetVersions[i].UpdatedAt = time.Time{}
// 		}
// 		r.Equal(expectedAssetVersions, assetVersions)
// 	})

// 	r.Run("should return current version of an assets", func() {
// 		expectedLatestVersion := asset.Asset{
// 			ID:          astVersioning.ID,
// 			URN:         assetURN,
// 			Type:        "table",
// 			Service:     "bigquery",
// 			Description: "new description in v0.2",
// 			Data:        map[string]interface{}{"data1": float64(12345)},
// 			Labels:      map[string]string{"key1": "value1"},
// 			Version:     "0.5",
// 			UpdatedBy:   r.users[1],
// 		}

// 		ast, err := r.repository.GetByID(r.ctx, astVersioning.ID)
// 		// hard to get the internally generated user id, we exclude the owners from the assertion
// 		astOwners := ast.Owners
// 		ast.Owners = nil
// 		r.NoError(err)
// 		// making updatedby user time empty to make ast comparable
// 		ast.UpdatedBy.CreatedAt = time.Time{}
// 		ast.UpdatedBy.UpdatedAt = time.Time{}
// 		ast.CreatedAt = time.Time{}
// 		ast.UpdatedAt = time.Time{}
// 		r.Equal(expectedLatestVersion, ast)

// 		r.Len(astOwners, 2)
// 	})

// 	r.Run("should return current version of an assets with by version", func() {
// 		expectedLatestVersion := asset.Asset{
// 			ID:          astVersioning.ID,
// 			URN:         assetURN,
// 			Type:        "table",
// 			Service:     "bigquery",
// 			Description: "new description in v0.2",
// 			Data:        map[string]interface{}{"data1": float64(12345)},
// 			Labels:      map[string]string{"key1": "value1"},
// 			Version:     "0.5",
// 			UpdatedBy:   r.users[1],
// 		}

// 		ast, err := r.repository.GetByVersion(r.ctx, astVersioning.ID, "0.5")
// 		// hard to get the internally generated user id, we exclude the owners from the assertion
// 		astOwners := ast.Owners
// 		ast.Owners = nil
// 		r.NoError(err)
// 		// making updatedby user time empty to make ast comparable
// 		ast.UpdatedBy.CreatedAt = time.Time{}
// 		ast.UpdatedBy.UpdatedAt = time.Time{}
// 		ast.CreatedAt = time.Time{}
// 		ast.UpdatedAt = time.Time{}
// 		r.Equal(expectedLatestVersion, ast)

// 		r.Len(astOwners, 2)
// 	})

// 	r.Run("should return a specific version of an asset", func() {
// 		selectedVersion := "0.3"
// 		expectedAsset := asset.Asset{
// 			ID:          astVersioning.ID,
// 			URN:         assetURN,
// 			Type:        "table",
// 			Service:     "bigquery",
// 			Description: "new description in v0.2",
// 			Version:     "0.3",
// 			Changelog: diff.Changelog{
// 				diff.Change{Type: "create", Path: []string{"data", "data1"}, From: interface{}(nil), To: float64(12345)},
// 			},
// 			UpdatedBy: r.users[1],
// 		}
// 		expectedOwners := []user.User{
// 			{
// 				Email:    "user@odpf.io",
// 				Provider: defaultProviderName,
// 			},
// 			{
// 				Email:    "meteor@odpf.io",
// 				Provider: defaultProviderName,
// 			},
// 		}
// 		ast, err := r.repository.GetByVersion(r.ctx, astVersioning.ID, selectedVersion)
// 		// hard to get the internally generated user id, we exclude the owners from the assertion
// 		astOwners := ast.Owners
// 		ast.Owners = nil
// 		r.Assert().NoError(err)
// 		// making updatedby user time empty to make ast comparable
// 		ast.UpdatedBy.CreatedAt = time.Time{}
// 		ast.UpdatedBy.UpdatedAt = time.Time{}
// 		ast.CreatedAt = time.Time{}
// 		ast.UpdatedAt = time.Time{}
// 		r.Assert().Equal(expectedAsset, ast)

// 		for i := 0; i < len(astOwners); i++ {
// 			astOwners[i].ID = ""
// 		}
// 		r.Assert().Equal(expectedOwners, astOwners)
// 	})
// }

// func (r *AssetRepositoryTestSuite) TestUpsert() {
// 	r.Run("on insert", func() {
// 		r.Run("set ID to asset and version to base version", func() {
// 			ast := asset.Asset{
// 				URN:       "urn-u-1",
// 				Type:      "table",
// 				Service:   "bigquery",
// 				Version:   "0.1",
// 				UpdatedBy: r.users[0],
// 			}
// 			id, err := r.repository.Upsert(r.ctx, &ast)
// 			r.NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			ast.ID = id

// 			r.Equal(r.lengthOfString(ast.ID), 36) // uuid

// 			assetInDB, err := r.repository.GetByID(r.ctx, ast.ID)
// 			r.Require().NoError(err)
// 			r.NotEqual(time.Time{}, assetInDB.CreatedAt)
// 			r.NotEqual(time.Time{}, assetInDB.UpdatedAt)
// 			r.assertAsset(&ast, &assetInDB)
// 		})

// 		r.Run("should store owners if any", func() {
// 			ast := asset.Asset{
// 				URN:     "urn-u-3",
// 				Type:    "table",
// 				Service: "bigquery",
// 				Owners: []user.User{
// 					r.users[1],
// 					r.users[2],
// 				},
// 				UpdatedBy: r.users[0],
// 			}

// 			id, err := r.repository.Upsert(r.ctx, &ast)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			ast.ID = id

// 			actual, err := r.repository.GetByID(r.ctx, ast.ID)
// 			r.NoError(err)

// 			r.Len(actual.Owners, len(ast.Owners))
// 			for i, owner := range actual.Owners {
// 				r.Equal(ast.Owners[i].ID, owner.ID)
// 			}
// 		})

// 		r.Run("should create owners as users if they do not exist yet", func() {
// 			ast := asset.Asset{
// 				URN:     "urn-u-3a",
// 				Type:    "table",
// 				Service: "bigquery",
// 				Owners: []user.User{
// 					{Email: "newuser@example.com", Provider: defaultProviderName},
// 				},
// 				UpdatedBy: r.users[0],
// 			}

// 			id, err := r.repository.Upsert(r.ctx, &ast)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)

// 			actual, err := r.repository.GetByID(r.ctx, id)
// 			r.NoError(err)

// 			r.Len(actual.Owners, len(ast.Owners))
// 			for i, owner := range actual.Owners {
// 				r.Equal(ast.Owners[i].Email, owner.Email)
// 				r.Equal(r.lengthOfString(owner.ID), 36) // uuid
// 			}
// 		})
// 	})

// 	r.Run("on update", func() {
// 		r.Run("should not create nor updating the asset if asset is identical", func() {
// 			ast := asset.Asset{
// 				URN:       "urn-u-2",
// 				Type:      "table",
// 				Service:   "bigquery",
// 				UpdatedBy: r.users[0],
// 			}
// 			identicalAsset := ast
// 			identicalAsset.Name = "some-name"

// 			id, err := r.repository.Upsert(r.ctx, &ast)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			ast.ID = id

// 			id, err = r.repository.Upsert(r.ctx, &identicalAsset)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			identicalAsset.ID = id

// 			r.Equal(ast.ID, identicalAsset.ID)
// 		})

// 		r.Run("should delete old owners if it does not exist on new asset", func() {
// 			ast := asset.Asset{
// 				URN:     "urn-u-4",
// 				Type:    "table",
// 				Service: "bigquery",
// 				Owners: []user.User{
// 					r.users[1],
// 					r.users[2],
// 				},
// 				UpdatedBy: r.users[0],
// 			}
// 			newAsset := ast
// 			newAsset.Owners = []user.User{
// 				r.users[2],
// 			}

// 			id, err := r.repository.Upsert(r.ctx, &ast)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			ast.ID = id

// 			id, err = r.repository.Upsert(r.ctx, &newAsset)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			newAsset.ID = id

// 			actual, err := r.repository.GetByID(r.ctx, ast.ID)
// 			r.NoError(err)
// 			r.Len(actual.Owners, len(newAsset.Owners))
// 			for i, owner := range actual.Owners {
// 				r.Equal(newAsset.Owners[i].ID, owner.ID)
// 			}
// 		})

// 		r.Run("should create new owners if it does not exist on old asset", func() {
// 			ast := asset.Asset{
// 				URN:     "urn-u-4",
// 				Type:    "table",
// 				Service: "bigquery",
// 				Owners: []user.User{
// 					r.users[1],
// 				},
// 				UpdatedBy: r.users[0],
// 			}
// 			newAsset := ast
// 			newAsset.Owners = []user.User{
// 				r.users[1],
// 				r.users[2],
// 			}

// 			id, err := r.repository.Upsert(r.ctx, &ast)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			ast.ID = id

// 			id, err = r.repository.Upsert(r.ctx, &newAsset)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			newAsset.ID = id

// 			actual, err := r.repository.GetByID(r.ctx, ast.ID)
// 			r.NoError(err)
// 			r.Len(actual.Owners, len(newAsset.Owners))
// 			for i, owner := range actual.Owners {
// 				r.Equal(newAsset.Owners[i].ID, owner.ID)
// 			}
// 		})

// 		r.Run("should create users from owners if owner emails do not exist yet", func() {
// 			ast := asset.Asset{
// 				URN:     "urn-u-4a",
// 				Type:    "table",
// 				Service: "bigquery",
// 				Owners: []user.User{
// 					r.users[1],
// 				},
// 				UpdatedBy: r.users[0],
// 			}
// 			newAsset := ast
// 			newAsset.Owners = []user.User{
// 				r.users[1],
// 				{Email: "newuser@example.com", Provider: defaultProviderName},
// 			}

// 			id, err := r.repository.Upsert(r.ctx, &ast)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			ast.ID = id

// 			id, err = r.repository.Upsert(r.ctx, &newAsset)
// 			r.Require().NoError(err)
// 			r.Require().Equal(r.lengthOfString(id), 36)
// 			newAsset.ID = id

// 			actual, err := r.repository.GetByID(r.ctx, ast.ID)
// 			r.NoError(err)
// 			r.Len(actual.Owners, len(newAsset.Owners))
// 			for i, owner := range actual.Owners {
// 				r.Equal(newAsset.Owners[i].Email, owner.Email)
// 				r.Equal(r.lengthOfString(owner.ID), 36) // uuid
// 			}
// 		})
// 	})
// }

// func (r *AssetRepositoryTestSuite) TestDelete() {
// 	r.Run("return error from client if any", func() {
// 		err := r.repository.Delete(r.ctx, "invalid-uuid")
// 		r.Error(err)
// 		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
// 	})

// 	r.Run("return NotFoundError if asset does not exist", func() {
// 		uuid := "2aabb450-f986-44e2-a6db-7996861d5004"
// 		err := r.repository.Delete(r.ctx, uuid)
// 		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuid})
// 	})

// 	r.Run("should delete correct asset", func() {
// 		asset1 := asset.Asset{
// 			URN:       "urn-del-1",
// 			Type:      "table",
// 			Service:   "bigquery",
// 			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
// 		}
// 		asset2 := asset.Asset{
// 			URN:       "urn-del-2",
// 			Type:      "topic",
// 			Service:   "kafka",
// 			Version:   asset.BaseVersion,
// 			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
// 		}

// 		var err error
// 		id, err := r.repository.Upsert(r.ctx, &asset1)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		asset1.ID = id

// 		id, err = r.repository.Upsert(r.ctx, &asset2)
// 		r.Require().NoError(err)
// 		r.Require().Equal(r.lengthOfString(id), 36)
// 		asset2.ID = id

// 		err = r.repository.Delete(r.ctx, asset1.ID)
// 		r.NoError(err)

// 		_, err = r.repository.GetByID(r.ctx, asset1.ID)
// 		r.ErrorAs(err, &asset.NotFoundError{AssetID: asset1.ID})

// 		asset2FromDB, err := r.repository.GetByID(r.ctx, asset2.ID)
// 		r.NoError(err)
// 		r.Equal(asset2.ID, asset2FromDB.ID)

// 		// cleanup
// 		err = r.repository.Delete(r.ctx, asset2.ID)
// 		r.NoError(err)
// 	})
// }

// func (r *AssetRepositoryTestSuite) assertAsset(expectedAsset *asset.Asset, actualAsset *asset.Asset) bool {
// 	// sanitize time to make the assets comparable
// 	expectedAsset.CreatedAt = time.Time{}
// 	expectedAsset.UpdatedAt = time.Time{}
// 	expectedAsset.UpdatedBy.CreatedAt = time.Time{}
// 	expectedAsset.UpdatedBy.UpdatedAt = time.Time{}

// 	actualAsset.CreatedAt = time.Time{}
// 	actualAsset.UpdatedAt = time.Time{}
// 	actualAsset.UpdatedBy.CreatedAt = time.Time{}
// 	actualAsset.UpdatedBy.UpdatedAt = time.Time{}

// 	return r.Equal(expectedAsset, actualAsset)
// }

// func (r *AssetRepositoryTestSuite) lengthOfString(s string) int {
// 	return len([]rune(s))
// }

func TestAssetEndToEnd(t *testing.T) {
	suite.Run(t, &AssetEndToEndTestSuite{})
}
