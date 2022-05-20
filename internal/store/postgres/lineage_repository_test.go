package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/internal/store/postgres"
	"github.com/odpf/salt/log"
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
	rootNode := r.bigquery("test-get-graph-root-node")
	// populate root node
	err := r.repository.Upsert(r.ctx,
		rootNode,
		[]asset.LineageNode{
			r.optimus("optimus-tgg-1"),
		},
		[]asset.LineageNode{
			r.metabase("metabase-tgg-99"),
		})
	r.Require().NoError(err)
	// populate upstream's node
	err = r.repository.Upsert(r.ctx,
		r.optimus("optimus-tgg-1"),
		[]asset.LineageNode{
			r.bigquery("table-50"),
			r.bigquery("table-51"),
		},
		[]asset.LineageNode{},
	)
	r.Require().NoError(err)
	// populate downstream's node
	err = r.repository.Upsert(r.ctx,
		r.metabase("metabase-tgg-99"),
		[]asset.LineageNode{},
		[]asset.LineageNode{
			r.metabase("metabase-tgg-51"),
			r.metabase("metabase-tgg-52"),
		},
	)
	r.Require().NoError(err)

	r.Run("should recursively fetch all graph", func() {
		expected := asset.LineageGraph{
			{Source: "optimus-tgg-1", Target: rootNode.URN},
			{Source: "table-50", Target: "optimus-tgg-1"},
			{Source: "table-51", Target: "optimus-tgg-1"},
			{Source: rootNode.URN, Target: "metabase-tgg-99"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-51"},
			{Source: "metabase-tgg-99", Target: "metabase-tgg-52"},
		}

		graph, err := r.repository.GetGraph(r.ctx, rootNode)
		r.Require().NoError(err)
		r.compareGraphs(expected, graph)
	})
}

func (r *LineageRepositoryTestSuite) TestUpsert() {
	r.Run("should insert all as graph if upstreams and downstreams are new", func() {
		nodeURN := "table-1"
		node := asset.LineageNode{
			URN:     nodeURN,
			Type:    "table",
			Service: "bigquery",
		}
		upstreams := []asset.LineageNode{
			{URN: "job-1", Type: asset.TypeJob, Service: "optimus"},
		}
		downstreams := []asset.LineageNode{
			{URN: "dashboard-1", Type: asset.TypeDashboard, Service: "metabase"},
			{URN: "dashboard-2", Type: asset.TypeDashboard, Service: "optimus"},
		}
		err := r.repository.Upsert(r.ctx, node, upstreams, downstreams)
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, node)
		r.Require().NoError(err)
		r.compareGraphs(asset.LineageGraph{
			{Source: "job-1", Target: nodeURN},
			{Source: nodeURN, Target: "dashboard-1"},
			{Source: nodeURN, Target: "dashboard-2"},
		}, graph)
	})

	r.Run("should insert or delete graph when updating existing graph", func() {
		nodeURN := "update-table"
		node := asset.LineageNode{
			URN:     nodeURN,
			Type:    "table",
			Service: "bigquery",
		}

		// create initial
		err := r.repository.Upsert(r.ctx, node,
			[]asset.LineageNode{
				{URN: "job-99", Type: asset.TypeJob, Service: "optimus"},
			},
			[]asset.LineageNode{
				{URN: "dashboard-99", Type: asset.TypeDashboard, Service: "metabase"},
			})
		r.NoError(err)

		// update
		err = r.repository.Upsert(r.ctx, node,
			[]asset.LineageNode{
				{URN: "job-99", Type: asset.TypeJob, Service: "optimus"},
				{URN: "job-100", Type: asset.TypeJob, Service: "optimus"},
			},
			[]asset.LineageNode{
				{URN: "dashboard-93", Type: asset.TypeDashboard, Service: "metabase"},
			})
		r.NoError(err)

		graph, err := r.repository.GetGraph(r.ctx, node)
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

func (r *LineageRepositoryTestSuite) bigquery(urn string) asset.LineageNode {
	return asset.LineageNode{
		URN:     urn,
		Type:    asset.TypeTable,
		Service: "bigquery",
	}
}

func (r *LineageRepositoryTestSuite) optimus(urn string) asset.LineageNode {
	return asset.LineageNode{
		URN:     urn,
		Type:    asset.TypeJob,
		Service: "optimus",
	}
}

func (r *LineageRepositoryTestSuite) metabase(urn string) asset.LineageNode {
	return asset.LineageNode{
		URN:     urn,
		Type:    asset.TypeDashboard,
		Service: "metabase",
	}
}

func TestLineageRepository(t *testing.T) {
	suite.Run(t, &LineageRepositoryTestSuite{})
}
