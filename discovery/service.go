package discovery

import (
	"context"
	"fmt"

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

func (s *Service) Upsert(ctx context.Context, t record.Type, records []record.Record) (err error) {
	repo, err := s.factory.For(t)
	if err != nil {
		return errors.Wrapf(err, "error building repo for type \"%s\"", t)
	}

	err = repo.CreateOrReplaceMany(ctx, records)
	if err != nil {
		return errors.Wrap(err, "error upserting records")
	}

	return nil
}

func (s *Service) DeleteRecord(ctx context.Context, t record.Type, recordURN string) error {
	repo, err := s.factory.For(t)
	if err != nil {
		return errors.Wrapf(err, "error building repo for type \"%s\"", t)
	}

	err = repo.Delete(ctx, recordURN)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Search(ctx context.Context, cfg SearchConfig) (records []record.Record, err error) {
	fmt.Printf("%+v\n", cfg.TypeWhiteList)
	return s.recordSearcher.Search(ctx, cfg)
}
