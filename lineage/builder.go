package lineage

import (
	"context"
	"fmt"

	"github.com/odpf/columbus/lib/set"
)

// Builder encapsulates the algorithm for building a graph
type Builder interface {
	Build(context.Context, Repository) (Graph, error)
}

var DefaultBuilder = defaultBuilder{}

type defaultBuilder struct{}

// Build the lineage graph for all available types
// The graph construction algorithm works roughly as follows:
// Given,
//   AT = collection of all types
//   ER = collection of assets belonging to a particular type
// We can express the algorithm as below (pseudo-code):
//   graph = {}
//   for type in AT
//     for record in ER(type)
//       create adjaceny entry from record
//       use record's lineage field to build upstreams & downstreams
//       add entry to graph
//   for entry in graph:
//     follow all refs in entry, adding backrefs where needed
// Explanation:
// We process every record once. If it has a proper lineage metadata configured, it will
// be used for obtaining the IDs of related resources. These are called internal or forward-refs. Once all document-refs are resolved
// and all assets have been added to the graph, we follow these links and check if the referred entry has a
// corresponding entry pointing back (called an external or back-ref). If not we add it.
// If any reference'd record is not found in the graph, the algorithm gives up and looks at the next related entry.
// This has the effect of phantom references in graph: A resource may refer another resource in the graph, but that resource
// may not be available in the graph
func (builder defaultBuilder) Build(ctx context.Context, repo Repository) (Graph, error) {
	edges, err := repo.GetEdges(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting edges: %w", err)
	}

	var graph = make(AdjacencyMap)
	builder.populateGraph(ctx, graph, edges)

	return NewCachedGraph(NewInMemoryGraph(graph)), nil
}

// load the assets for a type onto the graph
// if record has a valid "lineage" field
// it will be used for obtaining the values for downstreams and upstreams.
func (builder defaultBuilder) populateGraph(ctx context.Context, graph AdjacencyMap, edges []Edge) {
	nodes := builder.buildNodes(edges)
	for _, node := range nodes {
		builder.setEntry(graph, node)
	}
}

func (builder defaultBuilder) buildNodes(edges []Edge) []Node {
	nodes := []Node{}
	nodesIndex := map[string]int{}
	for _, edge := range edges {
		sourceIndex, exists := nodesIndex[edge.SourceID]
		if !exists {
			newSource := edge.Source
			newSource.ID = edge.SourceID
			nodes = append(nodes, newSource)
			sourceIndex = len(nodes) - 1
			nodesIndex[newSource.ID] = sourceIndex
		}
		targetIndex, exists := nodesIndex[edge.TargetID]
		if !exists {
			newTarget := edge.Target
			newTarget.ID = edge.TargetID
			nodes = append(nodes, newTarget)
			targetIndex = len(nodes) - 1
			nodesIndex[newTarget.ID] = targetIndex
		}

		source := &nodes[sourceIndex]
		target := &nodes[targetIndex]

		source.Downstreams = append(source.Downstreams, *target)
		target.Upstreams = append(target.Upstreams, *source)
	}

	return nodes
}

// add the corresponding assets to graph.
// Uses lineageProcessor to obtain information about upstreams/downstreams
func (builder defaultBuilder) setEntry(graph AdjacencyMap, node Node) {
	entry := AdjacencyEntry{
		ID:          node.ID,
		URN:         node.URN,
		Type:        node.Type.String(),
		Service:     node.Service,
		Downstreams: builder.buildAdjacents(node, dataflowDirDownstream),
		Upstreams:   builder.buildAdjacents(node, dataflowDirUpstream),
	}

	graph[entry.ID] = entry
}

func (builder defaultBuilder) buildAdjacents(node Node, dir dataflowDir) (adjacents set.StringSet) {
	adjacents = set.NewStringSet()

	var nodes []Node
	if dir == dataflowDirUpstream {
		nodes = node.Upstreams
	} else {
		nodes = node.Downstreams
	}

	for _, n := range nodes {
		adjacents.Add(n.ID)
	}

	return
}
