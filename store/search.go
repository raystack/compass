package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/models"
	"github.com/olivere/elastic/v7"
)

var (
	defaultMaxResults = 200
	defaultMinScore   = 0.01
)

type SearcherConfig struct {
	Client              *elasticsearch.Client
	TypeRepo            models.TypeRepository
	TypeWhiteList       []string
	CachedTypesDuration int
}

// Searcher is an implementation of models.RecordV1Searcher
type Searcher struct {
	cli                 *elasticsearch.Client
	typeWhiteList       []string
	typeWhiteListSet    map[string]bool
	typeRepository      models.TypeRepository
	cachedTypes         []models.Type
	cachedTypesMap      map[string]models.Type
	cachedTypeExpiredOn time.Time
	cachedTypesDuration int
}

// NewSearcher creates a new instance of Searcher
// You can optionally specify a list of type names to whitelist for search
// If the white list is nil (or has zero length), then the search will be run
// on all types. This can be further restricted by FilterConfig.TypeWhiteList
// in Search()
func NewSearcher(config SearcherConfig) (*Searcher, error) {
	var whiteListSet = make(map[string]bool)
	for _, ent := range config.TypeWhiteList {
		if isReservedName(ent) {
			return nil, fmt.Errorf("invalid type name in whitelist: %q: reserved for internal purposes", ent)
		}
		whiteListSet[ent] = true
	}

	return &Searcher{
		cli:                 config.Client,
		typeWhiteList:       config.TypeWhiteList,
		typeWhiteListSet:    whiteListSet,
		typeRepository:      config.TypeRepo,
		cachedTypesDuration: config.CachedTypesDuration,
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
func (sr *Searcher) Search(ctx context.Context, cfg models.SearchConfig) ([]models.SearchResult, error) {
	if strings.TrimSpace(cfg.Text) == "" {
		return nil, fmt.Errorf("search text cannot be empty")
	}

	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	indices := sr.searchIndices(cfg.TypeWhiteList)
	query, err := sr.buildQuery(ctx, cfg, indices)
	if err != nil {
		return nil, fmt.Errorf("error building query: %v", err)
	}
	res, err := sr.cli.Search(
		sr.cli.Search.WithBody(query),
		sr.cli.Search.WithIndex(indices...),
		sr.cli.Search.WithSize(maxResults),
		sr.cli.Search.WithIgnoreUnavailable(true),
		sr.cli.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %v", err)
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error decoding search response: %v", err)
	}

	return sr.toSearchResults(response.Hits.Hits), nil
}

func (sr *Searcher) buildQuery(ctx context.Context, cfg models.SearchConfig, indices []string) (io.Reader, error) {
	queries, err := sr.buildQueriesFromIndices(ctx, indices, cfg)
	if err != nil {
		return nil, err
	}
	query := elastic.NewBoolQuery().
		Should(queries...).
		Filter(sr.filterQuery(cfg.Filters)...)

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

func (sr *Searcher) buildQueriesFromIndices(ctx context.Context, indices []string, cfg models.SearchConfig) ([]elastic.Query, error) {
	types, err := sr.mapIndicesToTypes(indices)
	if err != nil {
		return nil, err
	}

	var queries []elastic.Query
	for _, typ := range types {
		fields, err := sr.buildTypeFields(typ)
		if err != nil {
			return nil, err
		}

		query := elastic.NewBoolQuery().
			Should(
				elastic.
					NewMultiMatchQuery(
						cfg.Text,
						fields...,
					),
				elastic.
					NewMultiMatchQuery(
						cfg.Text,
						fields...,
					).
					Fuzziness("AUTO"),
				elastic.
					NewMultiMatchQuery(
						cfg.Text,
					).
					Fuzziness("AUTO"),
			).
			Filter(
				elastic.NewTermQuery("_index", typ.Name),
			)
		queries = append(queries, query)
	}

	return queries, nil
}

func (sr *Searcher) buildTypeFields(resourceType models.Type) (fields []string, err error) {
	fields = append(
		fields,
		fmt.Sprintf("%s^10", resourceType.Fields.ID),
		fmt.Sprintf("%s^5", resourceType.Fields.Title),
	)

	return
}

func (sr *Searcher) filterQuery(filters map[string][]string) (filterQueries []elastic.Query) {
	for key, elements := range filters {
		if len(elements) < 1 {
			continue
		}
		filterQueries = append(
			filterQueries,
			elastic.NewTermQuery(key, elements[0]),
		)
	}
	return
}

func (sr *Searcher) toSearchResults(hits []searchHit) (results []models.SearchResult) {
	for _, hit := range hits {
		results = append(results, models.SearchResult{
			TypeName: hit.Index,
			RecordV1: hit.Source,
		})
	}
	return
}

func (sr *Searcher) searchIndices(localWhiteList []string) []string {
	hasGL := len(sr.typeWhiteList) > 0
	hasLL := len(localWhiteList) > 0
	switch {
	case hasGL && hasLL:
		var indices []string
		for _, idx := range localWhiteList {
			if sr.typeWhiteListSet[idx] {
				indices = append(indices, idx)
			}
		}
		return indices
	case hasGL || hasLL:
		return anyValidStringSlice(localWhiteList, sr.typeWhiteList)
	default:
		return []string{}
	}
}

func (sr *Searcher) mapIndicesToTypes(indices []string) ([]models.Type, error) {
	types, err := sr.getTypes(context.Background())
	if err != nil {
		return types, err
	}
	if len(indices) == 0 {
		return types, nil
	}

	whitelistedTypes := []models.Type{}
	for _, index := range indices {
		typ, ok := sr.cachedTypesMap[index]
		if ok {
			whitelistedTypes = append(whitelistedTypes, typ)
		}
	}

	return whitelistedTypes, nil
}

func (sr *Searcher) getTypes(ctx context.Context) ([]models.Type, error) {
	if sr.cachedTypes == nil || time.Now().After(sr.cachedTypeExpiredOn) {
		types, err := sr.typeRepository.GetAll(ctx)
		if err != nil {
			return nil, err
		}

		sr.cachedTypes = types
		sr.cachedTypesMap = sr.buildTypesMap(types)
		sr.cachedTypeExpiredOn = time.Now().Add(time.Duration(sr.cachedTypesDuration) * time.Second)
	}

	return sr.cachedTypes, nil
}

func (sr *Searcher) buildTypesMap(types []models.Type) map[string]models.Type {
	typesMap := map[string]models.Type{}
	for _, typ := range types {
		typesMap[typ.Name] = typ
	}

	return typesMap
}

func anyValidStringSlice(slices ...[]string) []string {
	for _, slice := range slices {
		if len(slice) > 0 {
			return slice
		}
	}
	return nil
}
