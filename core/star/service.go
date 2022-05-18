package star

import (
	"context"

	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/user"
)

func NewService(starRepository Repository) *Service {
	return &Service{
		starRepository: starRepository,
	}
}

type Service struct {
	starRepository Repository
}

func (s *Service) GetStarredAssetsByUserID(ctx context.Context, flt Filter, userID string) ([]asset.Asset, error) {
	return s.starRepository.GetAllAssetsByUserID(ctx, flt, userID)
}
func (s *Service) GetStarredAssetByUserID(ctx context.Context, userID, assetID string) (asset.Asset, error) {
	return s.starRepository.GetAssetByUserID(ctx, userID, assetID)
}
func (s *Service) GetStargazers(ctx context.Context, flt Filter, assetID string) ([]user.User, error) {
	return s.starRepository.GetStargazers(ctx, flt, assetID)
}
func (s *Service) Stars(ctx context.Context, userID, assetID string) (string, error) {
	return s.starRepository.Create(ctx, userID, assetID)
}
func (s *Service) Unstars(ctx context.Context, userID, assetID string) error {
	return s.starRepository.Delete(ctx, userID, assetID)
}
