package lineage

import "github.com/odpf/columbus/asset"

type Graph []Edge

type Edge struct {
	// Source represents source node ID
	Source string

	// Source represents source node ID
	Target string

	// Prop is a map containing extra information about the edge
	Prop map[string]interface{}
}

type Node struct {
	URN     string
	Type    asset.Type
	Service string
}
