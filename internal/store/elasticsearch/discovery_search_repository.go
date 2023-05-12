package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/goto/compass/core/asset"
	"github.com/olivere/elastic/v7"
)

const (
	defaultMaxResults                  = 200
	defaultGroupsSize                  = 10
	defaultMinScore                    = 0.01
	defaultFunctionScoreQueryScoreMode = "sum"
	suggesterName                      = "name-phrase-suggest"
)

var returnedAssetFieldsResult = []string{"id", "urn", "type", "service", "name", "description", "data", "labels", "created_at", "updated_at"}

// Search the asset store
func (repo *DiscoveryRepository) Search(ctx context.Context, cfg asset.SearchConfig) (results []asset.SearchResult, err error) {
	if strings.TrimSpace(cfg.Text) == "" {
		err = asset.DiscoveryError{Err: errors.New("search text cannot be empty")}
		return
	}
	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	query, err := repo.buildQuery(cfg)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error building query %w", err)}
		return
	}

	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(defaultSearchIndex),
		repo.cli.client.Search.WithSize(maxResults),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithSourceIncludes(returnedAssetFieldsResult...),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error executing search %w", err)}
		return
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error decoding search response %w", err)}
		return
	}

	results = repo.toSearchResults(response.Hits.Hits)
	return
}

func (repo *DiscoveryRepository) Suggest(ctx context.Context, config asset.SearchConfig) (results []string, err error) {
	maxResults := config.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}

	query, err := repo.buildSuggestQuery(config)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error building query: %s", err)}
		return
	}
	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(defaultSearchIndex),
		repo.cli.client.Search.WithSize(maxResults),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error executing search %w", err)}
		return
	}
	if res.IsError() {
		err = asset.DiscoveryError{Err: fmt.Errorf("error when searching %s", errorReasonFromResponse(res))}
		return
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error decoding search response %w", err)}
		return
	}
	results, err = repo.toSuggestions(response)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error mapping response to suggestion %w", err)}
	}

	return
}

func (repo *DiscoveryRepository) buildQuery(cfg asset.SearchConfig) (io.Reader, error) {
	var query elastic.Query

	query = repo.buildTextQuery(cfg.Text)
	if filterQueries, ok := repo.buildFilterTermQueries(cfg.Filters); ok {
		query = elastic.NewBoolQuery().Should(query).Filter(filterQueries...)
	}
	query = repo.buildFilterMatchQueries(query, cfg.Queries)
	query = repo.buildFunctionScoreQuery(query, cfg.RankBy)

	src, err := query.Source()
	if err != nil {
		return nil, err
	}

	payload := new(bytes.Buffer)
	q := &searchQuery{
		MinScore: defaultMinScore,
		Query:    src,
	}
	return payload, json.NewEncoder(payload).Encode(q)
}

func (repo *DiscoveryRepository) buildSuggestQuery(cfg asset.SearchConfig) (io.Reader, error) {
	suggester := elastic.NewCompletionSuggester(suggesterName).
		Field("name.suggest").
		SkipDuplicates(true).
		Size(5).
		Text(cfg.Text)
	src, err := elastic.NewSearchSource().
		Suggester(suggester).
		Source()
	if err != nil {
		return nil, fmt.Errorf("error building search source %w", err)
	}

	payload := new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(src)
	if err != nil {
		return payload, fmt.Errorf("error building reader %w", err)
	}

	return payload, err
}

func (repo *DiscoveryRepository) buildTextQuery(text string) elastic.Query {
	boostedFields := []string{
		"urn^10",
		"name^5",
	}

	return elastic.NewBoolQuery().
		Should(
			elastic.
				NewMultiMatchQuery(
					text,
					boostedFields...,
				),
			elastic.
				NewMultiMatchQuery(
					text,
					boostedFields...,
				).
				Fuzziness("AUTO"),
			elastic.
				NewMultiMatchQuery(
					text,
				).
				Fuzziness("AUTO"),
		)
}

func (repo *DiscoveryRepository) buildFilterMatchQueries(query elastic.Query, queries map[string]string) elastic.Query {
	if len(queries) == 0 {
		return query
	}

	esQueries := []elastic.Query{}
	for field, value := range queries {
		esQueries = append(esQueries,
			elastic.
				NewMatchQuery(field, value).
				Fuzziness("AUTO"))
	}

	return elastic.NewBoolQuery().
		Should(query).
		Filter(esQueries...)
}

//
func (repo *DiscoveryRepository) buildFilterTermQueries(filters map[string][]string) ([]elastic.Query, bool) {
	if len(filters) == 0 {
		return nil, false
	}

	var filterQueries []elastic.Query
	for key, rawValues := range filters {
		if len(rawValues) < 1 {
			continue
		}

		var values []interface{}
		for _, rawVal := range rawValues {
			values = append(values, rawVal)
		}

		key := fmt.Sprintf("%s.keyword", key)
		filterQueries = append(
			filterQueries,
			elastic.NewTermsQuery(key, values...),
		)
	}

	return filterQueries, true
}

func (repo *DiscoveryRepository) buildFilterExistsQueries(fields []string) ([]elastic.Query, bool) {
	if len(fields) == 0 {
		return nil, false
	}

	var filterQueries []elastic.Query
	for _, field := range fields {
		filterQueries = append(
			filterQueries,
			elastic.NewExistsQuery(fmt.Sprintf("%s.keyword", field)),
		)
	}

	return filterQueries, true
}

func (repo *DiscoveryRepository) buildFunctionScoreQuery(query elastic.Query, rankBy string) elastic.Query {
	if rankBy == "" {
		return query
	}

	factorFunc := elastic.NewFieldValueFactorFunction().
		Field(rankBy).
		Modifier("log1p").
		Missing(1.0).
		Weight(1.0)

	fsQuery := elastic.NewFunctionScoreQuery().
		ScoreMode(defaultFunctionScoreQueryScoreMode).
		AddScoreFunc(factorFunc).
		Query(query)

	return fsQuery
}

func (repo *DiscoveryRepository) toSearchResults(hits []searchHit) []asset.SearchResult {
	results := []asset.SearchResult{}
	for _, hit := range hits {
		r := hit.Source
		id := r.ID
		if id == "" { // this is for backward compatibility for asset without ID
			id = r.URN
		}
		results = append(results, asset.SearchResult{
			Type:        r.Type.String(),
			ID:          id,
			URN:         r.URN,
			Description: r.Description,
			Title:       r.Name,
			Service:     r.Service,
			Labels:      r.Labels,
			Data:        r.Data,
		})
	}
	return results
}

func (repo *DiscoveryRepository) toSuggestions(response searchResponse) (results []string, err error) {
	suggests, exists := response.Suggest[suggesterName]
	if !exists {
		err = errors.New("suggester key does not exist")
		return
	}
	results = []string{}
	for _, s := range suggests {
		for _, option := range s.Options {
			results = append(results, option.Text)
		}
	}

	return
}
func (repo *DiscoveryRepository) GroupAssets(ctx context.Context, cfg asset.GroupConfig) ([]asset.GroupResult, error) {
	if len(cfg.GroupBy) == 0 || cfg.GroupBy[0] == "" {
		err := asset.DiscoveryError{Op: "Group", Err: fmt.Errorf("group by field cannot be empty")}
		return nil, err
	}

	queryBody, err := repo.buildGroupQuery(cfg)
	if err != nil {
		return nil, asset.DiscoveryError{Op: "Group", Err: fmt.Errorf("build query: %w", err)}
	}
	repo.logger.Debug("group asset query", "query", queryBody, "config", cfg)

	search := repo.cli.client.Search
	res, err := search(
		search.WithFilterPath("aggregations"),
		search.WithBody(queryBody),
		search.WithIgnoreUnavailable(true),
		search.WithContext(ctx),
		search.WithSize(0),
	)

	if err != nil {
		err = asset.DiscoveryError{Op: "Group", Err: fmt.Errorf("execute group query: %w", err)}
		return nil, err
	}

	defer drainBody(res)
	if res.IsError() {
		return nil, asset.DiscoveryError{Op: "Group", Err: fmt.Errorf(errorReasonFromResponse(res))}
	}

	var response groupResponse

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = asset.DiscoveryError{Op: "Group", Err: fmt.Errorf("decode group response: %w", err)}
		return nil, err
	}
	results := repo.toGroupResults(response.Aggregations.CompositeAggregations.Buckets)
	return results, nil
}

func (repo *DiscoveryRepository) toGroupResults(buckets []aggregationBucket) []asset.GroupResult {
	groupResult := make([]asset.GroupResult, len(buckets))

	for i, bucket := range buckets {
		groupResult[i].Assets = make([]asset.Asset, len(bucket.Hits.Hits.Hits))
		for j, hit := range bucket.Hits.Hits.Hits {
			groupResult[i].Assets[j] = hit.Source
		}

		groupResult[i].Fields = make([]asset.GroupField, 0, len(bucket.Key))
		for key, value := range bucket.Key {
			groupResult[i].Fields = append(groupResult[i].Fields, asset.GroupField{Name: key, Value: value})
		}
	}
	return groupResult
}

func (repo *DiscoveryRepository) buildGroupQuery(cfg asset.GroupConfig) (*strings.Reader, error) {
	// This code takes care of creating filter term queries from the input filters mentioned in request.
	var filterQueries []elastic.Query
	if filterExistsQueries, ok := repo.buildFilterExistsQueries(cfg.GroupBy); ok {
		filterQueries = append(filterQueries, filterExistsQueries...)
	}
	if filterTermQueries, ok := repo.buildFilterTermQueries(cfg.Filters); ok {
		filterQueries = append(filterQueries, filterTermQueries...)
	}

	size := cfg.Size
	if size <= 0 {
		size = defaultGroupsSize
	}

	// By default, the groupby fields would be part of the response hence added them in the input included fields list.
	includedFields := cfg.GroupBy
	if len(cfg.IncludedFields) > 0 {
		includedFields = append(cfg.GroupBy, cfg.IncludedFields...)
	}

	compositeAggSources := make([]elastic.CompositeAggregationValuesSource, len(cfg.GroupBy))
	for i, group := range cfg.GroupBy {
		compositeAggSources[i] = elastic.NewCompositeAggregationTermsValuesSource(group).
			Field(fmt.Sprintf("%s.keyword", group))
	}

	// Hits aggregation helps to return the specific parts of _source in response.
	compositeAggregation := elastic.NewCompositeAggregation().Sources(compositeAggSources...).
		Size(size).
		SubAggregation("hits", elastic.NewTopHitsAggregation().
			SearchSource(elastic.NewSearchSource().
				FetchSourceContext(
					elastic.NewFetchSourceContext(true).
						Include(includedFields...),
				),
			))

	body, err := elastic.NewSearchRequest().
		Query(elastic.NewBoolQuery().Filter(filterQueries...)).
		Aggregation("composite-group", compositeAggregation).
		Body()
	if err != nil {
		return nil, fmt.Errorf("new search request body: %w", err)
	}

	return strings.NewReader(body), nil
}
