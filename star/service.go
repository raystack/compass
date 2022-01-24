package star

import (
	"context"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/user"
)

// Service is a business layer to manage stars
type Service struct {
	repository Repository
	assetRepo  asset.Repository
}

// Star bookmarks an asset for a user
func (s *Service) Star(ctx context.Context, userID string, starring *Star) (string, error) {
	if err := starring.ValidateAssetURN(); err != nil {
		return "", err
	}
	assetID, err := s.assetRepo.GetIDByURN(ctx, &starring.Asset)
	if err != nil {
		return "", err
	}
	starID, err := s.repository.Create(ctx, userID, assetID)
	if err != nil {
		return "", err
	}
	return starID, nil
}

// GetStargazersByURN returns all users that stars an asset by asset urn and type
func (s *Service) GetStargazersByURN(ctx context.Context, cfg Config, starring *Star) ([]user.User, error) {
	if err := starring.ValidateAssetURN(); err != nil {
		return nil, err
	}
	assetID, err := s.assetRepo.GetIDByURN(ctx, &starring.Asset)
	if err != nil {
		return nil, err
	}
	users, err := s.repository.GetStargazers(ctx, cfg, assetID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetStargazersByID returns all users that stars an asset by asset id
func (s *Service) GetStargazersByID(ctx context.Context, cfg Config, assetID string) ([]user.User, error) {
	if assetID == "" {
		return nil, InvalidError{}
	}
	users, err := s.repository.GetStargazers(ctx, cfg, assetID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetAllAssetsByUserID returns all assets starred by a user
func (s *Service) GetAllAssetsByUserID(ctx context.Context, cfg Config, userID string) ([]asset.Asset, error) {
	assets, err := s.repository.GetAllAssetsByUserID(ctx, cfg, userID)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// GetAssetByUserID returns an asset starred by a user
func (s *Service) GetAssetByUserID(ctx context.Context, userID string, starring *Star) (*asset.Asset, error) {
	if err := starring.ValidateAssetURN(); err != nil {
		return nil, err
	}
	assetID, err := s.assetRepo.GetIDByURN(ctx, &starring.Asset)
	if err != nil {
		return nil, err
	}
	ast, err := s.repository.GetAssetByUserID(ctx, userID, assetID)
	if err != nil {
		return nil, err
	}
	return ast, nil
}

// Unstar deletes a starred asset
func (s *Service) Unstar(ctx context.Context, userID string, starring *Star) error {
	if err := starring.ValidateAssetURN(); err != nil {
		return err
	}
	assetID, err := s.assetRepo.GetIDByURN(ctx, &starring.Asset)
	if err != nil {
		return err
	}
	err = s.repository.Delete(ctx, userID, assetID)
	if err != nil {
		return err
	}
	return nil
}

// NewService is a function to create star service
func NewService(repository Repository, assetRepo asset.Repository) *Service {
	return &Service{
		repository: repository,
		assetRepo:  assetRepo,
	}
}
