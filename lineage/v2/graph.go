package lineage

import "github.com/odpf/columbus/asset"

type Graph []Edge

type Edge struct {
	// Source represents source node ID
	Source string `json:"source"`

	// Source represents source node ID
	Target string `json:"target"`

	// Prop is a map containing extra information about the edge
	Prop map[string]interface{} `json:"prop"`
}

type Node struct {
	URN     string
	Type    asset.Type
	Service string
}
