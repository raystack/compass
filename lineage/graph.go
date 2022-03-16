package lineage

import "github.com/odpf/columbus/asset"

type Graph []Edge

type Edge struct {
	// Source represents source's node ID
	Source string `json:"source"`

	// Target represents target's node ID
	Target string `json:"target"`

	// Prop is a map containing extra information about the edge
	Prop map[string]interface{} `json:"prop"`
}

type Node struct {
	URN     string     `json:"urn"`
	Type    asset.Type `json:"type"`
	Service string     `json:"service"`
}
