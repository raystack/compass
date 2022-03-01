package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/olivere/elastic/v7"
)

// RecordRepository implements discovery.RecordRepository
// with elasticsearch as the backing store.
type RecordRepository struct {
	typeName string
	cli      *elasticsearch.Client
}

func (repo *RecordRepository) GetAll(ctx context.Context, cfg discovery.GetConfig) (recordList discovery.RecordList, err error) {
	// XXX(Aman): we should probably think about result ordering, if the client
	// is going to slice the data for pagination. Does ES guarantee the result order?
	body, err := repo.getAllQuery(cfg.Filters)
	if err != nil {
		err = fmt.Errorf("error building search query: %w", err)
		return
	}
	size := cfg.Size
	if size == 0 {
		size = defaultGetSize
	}

	resp, err := repo.cli.Search(
		repo.cli.Search.WithIndex(repo.typeName),
		repo.cli.Search.WithBody(body),
		repo.cli.Search.WithFrom(cfg.From),
		repo.cli.Search.WithSize(size),
		repo.cli.Search.WithSort(defaultSortField),
		repo.cli.Search.WithContext(ctx),
	)
	if err != nil {
		err = fmt.Errorf("error executing search: %w", err)
		return
	}
	defer resp.Body.Close()
	if resp.IsError() {
		err = fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(resp))
		return
	}

	var response searchResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		err = fmt.Errorf("error decoding es response: %w", err)
		return
	}
	var assets = repo.toRecordList(response)

	recordList.Data = assets
	recordList.Count = len(assets)
	recordList.Total = int(response.Hits.Total.Value)

	return
}

func (repo *RecordRepository) GetAllIterator(ctx context.Context) (discovery.RecordIterator, error) {
	body, err := repo.getAllQuery(discovery.RecordFilter{})
	if err != nil {
		return nil, fmt.Errorf("error building search query: %w", err)
	}

	resp, err := repo.cli.Search(
		repo.cli.Search.WithIndex(repo.typeName),
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

func (repo *RecordRepository) CreateOrReplaceMany(ctx context.Context, assets []asset.Asset) error {
	requestPayload, err := repo.createBulkInsertPayload(assets)
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

func (repo *RecordRepository) GetByID(ctx context.Context, id string) (r asset.Asset, err error) {
	res, err := repo.cli.Get(
		repo.typeName,
		url.PathEscape(id),
		repo.cli.Get.WithContext(ctx),
	)
	if err != nil {
		err = fmt.Errorf("error executing get: %w", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			err = asset.NotFoundError{AssetID: id}
			return
		}
		err = fmt.Errorf("got %s response from elasticsearch: %s", res.Status(), res)
		return
	}

	var response searchHit
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = fmt.Errorf("error parsing response: %w", err)
		return
	}

	r = response.Source
	return
}

func (repo *RecordRepository) Delete(ctx context.Context, id string) error {
	res, err := repo.cli.Delete(
		repo.typeName,
		url.PathEscape(id),
		repo.cli.Delete.WithRefresh("true"),
		repo.cli.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting record: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return asset.NotFoundError{AssetID: id}
		}
		return fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))
	}

	return nil
}

func (repo *RecordRepository) createBulkInsertPayload(assets []asset.Asset) (io.Reader, error) {
	payload := bytes.NewBuffer(nil)
	for _, ast := range assets {
		err := repo.writeInsertAction(payload, ast)
		if err != nil {
			return nil, fmt.Errorf("createBulkInsertPayload: %w", err)
		}
		err = json.NewEncoder(payload).Encode(ast)
		if err != nil {
			return nil, fmt.Errorf("error serialising record: %w", err)
		}
	}
	return payload, nil
}

func (repo *RecordRepository) writeInsertAction(w io.Writer, ast asset.Asset) error {
	if strings.TrimSpace(ast.URN) == "" {
		return fmt.Errorf("URN record field cannot be empty")
	}
	type obj map[string]interface{}
	action := obj{
		"index": obj{
			"_index": repo.typeName,
			"_id":    ast.URN,
		},
	}
	return json.NewEncoder(w).Encode(action)
}

func (repo *RecordRepository) toRecordList(res searchResponse) []asset.Asset {
	assets := []asset.Asset{}
	for _, entry := range res.Hits.Hits {
		assets = append(assets, entry.Source)
	}
	return assets
}

func (repo *RecordRepository) getAllQuery(filters discovery.RecordFilter) (io.Reader, error) {
	if len(filters) == 0 {
		return repo.matchAllQuery(), nil
	}
	return repo.termsQuery(filters)
}

func (repo *RecordRepository) matchAllQuery() io.Reader {
	return strings.NewReader(`{"query":{"match_all":{}}}`)
}

func (repo *RecordRepository) termsQuery(filters discovery.RecordFilter) (io.Reader, error) {
	var termQueries []elastic.Query
	for key, rawValues := range filters {
		var values []interface{}
		for _, val := range rawValues {
			values = append(values, val)
		}

		key := fmt.Sprintf("%s.keyword", key)
		termQueries = append(termQueries, elastic.NewTermsQuery(key, values...))
	}
	boolQuery := elastic.NewBoolQuery().Must(termQueries...)
	q := elastic.NewBoolQuery().Should(boolQuery)
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

func (repo *RecordRepository) scrollRecords(ctx context.Context, scrollID string) ([]asset.Asset, string, error) {
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

// recordIterator is the internal implementation of record.RecordIterator by RecordRepository
type recordIterator struct {
	resp     *esapi.Response
	records  []asset.Asset
	repo     *RecordRepository
	scrollID string
}

func (it *recordIterator) Scan() bool {
	return len(strings.TrimSpace(it.scrollID)) > 0
}

func (it *recordIterator) Next() (prev []asset.Asset) {
	prev = it.records
	var err error
	it.records, it.scrollID, err = it.repo.scrollRecords(context.Background(), it.scrollID)
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

func (factory *RecordRepositoryFactory) For(typeName string) (discovery.RecordRepository, error) {
	return &RecordRepository{
		cli:      factory.cli,
		typeName: typeName,
	}, nil
}

func NewRecordRepositoryFactory(cli *elasticsearch.Client) *RecordRepositoryFactory {
	return &RecordRepositoryFactory{
		cli: cli,
	}
}
