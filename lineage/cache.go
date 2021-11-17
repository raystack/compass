package lineage

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// CachedGraph is a add-on component (decorator) available
// that cache's the result of Query() depending on the cfg
// Use NewCachedGraph for building this value
type CachedGraph struct {
	source Graph
	mu     sync.RWMutex
	cache  map[string]AdjacencyMap
}

func (graph *CachedGraph) Query(cfg QueryCfg) (AdjacencyMap, error) {
	graph.mu.RLock()
	cfgHash := graph.hashCfg(cfg)
	memo, exists := graph.cache[cfgHash]
	graph.mu.RUnlock()
	if exists {
		return memo, nil
	}
	result, err := graph.source.Query(cfg)
	if err != nil {
		return nil, fmt.Errorf("CachedGraph: error calling Query() on source graph: %w", err)
	}
	graph.mu.Lock()
	graph.cache[cfgHash] = result
	graph.mu.Unlock()
	return result, nil
}

// a hashing function for obtaining the idtype of a queryCfg object
// subject to change in the future.
func (graph *CachedGraph) hashCfg(cfg QueryCfg) string {
	var values []string
	for _, t := range cfg.TypeWhitelist {
		values = append(values, t.String())
	}

	sort.Strings(values)
	values = append(values, cfg.Root+"%")
	if cfg.Collapse {
		values = append(values, "@")
	}

	return strings.Join(values, ",")
}

func NewCachedGraph(g Graph) *CachedGraph {
	return &CachedGraph{
		source: g,
		cache:  map[string]AdjacencyMap{},
	}
}
