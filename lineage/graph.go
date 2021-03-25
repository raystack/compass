package lineage

import (
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/models"
)

type QueryCfg struct {
	TypeWhitelist []string
	Collapse      bool
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

	if len(cfg.TypeWhitelist) == 0 {
		return graph.Supergraph, nil
	}

	var response = make(AdjacencyMap)
	for _, typ := range cfg.TypeWhitelist {
		if _, typExists := graph.typeIdx[typ]; !typExists {
			return nil, models.ErrNoSuchType{TypeName: typ}
		}
		for entry := range graph.typeIdx[typ] {
			response[entry] = graph.Supergraph[entry]
		}
	}

	if cfg.Collapse {
		graph.collapse(response, set.NewStringSet(cfg.TypeWhitelist...))
	}
	return response, nil
}

func (graph InMemoryGraph) collapse(subgraph AdjacencyMap, typeWhitelist set.StringSet) {
	for _, entry := range subgraph {
		entry.Upstreams = graph.collapseInDir(entry, models.DataflowDirUpstream, typeWhitelist)
		entry.Downstreams = graph.collapseInDir(entry, models.DataflowDirDownstream, typeWhitelist)
		subgraph[entry.ID()] = entry
	}
	return
}

func (graph InMemoryGraph) collapseInDir(root AdjacencyEntry, dir models.DataflowDir, types set.StringSet) set.StringSet {
	var (
		queue     = []AdjacencyEntry{root}
		collapsed = set.NewStringSet()
	)
	for len(queue) > 0 {
		n := len(queue)
		entry := queue[n-1]
		queue = queue[:n-1]
		adjacents := entry.AdjacentEntriesInDir(dir)
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

func NewInMemoryGraph(data AdjacencyMap) InMemoryGraph {
	graph := InMemoryGraph{
		Supergraph: data,
		typeIdx:    make(map[string]set.StringSet),
	}
	graph.init()
	return graph
}
