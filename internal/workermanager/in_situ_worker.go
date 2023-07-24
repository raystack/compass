package workermanager

import (
	"context"
	"fmt"

	"github.com/goto/compass/core/asset"
)

type InSituWorker struct {
	discoveryRepo DiscoveryRepository
}

func NewInSituWorker(deps Deps) *InSituWorker {
	return &InSituWorker{
		discoveryRepo: deps.DiscoveryRepo,
	}
}

func (m *InSituWorker) EnqueueIndexAssetJob(ctx context.Context, ast asset.Asset) error {
	if err := m.discoveryRepo.Upsert(ctx, ast); err != nil {
		return fmt.Errorf("index asset: upsert into discovery repo: %w: urn '%s'", err, ast.URN)
	}

	return nil
}

func (m *InSituWorker) EnqueueDeleteAssetJob(ctx context.Context, urn string) error {
	if err := m.discoveryRepo.DeleteByURN(ctx, urn); err != nil {
		return fmt.Errorf("delete asset from discovery repo: %w: urn '%s'", err, urn)
	}
	return nil
}

func (*InSituWorker) Close() error { return nil }
