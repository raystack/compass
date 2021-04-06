package store

import (
	"bytes"
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
	filterScriptBasePredicate       = `doc.containsKey(%q) == false || doc[%q].size() == 0`
	filterScriptMatchValuePredicate = `doc[%q].value == %q`
	defaultMaxResults               = 200
	defaultMinScore                 = 0.01
)

type SearcherConfig struct {
	Client                 *elasticsearch.Client
	TypeRepo               models.TypeRepository
	TypeWhiteList          []string
	CachedTypesMapDuration int
}

// Searcher is an implementation of models.RecordSearcher
type Searcher struct {
	cli                    *elasticsearch.Client
	typeWhiteList          []string
	typeWhiteListSet       map[string]bool
	typeRepository         models.TypeRepository
	cachedTypesMap         map[string]models.Type
	cachedTypeExpiredOn    time.Time
	cachedTypesMapDuration int
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
		cli:                    config.Client,
		typeWhiteList:          config.TypeWhiteList,
		typeWhiteListSet:       whiteListSet,
		typeRepository:         config.TypeRepo,
		cachedTypesMapDuration: config.CachedTypesMapDuration,
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
func (sr *Searcher) Search(cfg models.SearchConfig) ([]models.SearchResult, error) {
	if strings.TrimSpace(cfg.Text) == "" {
		return nil, fmt.Errorf("search text cannot be empty")
	}

	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	indices := sr.searchIndices(cfg.TypeWhiteList)
	query, err := sr.buildQuery(cfg, indices)
	if err != nil {
		return nil, fmt.Errorf("error building query: %v", err)
	}
	res, err := sr.cli.Search(
		sr.cli.Search.WithBody(query),
		sr.cli.Search.WithIndex(indices...),
		sr.cli.Search.WithSize(maxResults),
		sr.cli.Search.WithIgnoreUnavailable(true),
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

func (sr *Searcher) buildQuery(cfg models.SearchConfig, indices []string) (io.Reader, error) {
	queries, err := sr.buildQueriesFromIndices(indices, cfg)
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

func (sr *Searcher) buildQueriesFromIndices(indices []string, cfg models.SearchConfig) ([]elastic.Query, error) {
	var queries []elastic.Query
	for _, index := range indices {
		fields, err := sr.buildTypeFields(index)
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
				elastic.NewTermQuery("_index", index),
			)
		queries = append(queries, query)
	}

	return queries, nil
}

func (sr *Searcher) buildTypeFields(typeName string) (fields []string, err error) {
	resourceType, err := sr.getType(typeName)
	if err != nil {
		return fields, err
	}

	fields = append(
		fields,
		fmt.Sprintf("%s^10", resourceType.Fields.ID),
		fmt.Sprintf("%s^5", resourceType.Fields.Title),
	)

	return
}

func (sr *Searcher) filterQuery(filters map[string][]string) (filterQueries []elastic.Query) {
	for key, elements := range filters {
		filter := buildScriptFilter(key, elements)
		filterQueries = append(filterQueries, filter)
	}
	return
}

func (sr *Searcher) toSearchResults(hits []searchHit) (results []models.SearchResult) {
	for _, hit := range hits {
		results = append(results, models.SearchResult{
			TypeName: hit.Index,
			Record:   hit.Source,
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
		return []string{defaultSearchIndex}
	}
}

func (sr *Searcher) getType(typeName string) (models.Type, error) {
	if sr.cachedTypesMap == nil || time.Now().After(sr.cachedTypeExpiredOn) {
		typesMap, err := sr.buildTypesMap()
		if err != nil {
			return models.Type{}, err
		}

		sr.cachedTypesMap = typesMap
		sr.cachedTypeExpiredOn = time.Now().Add(time.Duration(sr.cachedTypesMapDuration) * time.Second)
	}

	resourceType, ok := sr.cachedTypesMap[typeName]
	if !ok {
		return models.Type{}, fmt.Errorf("type does not exist")
	}

	return resourceType, nil
}

func (sr *Searcher) buildTypesMap() (map[string]models.Type, error) {
	types, err := sr.typeRepository.GetAll()
	if err != nil {
		return nil, err
	}

	typesMap := map[string]models.Type{}
	for _, typ := range types {
		typesMap[typ.Name] = typ
	}

	return typesMap, nil
}

func anyValidStringSlice(slices ...[]string) []string {
	for _, slice := range slices {
		if len(slice) > 0 {
			return slice
		}
	}
	return nil
}

// buildScriptFilter builds a script that can be used
// within the filter context of a query. The script behaves
// mostly as a "terms" filter, except that it will also match documents
// that don't have the filter "key".
func buildScriptFilter(key string, values []string) *elastic.ScriptQuery {

	// by default, all string fields in the document are indexed as `text`, which is not suitable
	// for exact matches, since their contents are analyzed before being stored.
	// The "${key}.keyword" sub-field is created for each text field that has < 256 chars, and holds
	// the unprocessed contents, which are suitable for exact string matches
	key = fmt.Sprintf("%s.keyword", key)

	predicates := []string{
		fmt.Sprintf(filterScriptBasePredicate, key, key),
	}
	for _, value := range values {
		predicate := fmt.Sprintf(filterScriptMatchValuePredicate, key, value)
		predicates = append(predicates, predicate)
	}

	src := strings.Join(predicates, " || ")
	return elastic.NewScriptQuery(elastic.NewScriptInline(src))
}
