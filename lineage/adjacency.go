package lineage

import (
	"github.com/odpf/columbus/lib/set"
)

// AdjacencyEntry holds metadata about a resource
// (like it’s name and type) as well it’s relationship to other resources
type AdjacencyEntry struct {
	ID          string        `json:"id"`
	URN         string        `json:"urn"`
	Type        string        `json:"type"`
	Service     string        `json:"service"`
	Upstreams   set.StringSet `json:"upstreams"`
	Downstreams set.StringSet `json:"downstreams"`
}

func (e AdjacencyEntry) getAdjacents(dir dataflowDir) set.StringSet {
	if dir == dataflowDirDownstream {
		return e.Downstreams
	}

	return e.Upstreams
}

// AdjacencyMap is a composite representation of graph.
// An AdjacencyMap is a hashmap analogue to the Adjacency List data structure.
// The key of the hashmap is a unique resource identifier while the value is an instance of AdjacencyData
type AdjacencyMap map[string]AdjacencyEntry
