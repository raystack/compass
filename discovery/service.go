package discovery

import (
	"context"

	"github.com/odpf/columbus/record"
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

func (s *Service) Upsert(ctx context.Context, typeName string, records []record.Record) (err error) {
	repo, err := s.factory.For(typeName)
	if err != nil {
		return errors.Wrapf(err, "error building repo for type \"%s\"", typeName)
	}

	err = repo.CreateOrReplaceMany(ctx, records)
	if err != nil {
		return errors.Wrap(err, "error upserting records")
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
