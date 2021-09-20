package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/models"
	"github.com/olivere/elastic/v7"
)

const (
	defaultScrollTimeout   = 30 * time.Second
	defaultScrollBatchSize = 1000
)

type getResponse struct {
	Source models.RecordV1 `json:"_source"`
}

// RecordV1Repository implements models.RecordV1Repository
// with elasticsearch as the backing store.
type RecordV1Repository struct {
	recordType models.Type
	cli        *elasticsearch.Client
}

func (repo *RecordV1Repository) CreateOrReplaceMany(ctx context.Context, records []models.RecordV1) error {
	idxExists, err := indexExists(ctx, repo.cli, repo.recordType.Name)
	if err != nil {
		return err
	}
	if !idxExists {
		return models.ErrNoSuchType{TypeName: repo.recordType.Name}
	}

	requestPayload, err := repo.createBulkInsertPayload(records)
	if err != nil {
		return fmt.Errorf("error serialising payload: %w", err)
	}
	res, err := repo.cli.Bulk(
		requestPayload,
		repo.cli.Bulk.WithRefresh("true"),
		repo.cli.Bulk.WithContext(ctx),
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

func (repo *RecordV1Repository) createBulkInsertPayload(records []models.RecordV1) (io.Reader, error) {
	payload := bytes.NewBuffer(nil)
	for _, record := range records {
		err := repo.writeInsertAction(payload, record)
		if err != nil {
			return nil, fmt.Errorf("createBulkInsertPayload: %w", err)
		}
		err = json.NewEncoder(payload).Encode(record)
		if err != nil {
			return nil, fmt.Errorf("error serialising record: %w", err)
		}
	}
	return payload, nil
}

func (repo *RecordV1Repository) writeInsertAction(w io.Writer, record models.RecordV1) error {
	id, ok := record[repo.recordType.Fields.ID].(string)
	if !ok {
		return fmt.Errorf("record must have a %q string field", repo.recordType.Fields.ID)
	}
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%q record field cannot be empty", repo.recordType.Fields.ID)
	}
	type obj map[string]interface{}
	action := obj{
		"index": obj{
			"_index": repo.recordType.Name,
			"_id":    id,
		},
	}
	return json.NewEncoder(w).Encode(action)
}

func (repo *RecordV1Repository) GetAll(ctx context.Context, filters models.RecordV1Filter) ([]models.RecordV1, error) {
	// XXX(Aman): we should probably think about result ordering, if the client
	// is going to slice the data for pagination. Does ES guarantee the result order?
	body, err := repo.getAllQuery(filters)
	if err != nil {
		return nil, fmt.Errorf("error building search query: %w", err)
	}

	resp, err := repo.cli.Search(
		repo.cli.Search.WithIndex(repo.recordType.Name),
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

	var response searchResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error decoding es response: %w", err)
	}
	var results = repo.toRecordV1List(response)
	var scrollID = response.ScrollID
	for {
		var nextResults []models.RecordV1
		nextResults, scrollID, err = repo.scrollRecordV1s(ctx, scrollID)
		if err != nil {
			return nil, fmt.Errorf("error scrolling results: %v", err)
		}
		if len(nextResults) == 0 {
			break
		}
		results = append(results, nextResults...)
	}
	return results, nil
}

func (repo *RecordV1Repository) scrollRecordV1s(ctx context.Context, scrollID string) ([]models.RecordV1, string, error) {
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
	var response searchResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	resp.Body.Close()
	if err != nil {
		return nil, "", fmt.Errorf("error decoding es response: %w", err)
	}
	return repo.toRecordV1List(response), response.ScrollID, nil
}

func (repo *RecordV1Repository) toRecordV1List(res searchResponse) (records []models.RecordV1) {
	for _, entry := range res.Hits.Hits {
		records = append(records, entry.Source)
	}
	return
}

func (repo *RecordV1Repository) getAllQuery(filters models.RecordV1Filter) (io.Reader, error) {
	if len(filters) == 0 {
		return repo.matchAllQuery(), nil
	}
	return repo.termsQuery(filters)
}

func (repo *RecordV1Repository) matchAllQuery() io.Reader {
	return strings.NewReader(`{"query":{"match_all":{}}}`)
}

func (repo *RecordV1Repository) termsQuery(filters models.RecordV1Filter) (io.Reader, error) {
	var termsQueries []elastic.Query
	for key, rawValues := range filters {
		var values []interface{}
		for _, val := range rawValues {
			values = append(values, val)
		}
		key = fmt.Sprintf("%s.keyword", key)
		termsQueries = append(termsQueries, elastic.NewTermsQuery(key, values...))
	}
	q := elastic.NewBoolQuery().Must(termsQueries...)
	src, err := q.Source()
	if err != nil {
		return nil, fmt.Errorf("error building terms query: %w", err)
	}
	raw := searchQuery{
		Query:    src,
		MinScore: defaultMinScore,
	}
	payload := bytes.NewBuffer(nil)
	return payload, json.NewEncoder(payload).Encode(raw)
}

func (repo *RecordV1Repository) GetByID(ctx context.Context, id string) (models.RecordV1, error) {
	res, err := repo.cli.Get(
		repo.recordType.Name,
		id,
		repo.cli.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing get: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return nil, models.ErrNoSuchRecordV1{RecordV1ID: id}
		}
		return nil, fmt.Errorf("error response from elasticsearch: %s", res.Status())
	}

	var response getResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}
	return response.Source, nil
}

func (repo *RecordV1Repository) Delete(ctx context.Context, id string) error {
	res, err := repo.cli.Delete(
		repo.recordType.Name,
		id,
		repo.cli.Delete.WithRefresh("true"),
		repo.cli.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting record: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return models.ErrNoSuchRecordV1{RecordV1ID: id}
		}
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}

	return nil
}

// RecordV1RepositoryFactory can be used to construct a RecordV1Repository
// for a certain type
type RecordV1RepositoryFactory struct {
	cli *elasticsearch.Client
}

func (factory *RecordV1RepositoryFactory) For(recordType models.Type) (models.RecordV1Repository, error) {
	return &RecordV1Repository{
		cli:        factory.cli,
		recordType: recordType.Normalise(),
	}, nil
}

func NewRecordV1RepositoryFactory(cli *elasticsearch.Client) *RecordV1RepositoryFactory {
	return &RecordV1RepositoryFactory{
		cli: cli,
	}
}
