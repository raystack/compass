package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/store/postgres"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

const (
	defaultGetMaxSize = 7
)

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

	logger := logrus.New()
	// logger.SetLevel(logrus.DebugLevel)
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		logger.Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewAssetRepository(r.client, defaultGetMaxSize)
	if err != nil {
		logger.Fatal(err)
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
		}
		err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		assets = append(assets, ast)
	}

	r.Run("should return all assets limited by default size", func() {
		results, err := r.repository.Get(r.ctx, asset.GetConfig{})
		r.Require().NoError(err)
		r.Require().Len(results, defaultGetMaxSize)
		r.Equal(assets[:defaultGetMaxSize], results)
	})

	r.Run("should override default size using GetConfig.Size", func() {
		size := 8
		results, err := r.repository.Get(r.ctx, asset.GetConfig{
			Size: size,
		})
		r.Require().NoError(err)
		r.Require().Len(results, size)
		r.Equal(assets[:size], results)
	})

	r.Run("should fetch assets by offset defined in GetConfig.Offset", func() {
		offset := 2
		results, err := r.repository.Get(r.ctx, asset.GetConfig{
			Offset: offset,
		})
		r.Require().NoError(err)
		r.Equal(assets[offset:defaultGetMaxSize+offset], results)
	})

	r.Run("should filter using type", func() {
		results, err := r.repository.Get(r.ctx, asset.GetConfig{
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
		results, err := r.repository.Get(r.ctx, asset.GetConfig{
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
		err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
	}

	r.Run("should return total assets with filter", func() {
		actual, err := r.repository.GetCount(r.ctx, asset.GetConfig{
			Type:    typ,
			Service: service,
		})
		r.Require().NoError(err)
		r.Equal(total, actual)
	})
}

func (r *AssetRepositoryTestSuite) TestGetByID() {
	r.Run("return error from client if any", func() {
		_, err := r.repository.GetByID(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "error getting asset with ID = \"invalid-uuid\" from DB")
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
		}
		asset2 := asset.Asset{
			URN:     "urn-gbi-2",
			Type:    "topic",
			Service: "kafka",
		}

		var err error
		err = r.repository.Upsert(r.ctx, &asset1)
		r.Require().NoError(err)
		err = r.repository.Upsert(r.ctx, &asset2)
		r.Require().NoError(err)

		result, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		r.Equal(asset2, result)
	})
}

func (r *AssetRepositoryTestSuite) TestUpsert() {
	r.Run("set ID to asset on inserting", func() {
		ast := asset.Asset{
			URN:     "urn-u-1",
			Type:    "table",
			Service: "bigquery",
		}
		err := r.repository.Upsert(r.ctx, &ast)
		r.NoError(err)
		r.Equal(r.lengthOfString(ast.ID), 36) // uuid

		assetInDB, err := r.repository.GetByID(r.ctx, ast.ID)
		r.Require().NoError(err)
		r.Equal(ast, assetInDB)
	})

	r.Run("should not create but update existing asset if urn, type and service match", func() {
		ast := asset.Asset{
			URN:     "urn-u-1",
			Type:    "table",
			Service: "bigquery",
		}
		identicalAsset := ast
		identicalAsset.Name = "some-name"

		err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		err = r.repository.Upsert(r.ctx, &identicalAsset)
		r.Require().NoError(err)

		r.Equal(ast.ID, identicalAsset.ID)
	})
}

func (r *AssetRepositoryTestSuite) TestDelete() {
	r.Run("return error from client if any", func() {
		err := r.repository.Delete(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "error deleting asset with ID = \"invalid-uuid\"")
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
		}

		var err error
		err = r.repository.Upsert(r.ctx, &asset1)
		r.Require().NoError(err)
		err = r.repository.Upsert(r.ctx, &asset2)
		r.Require().NoError(err)

		err = r.repository.Delete(r.ctx, asset1.ID)
		r.NoError(err)

		_, err = r.repository.GetByID(r.ctx, asset1.ID)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: asset1.ID})

		asset2FromDB, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		r.Equal(asset2, asset2FromDB)

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
