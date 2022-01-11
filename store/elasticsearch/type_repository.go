package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/record"
	"github.com/pkg/errors"
)

const (
	// name of the metadata index
	defaultMetaIndex = "meta"

	// name of the search index
	defaultSearchIndex = "universe"

	defaultScrollTimeout   = 30 * time.Second
	defaultScrollBatchSize = 20

	getAllQuery = `{"query":{"match_all":{}}}`
)

func isReservedName(name string) bool {
	name = strings.ToLower(name)
	return name == defaultMetaIndex || name == defaultSearchIndex
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

	// checking for the existence of index before adding the metadata
	// entry, because if this operation fails, we don't have to do a rollback
	// for the addTypeToMetaIdx()
	idxExists, err := indexExists(ctx, repo.cli, recordTypeName.String())
	if err != nil {
		return errors.Wrap(err, "error checking index existance")
	}

	// save the type information to meta index
	if err := repo.addTypeToMetaIdx(ctx, recordTypeName); err != nil {
		return errors.Wrap(err, "error storing type")
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
	res, err := repo.cli.Get(
		defaultMetaIndex,
		name,
		repo.cli.Get.WithRefresh(true),
		repo.cli.Get.WithContext(ctx),
	)
	if err != nil {
		return record.TypeName(""), elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return record.TypeName(""), record.ErrNoSuchType{TypeName: name}
		}
		return record.TypeName(""), fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}

	var response = struct {
		Source struct {
			Name string `json:"name"`
		} `json:"_source"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return record.TypeName(""), fmt.Errorf("error parsing elasticsearch response: %w", err)
	}
	return record.TypeName(response.Source.Name), nil
}

func (repo *TypeRepository) GetAll(ctx context.Context) ([]record.TypeName, error) {
	body := strings.NewReader(getAllQuery)
	resp, err := repo.cli.Search(
		repo.cli.Search.WithIndex(defaultMetaIndex),
		repo.cli.Search.WithBody(body),
		repo.cli.Search.WithScroll(defaultScrollTimeout),
		repo.cli.Search.WithSize(defaultScrollBatchSize),
		repo.cli.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(resp))
	}

	var response typeResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error decoding es response: %w", err)
	}
	var results = repo.toTypeList(response)
	var scrollID = response.ScrollID
	for {
		var types []record.TypeName
		types, scrollID, err = repo.scrollRecord(ctx, scrollID)
		if err != nil {
			return nil, fmt.Errorf("error scrolling results: %v", err)
		}
		if len(types) == 0 {
			break
		}
		results = append(results, types...)
	}
	return results, nil
}

func (repo *TypeRepository) GetRecordsCount(ctx context.Context) (map[string]int, error) {
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

	results := map[string]int{}
	for _, index := range indices {
		if index.Index == defaultMetaIndex {
			continue
		}

		count, err := strconv.Atoi(index.DocsCount)
		if err != nil {
			return results, errors.Wrap(err, "error converting docs count to a number")
		}
		results[index.Index] = count
	}

	return results, nil
}

func (repo *TypeRepository) addTypeToMetaIdx(ctx context.Context, recordTypeName record.TypeName) error {
	raw := bytes.NewBuffer(nil)
	err := json.NewEncoder(raw).Encode(record.Type{Name: recordTypeName})
	if err != nil {
		return fmt.Errorf("error encoding type: %w", err)
	}

	res, err := repo.cli.Index(
		defaultMetaIndex,
		raw,
		repo.cli.Index.WithDocumentID(recordTypeName.String()),
		repo.cli.Index.WithRefresh("true"),
		repo.cli.Index.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}
	return nil
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

func (repo *TypeRepository) scrollRecord(ctx context.Context, scrollID string) ([]record.TypeName, string, error) {
	resp, err := repo.cli.Scroll(
		repo.cli.Scroll.WithScrollID(scrollID),
		repo.cli.Scroll.WithScroll(defaultScrollTimeout),
		repo.cli.Scroll.WithContext(ctx),
	)
	if err != nil {
		return nil, "", fmt.Errorf("error executing scroll: %v", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, "", fmt.Errorf("error response from elasticsearch: %v", errorReasonFromResponse(resp))
	}

	var response typeResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	resp.Body.Close()
	if err != nil {
		return nil, "", fmt.Errorf("error decoding es response: %w", err)
	}

	return repo.toTypeList(response), response.ScrollID, nil
}

func (repo *TypeRepository) toTypeList(response typeResponse) []record.TypeName {
	types := []record.TypeName{}
	for _, hit := range response.Hits.Hits {
		types = append(types, record.TypeName(hit.Source.Name))
	}

	return types
}

func (repo *TypeRepository) Delete(ctx context.Context, typeName string) error {
	if isReservedName(typeName) {
		return record.ErrReservedTypeName{TypeName: typeName}
	}

	res, err := repo.cli.Delete(
		defaultMetaIndex,
		typeName,
		repo.cli.Delete.WithRefresh("true"),
		repo.cli.Delete.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	if res.IsError() && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}
	res.Body.Close()

	res, err = repo.cli.Indices.Delete(
		[]string{typeName},
		repo.cli.Indices.Delete.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	if res.IsError() && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}
	res.Body.Close()

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

type typeResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []struct {
			Index  string `json:"_index"`
			Source struct {
				Name string `json:"name"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
