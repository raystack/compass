package discovery

import (
	"context"

	"github.com/odpf/columbus/asset"
	"github.com/pkg/errors"
)

type Service struct {
	factory        RecordRepositoryFactory
	recordSearcher RecordSearcher
}

func NewService(factory RecordRepositoryFactory, recordSearcher RecordSearcher) *Service {
	return &Service{
		factory:        factory,
		recordSearcher: recordSearcher,
	}
}

func (s *Service) Upsert(ctx context.Context, typeName string, assets []asset.Asset) (err error) {
	repo, err := s.factory.For(typeName)
	if err != nil {
		return errors.Wrapf(err, "error building repo for type \"%s\"", typeName)
	}

	err = repo.CreateOrReplaceMany(ctx, assets)
	if err != nil {
		return errors.Wrap(err, "error upserting assets")
	}

	return nil
}

func (s *Service) DeleteRecord(ctx context.Context, typeName string, recordURN string) error {
	repo, err := s.factory.For(typeName)
	if err != nil {
		return errors.Wrapf(err, "error building repo for type \"%s\"", typeName)
	}

	err = repo.Delete(ctx, recordURN)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Search(ctx context.Context, cfg SearchConfig) ([]SearchResult, error) {
	return s.recordSearcher.Search(ctx, cfg)
}

func (s *Service) Suggest(ctx context.Context, cfg SearchConfig) ([]string, error) {
	return s.recordSearcher.Suggest(ctx, cfg)
}
