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

	// we'll reuse record repositories' scrolling capabilities
	// to obtain types, instead of reimplementing it.
	// This will slow down this operation a bit (due to the JSON conversions necessary)
	// but is a very efficient trade-off considering we don't have to re-implement
	// scrolling. Or we could generalise the scrolling operation on elasticsearch response
	// and then both of them could use it.
	recordRepo := RecordV1Repository{
		cli:        repo.cli,
		recordType: models.Type{Name: "meta"},
	}

	rawEntities, err := recordRepo.GetAll(ctx, models.RecordV1Filter{})
	if err != nil {
		return nil, err
	}

	types := []models.Type{}
	for _, rawType := range rawEntities {
		var serialised = new(bytes.Buffer)
		if err := json.NewEncoder(serialised).Encode(rawType); err != nil {
			return nil, fmt.Errorf("internal: error serialising record to JSON: %w", err)
		}
		var recordType models.Type
		if err := json.NewDecoder(serialised).Decode(&recordType); err != nil {
			return nil, fmt.Errorf("internal: error deserialising JSON to type: %w", err)
		}
		types = append(types, recordType)
	}
	return types, nil
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
