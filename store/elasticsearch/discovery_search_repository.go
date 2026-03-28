package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
	boolQ := newBoolQuery()

	buildTextQuery(boolQ, cfg)
	buildFilterTermQueries(boolQ, cfg.Filters)
	buildMustMatchQueries(boolQ, cfg)

	query := buildFunctionScoreQuery(boolQ, cfg)

	searchBody := map[string]interface{}{
		"query":     query,
		"min_score": defaultMinScore,
	}

	if cfg.Flags.EnableHighlight {
		searchBody["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"*": map[string]interface{}{},
			},
		}
	}

	payload, err := json.Marshal(searchBody)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(string(payload)), nil
}

func (repo *DiscoveryRepository) buildSuggestQuery(ctx context.Context, cfg asset.SearchConfig) (io.Reader, error) {
	searchBody := map[string]interface{}{
		"suggest": map[string]interface{}{
			suggesterName: map[string]interface{}{
				"prefix": cfg.Text,
				"completion": map[string]interface{}{
					"field":           "name.suggest",
					"skip_duplicates": true,
					"size":            5,
				},
			},
		},
	}

	payload, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("error building reader %w", err)
	}

	return strings.NewReader(string(payload)), nil
}

// boolQuery is a JSON map builder for ES bool queries
type boolQuery struct {
	should             []interface{}
	filter             []interface{}
	must               []interface{}
	minimumShouldMatch string
}

func newBoolQuery() *boolQuery {
	return &boolQuery{}
}

func (q *boolQuery) toMap() map[string]interface{} {
	m := map[string]interface{}{}
	if len(q.should) > 0 {
		m["should"] = q.should
	}
	if len(q.filter) > 0 {
		m["filter"] = q.filter
	}
	if len(q.must) > 0 {
		m["must"] = q.must
	}
	if q.minimumShouldMatch != "" {
		m["minimum_should_match"] = q.minimumShouldMatch
	}
	return map[string]interface{}{"bool": m}
}

func buildTextQuery(boolQ *boolQuery, cfg asset.SearchConfig) {
	text := strings.TrimSpace(cfg.Text)
	if text == "" {
		boolQ.should = append(boolQ.should, map[string]interface{}{"match_all": map[string]interface{}{}})
		return
	}

	boostedFields := []string{
		"urn^10",
		"name^5",
	}

	// Phrase match for highest relevance
	boolQ.should = append(boolQ.should, map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  text,
			"fields": boostedFields,
			"type":   "phrase",
		},
	})

	// Multi match with AND operator
	boolQ.should = append(boolQ.should, map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":    text,
			"fields":   boostedFields,
			"operator": "and",
		},
	})

	// Standard multi match
	boolQ.should = append(boolQ.should, map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  text,
			"fields": boostedFields,
		},
	})

	if !cfg.Flags.DisableFuzzy {
		// Fuzzy match on boosted fields
		boolQ.should = append(boolQ.should, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     text,
				"fields":    boostedFields,
				"fuzziness": "AUTO",
			},
		})

		// Fuzzy match on all fields
		boolQ.should = append(boolQ.should, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     text,
				"fuzziness": "AUTO",
			},
		})
	}

	boolQ.minimumShouldMatch = "1"
}

func buildMustMatchQueries(boolQ *boolQuery, cfg asset.SearchConfig) {
	if len(cfg.Queries) == 0 {
		return
	}

	for field, value := range cfg.Queries {
		matchQ := map[string]interface{}{
			"query": value,
		}
		if !cfg.Flags.DisableFuzzy {
			matchQ["fuzziness"] = "AUTO"
		}
		boolQ.must = append(boolQ.must, map[string]interface{}{
			"match": map[string]interface{}{
				field: matchQ,
			},
		})
	}
}

func buildFilterTermQueries(boolQ *boolQuery, filters map[string][]string) {
	if len(filters) == 0 {
		return
	}

	for key, rawValues := range filters {
		if len(rawValues) < 1 {
			continue
		}

		keywordKey := fmt.Sprintf("%s.keyword", key)
		if len(rawValues) == 1 {
			boolQ.filter = append(boolQ.filter, map[string]interface{}{
				"term": map[string]interface{}{
					keywordKey: rawValues[0],
				},
			})
		} else {
			boolQ.filter = append(boolQ.filter, map[string]interface{}{
				"terms": map[string]interface{}{
					keywordKey: rawValues,
				},
			})
		}
	}
}

func buildFunctionScoreQuery(boolQ *boolQuery, cfg asset.SearchConfig) interface{} {
	text := strings.TrimSpace(cfg.Text)

	// Add exact match boost
	if text != "" {
		boolQ.should = append(boolQ.should, map[string]interface{}{
			"term": map[string]interface{}{
				"name.keyword": map[string]interface{}{
					"value": text,
					"boost": 100,
				},
			},
		})
	}

	queryMap := boolQ.toMap()

	if cfg.RankBy == "" {
		return queryMap
	}

	return map[string]interface{}{
		"function_score": map[string]interface{}{
			"query":      queryMap,
			"score_mode": defaultFunctionScoreQueryScoreMode,
			"functions": []interface{}{
				map[string]interface{}{
					"field_value_factor": map[string]interface{}{
						"field":    cfg.RankBy,
						"modifier": "log1p",
						"missing":  1.0,
					},
					"weight": 1.0,
				},
			},
		},
	}
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

const defaultGroupsSize = 10

// GroupAssets groups assets by specified fields using ES composite aggregation
func (repo *DiscoveryRepository) GroupAssets(ctx context.Context, cfg asset.GroupConfig) ([]asset.GroupResult, error) {
	if cfg.Namespace == nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("namespace cannot be empty")}
	}

	size := cfg.Size
	if size <= 0 {
		size = defaultGroupsSize
	}

	// Build composite aggregation sources from group-by fields
	sources := make([]map[string]interface{}, 0, len(cfg.GroupBy))
	for _, field := range cfg.GroupBy {
		sources = append(sources, map[string]interface{}{
			field: map[string]interface{}{
				"terms": map[string]interface{}{
					"field": fmt.Sprintf("%s.keyword", field),
				},
			},
		})
	}

	// Build filter query
	boolQ := newBoolQuery()
	for _, field := range cfg.GroupBy {
		boolQ.filter = append(boolQ.filter, map[string]interface{}{
			"exists": map[string]interface{}{
				"field": fmt.Sprintf("%s.keyword", field),
			},
		})
	}
	buildFilterTermQueries(boolQ, cfg.Filters)

	includedFields := defaultIncludedFields
	if len(cfg.IncludeFields) > 0 {
		includedFields = cfg.IncludeFields
	}

	payload := map[string]interface{}{
		"size":  0,
		"query": boolQ.toMap(),
		"aggs": map[string]interface{}{
			"group_result": map[string]interface{}{
				"composite": map[string]interface{}{
					"size":    size,
					"sources": sources,
				},
				"aggs": map[string]interface{}{
					"top_assets": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size":    size,
							"_source": includedFields,
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("error encoding query: %w", err)}
	}

	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(strings.NewReader(string(body))),
		repo.cli.client.Search.WithIndex(BuildAliasNameFromNamespace(cfg.Namespace)),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("error executing group query: %w", err)}
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("error response from elasticsearch: %s", errorReasonFromResponse(res))}
	}

	var response groupResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, asset.DiscoveryError{Err: fmt.Errorf("error decoding group response: %w", err)}
	}

	return repo.toGroupResults(response, cfg.GroupBy), nil
}

type groupResponse struct {
	Aggregations struct {
		GroupResult struct {
			Buckets []groupBucket `json:"buckets"`
		} `json:"group_result"`
	} `json:"aggregations"`
}

type groupBucket struct {
	Key       map[string]string `json:"key"`
	DocCount  int               `json:"doc_count"`
	TopAssets struct {
		Hits struct {
			Hits []searchHit `json:"hits"`
		} `json:"hits"`
	} `json:"top_assets"`
}

func (repo *DiscoveryRepository) toGroupResults(response groupResponse, groupBy []string) []asset.GroupResult {
	var results []asset.GroupResult
	for _, bucket := range response.Aggregations.GroupResult.Buckets {
		var fields []asset.GroupField
		for _, key := range groupBy {
			fields = append(fields, asset.GroupField{
				Key:   key,
				Value: bucket.Key[key],
			})
		}

		var assets []asset.Asset
		for _, hit := range bucket.TopAssets.Hits.Hits {
			r := hit.Source
			assets = append(assets, asset.Asset{
				ID:          r.ID,
				URN:         r.URN,
				Type:        r.Type,
				Name:        r.Name,
				Service:     r.Service,
				Description: r.Description,
				Data:        r.Data,
				Labels:      r.Labels,
			})
		}

		results = append(results, asset.GroupResult{
			Fields: fields,
			Assets: assets,
		})
	}
	return results
}
