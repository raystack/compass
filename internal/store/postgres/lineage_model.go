package postgres

import (
	"github.com/odpf/compass/core/asset"
)

type GraphModel []EdgeModel

func (gm GraphModel) toGraph() asset.Graph {
	graph := asset.Graph{}
	for _, em := range gm {
		graph = append(graph, em.toEdge())
	}

	return graph
}

type EdgeModel struct {
	Source string  `db:"source"`
	Target string  `db:"target"`
	Prop   JSONMap `db:"prop"`
}

func (m EdgeModel) toEdge() asset.Edge {
	edge := asset.Edge{
		Source: m.Source,
		Target: m.Target,
		Prop:   m.Prop,
	}

	return edge
}
