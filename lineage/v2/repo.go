package lineage

import (
	"context"
)

//go:generate mockery --name Repository --outpkg mocks --output ../../lib/mocks/ --structname LineageRepository --filename lineage_repository.go
type Repository interface {
	GetGraph(ctx context.Context, node Node) (Graph, error)
	Upsert(ctx context.Context, node Node, upstreams, downstreams []Node) error
}
