package asset

import (
	"context"
)

//go:generate mockery --name=LineageRepository -r --case underscore --with-expecter --structname LineageRepository --filename lineage_repository.go --output=./mocks
type LineageRepository interface {
	GetGraph(ctx context.Context, node LineageNode) (LineageGraph, error)
	Upsert(ctx context.Context, node LineageNode, upstreams, downstreams []LineageNode) error
}

type LineageGraph []LineageEdge

type LineageEdge struct {
	// Source represents source's node ID
	Source string `json:"source"`

	// Target represents target's node ID
	Target string `json:"target"`

	// Prop is a map containing extra information about the edge
	Prop map[string]interface{} `json:"prop"`
}

type LineageNode struct {
	URN     string `json:"urn"`
	Type    Type   `json:"type"`
	Service string `json:"service"`
}
