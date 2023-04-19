package asset

import (
	"context"
)

type LineageDirection string

func (dir LineageDirection) IsValid() bool {
	switch dir {
	case LineageDirectionUpstream, LineageDirectionDownstream, "":
		return true
	default:
		return false
	}
}

const (
	LineageDirectionUpstream   LineageDirection = "upstream"
	LineageDirectionDownstream LineageDirection = "downstream"
)

type LineageQuery struct {
	Level     int
	Direction LineageDirection
}

//go:generate mockery --name=LineageRepository -r --case underscore --with-expecter --structname=LineageRepository --filename=lineage_repository.go --output=./mocks
type LineageRepository interface {
	GetGraph(ctx context.Context, urn string, query LineageQuery) (LineageGraph, error)
	Upsert(ctx context.Context, urn string, upstreams, downstreams []string) error
	DeleteByURN(ctx context.Context, urn string) error
}

type LineageGraph []LineageEdge

type Lineage struct {
	Edges     []LineageEdge             `json:"edges"`
	NodeAttrs map[string]NodeAttributes `json:"node_attrs"`
}

type LineageEdge struct {
	// Source represents source's node ID
	Source string `json:"source"`

	// Target represents target's node ID
	Target string `json:"target"`

	// Prop is a map containing extra information about the edge
	Prop map[string]interface{} `json:"prop"`
}

type NodeAttributes struct {
	Probes ProbesInfo `json:"probes"`
}

type ProbesInfo struct {
	Latest Probe `json:"latest"`
}
