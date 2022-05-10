package discovery

import (
	"context"
	"fmt"

	"github.com/odpf/compass/asset"
)

type Service struct {
	factory       AssetRepositoryFactory
	assetSearcher AssetSearcher
}

func NewService(factory AssetRepositoryFactory, assetSearcher AssetSearcher) *Service {
	return &Service{
		factory:       factory,
		assetSearcher: assetSearcher,
	}
}

func (s *Service) Upsert(ctx context.Context, typeName string, assets []asset.Asset) (err error) {
	repo, err := s.factory.For(typeName)
	if err != nil {
		return fmt.Errorf("error building repo for type \"%s\": %w", typeName, err)
	}

	err = repo.CreateOrReplaceMany(ctx, assets)
	if err != nil {
		return fmt.Errorf("error upserting assets: %w", err)
	}

	return nil
}

func (s *Service) DeleteAsset(ctx context.Context, typeName string, assetURN string) error {
	repo, err := s.factory.For(typeName)
	if err != nil {
		return fmt.Errorf("error building repo for type \"%s\": %w", typeName, err)
	}

	err = repo.Delete(ctx, assetURN)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Search(ctx context.Context, cfg SearchConfig) ([]SearchResult, error) {
	return s.assetSearcher.Search(ctx, cfg)
}

func (s *Service) Suggest(ctx context.Context, cfg SearchConfig) ([]string, error) {
	return s.assetSearcher.Suggest(ctx, cfg)
}
