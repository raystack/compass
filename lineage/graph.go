package lineage

import (
	"fmt"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/record"
	"github.com/pkg/errors"
)

type QueryCfg struct {
	TypeWhitelist []record.Type
	Collapse      bool
	Root          string
}

type Graph interface {
	Query(QueryCfg) (AdjacencyMap, error)
}

type InMemoryGraph struct {
	Supergraph AdjacencyMap
	typeIdx    map[record.Type]set.StringSet
}

func (graph *InMemoryGraph) init() {
	for _, typ := range record.TypeList {
		graph.typeIdx[typ] = make(set.StringSet)
	}

	for id, entry := range graph.Supergraph {
		graph.typeIdx[entry.Type].Add(id)
	}
}

func (graph InMemoryGraph) Query(cfg QueryCfg) (AdjacencyMap, error) {
	var supergraph = graph.Supergraph

	if len(cfg.TypeWhitelist) != 0 {
		var tempSupergraph = make(AdjacencyMap)
		for _, typ := range cfg.TypeWhitelist {
			for entry := range graph.typeIdx[typ] {
				tempSupergraph[entry] = supergraph[entry]
			}
		}
		supergraph = tempSupergraph
	}

	if cfg.Collapse {
		stringList := []string{}
		for _, t := range cfg.TypeWhitelist {
			stringList = append(stringList, string(t))
		}
		graph.collapse(supergraph, set.NewStringSet(stringList...))
	}

	if cfg.Root != "" {
		var err error
		supergraph, err = graph.buildSubgraphFromRoot(supergraph, cfg.Root)
		if err != nil {
			return supergraph, errors.Wrap(err, "error building subgraph")
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
	result[rootElm.ID()] = rootElm
	graph.addAdjacentsInDir(result, subgraph, rootElm, dataflowDirUpstream)
	graph.addAdjacentsInDir(result, subgraph, rootElm, dataflowDirDownstream)

	return result, nil
}

func (graph InMemoryGraph) collapse(subgraph AdjacencyMap, typeWhitelist set.StringSet) {
	for _, entry := range subgraph {
		entry.Upstreams = graph.collapseInDir(entry, dataflowDirUpstream, typeWhitelist)
		entry.Downstreams = graph.collapseInDir(entry, dataflowDirDownstream, typeWhitelist)
		subgraph[entry.ID()] = entry
	}
}

func (graph InMemoryGraph) collapseInDir(root AdjacencyEntry, dir dataflowDir, types set.StringSet) set.StringSet {
	var (
		queue     = []AdjacencyEntry{root}
		collapsed = set.NewStringSet()
	)
	for len(queue) > 0 {
		n := len(queue)
		entry := queue[n-1]
		queue = queue[:n-1]
		adjacents := entry.getAdjacents(dir)
		for adjacent := range adjacents {
			adjEntry, exists := graph.Supergraph[adjacent]
			if !exists {
				continue
			}
			if types[string(adjEntry.Type)] {
				collapsed.Add(adjEntry.ID())
				continue
			}
			queue = append(queue, adjEntry)
		}
	}
	return collapsed
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
			subgraph[adjacentEl.ID()] = adjacentEl
			queue = append(queue, adjacentEl)
		}
	}
}

func NewInMemoryGraph(data AdjacencyMap) InMemoryGraph {
	graph := InMemoryGraph{
		Supergraph: data,
		typeIdx:    make(map[record.Type]set.StringSet),
	}
	graph.init()

	return graph
}
