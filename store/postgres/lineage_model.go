package postgres

import (
	"github.com/raystack/compass/core/asset"
)

type LineageGraphModel []LineageEdgeModel

func (gm LineageGraphModel) toGraph() asset.LineageGraph {
	graph := asset.LineageGraph{}
	for _, em := range gm {
		graph = append(graph, em.toEdge())
	}

	return graph
}

type LineageEdgeModel struct {
	NamespaceID string  `db:"namespace_id"`
	Source      string  `db:"source"`
	Target      string  `db:"target"`
	Prop        JSONMap `db:"prop"`
}

func (m LineageEdgeModel) toEdge() asset.LineageEdge {
	edge := asset.LineageEdge{
		Source: m.Source,
		Target: m.Target,
		Prop:   m.Prop,
	}

	return edge
}
