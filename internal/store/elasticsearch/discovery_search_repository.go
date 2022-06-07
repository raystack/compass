package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/odpf/compass/core/asset"
	"github.com/olivere/elastic/v7"
)

const (
	defaultMaxResults                  = 200
	defaultMinScore                    = 0.01
	defaultFunctionScoreQueryScoreMode = "sum"
	allIndices                         = "_all"
	suggesterName                      = "name-phrase-suggest"
)

var returnedAssetFieldsResult = []string{"id", "urn", "type", "service", "name", "description", "data", "labels", "created_at", "updated_at"}

// Search the asset store
func (repo *DiscoveryRepository) Search(ctx context.Context, cfg asset.SearchConfig) (results []asset.SearchResult, err error) {
	if strings.TrimSpace(cfg.Text) == "" {
		err = errors.New("search text cannot be empty")
		return
	}
	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	query, err := repo.buildQuery(ctx, cfg)
	if err != nil {
		err = fmt.Errorf("error building query %w", err)
		return
	}

	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(allIndices),
		repo.cli.client.Search.WithSize(maxResults),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithSourceIncludes(returnedAssetFieldsResult...),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		err = fmt.Errorf("error executing search %w", err)
		return
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = fmt.Errorf("error decoding search response %w", err)
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

	query, err := repo.buildSuggestQuery(ctx, config)
	if err != nil {
		err = fmt.Errorf("error building query: %s", err)
		return
	}
	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(allIndices),
		repo.cli.client.Search.WithSize(maxResults),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		err = fmt.Errorf("error executing search %w", err)
		return
	}
	if res.IsError() {
		err = fmt.Errorf("error when searching %s", errorReasonFromResponse(res))
		return
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = fmt.Errorf("error decoding search response %w", err)
		return
	}
	results, err = repo.toSuggestions(response)
	if err != nil {
		err = fmt.Errorf("error mapping response to suggestion %w", err)
	}

	return
}

func (repo *DiscoveryRepository) buildQuery(ctx context.Context, cfg asset.SearchConfig) (io.Reader, error) {
	var query elastic.Query

	query = repo.buildTextQuery(ctx, cfg.Text)
	query = repo.buildFilterTermQueries(query, cfg.Filters)
	query = repo.buildFilterMatchQueries(ctx, query, cfg.Queries)
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

func (repo *DiscoveryRepository) buildSuggestQuery(ctx context.Context, cfg asset.SearchConfig) (io.Reader, error) {
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

func (repo *DiscoveryRepository) buildTextQuery(ctx context.Context, text string) elastic.Query {
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

func (repo *DiscoveryRepository) buildFilterMatchQueries(ctx context.Context, query elastic.Query, queries map[string]string) elastic.Query {
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

func (repo *DiscoveryRepository) buildFilterTermQueries(query elastic.Query, filters map[string][]string) elastic.Query {
	if len(filters) == 0 {
		return query
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

	newQuery := elastic.NewBoolQuery().
		Should(query).
		Filter(filterQueries...)

	return newQuery
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
