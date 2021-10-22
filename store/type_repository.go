package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/models"
)

const (
	// name of the metadata index
	defaultMetaIndex = "meta"

	// name of the search index
	defaultSearchIndex = "universe"
)

// used as body to create index requests
// aliases the index to defaultSearchIndex
// and sets up the camelcase analyzer
var indexSettingsTemplate = `{
		"mappings": %s,
		"aliases": {
			%q: {}
		},
		"settings": {
			"analysis": {
				"analyzer": {
					"default": {
						"type": "pattern",
						"pattern": "([^\\p{L}\\d]+)|(?<=\\D)(?=\\d)|(?<=\\d)(?=\\D)|(?<=[\\p{L}&&[^\\p{Lu}]])(?=\\p{Lu})|(?<=\\p{Lu})(?=\\p{Lu}[\\p{L}&&[^\\p{Lu}]])"
					}
				}
			}
		}
	}`

func createIndexSettings(recordType models.Type) (string, error) {
	mappings, err := createIndexMapping(recordType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(indexSettingsTemplate, mappings, defaultSearchIndex), nil
}

func isReservedName(name string) bool {
	name = strings.ToLower(name)
	return name == defaultMetaIndex || name == defaultSearchIndex
}

// TypeRepository is an implementation of models.TypeRepository
// that uses elasticsearch as a backing store
type TypeRepository struct {
	cli *elasticsearch.Client
}

func (repo *TypeRepository) addTypeToMetaIdx(ctx context.Context, recordType models.Type) error {
	raw := bytes.NewBuffer(nil)
	err := json.NewEncoder(raw).Encode(recordType)
	if err != nil {
		return fmt.Errorf("error encoding type: %w", err)
	}

	res, err := repo.cli.Index(
		defaultMetaIndex,
		raw,
		repo.cli.Index.WithDocumentID(recordType.Name),
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

func (repo *TypeRepository) createIdx(ctx context.Context, recordType models.Type) error {
	indexSettings, err := createIndexSettings(recordType)
	if err != nil {
		return fmt.Errorf("error building index settings: %v", err)
	}
	res, err := repo.cli.Indices.Create(
		recordType.Name,
		repo.cli.Indices.Create.WithBody(strings.NewReader(indexSettings)),
		repo.cli.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error creating index %q: %s", recordType.Name, errorReasonFromResponse(res))
	}
	return nil
}

func (repo *TypeRepository) updateIdx(ctx context.Context, recordType models.Type) error {
	mappings, err := createIndexMapping(recordType)
	if err != nil {
		return fmt.Errorf("updateIdx: %v", err)
	}
	res, err := repo.cli.Indices.PutMapping(
		strings.NewReader(mappings),
		repo.cli.Indices.PutMapping.WithIndex(recordType.Name),
		repo.cli.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error updating index %q: %s", recordType.Name, errorReasonFromResponse(res))
	}
	return nil
}

func (repo *TypeRepository) CreateOrReplace(ctx context.Context, recordType models.Type) error {
	if isReservedName(recordType.Name) {
		return models.ErrReservedTypeName{TypeName: recordType.Name}
	}

	// checking for the existence of index before adding the metadata
	// entry, because if this operation fails, we don't have to do a rollback
	// for the addTypeToMetaIdx()
	idxExists, err := indexExists(ctx, repo.cli, recordType.Name)
	if err != nil {
		return err
	}

	// save the type information to meta index
	if err := repo.addTypeToMetaIdx(ctx, recordType); err != nil {
		return err
	}

	// update/create the index
	if idxExists {
		err = repo.updateIdx(ctx, recordType)
	} else {
		err = repo.createIdx(ctx, recordType)
	}
	if err != nil {
		return err
	}

	return nil
}

func (repo *TypeRepository) GetByName(ctx context.Context, name string) (models.Type, error) {
	res, err := repo.cli.Get(
		defaultMetaIndex,
		name,
		repo.cli.Get.WithRefresh(true),
		repo.cli.Get.WithContext(ctx),
	)
	if err != nil {
		return models.Type{}, elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return models.Type{}, models.ErrNoSuchType{TypeName: name}
		}
		return models.Type{}, fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}

	var response = struct {
		Source models.Type `json:"_source"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return models.Type{}, fmt.Errorf("error parsing elasticsearch response: %w", err)
	}
	return response.Source, nil
}

func (repo *TypeRepository) getAllQuery() io.Reader {
	return strings.NewReader(`{"query":{"match_all":{}}}`)
}

func (repo *TypeRepository) GetAll(ctx context.Context) ([]models.Type, error) {
	body := strings.NewReader(`{"query":{"match_all":{}}}`)
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
		var types []models.Type
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

func (repo *TypeRepository) scrollRecord(ctx context.Context, scrollID string) ([]models.Type, string, error) {
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

func (repo *TypeRepository) toTypeList(response typeResponse) []models.Type {
	types := []models.Type{}
	for _, hit := range response.Hits.Hits {
		types = append(types, hit.Source)
	}

	return types
}

func (repo *TypeRepository) Delete(ctx context.Context, typeName string) error {
	if isReservedName(typeName) {
		return models.ErrReservedTypeName{TypeName: typeName}
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

func NewTypeRepository(cli *elasticsearch.Client) *TypeRepository {
	return &TypeRepository{
		cli: cli,
	}
}

type typeResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []struct {
			Index  string      `json:"_index"`
			Source models.Type `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
