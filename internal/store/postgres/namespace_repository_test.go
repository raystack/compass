package postgres_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/odpf/compass/core/namespace"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
	"testing"
)

type NamespaceRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.NamespaceRepository
	ns         *namespace.Namespace
}

func (r *NamespaceRepositoryTestSuite) SetupSuite() {
	var err error
	r.ns = &namespace.Namespace{
		ID:       uuid.New(),
		Name:     "umbrella",
		State:    namespace.SharedState,
		Metadata: nil,
	}

	logger := log.NewLogrus()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository = postgres.NewNamespaceRepository(r.client)
}

func (r *NamespaceRepositoryTestSuite) TearDownSuite() {
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

func (r *NamespaceRepositoryTestSuite) SetupTest() {
}

func (r *NamespaceRepositoryTestSuite) TearDownTest() {
}

func (r *NamespaceRepositoryTestSuite) cleanup() error {
	queries := []string{
		"TRUNCATE TABLE namespaces CASCADE",
	}
	return r.client.ExecQueries(r.ctx, queries)
}

func (r *NamespaceRepositoryTestSuite) TestCreate() {
	r.Run("should fail to create a new namespace if name is empty", func() {
		_ = r.cleanup()
		_, err := r.repository.Create(r.ctx, &namespace.Namespace{
			State:    namespace.SharedState,
			Metadata: nil,
		})
		r.Error(err)
	})

	r.Run("should create a new namespace successfully", func() {
		_ = r.cleanup()
		ns := &namespace.Namespace{
			ID:    uuid.New(),
			Name:  "umbrella",
			State: namespace.SharedState,
			Metadata: map[string]interface{}{
				"hello": "world",
			},
		}
		id, err := r.repository.Create(r.ctx, ns)
		r.NoError(err)
		r.Equal(ns.ID.String(), id)
		fetched, err := r.repository.GetByName(r.ctx, ns.Name)
		r.NoError(err)
		r.EqualValues(ns, fetched)
	})

	r.Run("should fail to insert duplicate namespace", func() {
		_ = r.cleanup()
		ns := &namespace.Namespace{
			ID:    uuid.New(),
			Name:  "umbrella",
			State: namespace.SharedState,
			Metadata: map[string]interface{}{
				"hello": "world",
			},
		}
		_, err := r.repository.Create(r.ctx, ns)
		r.NoError(err)
		_, err = r.repository.Create(r.ctx, ns)
		r.Error(err)
	})
}

func (r *NamespaceRepositoryTestSuite) TestList() {
	r.Run("should return list of namespaces", func() {
		_ = r.cleanup()
		ns1 := &namespace.Namespace{
			ID:    uuid.New(),
			Name:  "umbrella",
			State: namespace.SharedState,
			Metadata: map[string]interface{}{
				"hello": "world",
			},
		}
		ns2 := &namespace.Namespace{
			ID:    uuid.New(),
			Name:  "umbrella-2",
			State: namespace.SharedState,
			Metadata: map[string]interface{}{
				"hello": "world-2",
			},
		}
		_, err := r.repository.Create(r.ctx, ns1)
		r.NoError(err)
		_, err = r.repository.Create(r.ctx, ns2)
		r.NoError(err)

		nss, err := r.repository.List(r.ctx)
		r.NoError(err)
		r.Contains(nss, ns1)
		r.Contains(nss, ns2)
	})
}

func (r *NamespaceRepositoryTestSuite) TestGetByID() {
	r.Run("should fetch namespace by id successfully", func() {
		_ = r.cleanup()
		ns := &namespace.Namespace{
			ID:    uuid.New(),
			Name:  "umbrella",
			State: namespace.SharedState,
			Metadata: map[string]interface{}{
				"hello": "world",
			},
		}
		id, err := r.repository.Create(r.ctx, ns)
		r.NoError(err)
		r.Equal(ns.ID.String(), id)
		fetched, err := r.repository.GetByID(r.ctx, ns.ID)
		r.NoError(err)
		r.EqualValues(ns, fetched)
	})
}

func (r *NamespaceRepositoryTestSuite) TestUpdate() {
	r.Run("should update an existing namespace successfully", func() {
		_ = r.cleanup()
		ns := &namespace.Namespace{
			ID:    uuid.New(),
			Name:  "umbrella",
			State: namespace.SharedState,
			Metadata: map[string]interface{}{
				"hello": "world",
			},
		}
		id, err := r.repository.Create(r.ctx, ns)
		r.NoError(err)
		r.Equal(ns.ID.String(), id)

		nsUpdated := &namespace.Namespace{
			ID:    ns.ID,
			Name:  ns.Name,
			State: namespace.SharedState,
			Metadata: map[string]interface{}{
				"hello": "xworldx",
				"bye":   "world",
			},
		}
		err = r.repository.Update(r.ctx, nsUpdated)
		r.NoError(err)

		fetched, err := r.repository.GetByID(r.ctx, nsUpdated.ID)
		r.NoError(err)
		r.EqualValues(nsUpdated, fetched)
	})
}

func TestNamespaceRepository(t *testing.T) {
	suite.Run(t, &NamespaceRepositoryTestSuite{})
}
