package postgres

import (
	"github.com/odpf/columbus/lineage/v2"
)

type GraphModel []EdgeModel

func (gm GraphModel) toGraph() lineage.Graph {
	graph := lineage.Graph{}
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

func (m EdgeModel) toEdge() lineage.Edge {
	edge := lineage.Edge{
		Source: m.Source,
		Target: m.Target,
		Prop:   m.Prop,
	}

	return edge
}
