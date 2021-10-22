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
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/odpf/columbus/models"
	"github.com/olivere/elastic/v7"
)

const (
	defaultScrollTimeout   = 30 * time.Second
	defaultScrollBatchSize = 1000
)

type getResponse struct {
	Source models.RecordV2 `json:"_source"`
}

// RecordRepository implements models.RecordRepository
// with elasticsearch as the backing store.
type RecordRepository struct {
	recordType models.Type
	cli        *elasticsearch.Client
}

func (repo *RecordRepository) CreateOrReplaceMany(ctx context.Context, records []models.RecordV2) error {
	idxExists, err := indexExists(ctx, repo.cli, repo.recordType.Name)
	if err != nil {
		return err
	}
	if !idxExists {
		return models.ErrNoSuchType{TypeName: repo.recordType.Name}
	}

	requestPayload, err := repo.createBulkInsertPayloadV2(records)
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

func (repo *RecordRepository) createBulkInsertPayloadV2(records []models.RecordV2) (io.Reader, error) {
	payload := bytes.NewBuffer(nil)
	for _, record := range records {
		err := repo.writeInsertActionV2(payload, record)
		if err != nil {
			return nil, fmt.Errorf("createBulkInsertPayloadV2: %w", err)
		}
		err = json.NewEncoder(payload).Encode(record)
		if err != nil {
			return nil, fmt.Errorf("error serialising record: %w", err)
		}
	}
	return payload, nil
}

func (repo *RecordRepository) writeInsertActionV2(w io.Writer, record models.RecordV2) error {
	if strings.TrimSpace(record.Urn) == "" {
		return fmt.Errorf("URN record field cannot be empty")
	}
	type obj map[string]interface{}
	action := obj{
		"index": obj{
			"_index": repo.recordType.Name,
			"_id":    record.Urn,
		},
	}
	return json.NewEncoder(w).Encode(action)
}

func (repo *RecordRepository) GetAllIterator(ctx context.Context) (models.RecordIterator, error) {
	body, err := repo.getAllQuery(models.RecordFilter{})
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
	if resp.IsError() {
		return nil, fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(resp))
	}

	var response searchResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error decoding es response: %w", err)
	}
	var results = repo.toRecordList(response)
	it := recordIterator{
		resp:     resp,
		records:  results,
		scrollID: response.ScrollID,
		repo:     repo,
	}
	return &it, nil
}

func (repo *RecordRepository) GetAll(ctx context.Context, filters models.RecordFilter) ([]models.RecordV2, error) {
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
	var results = repo.toRecordList(response)
	var scrollID = response.ScrollID
	for {
		var nextResults []models.RecordV2
		nextResults, scrollID, err = repo.scrollRecordV2s(ctx, scrollID)
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

func (repo *RecordRepository) scrollRecordV2s(ctx context.Context, scrollID string) ([]models.RecordV2, string, error) {
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
	return repo.toRecordList(response), response.ScrollID, nil
}

func (repo *RecordRepository) toRecordList(res searchResponse) (records []models.RecordV2) {
	for _, entry := range res.Hits.Hits {
		record, _ := mapToV2(entry.Source)
		records = append(records, record)
	}
	return
}

func (repo *RecordRepository) getAllQuery(filters models.RecordFilter) (io.Reader, error) {
	if len(filters) == 0 {
		return repo.matchAllQuery(), nil
	}
	return repo.termsQuery(filters)
}

func (repo *RecordRepository) matchAllQuery() io.Reader {
	return strings.NewReader(`{"query":{"match_all":{}}}`)
}

func (repo *RecordRepository) termsQuery(filters models.RecordFilter) (io.Reader, error) {
	var termsQueries []elastic.Query

	var termsQueriesV1 []elastic.Query
	for key, rawValues := range filters {
		var values []interface{}
		for _, val := range rawValues {
			values = append(values, val)
		}
		key = fmt.Sprintf("%s.keyword", key)
		termsQueriesV1 = append(termsQueriesV1, elastic.NewTermsQuery(key, values...))
	}
	termsQueries = append(termsQueries, elastic.NewBoolQuery().Must(termsQueriesV1...))

	var termsQueriesV2 []elastic.Query
	for key, rawValues := range filters {
		var values []interface{}
		for _, val := range rawValues {
			values = append(values, val)
		}

		key := fmt.Sprintf("data.%s.keyword", key)
		termsQueriesV2 = append(termsQueriesV2, elastic.NewTermsQuery(key, values...))
	}
	termsQueries = append(termsQueries, elastic.NewBoolQuery().Must(termsQueriesV2...))

	q := elastic.NewBoolQuery().Should(termsQueries...)
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

func (repo *RecordRepository) GetByID(ctx context.Context, id string) (record models.RecordV2, err error) {
	res, err := repo.cli.Get(
		repo.recordType.Name,
		id,
		repo.cli.Get.WithContext(ctx),
	)
	if err != nil {
		err = fmt.Errorf("error executing get: %w", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			err = models.ErrNoSuchRecord{RecordID: id}
			return
		}
		err = fmt.Errorf("error response from elasticsearch: %s", res.Status())
		return
	}

	var response getResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = fmt.Errorf("error parsing response: %w", err)
		return
	}

	record = response.Source
	return
}

func (repo *RecordRepository) Delete(ctx context.Context, id string) error {
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
			return models.ErrNoSuchRecord{RecordID: id}
		}
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}

	return nil
}

// recordIterator is the internal implementation of models.RecordIterator by RecordRepository
type recordIterator struct {
	resp     *esapi.Response
	records  []models.RecordV2
	repo     *RecordRepository
	scrollID string
}

func (it *recordIterator) Scan() bool {
	return len(strings.TrimSpace(it.scrollID)) > 0
}

func (it *recordIterator) Next() (prev []models.RecordV2) {
	prev = it.records
	var err error
	it.records, it.scrollID, err = it.repo.scrollRecordV2s(context.Background(), it.scrollID)
	if err != nil {
		panic("error scrolling results:" + err.Error())
	}
	if len(it.records) == 0 {
		it.scrollID = ""
	}
	return
}

func (it *recordIterator) Close() error {
	return it.resp.Body.Close()
}

// RecordRepositoryFactory can be used to construct a RecordRepository
// for a certain type
type RecordRepositoryFactory struct {
	cli *elasticsearch.Client
}

func (factory *RecordRepositoryFactory) For(recordType models.Type) (models.RecordRepository, error) {
	return &RecordRepository{
		cli:        factory.cli,
		recordType: recordType.Normalise(),
	}, nil
}

func NewRecordRepositoryFactory(cli *elasticsearch.Client) *RecordRepositoryFactory {
	return &RecordRepositoryFactory{
		cli: cli,
	}
}
