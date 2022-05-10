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
	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/discovery"
	"github.com/olivere/elastic/v7"
)

// AssetRepository implements discovery.AssetRepository
// with elasticsearch as the backing store.
type AssetRepository struct {
	typeName string
	cli      *elasticsearch.Client
}

func (repo *AssetRepository) GetAll(ctx context.Context, cfg discovery.GetConfig) (assetList discovery.AssetList, err error) {
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
	var assets = repo.toAssetList(response)

	assetList.Data = assets
	assetList.Count = len(assets)
	assetList.Total = int(response.Hits.Total.Value)

	return
}

func (repo *AssetRepository) GetAllIterator(ctx context.Context) (discovery.AssetIterator, error) {
	body, err := repo.getAllQuery(discovery.AssetFilter{})
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
	var results = repo.toAssetList(response)
	it := assetIterator{
		resp:     resp,
		assets:   results,
		scrollID: response.ScrollID,
		repo:     repo,
	}
	return &it, nil
}

func (repo *AssetRepository) CreateOrReplaceMany(ctx context.Context, assets []asset.Asset) error {
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

func (repo *AssetRepository) GetByID(ctx context.Context, id string) (r asset.Asset, err error) {
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

func (repo *AssetRepository) Delete(ctx context.Context, id string) error {
	res, err := repo.cli.Delete(
		repo.typeName,
		url.PathEscape(id),
		repo.cli.Delete.WithRefresh("true"),
		repo.cli.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting asset: %w", err)
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

func (repo *AssetRepository) createBulkInsertPayload(assets []asset.Asset) (io.Reader, error) {
	payload := bytes.NewBuffer(nil)
	for _, ast := range assets {
		err := repo.writeInsertAction(payload, ast)
		if err != nil {
			return nil, fmt.Errorf("createBulkInsertPayload: %w", err)
		}
		err = json.NewEncoder(payload).Encode(ast)
		if err != nil {
			return nil, fmt.Errorf("error serialising asset: %w", err)
		}
	}
	return payload, nil
}

func (repo *AssetRepository) writeInsertAction(w io.Writer, ast asset.Asset) error {
	if strings.TrimSpace(ast.URN) == "" {
		return fmt.Errorf("URN asset field cannot be empty")
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

func (repo *AssetRepository) toAssetList(res searchResponse) []asset.Asset {
	assets := []asset.Asset{}
	for _, entry := range res.Hits.Hits {
		assets = append(assets, entry.Source)
	}
	return assets
}

func (repo *AssetRepository) getAllQuery(filters discovery.AssetFilter) (io.Reader, error) {
	if len(filters) == 0 {
		return repo.matchAllQuery(), nil
	}
	return repo.termsQuery(filters)
}

func (repo *AssetRepository) matchAllQuery() io.Reader {
	return strings.NewReader(`{"query":{"match_all":{}}}`)
}

func (repo *AssetRepository) termsQuery(filters discovery.AssetFilter) (io.Reader, error) {
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

func (repo *AssetRepository) scrollAssets(ctx context.Context, scrollID string) ([]asset.Asset, string, error) {
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
	return repo.toAssetList(response), response.ScrollID, nil
}

// assetIterator is the internal implementation of AssetIterator by AssetRepository
type assetIterator struct {
	resp     *esapi.Response
	assets   []asset.Asset
	repo     *AssetRepository
	scrollID string
}

func (it *assetIterator) Scan() bool {
	return len(strings.TrimSpace(it.scrollID)) > 0
}

func (it *assetIterator) Next() (prev []asset.Asset) {
	prev = it.assets
	var err error
	it.assets, it.scrollID, err = it.repo.scrollAssets(context.Background(), it.scrollID)
	if err != nil {
		panic("error scrolling results:" + err.Error())
	}
	if len(it.assets) == 0 {
		it.scrollID = ""
	}
	return
}

func (it *assetIterator) Close() error {
	return it.resp.Body.Close()
}

// AssetRepositoryFactory can be used to construct a AssetRepository
// for a certain type
type AssetRepositoryFactory struct {
	cli *elasticsearch.Client
}

func (factory *AssetRepositoryFactory) For(typeName string) (discovery.AssetRepository, error) {
	return &AssetRepository{
		cli:      factory.cli,
		typeName: typeName,
	}, nil
}

func NewAssetRepositoryFactory(cli *elasticsearch.Client) *AssetRepositoryFactory {
	return &AssetRepositoryFactory{
		cli: cli,
	}
}
