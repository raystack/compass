package mocks

import (
	"context"

	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/discovery"
	"github.com/stretchr/testify/mock"
)

type TypeRepository struct {
	mock.Mock
}

func (repo *TypeRepository) CreateOrReplace(ctx context.Context, e asset.Type) error {
	args := repo.Called(ctx, e)
	return args.Error(0)
}

func (repo *TypeRepository) GetByName(ctx context.Context, name string) (asset.Type, error) {
	args := repo.Called(ctx, name)
	return args.Get(0).(asset.Type), args.Error(1)
}

func (repo *TypeRepository) GetAll(ctx context.Context) (map[asset.Type]int, error) {
	args := repo.Called(ctx)
	return args.Get(0).(map[asset.Type]int), args.Error(1)
}

type RecordRepositoryFactory struct {
	mock.Mock
}

func (fac *RecordRepositoryFactory) For(typeName string) (discovery.RecordRepository, error) {
	args := fac.Called(typeName)
	return args.Get(0).(discovery.RecordRepository), args.Error(1)
}

type RecordRepository struct {
	mock.Mock
}

func (repo *RecordRepository) CreateOrReplaceMany(ctx context.Context, assets []asset.Asset) error {
	args := repo.Called(ctx, assets)
	return args.Error(0)
}

func (repo *RecordRepository) GetAll(ctx context.Context, cfg discovery.GetConfig) (discovery.RecordList, error) {
	args := repo.Called(ctx, cfg)
	return args.Get(0).(discovery.RecordList), args.Error(1)
}

func (repo *RecordRepository) GetAllIterator(ctx context.Context) (discovery.RecordIterator, error) {
	args := repo.Called(ctx)
	return args.Get(0).(discovery.RecordIterator), args.Error(1)
}

func (repo *RecordRepository) GetByID(ctx context.Context, id string) (asset.Asset, error) {
	args := repo.Called(ctx, id)
	return args.Get(0).(asset.Asset), args.Error(1)
}

func (repo *RecordRepository) Delete(ctx context.Context, id string) error {
	args := repo.Called(ctx, id)
	return args.Error(0)
}

type RecordIterator struct {
	mock.Mock
}

func (m *RecordIterator) Scan() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *RecordIterator) Next() []asset.Asset {
	args := m.Called()
	return args.Get(0).([]asset.Asset)
}

func (m *RecordIterator) Close() error {
	args := m.Called()
	return args.Error(0)
}

type RecordSearcher struct {
	mock.Mock
}

func (searcher *RecordSearcher) Search(ctx context.Context, cfg discovery.SearchConfig) ([]discovery.SearchResult, error) {
	args := searcher.Called(ctx, cfg)
	return args.Get(0).([]discovery.SearchResult), args.Error(1)
}

func (searcher *RecordSearcher) Suggest(ctx context.Context, cfg discovery.SearchConfig) ([]string, error) {
	args := searcher.Called(ctx, cfg)
	return args.Get(0).([]string), args.Error(1)
}
