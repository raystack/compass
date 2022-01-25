package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/discovery"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

const (
	defaultMaxResults                  = 200
	defaultMinScore                    = 0.01
	defaultFunctionScoreQueryScoreMode = "sum"
	suggesterName                      = "name-phrase-suggest"
)

type SearcherConfig struct {
	Client *elasticsearch.Client
}

// Searcher is an implementation of record.RecordSearcher
type Searcher struct {
	cli              *elasticsearch.Client
	typeWhiteList    []string
	typeWhiteListSet map[string]bool
}

// NewSearcher creates a new instance of Searcher
// You can optionally specify a list of type names to whitelist for search
// If the white list is nil (or has zero length), then the search will be run
// on all types. This can be further restricted by FilterConfig.TypeWhiteList
// in Search()
func NewSearcher(config SearcherConfig) (*Searcher, error) {
	return &Searcher{
		cli: config.Client,
	}, nil
}

// Search the record store
// Note that Searcher accepts 2 different forms of type white list,
// depending on how it is passed
// (1) when passed to NewSearcher, this is called the "Global White List" or GL for short
// (2) when passed to Search() as models.SearchConfig.TypeWhiteList, it's called "Local White List" or LL
// GL dictates the superset of all type types that should searched, while LL can only be a subset.
// To demonstrate:
// GL : {A, B, C}
// LL : {C, D}
// Entities searched : {C}
// GL specified that search can only be done for {A, B, C} types, while LL requested
// the search for {C, D} types. Since {D} doesn't belong to GL's set, it won't be searched
func (sr *Searcher) Search(ctx context.Context, cfg discovery.SearchConfig) (results []discovery.SearchResult, err error) {
	if strings.TrimSpace(cfg.Text) == "" {
		err = errors.New("search text cannot be empty")
		return
	}
	indices := sr.buildIndices(cfg)

	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	query, err := sr.buildQuery(ctx, cfg, indices)
	if err != nil {
		err = fmt.Errorf("error building query %w", err)
		return
	}

	res, err := sr.cli.Search(
		sr.cli.Search.WithBody(query),
		sr.cli.Search.WithIndex(indices...),
		sr.cli.Search.WithSize(maxResults),
		sr.cli.Search.WithIgnoreUnavailable(true),
		sr.cli.Search.WithContext(ctx),
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

	results = sr.toSearchResults(response.Hits.Hits)
	return
}

func (sr *Searcher) Suggest(ctx context.Context, config discovery.SearchConfig) (results []string, err error) {
	maxResults := config.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}

	indices := sr.buildIndices(config)
	query, err := sr.buildSuggestQuery(ctx, config, indices)
	if err != nil {
		err = fmt.Errorf("error building query: %s", err)
		return
	}
	res, err := sr.cli.Search(
		sr.cli.Search.WithBody(query),
		sr.cli.Search.WithIndex(indices...),
		sr.cli.Search.WithSize(maxResults),
		sr.cli.Search.WithIgnoreUnavailable(true),
		sr.cli.Search.WithContext(ctx),
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
	results, err = sr.toSuggestions(response)
	if err != nil {
		err = fmt.Errorf("error mapping response to suggestion %w", err)
	}

	return
}

func (sr *Searcher) buildIndices(cfg discovery.SearchConfig) []string {
	hasGL := len(sr.typeWhiteList) > 0
	hasLL := len(cfg.TypeWhiteList) > 0
	switch {
	case hasGL && hasLL:
		var indices []string
		for _, idx := range cfg.TypeWhiteList {
			if sr.typeWhiteListSet[idx] {
				indices = append(indices, idx)
			}
		}
		return indices
	case hasGL || hasLL:
		return anyValidStringSlice(cfg.TypeWhiteList, sr.typeWhiteList)
	default:
		return []string{}
	}
}

func anyValidStringSlice(slices ...[]string) []string {
	for _, slice := range slices {
		if len(slice) > 0 {
			return slice
		}
	}
	return nil
}

func (sr *Searcher) buildQuery(ctx context.Context, cfg discovery.SearchConfig, indices []string) (io.Reader, error) {
	var query elastic.Query

	query = sr.buildTextQuery(ctx, cfg.Text)
	query = sr.buildFilterTermQueries(query, cfg.Filters)
	query = sr.buildFilterMatchQueries(ctx, query, cfg.Queries)
	query = sr.buildFunctionScoreQuery(query, cfg.RankBy)

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

func (sr *Searcher) buildSuggestQuery(ctx context.Context, cfg discovery.SearchConfig, indices []string) (io.Reader, error) {
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

func (sr *Searcher) buildTextQuery(ctx context.Context, text string) elastic.Query {
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

func (sr *Searcher) buildFilterMatchQueries(ctx context.Context, query elastic.Query, queries map[string]string) elastic.Query {
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

func (sr *Searcher) buildFilterTermQueries(query elastic.Query, filters map[string][]string) elastic.Query {
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

func (sr *Searcher) buildFunctionScoreQuery(query elastic.Query, rankBy string) elastic.Query {
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

func (sr *Searcher) toSearchResults(hits []searchHit) []discovery.SearchResult {
	results := []discovery.SearchResult{}
	for _, hit := range hits {
		r := hit.Source
		results = append(results, discovery.SearchResult{
			Type:        hit.Index,
			ID:          r.Urn,
			Description: r.Description,
			Title:       r.Name,
			Service:     r.Service,
			Labels:      r.Labels,
		})
	}
	return results
}

func (sr *Searcher) toSuggestions(response searchResponse) (results []string, err error) {
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
