package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/olivere/elastic/v7"
	"github.com/raystack/compass/core/asset"
)

const (
	defaultMaxResults                  = 200
	defaultMinScore                    = 0.01
	defaultFunctionScoreQueryScoreMode = "sum"
	suggesterName                      = "name-phrase-suggest"
)

var defaultIncludedFields = []string{"id", "namespace_id", "urn", "type", "service", "name", "description", "data", "labels", "created_at", "updated_at"}

// Search the asset store
func (repo *DiscoveryRepository) Search(ctx context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error) {
	if cfg.Namespace == nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("namespace cannot be empty")}
	}
	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	offset := cfg.Offset
	if offset < 0 {
		offset = 0
	}

	includedFields := defaultIncludedFields
	if len(cfg.IncludeFields) > 0 {
		includedFields = cfg.IncludeFields
	}

	query, err := repo.buildQuery(ctx, cfg)
	if err != nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("error building query %w", err)}
	}

	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(BuildAliasNameFromNamespace(cfg.Namespace)),
		repo.cli.client.Search.WithSize(maxResults),
		repo.cli.client.Search.WithFrom(offset),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithSourceIncludes(includedFields...),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("error executing search %w", err)}
	}

	defer res.Body.Close()
	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("error decoding search response %w", err)}
	}

	return repo.toSearchResults(response.Hits.Hits), nil
}

func (repo *DiscoveryRepository) Suggest(ctx context.Context, cfg asset.SearchConfig) (results []string, err error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}

	query, err := repo.buildSuggestQuery(ctx, cfg)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error building query: %s", err)}
		return
	}
	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(BuildAliasNameFromNamespace(cfg.Namespace)),
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

func (repo *DiscoveryRepository) buildQuery(ctx context.Context, cfg asset.SearchConfig) (io.Reader, error) {
	boolQuery := elastic.NewBoolQuery()

	repo.buildTextQuery(boolQuery, cfg)
	repo.buildFilterTermQueries(boolQuery, cfg.Filters)
	repo.buildMustMatchQueries(boolQuery, cfg)

	query := repo.buildFunctionScoreQuery(boolQuery, cfg)
	highlight := repo.buildHighlightQuery(cfg)

	searchSource := elastic.NewSearchSource().
		Query(query).
		MinScore(defaultMinScore)

	if highlight != nil {
		searchSource = searchSource.Highlight(highlight)
	}

	src, err := searchSource.Source()
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(payload)), nil
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

	payload, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("error building reader %w", err)
	}

	return strings.NewReader(string(payload)), err
}

func (repo *DiscoveryRepository) buildTextQuery(boolQuery *elastic.BoolQuery, cfg asset.SearchConfig) {
	text := strings.TrimSpace(cfg.Text)
	if text == "" {
		boolQuery.Should(elastic.NewMatchAllQuery())
		return
	}

	boostedFields := []string{
		"urn^10",
		"name^5",
	}

	// Phrase match for highest relevance
	boolQuery.Should(
		elastic.NewMultiMatchQuery(text, boostedFields...).
			Type("phrase"),
	)

	// Multi match with AND operator — all terms must match
	boolQuery.Should(
		elastic.NewMultiMatchQuery(text, boostedFields...).
			Operator("and"),
	)

	// Standard multi match without fuzziness
	boolQuery.Should(
		elastic.NewMultiMatchQuery(text, boostedFields...),
	)

	if !cfg.Flags.DisableFuzzy {
		// Fuzzy match on boosted fields
		boolQuery.Should(
			elastic.NewMultiMatchQuery(text, boostedFields...).
				Fuzziness("AUTO"),
		)

		// Fuzzy match on all fields
		boolQuery.Should(
			elastic.NewMultiMatchQuery(text).
				Fuzziness("AUTO"),
		)
	}

	boolQuery.MinimumShouldMatch("1")
}

func (repo *DiscoveryRepository) buildMustMatchQueries(boolQuery *elastic.BoolQuery, cfg asset.SearchConfig) {
	if len(cfg.Queries) == 0 {
		return
	}

	for field, value := range cfg.Queries {
		matchQuery := elastic.NewMatchQuery(field, value)
		if !cfg.Flags.DisableFuzzy {
			matchQuery = matchQuery.Fuzziness("AUTO")
		}
		boolQuery.Must(matchQuery)
	}
}

func (repo *DiscoveryRepository) buildFilterTermQueries(boolQuery *elastic.BoolQuery, filters map[string][]string) {
	if len(filters) == 0 {
		return
	}

	for key, rawValues := range filters {
		if len(rawValues) < 1 {
			continue
		}

		key := fmt.Sprintf("%s.keyword", key)
		if len(rawValues) == 1 {
			boolQuery.Filter(elastic.NewTermQuery(key, rawValues[0]))
		} else {
			var values []interface{}
			for _, rawVal := range rawValues {
				values = append(values, rawVal)
			}
			boolQuery.Filter(elastic.NewTermsQuery(key, values...))
		}
	}
}

func (repo *DiscoveryRepository) buildFunctionScoreQuery(query elastic.Query, cfg asset.SearchConfig) elastic.Query {
	text := strings.TrimSpace(cfg.Text)

	// Add exact match boost directly as a should clause with high boost
	if text != "" {
		if bq, ok := query.(*elastic.BoolQuery); ok {
			bq.Should(
				elastic.NewTermQuery("name.keyword", text).Boost(100),
			)
		}
	}

	if cfg.RankBy == "" {
		return query
	}

	factorFunc := elastic.NewFieldValueFactorFunction().
		Field(cfg.RankBy).
		Modifier("log1p").
		Missing(1.0).
		Weight(1.0)

	fsQuery := elastic.NewFunctionScoreQuery().
		ScoreMode(defaultFunctionScoreQueryScoreMode).
		AddScoreFunc(factorFunc).
		Query(query)

	return fsQuery
}

func (repo *DiscoveryRepository) buildHighlightQuery(cfg asset.SearchConfig) *elastic.Highlight {
	if !cfg.Flags.EnableHighlight {
		return nil
	}

	return elastic.NewHighlight().Field("*")
}

func (repo *DiscoveryRepository) toSearchResults(hits []searchHit) []asset.SearchResult {
	results := make([]asset.SearchResult, len(hits))
	for i, hit := range hits {
		r := hit.Source
		id := r.ID
		if id == "" { // this is for backward compatibility for asset without ID
			id = r.URN
		}

		data := r.Data
		if len(hit.HighLight) > 0 {
			if data == nil {
				data = make(map[string]interface{})
			}
			data["_highlight"] = hit.HighLight
		}

		results[i] = asset.SearchResult{
			Type:        r.Type.String(),
			ID:          id,
			URN:         r.URN,
			Description: r.Description,
			Title:       r.Name,
			Service:     r.Service,
			Labels:      r.Labels,
			Data:        data,
		}
	}
	return results
}

func (repo *DiscoveryRepository) toSuggestions(response searchResponse) (results []string, err error) {
	suggests, exists := response.Suggest[suggesterName]
	if !exists {
		err = fmt.Errorf("suggester key does not exist")
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
