package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/record"
	"github.com/pkg/errors"
)

const (
	// name of the search index
	defaultSearchIndex = "universe"

	defaultScrollTimeout   = 30 * time.Second
	defaultScrollBatchSize = 20

	getAllQuery = `{"query":{"match_all":{}}}`
)

func isReservedName(name string) bool {
	name = strings.ToLower(name)
	return name == defaultSearchIndex
}

// TypeRepository is an implementation of record.TypeRepository
// that uses elasticsearch as a backing store
type TypeRepository struct {
	cli *elasticsearch.Client
}

func (repo *TypeRepository) CreateOrReplace(ctx context.Context, recordTypeName record.TypeName) error {
	if isReservedName(recordTypeName.String()) {
		return record.ErrReservedTypeName{TypeName: recordTypeName.String()}
	}

	// checking for the existence of index before adding the metadata entry
	idxExists, err := indexExists(ctx, repo.cli, recordTypeName.String())
	if err != nil {
		return errors.Wrap(err, "error checking index existance")
	}

	// update/create the index
	if idxExists {
		err = repo.updateIdx(ctx, recordTypeName)
		if err != nil {
			err = errors.Wrap(err, "error updating index")
		}
	} else {
		err = repo.createIdx(ctx, recordTypeName)
		if err != nil {
			err = errors.Wrap(err, "error creating index")
		}
	}

	return err
}

func (repo *TypeRepository) GetByName(ctx context.Context, name string) (record.TypeName, error) {
	isExist, err := indexExists(ctx, repo.cli, name)
	if err != nil {
		return record.TypeName(""), errors.Wrapf(err, "error checking index type existence: %s", name)
	}
	if !isExist {
		return record.TypeName(""), errors.Wrapf(err, "index does not exist: %s", name)
	}

	return record.TypeName(name), nil
}

func (repo *TypeRepository) GetAll(ctx context.Context) (map[record.TypeName]int, error) {
	resp, err := repo.cli.Cat.Indices(
		repo.cli.Cat.Indices.WithFormat("json"),
		repo.cli.Cat.Indices.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error from es client")
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, fmt.Errorf("error from es server: %s", errorReasonFromResponse(resp))
	}
	var indices []esIndex
	err = json.NewDecoder(resp.Body).Decode(&indices)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding es response")
	}

	results := map[record.TypeName]int{}
	for _, index := range indices {
		count, err := strconv.Atoi(index.DocsCount)
		if err != nil {
			return results, errors.Wrap(err, "error converting docs count to a number")
		}
		results[record.TypeName(index.Index)] = count
	}

	return results, nil
}

func (repo *TypeRepository) createIdx(ctx context.Context, recordTypeName record.TypeName) error {
	indexSettings := buildTypeIndexSettings()
	res, err := repo.cli.Indices.Create(
		recordTypeName.String(),
		repo.cli.Indices.Create.WithBody(strings.NewReader(indexSettings)),
		repo.cli.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error creating index %q: %s", recordTypeName, errorReasonFromResponse(res))
	}
	return nil
}

func (repo *TypeRepository) updateIdx(ctx context.Context, recordTypeName record.TypeName) error {
	res, err := repo.cli.Indices.PutMapping(
		strings.NewReader(typeIndexMapping),
		repo.cli.Indices.PutMapping.WithIndex(recordTypeName.String()),
		repo.cli.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error updating index %q: %s", recordTypeName, errorReasonFromResponse(res))
	}
	return nil
}

// checks for the existence of an index
func indexExists(ctx context.Context, cli *elasticsearch.Client, name string) (bool, error) {
	res, err := cli.Indices.Exists(
		[]string{name},
		cli.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("indexExists: %w", elasticSearchError(err))
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}

func buildTypeIndexSettings() string {
	return fmt.Sprintf(indexSettingsTemplate, typeIndexMapping, defaultSearchIndex)
}

func NewTypeRepository(cli *elasticsearch.Client) *TypeRepository {
	return &TypeRepository{
		cli: cli,
	}
}
