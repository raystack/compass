package discovery

import (
	"context"
	"fmt"

	"github.com/odpf/columbus/record"
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
		return fmt.Errorf("error building repo for type \"%s\": %w", typeName, err)
	}

	err = repo.CreateOrReplaceMany(ctx, records)
	if err != nil {
		return fmt.Errorf("error upserting records: %w", err)
	}

	return nil
}

func (s *Service) DeleteRecord(ctx context.Context, typeName string, recordURN string) error {
	repo, err := s.factory.For(typeName)
	if err != nil {
		return fmt.Errorf("error building repo for type \"%s\": %w", typeName, err)
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
