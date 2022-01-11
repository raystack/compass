package lineage

import (
	"context"
	"fmt"

	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/record"
)

// Builder encapsulates the algorithm for building a graph
type Builder interface {
	Build(context.Context, record.TypeRepository, discovery.RecordRepositoryFactory) (Graph, error)
}

var DefaultBuilder = defaultBuilder{}

type defaultBuilder struct{}

// Build the lineage graph for all available types
// The graph construction algorithm works roughly as follows:
// Given,
//   AT = collection of all types
//   ER = collection of records belonging to a particular type
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
// and all records have been added to the graph, we follow these links and check if the referred entry has a
// corresponding entry pointing back (called an external or back-ref). If not we add it.
// If any reference'd record is not found in the graph, the algorithm gives up and looks at the next related entry.
// This has the effect of phantom references in graph: A resource may refer another resource in the graph, but that resource
// may not be available in the graph
func (builder defaultBuilder) Build(ctx context.Context, tr record.TypeRepository, rrf discovery.RecordRepositoryFactory) (Graph, error) {
	typs, err := tr.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading type metadata: %w", err)
	}

	var graph = make(AdjacencyMap)
	for typName := range typs {
		if err := builder.populateTypeRecords(ctx, graph, typName.String(), rrf); err != nil {
			return nil, fmt.Errorf("error parsing type records: %w", err)
		}
	}

	builder.addBackRefs(graph)
	return NewCachedGraph(NewInMemoryGraph(graph)), nil
}

// load the records for a type onto the graph
// if record has a valid "lineage" field
// it will be used for obtaining the values for downstreams and upstreams.
func (builder defaultBuilder) populateTypeRecords(ctx context.Context, graph AdjacencyMap, typName string, rrf discovery.RecordRepositoryFactory) error {
	recordRepository, err := rrf.For(typName)
	if err != nil {
		return fmt.Errorf("error obtaing record repository: %w", err)
	}

	recordIter, err := recordRepository.GetAllIterator(ctx)
	if err != nil {
		return fmt.Errorf("error getting record iterator: %w", err)
	}

	defer recordIter.Close()
	for recordIter.Scan() {
		for _, record := range recordIter.Next() {
			builder.addRecord(graph, typName, record)
		}
	}
	return nil
}

// add the corresponding records to graph.
// Uses lineageProcessor to obtain information about upstreams/downstreams
func (builder defaultBuilder) addRecord(graph AdjacencyMap, typeName string, record record.Record) {
	entry := AdjacencyEntry{
		Type:        typeName,
		URN:         record.Urn,
		Service:     record.Service,
		Downstreams: builder.buildAdjacents(record, dataflowDirDownstream),
		Upstreams:   builder.buildAdjacents(record, dataflowDirUpstream),
	}

	graph[entry.ID()] = entry
}

// Follow and validate all references, adding back-refs where needed
func (builder defaultBuilder) addBackRefs(graph AdjacencyMap) {
	for _, entry := range graph {
		// for {upstream, downstream}, find adjacents and
		// validate refs on both nodes. If any node is missing
		// a corresponding entry, add it.

		entryID := entry.ID()
		for upstream := range entry.Upstreams {
			builder.addBackRef(graph, entryID, upstream, dataflowDirDownstream)
		}
		for downstream := range entry.Downstreams {
			builder.addBackRef(graph, entryID, downstream, dataflowDirUpstream)
		}
	}
}

func (builder defaultBuilder) addBackRef(graph AdjacencyMap, refID string, backRefID string, dir dataflowDir) {
	backRefEntry, exists := graph[backRefID]
	if !exists {
		return
	}

	adjacents := backRefEntry.getAdjacents(dir)
	if _, exists := adjacents[refID]; !exists {
		adjacents.Add(refID)
	}

	return
}

func (builder defaultBuilder) buildAdjacents(r record.Record, dir dataflowDir) (adjacents set.StringSet) {
	adjacents = set.NewStringSet()

	var lineageRecords []record.LineageRecord
	if dir == dataflowDirUpstream {
		lineageRecords = r.Upstreams
	} else {
		lineageRecords = r.Downstreams
	}

	for _, lr := range lineageRecords {
		urnWithType := fmt.Sprintf("%s/%s", lr.Type, lr.Urn)
		adjacents.Add(urnWithType)
	}

	return
}
