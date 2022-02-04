package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStarModel(t *testing.T) {
	t.Run("successfully build  star from star model", func(t *testing.T) {
		assetModel := &AssetModel{
			URN:  "asseturn",
			Type: "assettype",
		}
		sm := &StarModel{
			ID:        "id",
			UserID:    "userid",
			AssetURN:  "asseturn",
			AssetType: "assettype",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s := sm.toStar(assetModel)

		assert.Equal(t, s.ID, sm.ID)
		assert.Equal(t, s.Asset.URN, sm.AssetURN)
		assert.Equal(t, s.Asset.Type.String(), sm.AssetType)
		assert.True(t, s.CreatedAt.Equal(sm.CreatedAt))
		assert.True(t, s.UpdatedAt.Equal(sm.UpdatedAt))
	})
}
