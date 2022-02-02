package lineage

import (
	"fmt"
)

type QueryCfg struct {
	Root     string
	Collapse bool
}

type Graph interface {
	Query(QueryCfg) (AdjacencyMap, error)
}

type InMemoryGraph struct {
	Supergraph AdjacencyMap
}

func (graph InMemoryGraph) Query(cfg QueryCfg) (AdjacencyMap, error) {
	var supergraph = graph.Supergraph

	if cfg.Root != "" {
		var err error
		supergraph, err = graph.buildSubgraphFromRoot(supergraph, cfg.Root)
		if err != nil {
			return supergraph, fmt.Errorf("error building subgraph: %w", err)
		}
	}

	return supergraph, nil
}

func (graph InMemoryGraph) buildSubgraphFromRoot(subgraph AdjacencyMap, root string) (result AdjacencyMap, err error) {
	rootElm, exists := subgraph[root]
	if !exists {
		return result, fmt.Errorf("no such resource %q", root)
	}

	result = make(AdjacencyMap)
	result[rootElm.ID] = rootElm
	graph.addAdjacentsInDir(result, subgraph, rootElm, dataflowDirUpstream)
	graph.addAdjacentsInDir(result, subgraph, rootElm, dataflowDirDownstream)

	return result, nil
}

func (graph InMemoryGraph) addAdjacentsInDir(subgraph AdjacencyMap, superGraph AdjacencyMap, root AdjacencyEntry, dir dataflowDir) {
	queue := []AdjacencyEntry{root}
	for len(queue) > 0 {
		n := len(queue)
		el := queue[n-1]
		queue = queue[:n-1]
		for adjacent := range el.getAdjacents(dir) {
			adjacentEl, exists := superGraph[adjacent]
			if !exists {
				continue
			}
			subgraph[adjacentEl.ID] = adjacentEl
			queue = append(queue, adjacentEl)
		}
	}
}

func NewInMemoryGraph(data AdjacencyMap) InMemoryGraph {
	graph := InMemoryGraph{
		Supergraph: data,
	}

	return graph
}
