package lineage

import (
	"context"
	"fmt"

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
//       use lineage processor to find upstreams & downstreams
//       add entry to graph
//   for entry in graph:
//     follow all refs in entry, adding backrefs where needed
// Explanation:
// We process every record for every type once. If it has a proper lineage metadata configured, it will
// be used for obtaining the IDs of related resources. These are called internal or forward-refs. Once all document-refs are resolved
// and all records have been added to the graph, we follow these links and check if the referred entry has a
// corresponding entry pointing back (called an external or back-ref). If not we add it.
// If any reference'd record is not found in the graph, the algorithm gives up and looks at the next related entry.
// This has the effect of phatom references in graph: A resource may refer another resource in the graph, but that resource
// may not be available in the graph
func (builder defaultBuilder) Build(ctx context.Context, er models.TypeRepository, rrf models.RecordRepositoryFactory) (Graph, error) {
	typs, err := er.GetAll(ctx)
	if err != nil {
		return InMemoryGraph{}, fmt.Errorf("error loading type metadata: %w", err)
	}

	var graph = make(AdjacencyMap)
	for _, typ := range typs {
		if err := builder.populateTypeRecords(ctx, graph, typ, rrf); err != nil {
			return InMemoryGraph{}, fmt.Errorf("error parsing type records: %w", err)
		}
	}

	builder.addBackRefs(graph)
	return NewCachedGraph(NewInMemoryGraph(graph)), nil
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

// load the records for a type onto the graph
// if type definition has a valid "lineage" field
// it will be used for obtaining the values for downstreams and upstreams.
func (builder defaultBuilder) populateTypeRecords(ctx context.Context, graph AdjacencyMap, typ models.Type, rrf models.RecordRepositoryFactory) error {
	lineageProc := newLineageProcessor(typ.Lineage)
	recordRepository, err := rrf.For(typ)
	if err != nil {
		return fmt.Errorf("error obtaing record repository: %w", err)
	}

	recordIter, err := recordRepository.GetAllIterator(ctx)
	if err != nil {
		panic(err)
	}
	defer recordIter.Close()
	for recordIter.Scan() {
		for _, record := range recordIter.Next() {
			if err := builder.addRecord(graph, typ, record, lineageProc); err != nil {
				return fmt.Errorf("error adding record to graph: %w", err)
			}
		}
	}
	return nil
}

// add the corresponding records to graph.
// Uses lineageProcessor to obtain information about upstreams/downstreams
func (builder defaultBuilder) addRecord(graph AdjacencyMap, typ models.Type, record models.Record, lineageProcessor lineageProcessor) error {
	recordID, ok := record[typ.Fields.ID].(string)
	if !ok {
		return fmt.Errorf("record missing ID field; record=%#v type=%#v", record, typ)
	}
	upstreams, downstreams, err := lineageProcessor.LineageOf(record)
	if err != nil {
		return fmt.Errorf("error obtaining lineage for record %q: %w", recordID, err)
	}
	var (
		entry = AdjacencyEntry{
			Type:        typ.Name,
			URN:         recordID,
			Downstreams: downstreams,
			Upstreams:   upstreams,
		}
	)
	graph[entry.ID()] = entry
	return nil
}

type StaticGraphBuilder struct {
	Graph InMemoryGraph
}

func (bld StaticGraphBuilder) Build(models.TypeRepository, models.RecordRepositoryFactory) (Graph, error) {
	return bld.Graph, nil
}
