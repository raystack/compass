package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

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

// Search the asset store
func (repo *DiscoveryRepository) Search(ctx context.Context, cfg asset.SearchConfig) (results []asset.SearchResult, err error) {
	if strings.TrimSpace(cfg.Text) == "" {
		return nil, asset.DiscoveryError{Op: "Search", Err: errors.New("search text cannot be empty")}
	}
	var returnedAssetFieldsResult []string

	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	offset := cfg.Offset
	if offset < 0 {
		offset = 0
	}

	if len(cfg.IncludeFields) == 0 {
		returnedAssetFieldsResult = []string{
			"id", "urn", "type", "service", "name", "description", "data", "labels",
			"created_at", "updated_at",
		}
	} else {
		returnedAssetFieldsResult = cfg.IncludeFields
	}

	defer func(start time.Time) {
		const op = "search"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: "Search",
			start:       start,
			err:         err,
		})
	}(time.Now())

	query, err := buildQuery(cfg)
	if err != nil {
		return nil, asset.DiscoveryError{Op: "Search", Err: fmt.Errorf("build query: %w", err)}
	}

	search := repo.cli.client.Search
	res, err := search(
		search.WithBody(query),
		search.WithIndex(defaultSearchIndex),
		search.WithSize(maxResults),
		search.WithFrom(offset),
		search.WithIgnoreUnavailable(true),
		search.WithSourceIncludes(returnedAssetFieldsResult...),
		search.WithContext(ctx),
	)
	if err != nil {
		return nil, asset.DiscoveryError{Op: "Search", Err: fmt.Errorf("execute search: %w", err)}
	}
	if res.IsError() {
		code, reason := errorCodeAndReason(res)
		return nil, asset.DiscoveryError{
			Op:     "Search",
			ESCode: code,
			Err:    fmt.Errorf("execute search: %s", reason),
		}
	}

	var response searchResponse
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, asset.DiscoveryError{Op: "Search", Err: fmt.Errorf("decode search response: %w", err)}
	}

	return toSearchResults(response.Hits.Hits), nil
}

func (repo *DiscoveryRepository) GroupAssets(ctx context.Context, cfg asset.GroupConfig) (results []asset.GroupResult, err error) {
	if len(cfg.GroupBy) == 0 || cfg.GroupBy[0] == "" {
		err := asset.DiscoveryError{Op: "Group", Err: fmt.Errorf("group by field cannot be empty")}
		return nil, err
	}

	defer func(start time.Time) {
		const op = "search"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: "GroupAssets",
			start:       start,
			err:         err,
		})
	}(time.Now())

	queryBody, err := buildGroupQuery(cfg)
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
		return nil, asset.DiscoveryError{Op: "Group", Err: fmt.Errorf("execute group query: %w", err)}
	}

	defer drainBody(res)
	if res.IsError() {
		code, reason := errorCodeAndReason(res)
		return nil, asset.DiscoveryError{
			Op:     "Group",
			ESCode: code,
			Err:    errors.New(reason),
		}
	}

	var response groupResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, asset.DiscoveryError{Op: "Group", Err: fmt.Errorf("decode group response: %w", err)}
	}

	return toGroupResults(response.Aggregations.CompositeAggregations.Buckets), nil
}

func (repo *DiscoveryRepository) Suggest(ctx context.Context, config asset.SearchConfig) (results []string, err error) {
	maxResults := config.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}

	defer func(start time.Time) {
		const op = "search"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: "Suggest",
			start:       start,
			err:         err,
		})
	}(time.Now())

	query, err := buildSuggestQuery(config)
	if err != nil {
		return nil, asset.DiscoveryError{Op: "Suggest", Err: fmt.Errorf("build query: %w", err)}
	}

	search := repo.cli.client.Search
	res, err := search(
		search.WithBody(query),
		search.WithIndex(defaultSearchIndex),
		search.WithSize(maxResults),
		search.WithIgnoreUnavailable(true),
		search.WithContext(ctx),
	)
	if err != nil {
		return nil, asset.DiscoveryError{Op: "Suggest", Err: fmt.Errorf("execute search: %w", err)}
	}
	if res.IsError() {
		code, reason := errorCodeAndReason(res)
		return nil, asset.DiscoveryError{
			Op:     "Suggest",
			ESCode: code,
			Err:    fmt.Errorf("execute search: %s", reason),
		}
	}

	var response searchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, asset.DiscoveryError{Op: "Suggest", Err: fmt.Errorf("decode search response: %w", err)}
	}

	results, err = toSuggestions(response)
	if err != nil {
		return nil, asset.DiscoveryError{Op: "Suggest", Err: fmt.Errorf("map response to suggestion: %w", err)}
	}

	return results, nil
}

func buildQuery(cfg asset.SearchConfig) (io.Reader, error) {
	boolQuery := elastic.NewBoolQuery()
	buildTextQuery(boolQuery, cfg)
	buildFilterTermQueries(boolQuery, cfg.Filters)
	buildMustMatchQueries(boolQuery, cfg)
	query := buildFunctionScoreQuery(boolQuery, cfg.RankBy, cfg.Text)
	highlight := buildHighlightQuery(cfg)

	body, err := elastic.NewSearchRequest().
		Query(query).
		Highlight(highlight).
		MinScore(defaultMinScore).
		Body()
	if err != nil {
		return nil, fmt.Errorf("build query: new search request: %w", err)
	}

	return strings.NewReader(body), nil
}

func buildSuggestQuery(cfg asset.SearchConfig) (io.Reader, error) {
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

func buildTextQuery(q *elastic.BoolQuery, cfg asset.SearchConfig) {
	boostedFields := []string{"urn^10", "name^5"}
	q.Should(
		// Phrase query cannot have `FUZZINESS`
		elastic.NewMultiMatchQuery(cfg.Text, boostedFields...).
			Type("phrase"),
	)
	for _, mq := range []*elastic.MultiMatchQuery{
		elastic.NewMultiMatchQuery(cfg.Text, boostedFields...).
			Operator("and"),
		elastic.NewMultiMatchQuery(cfg.Text, boostedFields...),
		elastic.NewMultiMatchQuery(cfg.Text),
	} {
		if !cfg.Flags.DisableFuzzy {
			mq.Fuzziness("AUTO")
		}

		q.Should(mq)
	}
}

func buildMustMatchQueries(q *elastic.BoolQuery, cfg asset.SearchConfig) {
	if len(cfg.Queries) == 0 {
		return
	}

	for field, value := range cfg.Queries {
		if cfg.Flags.DisableFuzzy {
			q.Must(elastic.NewMatchQuery(field, value))
			continue
		}

		q.Must(elastic.NewMatchQuery(field, value).
			Fuzziness("AUTO"))
	}
}

func buildFilterTermQueries(q *elastic.BoolQuery, filters map[string][]string) {
	if len(filters) == 0 {
		return
	}

	for field, rawValues := range filters {
		if len(rawValues) < 1 {
			continue
		}

		key := fmt.Sprintf("%s.keyword", field)
		if len(rawValues) == 1 {
			q.Filter(elastic.NewTermQuery(key, rawValues[0]))
			continue
		}

		var values []interface{}
		for _, rawVal := range rawValues {
			values = append(values, rawVal)
		}
		q.Filter(elastic.NewTermsQuery(key, values...))
	}
}

func buildFilterExistsQueries(q *elastic.BoolQuery, fields []string) {
	if len(fields) == 0 {
		return
	}

	for _, field := range fields {
		q.Filter(elastic.NewExistsQuery(fmt.Sprintf("%s.keyword", field)))
	}
}

func buildFunctionScoreQuery(query elastic.Query, rankBy, text string) elastic.Query {
	// Added exact match term query here so that exact match gets higher priority.
	fsQuery := elastic.NewFunctionScoreQuery().
		Add(
			elastic.NewTermQuery("name.keyword", text),
			elastic.NewWeightFactorFunction(2),
		)

	if rankBy != "" {
		fsQuery.AddScoreFunc(
			elastic.NewFieldValueFactorFunction().
				Field(rankBy).
				Modifier("log1p").
				Missing(1.0).
				Weight(1.0),
		)
	}

	fsQuery.Query(query).ScoreMode(defaultFunctionScoreQueryScoreMode)
	return fsQuery
}

func buildHighlightQuery(cfg asset.SearchConfig) *elastic.Highlight {
	if cfg.Flags.EnableHighlight {
		return elastic.NewHighlight().Field("*")
	}
	return nil
}

func toSearchResults(hits []searchHit) []asset.SearchResult {
	results := make([]asset.SearchResult, len(hits))
	for i, hit := range hits {
		r := hit.Source
		id := r.ID
		if id == "" { // this is for backward compatibility for asset without ID
			id = r.URN
		}

		data := r.Data

		if data != nil && hit.HighLight != nil {
			data["_highlight"] = hit.HighLight
		} else if data == nil && hit.HighLight != nil {
			data = map[string]interface{}{
				"_highlight": hit.HighLight,
			}
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

func toSuggestions(response searchResponse) ([]string, error) {
	suggests, exists := response.Suggest[suggesterName]
	if !exists {
		return nil, errors.New("suggester key does not exist")
	}

	var results []string
	for _, s := range suggests {
		for _, option := range s.Options {
			results = append(results, option.Text)
		}
	}
	return results, nil
}

func toGroupResults(buckets []aggregationBucket) []asset.GroupResult {
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

func buildGroupQuery(cfg asset.GroupConfig) (*strings.Reader, error) {
	boolQuery := elastic.NewBoolQuery()
	// This code takes care of creating filter term queries from the input filters mentioned in request.
	buildFilterExistsQueries(boolQuery, cfg.GroupBy)
	buildFilterTermQueries(boolQuery, cfg.Filters)

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
		Query(boolQuery).
		Aggregation("composite-group", compositeAggregation).
		Body()
	if err != nil {
		return nil, fmt.Errorf("new search request body: %w", err)
	}

	return strings.NewReader(body), nil
}
