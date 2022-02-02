package lineage

import (
	"context"
	"time"

	"github.com/odpf/columbus/asset"
)

// dataflowDir describes the direction of data IO with regards to another resource
// it's "upstream" if the application only reads, "downstream" if the application only writes
// "bidirectional" if the application both reads and writes to another resource.
// "bidirectional" direction is currently not supported.
type dataflowDir string

const (
	dataflowDirUpstream   = dataflowDir("upstream")
	dataflowDirDownstream = dataflowDir("downstream")
)

type Node struct {
	asset.Asset
	Upstreams   []Node
	Downstreams []Node
}

type Edge struct {
	ID        string    `json:"id"`
	SourceID  string    `json:"source_id"`
	TargetID  string    `json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Source    Node      `json:"source"`
	Target    Node      `json:"target"`
}

type Repository interface {
	GetEdges(context.Context) ([]Edge, error)
}
