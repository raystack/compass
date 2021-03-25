package lineage

import (
	"fmt"
	"strings"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/models"
)

// AdjacencyEntry holds metadata about a resource
// (like it’s name and type) as well it’s relationship to other resources
type AdjacencyEntry struct {
	Type        string        `json:"type"`
	URN         string        `json:"urn"`
	Upstreams   set.StringSet `json:"upstreams"`
	Downstreams set.StringSet `json:"downstreams"`
}

func (e AdjacencyEntry) AdjacentEntriesInDir(dir models.DataflowDir) set.StringSet {
	switch dir {
	case models.DataflowDirDownstream:
		return e.Downstreams
	case models.DataflowDirUpstream:
		return e.Upstreams
	default:
		return set.StringSet{}
	}
}

func (e AdjacencyEntry) ID() string {
	var (
		typ = strings.TrimSpace(e.Type)
		urn = strings.TrimSpace(e.URN)
	)
	if typ == "" && urn == "" {
		return "<unknown>/<unknown>"
	}
	return fmt.Sprintf("%s/%s", typ, urn)
}

// AdjacencyMap is a composite representation of graph.
// An AdjacencyMap is a hashmap analogue to the Adjacency List data structure.
// The key of the hashmap is a unique resource identifier while the value is an instance of AdjacencyData
type AdjacencyMap map[string]AdjacencyEntry
