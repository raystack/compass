package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/store/postgres"
	"github.com/odpf/columbus/user"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/suite"
)

var defaultAssetUpdaterUserID = uuid.NewString()

type AssetRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.AssetRepository
}

func (r *AssetRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	userRepo, err := postgres.NewUserRepository(r.client, user.Config{IdentityProviderDefaultName: defaultProviderName})
	if err != nil {
		r.T().Fatal(err)
	}
	r.repository, err = postgres.NewAssetRepository(r.client, userRepo, defaultGetMaxSize)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *AssetRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	err := r.client.Close()
	if err != nil {
		r.T().Fatal(err)
	}
	err = purgeDocker(r.pool, r.resource)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *AssetRepositoryTestSuite) TestGet() {
	// populate assets
	total := 12
	assets := []asset.Asset{}
	for i := 0; i < total; i++ {
		var typ asset.Type
		var service string
		if (i % 2) == 0 { // if even
			typ = asset.TypeJob
			service = "postgres"
		} else {
			typ = asset.TypeDashboard
			service = "metabase"
		}

		ast := asset.Asset{
			URN:     fmt.Sprintf("urn-get-%d", i),
			Type:    typ,
			Service: service,
			Version: asset.BaseVersion,
		}
		id, err := r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &ast)
		r.Require().NoError(err)
		r.Require().Equal(r.lengthOfString(id), 36)
		ast.ID = id
		assets = append(assets, ast)
	}

	r.Run("should return all assets limited by default size", func() {
		results, err := r.repository.Get(r.ctx, asset.Config{})
		r.Require().NoError(err)
		r.Require().Len(results, defaultGetMaxSize)
		// making time empty to make ast comparable
		for i := 0; i < len(results); i++ {
			results[i].CreatedAt = time.Time{}
			results[i].UpdatedAt = time.Time{}
		}
		r.Equal(assets[:defaultGetMaxSize], results)
	})

	r.Run("should override default size using GetConfig.Size", func() {
		size := 8
		results, err := r.repository.Get(r.ctx, asset.Config{
			Size: size,
		})
		r.Require().NoError(err)
		r.Require().Len(results, size)
		// making time empty to make ast comparable
		for i := 0; i < len(results); i++ {
			results[i].CreatedAt = time.Time{}
			results[i].UpdatedAt = time.Time{}
		}
		r.Equal(assets[:size], results)
	})

	r.Run("should fetch assets by offset defined in GetConfig.Offset", func() {
		offset := 2
		results, err := r.repository.Get(r.ctx, asset.Config{
			Offset: offset,
		})
		r.Require().NoError(err)
		// making time empty to make ast comparable
		for i := 0; i < len(results); i++ {
			results[i].CreatedAt = time.Time{}
			results[i].UpdatedAt = time.Time{}
		}
		r.Equal(assets[offset:defaultGetMaxSize+offset], results)
	})

	r.Run("should filter using type", func() {
		results, err := r.repository.Get(r.ctx, asset.Config{
			Type: asset.TypeDashboard,
			Size: total,
		})
		r.Require().NoError(err)
		r.Require().Len(results, total/2)
		for _, ast := range results {
			r.Equal(asset.TypeDashboard, ast.Type)
		}
	})

	r.Run("should filter using service", func() {
		results, err := r.repository.Get(r.ctx, asset.Config{
			Service: "postgres",
			Size:    total,
		})
		r.Require().NoError(err)
		r.Require().Len(results, total/2)
		for _, ast := range results {
			r.Equal("postgres", ast.Service)
		}
	})
}

func (r *AssetRepositoryTestSuite) TestGetCount() {
	// populate assets
	total := 12
	typ := asset.TypeJob
	service := "service-getcount"
	for i := 0; i < total; i++ {
		ast := asset.Asset{
			URN:     fmt.Sprintf("urn-getcount-%d", i),
			Type:    typ,
			Service: service,
		}
		id, err := r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &ast)
		r.Require().NoError(err)
		r.Require().Equal(r.lengthOfString(id), 36)
		ast.ID = id
	}

	r.Run("should return total assets with filter", func() {
		actual, err := r.repository.GetCount(r.ctx, asset.Config{
			Type:    typ,
			Service: service,
		})
		r.Require().NoError(err)
		r.Equal(total, actual)
	})
}

func (r *AssetRepositoryTestSuite) TestGetByID() {

	// create users
	user1 := user.User{Email: "johndoe@example.com", Provider: defaultProviderName}
	user2 := user.User{Email: "janedoe@example.com", Provider: defaultProviderName}
	userRepo, err := postgres.NewUserRepository(r.client, user.Config{IdentityProviderDefaultName: defaultProviderName})
	r.Require().NoError(err)
	user1.ID, err = userRepo.Create(r.ctx, &user1)
	r.Require().NoError(err)
	user2.ID, err = userRepo.Create(r.ctx, &user2)
	r.Require().NoError(err)
	// clean up
	defer func() {
		err = r.client.ExecQueries(r.ctx, []string{"DELETE from users"})
		r.Require().NoError(err)
	}()

	r.Run("return error from client if asset not an uuid", func() {
		_, err := r.repository.GetByID(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
	})

	r.Run("return NotFoundError if asset does not exist", func() {
		uuid := "2aabb450-f986-44e2-a6db-7996861d5004"
		_, err := r.repository.GetByID(r.ctx, uuid)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuid})
	})

	r.Run("return correct asset from db", func() {
		asset1 := asset.Asset{
			URN:     "urn-gbi-1",
			Type:    "table",
			Service: "bigquery",
			Version: asset.BaseVersion,
		}
		asset2 := asset.Asset{
			URN:     "urn-gbi-2",
			Type:    "topic",
			Service: "kafka",
			Version: asset.BaseVersion,
		}

		var err error
		id, err := r.repository.Upsert(r.ctx, user1.ID, &asset1)
		r.Require().NoError(err)
		r.Require().Equal(r.lengthOfString(id), 36)
		asset1.ID = id

		id, err = r.repository.Upsert(r.ctx, user1.ID, &asset2)
		r.Require().NoError(err)
		r.Require().Equal(r.lengthOfString(id), 36)
		asset2.ID = id

		result, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		asset2.UpdatedBy = user1
		// making time empty to make ast comparable
		result.UpdatedBy.CreatedAt = time.Time{}
		result.UpdatedBy.UpdatedAt = time.Time{}
		result.CreatedAt = time.Time{}
		result.UpdatedAt = time.Time{}
		r.Equal(asset2, result)
	})

	r.Run("return owners if any", func() {

		ast := asset.Asset{
			URN:     "urn-gbi-3",
			Type:    "table",
			Service: "bigquery",
			Owners: []user.User{
				user1,
				user2,
			},
		}

		id, err := r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &ast)
		r.Require().NoError(err)
		r.Require().Equal(r.lengthOfString(id), 36)
		ast.ID = id

		result, err := r.repository.GetByID(r.ctx, ast.ID)
		r.NoError(err)
		r.Len(result.Owners, len(ast.Owners))
		for i, owner := range result.Owners {
			r.Equal(ast.Owners[i].ID, owner.ID)
		}
	})
}

func (r *AssetRepositoryTestSuite) TestVersions() {
	// create users
	user1 := user.User{Email: "user@odpf.io", Provider: defaultProviderName}
	userRepo, err := postgres.NewUserRepository(r.client, user.Config{IdentityProviderDefaultName: defaultProviderName})
	r.Require().NoError(err)
	user1.ID, err = userRepo.Create(r.ctx, &user1)
	// clean up
	defer func() {
		err = r.client.ExecQueries(r.ctx, []string{"DELETE from users"})
		r.Require().NoError(err)
	}()

	assetURN := "urn-u-2-version"
	// v0.1
	astVersioning := asset.Asset{
		URN:     assetURN,
		Type:    "table",
		Service: "bigquery",
	}

	id, err := r.repository.Upsert(r.ctx, user1.ID, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(r.lengthOfString(id), 36)
	astVersioning.ID = id

	// v0.2
	astVersioning.Description = "new description in v0.2"
	id, err = r.repository.Upsert(r.ctx, user1.ID, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

	// v0.3
	astVersioning.Owners = []user.User{
		{
			Email: "user@odpf.io",
		},
		{
			Email:    "meteor@odpf.io",
			Provider: "meteor",
		},
	}
	id, err = r.repository.Upsert(r.ctx, user1.ID, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

	// v0.4
	astVersioning.Data = map[string]interface{}{
		"data1": float64(12345),
	}
	id, err = r.repository.Upsert(r.ctx, user1.ID, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

	// v0.5
	astVersioning.Labels = map[string]string{
		"key1": "value1",
	}

	id, err = r.repository.Upsert(r.ctx, user1.ID, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

	r.Run("should return 3 last versions of an assets if there are exist", func() {

		expectedAssetVersions := []asset.AssetVersion{
			{
				ID:      astVersioning.ID,
				URN:     assetURN,
				Type:    "table",
				Service: "bigquery",
				Version: "0.4",
				Changelog: diff.Changelog{
					diff.Change{Type: "create", Path: []string{"labels", "key1"}, From: interface{}(nil), To: "value1"},
				},
				UpdatedBy: user1,
			},
			{
				ID:      astVersioning.ID,
				URN:     assetURN,
				Type:    "table",
				Service: "bigquery",
				Version: "0.3",
				Changelog: diff.Changelog{
					diff.Change{Type: "create", Path: []string{"data", "data1"}, From: interface{}(nil), To: float64(12345)},
				},
				UpdatedBy: user1,
			},
			{
				ID:      astVersioning.ID,
				URN:     assetURN,
				Type:    "table",
				Service: "bigquery",
				Version: "0.2",
				Changelog: diff.Changelog{
					diff.Change{Type: "create", Path: []string{"owners", "0", "email"}, From: interface{}(nil), To: "user@odpf.io"},
					diff.Change{Type: "create", Path: []string{"owners", "1", "email"}, From: interface{}(nil), To: "meteor@odpf.io"},
				},
				UpdatedBy: user1,
			},
		}

		assetVersions, err := r.repository.GetPrevVersions(r.ctx, asset.Config{Size: 3}, astVersioning.ID)
		r.NoError(err)
		// making updatedby user time empty to make ast comparable
		for i := 0; i < len(assetVersions); i++ {
			assetVersions[i].UpdatedBy.CreatedAt = time.Time{}
			assetVersions[i].UpdatedBy.UpdatedAt = time.Time{}
			assetVersions[i].CreatedAt = time.Time{}
			assetVersions[i].UpdatedAt = time.Time{}
		}
		r.Equal(expectedAssetVersions, assetVersions)
	})

	r.Run("should return current version of an assets", func() {
		expectedLatestVersion := asset.Asset{
			ID:          astVersioning.ID,
			URN:         assetURN,
			Type:        "table",
			Service:     "bigquery",
			Description: "new description in v0.2",
			Data:        map[string]interface{}{"data1": float64(12345)},
			Labels:      map[string]string{"key1": "value1"},
			Version:     "0.5",
			UpdatedBy:   user1,
		}

		ast, err := r.repository.GetByID(r.ctx, astVersioning.ID)
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners := ast.Owners
		ast.Owners = nil
		r.NoError(err)
		// making updatedby user time empty to make ast comparable
		ast.UpdatedBy.CreatedAt = time.Time{}
		ast.UpdatedBy.UpdatedAt = time.Time{}
		ast.CreatedAt = time.Time{}
		ast.UpdatedAt = time.Time{}
		r.Equal(expectedLatestVersion, ast)

		r.Len(astOwners, 2)
	})

	r.Run("should return current version of an assets with by version", func() {
		expectedLatestVersion := asset.Asset{
			ID:          astVersioning.ID,
			URN:         assetURN,
			Type:        "table",
			Service:     "bigquery",
			Description: "new description in v0.2",
			Data:        map[string]interface{}{"data1": float64(12345)},
			Labels:      map[string]string{"key1": "value1"},
			Version:     "0.5",
			UpdatedBy:   user1,
		}

		ast, err := r.repository.GetByVersion(r.ctx, astVersioning.ID, "0.5")
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners := ast.Owners
		ast.Owners = nil
		r.NoError(err)
		// making updatedby user time empty to make ast comparable
		ast.UpdatedBy.CreatedAt = time.Time{}
		ast.UpdatedBy.UpdatedAt = time.Time{}
		ast.CreatedAt = time.Time{}
		ast.UpdatedAt = time.Time{}
		r.Equal(expectedLatestVersion, ast)

		r.Len(astOwners, 2)
	})

	r.Run("should return a specific version of an asset", func() {
		selectedVersion := "0.3"
		expectedAsset := asset.Asset{
			ID:          astVersioning.ID,
			URN:         assetURN,
			Type:        "table",
			Service:     "bigquery",
			Description: "new description in v0.2",
			Version:     "0.3",
			Changelog: diff.Changelog{
				diff.Change{Type: "create", Path: []string{"data", "data1"}, From: interface{}(nil), To: float64(12345)},
			},
			UpdatedBy: user1,
		}
		expectedOwners := []user.User{
			{
				Email:    "user@odpf.io",
				Provider: "shield",
			},
			{
				Email:    "meteor@odpf.io",
				Provider: "meteor",
			},
		}
		ast, err := r.repository.GetByVersion(r.ctx, astVersioning.ID, selectedVersion)
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners := ast.Owners
		ast.Owners = nil
		r.Assert().NoError(err)
		// making updatedby user time empty to make ast comparable
		ast.UpdatedBy.CreatedAt = time.Time{}
		ast.UpdatedBy.UpdatedAt = time.Time{}
		ast.CreatedAt = time.Time{}
		ast.UpdatedAt = time.Time{}
		r.Assert().Equal(expectedAsset, ast)

		for i := 0; i < len(astOwners); i++ {
			astOwners[i].ID = ""
		}
		r.Assert().Equal(expectedOwners, astOwners)
	})
}

func (r *AssetRepositoryTestSuite) TestUpsert() {
	// create users
	user1 := user.User{Email: "johndoe@example.com", Provider: defaultProviderName}
	user2 := user.User{Email: "janedoe@example.com", Provider: defaultProviderName}
	userRepo, err := postgres.NewUserRepository(r.client, user.Config{IdentityProviderDefaultName: defaultProviderName})
	r.Require().NoError(err)
	user1.ID, err = userRepo.Create(r.ctx, &user1)
	r.Require().NoError(err)
	user2.ID, err = userRepo.Create(r.ctx, &user2)
	r.Require().NoError(err)
	// clean up
	defer func() {
		err = r.client.ExecQueries(r.ctx, []string{"DELETE from users"})
		r.Require().NoError(err)
	}()

	r.Run("on insert", func() {
		r.Run("set ID to asset and version to base version", func() {
			ast := asset.Asset{
				URN:     "urn-u-1",
				Type:    "table",
				Service: "bigquery",
				Version: "0.1",
			}
			id, err := r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &ast)
			r.NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			ast.ID = id

			r.Equal(r.lengthOfString(ast.ID), 36) // uuid

			assetInDB, err := r.repository.GetByID(r.ctx, ast.ID)
			r.Require().NoError(err)
			r.NotEqual(time.Time{}, assetInDB.CreatedAt)
			r.NotEqual(time.Time{}, assetInDB.UpdatedAt)
			assetInDB.CreatedAt = time.Time{}
			assetInDB.UpdatedAt = time.Time{}
			r.Equal(ast, assetInDB)
		})

		r.Run("should store owners if any", func() {
			ast := asset.Asset{
				URN:     "urn-u-3",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					user1,
					user2,
				},
			}

			id, err := r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &ast)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			ast.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)

			r.Len(actual.Owners, len(ast.Owners))
			for i, owner := range actual.Owners {
				r.Equal(ast.Owners[i].ID, owner.ID)
			}
		})

		r.Run("should create owners as users if they do not exist yet", func() {
			ast := asset.Asset{
				URN:     "urn-u-3a",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					{Email: "newuser@example.com", Provider: defaultProviderName},
				},
			}

			id, err := r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &ast)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)

			actual, err := r.repository.GetByID(r.ctx, id)
			r.NoError(err)

			r.Len(actual.Owners, len(ast.Owners))
			for i, owner := range actual.Owners {
				r.Equal(ast.Owners[i].Email, owner.Email)
				r.Equal(r.lengthOfString(owner.ID), 36) // uuid
			}
		})
	})

	r.Run("on update", func() {
		r.Run("should not create nor updating the asset if asset is identical", func() {
			ast := asset.Asset{
				URN:     "urn-u-2",
				Type:    "table",
				Service: "bigquery",
			}
			identicalAsset := ast
			identicalAsset.Name = "some-name"

			id, err := r.repository.Upsert(r.ctx, user1.ID, &ast)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, user1.ID, &identicalAsset)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			identicalAsset.ID = id

			r.Equal(ast.ID, identicalAsset.ID)
		})

		r.Run("should delete old owners if it does not exist on new asset", func() {
			ast := asset.Asset{
				URN:     "urn-u-4",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					user1,
					user2,
				},
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				user2,
			}

			id, err := r.repository.Upsert(r.ctx, user1.ID, &ast)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, user1.ID, &newAsset)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			newAsset.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)
			r.Len(actual.Owners, len(newAsset.Owners))
			for i, owner := range actual.Owners {
				r.Equal(newAsset.Owners[i].ID, owner.ID)
			}
		})

		r.Run("should create new owners if it does not exist on old asset", func() {
			ast := asset.Asset{
				URN:     "urn-u-4",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					user1,
				},
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				user1,
				user2,
			}

			id, err := r.repository.Upsert(r.ctx, user1.ID, &ast)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, user1.ID, &newAsset)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			newAsset.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)
			r.Len(actual.Owners, len(newAsset.Owners))
			for i, owner := range actual.Owners {
				r.Equal(newAsset.Owners[i].ID, owner.ID)
			}
		})

		r.Run("should create users from owners if owner emails do not exist yet", func() {
			ast := asset.Asset{
				URN:     "urn-u-4a",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					user1,
				},
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				user1,
				{Email: "newuser@example.com", Provider: defaultProviderName},
			}

			id, err := r.repository.Upsert(r.ctx, user1.ID, &ast)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, user1.ID, &newAsset)
			r.Require().NoError(err)
			r.Require().Equal(r.lengthOfString(id), 36)
			newAsset.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)
			r.Len(actual.Owners, len(newAsset.Owners))
			for i, owner := range actual.Owners {
				r.Equal(newAsset.Owners[i].Email, owner.Email)
				r.Equal(r.lengthOfString(owner.ID), 36) // uuid
			}
		})
	})
}

func (r *AssetRepositoryTestSuite) TestDelete() {
	r.Run("return error from client if any", func() {
		err := r.repository.Delete(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
	})

	r.Run("return NotFoundError if asset does not exist", func() {
		uuid := "2aabb450-f986-44e2-a6db-7996861d5004"
		err := r.repository.Delete(r.ctx, uuid)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuid})
	})

	r.Run("should delete correct asset", func() {
		asset1 := asset.Asset{
			URN:     "urn-del-1",
			Type:    "table",
			Service: "bigquery",
		}
		asset2 := asset.Asset{
			URN:     "urn-del-2",
			Type:    "topic",
			Service: "kafka",
			Version: asset.BaseVersion,
		}

		var err error
		id, err := r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &asset1)
		r.Require().NoError(err)
		r.Require().Equal(r.lengthOfString(id), 36)
		asset1.ID = id

		id, err = r.repository.Upsert(r.ctx, defaultAssetUpdaterUserID, &asset2)
		r.Require().NoError(err)
		r.Require().Equal(r.lengthOfString(id), 36)
		asset2.ID = id

		err = r.repository.Delete(r.ctx, asset1.ID)
		r.NoError(err)

		_, err = r.repository.GetByID(r.ctx, asset1.ID)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: asset1.ID})

		asset2FromDB, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		r.Equal(asset2.ID, asset2FromDB.ID)

		// cleanup
		err = r.repository.Delete(r.ctx, asset2.ID)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) lengthOfString(s string) int {
	return len([]rune(s))
}

func TestAssetRepository(t *testing.T) {
	suite.Run(t, &AssetRepositoryTestSuite{})
}
