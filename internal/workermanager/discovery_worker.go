package workermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/pkg/worker"
)

//go:generate mockery --name=DiscoveryRepository -r --case underscore --with-expecter --structname DiscoveryRepository --filename discovery_repository_mock.go --output=./mocks

type DiscoveryRepository interface {
	Upsert(context.Context, asset.Asset) error
	DeleteByURN(ctx context.Context, assetURN string) error
}

func (m *Manager) EnqueueIndexAssetJob(ctx context.Context, ast asset.Asset) error {
	payload, err := json.Marshal(ast)
	if err != nil {
		return fmt.Errorf("enqueue index asset job: serialize payload: %w", err)
	}

	err = m.worker.Enqueue(ctx, worker.JobSpec{
		Type:    jobIndexAsset,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("enqueue index asset job: %w", err)
	}

	return nil
}

func (m *Manager) indexAssetHandler() worker.JobHandler {
	return worker.JobHandler{
		Handle: m.IndexAsset,
		JobOpts: worker.JobOptions{
			MaxAttempts:     3,
			Timeout:         5 * time.Second,
			BackoffStrategy: worker.DefaultExponentialBackoff,
		},
	}
}

func (m *Manager) IndexAsset(ctx context.Context, job worker.JobSpec) error {
	var ast asset.Asset
	if err := json.Unmarshal(job.Payload, &ast); err != nil {
		return fmt.Errorf("index asset: deserialise payload: %w", err)
	}

	if err := m.discoveryRepo.Upsert(ctx, ast); err != nil {
		return &worker.RetryableError{
			Cause: fmt.Errorf("index asset: upsert into discovery repo: %w: urn '%s'", err, ast.URN),
		}
	}
	return nil
}

func (m *Manager) EnqueueDeleteAssetJob(ctx context.Context, urn string) error {
	err := m.worker.Enqueue(ctx, worker.JobSpec{
		Type:    jobDeleteAsset,
		Payload: ([]byte)(urn),
	})
	if err != nil {
		return fmt.Errorf("enqueue delete asset job: %w: urn '%s'", err, urn)
	}

	return nil
}

func (m *Manager) deleteAssetHandler() worker.JobHandler {
	return worker.JobHandler{
		Handle: m.DeleteAsset,
		JobOpts: worker.JobOptions{
			MaxAttempts:     3,
			Timeout:         5 * time.Second,
			BackoffStrategy: worker.DefaultExponentialBackoff,
		},
	}
}

func (m *Manager) DeleteAsset(ctx context.Context, job worker.JobSpec) error {
	urn := (string)(job.Payload)
	if err := m.discoveryRepo.DeleteByURN(ctx, urn); err != nil {
		return &worker.RetryableError{
			Cause: fmt.Errorf("delete asset from discovery repo: %w: urn '%s'", err, urn),
		}
	}
	return nil
}
