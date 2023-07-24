package workermanager_test

import (
	"errors"
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/internal/workermanager/mocks"
	"github.com/stretchr/testify/assert"
)

func TestInSituWorker_EnqueueIndexAssetJob(t *testing.T) {
	sampleAsset := asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.TypeDashboard, Service: "some-service"}

	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "Failure",
			discoveryErr: errors.New("fail"),
			expectedErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				Upsert(ctx, sampleAsset).
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueIndexAssetJob(ctx, sampleAsset)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInSituWorker_EnqueueDeleteAssetJob(t *testing.T) {
	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "Failure",
			discoveryErr: errors.New("fail"),
			expectedErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				DeleteByURN(ctx, "some-urn").
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueDeleteAssetJob(ctx, "some-urn")
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
