package lineage

import (
	"context"
	"fmt"

	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/models"
)

// Builder encapsulates the algorithm for building a graph
type Builder interface {
	Build(context.Context, models.TypeRepository, models.RecordRepositoryFactory) (Graph, error)
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
func (builder defaultBuilder) Build(ctx context.Context, er models.TypeRepository, rrf models.RecordRepositoryFactory) (Graph, error) {
	typs, err := er.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading type metadata: %w", err)
	}

	var graph = make(AdjacencyMap)
	for _, typ := range typs {
		if err := builder.populateTypeRecords(ctx, graph, typ, rrf); err != nil {
			return nil, fmt.Errorf("error parsing type records: %w", err)
		}
	}

	builder.addBackRefs(graph)
	return NewCachedGraph(NewInMemoryGraph(graph)), nil
}

// load the records for a type onto the graph
// if record has a valid "lineage" field
// it will be used for obtaining the values for downstreams and upstreams.
func (builder defaultBuilder) populateTypeRecords(ctx context.Context, graph AdjacencyMap, typ models.Type, rrf models.RecordRepositoryFactory) error {
	recordRepository, err := rrf.For(typ)
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
			builder.addRecord(graph, typ.Name, record)
		}
	}
	return nil
}

// add the corresponding records to graph.
// Uses lineageProcessor to obtain information about upstreams/downstreams
func (builder defaultBuilder) addRecord(graph AdjacencyMap, typeName string, record models.Record) {
	entry := AdjacencyEntry{
		Type:        typeName,
		URN:         record.Urn,
		Downstreams: builder.buildDownstreams(record),
		Upstreams:   builder.buildUpstreams(record),
	}

	graph[entry.ID()] = entry
}

// Follow and validate all references, adding back-refs where needed
func (builder defaultBuilder) addBackRefs(graph AdjacencyMap) {
	for _, entry := range graph {
		// for {upstream, downstream}, find adjacents and
		// validate refs on both nodes. If any node is missing
		// a corresponding entry, add it.
		for _, dir := range models.AllDataflowDir {
			for adjacent := range entry.AdjacentEntriesInDir(dir) {
				builder.addBackRefIfNotFound(graph, entry, adjacent, dir)
			}
		}
	}
}

func (builder defaultBuilder) addBackRefIfNotFound(graph AdjacencyMap, entry AdjacencyEntry, adjacent string, dir models.DataflowDir) {
	adjacentEntry, exists := graph[adjacent]
	if !exists {
		return
	}
	oppositeDir, known := builder.opposite(dir)
	if !known {
		return
	}
	entries := adjacentEntry.AdjacentEntriesInDir(oppositeDir)
	if _, exists := entries[entry.ID()]; !exists {
		entries.Add(entry.ID())
	}
	return
}

func (builder defaultBuilder) opposite(dir models.DataflowDir) (models.DataflowDir, bool) {
	switch dir {
	case models.DataflowDirUpstream:
		return models.DataflowDirDownstream, true
	case models.DataflowDirDownstream:
		return models.DataflowDirUpstream, true
	default:
		return models.DataflowDir("unknown"), false
	}
}

func (builder defaultBuilder) buildUpstreams(record models.Record) (upstreams set.StringSet) {
	upstreams = set.NewStringSet()

	for _, lr := range record.Upstreams {
		urnWithType := builder.addTypePrefix(lr.Urn, lr.Type)
		upstreams.Add(urnWithType)
	}

	return
}

func (builder defaultBuilder) buildDownstreams(record models.Record) (downstreams set.StringSet) {
	downstreams = set.NewStringSet()

	for _, lr := range record.Downstreams {
		urnWithType := builder.addTypePrefix(lr.Urn, lr.Type)
		downstreams.Add(urnWithType)
	}

	return
}

func (builder defaultBuilder) addTypePrefix(urn string, typ string) (rv string) {
	return fmt.Sprintf("%s/%s", typ, urn)
}
