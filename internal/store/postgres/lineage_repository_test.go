package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/salt/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type LineageRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.LineageRepository
}

func (r *LineageRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus()
	r.client, r.pool, r.resource, err = newTestClient(logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()

	r.repository, err = postgres.NewLineageRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *LineageRepositoryTestSuite) TearDownSuite() {
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

func (r *LineageRepositoryTestSuite) TestGetGraph() {
	rootNode := "test-get-graph-root-node"

	// populate root node
	// Graph:
	//
	// table-50																							  metabase-tgg-51
	//  				> optimus-tgg-1 >	rootNode > metabase-tgg-99 >
	// table-51 																							metabase-tgg-52
	//
	err := r.repository.Upsert(r.ctx, rootNode, []string{"optimus-tgg-1"}, []string{"metabase-tgg-99"})
	r.Require().NoError(err)
	// populate upstream's node
	err = r.repository.Upsert(r.ctx, "optimus-tgg-1", []string{"table-50", "table-51"}, nil)
	r.Require().NoError(err)
	// populate downstream's node
	err = r.repository.Upsert(r.ctx, "metabase-tgg-99", nil, []string{"metabase-tgg-51", "metabase-tgg-52"})
	r.Require().NoError(err)

	r.Run("should recursively fetch all graph", func() {
		expected := asset.LineageGraph{
			{Source: "optimus-tgg-1", Target: rootNode},
			{Source: "table-50", Target: "optimus-tgg-1"},
			{Source: "table-51", Target: "optimus-tgg-1"},
			{Source: rootNode, Target: "metabase-tgg-99"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-51"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-52"},
		}

		graph, err := r.repository.GetGraph(r.ctx, rootNode, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(expected, graph)
	})

	r.Run("should fetch based on the level given in config if any", func() {
		expected := asset.LineageGraph{
			{Source: "optimus-tgg-1", Target: rootNode},
			{Source: rootNode, Target: "metabase-tgg-99"},
		}

		graph, err := r.repository.GetGraph(r.ctx, rootNode, asset.LineageQuery{
			Level: 1,
		})
		r.Require().NoError(err)
		r.compareGraphs(expected, graph)
	})

	r.Run("should fetch based on the direction given in config if any", func() {
		expected := asset.LineageGraph{
			{Source: rootNode, Target: "metabase-tgg-99"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-51"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-52"},
		}

		graph, err := r.repository.GetGraph(r.ctx, rootNode, asset.LineageQuery{
			Direction: asset.LineageDirectionDownstream,
		})
		r.Require().NoError(err)
		r.compareGraphs(expected, graph)
	})
}

func (r *LineageRepositoryTestSuite) TestUpsert() {
	r.Run("should insert all as graph if upstreams and downstreams are new", func() {
		nodeURN := "table-1"
		upstreams := []string{"job-1"}
		downstreams := []string{"dashboard-1", "dashboard-2"}
		err := r.repository.Upsert(r.ctx, nodeURN, upstreams, downstreams)
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{
			{Source: "job-1", Target: nodeURN},
			{Source: nodeURN, Target: "dashboard-1"},
			{Source: nodeURN, Target: "dashboard-2"},
		}, graph)
	})

	r.Run("should insert or delete graph when updating existing graph", func() {
		nodeURN := "update-table"

		// create initial
		err := r.repository.Upsert(r.ctx, nodeURN, []string{"job-99"}, []string{"dashboard-99"})
		r.NoError(err)

		// update
		err = r.repository.Upsert(r.ctx, nodeURN, []string{"job-99", "job-100"}, []string{"dashboard-93"})
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, nodeURN, asset.LineageQuery{})
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{
			{Source: "job-99", Target: nodeURN},
			{Source: "job-100", Target: nodeURN},
			{Source: nodeURN, Target: "dashboard-93"},
		}, graph)
	})
}

func (r *LineageRepositoryTestSuite) compareGraphs(expected, actual asset.LineageGraph) {
	expLen := len(expected)
	r.Require().Len(actual, expLen)

	for i := 0; i < expLen; i++ {
		r.Equal(expected[i].Source, actual[i].Source, fmt.Sprintf("different source on index %d", i))
		r.Equal(expected[i].Target, actual[i].Target, fmt.Sprintf("different target on index %d", i))
	}
}

func TestLineageRepository(t *testing.T) {
	suite.Run(t, &LineageRepositoryTestSuite{})
}
