package mock

import (
	"context"
	"io/ioutil"

	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/lineage"
	"github.com/odpf/columbus/record"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

type TypeRepository struct {
	mock.Mock
}

func (repo *TypeRepository) CreateOrReplace(ctx context.Context, e record.TypeName) error {
	args := repo.Called(ctx, e)
	return args.Error(0)
}

func (repo *TypeRepository) GetByName(ctx context.Context, name string) (record.TypeName, error) {
	args := repo.Called(ctx, name)
	return args.Get(0).(record.TypeName), args.Error(1)
}

func (repo *TypeRepository) GetAll(ctx context.Context) (map[record.TypeName]int, error) {
	args := repo.Called(ctx)
	return args.Get(0).(map[record.TypeName]int), args.Error(1)
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

func (repo *RecordRepository) CreateOrReplaceMany(ctx context.Context, records []record.Record) error {
	args := repo.Called(ctx, records)
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

func (repo *RecordRepository) GetByID(ctx context.Context, id string) (record.Record, error) {
	args := repo.Called(ctx, id)
	return args.Get(0).(record.Record), args.Error(1)
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

func (m *RecordIterator) Next() []record.Record {
	args := m.Called()
	return args.Get(0).([]record.Record)
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

type LineageProvider struct {
	mock.Mock
}

func (lp *LineageProvider) Graph() (lineage.Graph, error) {
	args := lp.Called()
	return args.Get(0).(lineage.Graph), args.Error(1)
}

type Graph struct {
	mock.Mock
}

func (graph *Graph) Query(cfg lineage.QueryCfg) (lineage.AdjacencyMap, error) {
	args := graph.Called(cfg)
	return args.Get(0).(lineage.AdjacencyMap), args.Error(1)
}

type Logger struct{}

func (l Logger) WithField(key string, value interface{}) *logrus.Entry {
	return logrus.NewEntry(&logrus.Logger{Out: ioutil.Discard})
}
func (l Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return logrus.NewEntry(&logrus.Logger{Out: ioutil.Discard})
}
func (l Logger) WithError(err error) *logrus.Entry {
	return logrus.NewEntry(&logrus.Logger{Out: ioutil.Discard})
}
func (l Logger) Debugf(format string, args ...interface{})   {}
func (l Logger) Infof(format string, args ...interface{})    {}
func (l Logger) Printf(format string, args ...interface{})   {}
func (l Logger) Warnf(format string, args ...interface{})    {}
func (l Logger) Warningf(format string, args ...interface{}) {}
func (l Logger) Errorf(format string, args ...interface{})   {}
func (l Logger) Fatalf(format string, args ...interface{})   {}
func (l Logger) Panicf(format string, args ...interface{})   {}
func (l Logger) Debug(args ...interface{})                   {}
func (l Logger) Info(args ...interface{})                    {}
func (l Logger) Print(args ...interface{})                   {}
func (l Logger) Warn(args ...interface{})                    {}
func (l Logger) Warning(args ...interface{})                 {}
func (l Logger) Error(args ...interface{})                   {}
func (l Logger) Fatal(args ...interface{})                   {}
func (l Logger) Panic(args ...interface{})                   {}
func (l Logger) Debugln(args ...interface{})                 {}
func (l Logger) Infoln(args ...interface{})                  {}
func (l Logger) Println(args ...interface{})                 {}
func (l Logger) Warnln(args ...interface{})                  {}
func (l Logger) Warningln(args ...interface{})               {}
func (l Logger) Errorln(args ...interface{})                 {}
func (l Logger) Fatalln(args ...interface{})                 {}
func (l Logger) Panicln(args ...interface{})                 {}
