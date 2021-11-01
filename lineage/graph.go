package lineage

import (
	"fmt"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/models"
	"github.com/pkg/errors"
)

type QueryCfg struct {
	TypeWhitelist []string
	Collapse      bool
	Root          string
}

type Graph interface {
	Query(QueryCfg) (AdjacencyMap, error)
}

type InMemoryGraph struct {
	Supergraph AdjacencyMap
	typeIdx    map[string]set.StringSet
}

func (graph *InMemoryGraph) init() {
	for id, entry := range graph.Supergraph {
		values, exists := graph.typeIdx[entry.Type]
		if !exists {
			values = make(set.StringSet)
			graph.typeIdx[entry.Type] = values
		}
		values.Add(id)
	}
	return
}

func (graph InMemoryGraph) Query(cfg QueryCfg) (AdjacencyMap, error) {
	var supergraph = graph.Supergraph

	if len(cfg.TypeWhitelist) != 0 {
		var tempSupergraph = make(AdjacencyMap)
		for _, typ := range cfg.TypeWhitelist {
			if _, typExists := graph.typeIdx[typ]; !typExists {
				return nil, models.ErrNoSuchType{TypeName: typ}
			}
			for entry := range graph.typeIdx[typ] {
				tempSupergraph[entry] = supergraph[entry]
			}
		}
		supergraph = tempSupergraph
	}

	if cfg.Collapse {
		graph.collapse(supergraph, set.NewStringSet(cfg.TypeWhitelist...))
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
	graph.addAdjacentsInDir(result, graph.Supergraph, rootElm, dataflowDirUpstream)
	graph.addAdjacentsInDir(result, graph.Supergraph, rootElm, dataflowDirDownstream)

	return result, nil
}

func (graph InMemoryGraph) collapse(subgraph AdjacencyMap, typeWhitelist set.StringSet) {
	for _, entry := range subgraph {
		entry.Upstreams = graph.collapseInDir(entry, dataflowDirUpstream, typeWhitelist)
		entry.Downstreams = graph.collapseInDir(entry, dataflowDirDownstream, typeWhitelist)
		subgraph[entry.ID()] = entry
	}
	return
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
			if types[adjEntry.Type] {
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
	return
}

// remove refs for each entry that don't exist in the built graph
// during handler.addAdjacentsInDir we follow the declarations to add
// upstreams/downstreams to the graph. Once that is done, we've constructed the lineage
// graph for the requested resource. During this step, we remove outgoing refs to any
// resource that doesn't belong in the graph
func (graph InMemoryGraph) pruneRefs(subgraph AdjacencyMap) AdjacencyMap {
	pruned := make(AdjacencyMap)
	for _, entry := range subgraph {
		entry.Upstreams = graph.filterRefs(entry.Upstreams, subgraph)
		entry.Downstreams = graph.filterRefs(entry.Downstreams, subgraph)
		pruned[entry.ID()] = entry
	}
	return pruned
}

func (graph InMemoryGraph) filterRefs(refs set.StringSet, subgraph AdjacencyMap) set.StringSet {
	rv := set.NewStringSet()
	for ref := range refs {
		if _, exists := subgraph[ref]; exists {
			rv.Add(ref)
		}
	}
	return rv
}

func NewInMemoryGraph(data AdjacencyMap) InMemoryGraph {
	graph := InMemoryGraph{
		Supergraph: data,
		typeIdx:    make(map[string]set.StringSet),
	}
	graph.init()
	return graph
}
