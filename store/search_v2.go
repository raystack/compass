package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/models"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

var (
	defaultMaxResults = 200
	defaultMinScore   = 0.01
)

type SearcherConfig struct {
	Client        *elasticsearch.Client
	TypeWhiteList []string
}

// Searcher is an implementation of models.RecordV2Searcher
type SearcherV2 struct {
	cli              *elasticsearch.Client
	typeWhiteList    []string
	typeWhiteListSet map[string]bool
}

// NewSearcherV2 creates a new instance of SearcherV2
// You can optionally specify a list of type names to whitelist for search
// If the white list is nil (or has zero length), then the search will be run
// on all types. This can be further restricted by FilterConfig.TypeWhiteList
// in Search()
func NewSearcherV2(config SearcherConfig) (*SearcherV2, error) {
	var whiteListSet = make(map[string]bool)
	for _, ent := range config.TypeWhiteList {
		if isReservedName(ent) {
			return nil, fmt.Errorf("invalid type name in whitelist: %q: reserved for internal purposes", ent)
		}
		whiteListSet[ent] = true
	}

	return &SearcherV2{
		cli:              config.Client,
		typeWhiteList:    config.TypeWhiteList,
		typeWhiteListSet: whiteListSet,
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
func (sr *SearcherV2) Search(ctx context.Context, cfg models.SearchConfig) (results []models.SearchResultV2, err error) {
	if strings.TrimSpace(cfg.Text) == "" {
		err = errors.New("search text cannot be empty")
		return
	}

	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	indices := sr.searchIndices(cfg.TypeWhiteList)
	query, err := sr.buildQuery(ctx, cfg, indices)
	if err != nil {
		err = errors.Wrap(err, "error building query")
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
		err = errors.Wrap(err, "error executing search")
		return
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = errors.Wrap(err, "error decoding search response")
		return
	}

	results, err = sr.toSearchResults(response.Hits.Hits)
	if err != nil {
		err = errors.Wrap(err, "error building search results")
		return
	}

	return
}

func (sr *SearcherV2) buildQuery(ctx context.Context, cfg models.SearchConfig, indices []string) (io.Reader, error) {
	textQuery := sr.buildTextQuery(ctx, cfg.Text)
	filterQueries := sr.buildFilterQueries(cfg.Filters)
	query := elastic.NewBoolQuery().
		Should(textQuery).
		Filter(filterQueries...)

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

func (sr *SearcherV2) buildTextQuery(ctx context.Context, text string) elastic.Query {
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

func (sr *SearcherV2) buildFilterQueries(filters map[string][]string) (filterQueries []elastic.Query) {
	for key, elements := range filters {
		if len(elements) < 1 {
			continue
		}
		filterQueries = append(
			filterQueries,
			elastic.NewTermQuery("data."+key, elements[0]),
		)
	}
	return
}

func (sr *SearcherV2) toSearchResults(hits []searchHit) (results []models.SearchResultV2, err error) {
	for _, hit := range hits {
		var record models.RecordV2
		record, err = mapToV2(hit.Source)
		if err != nil {
			err = errors.Wrap(err, "error transforming to record v2")
			return
		}

		results = append(results, models.SearchResultV2{
			TypeName: hit.Index,
			RecordV2: record,
		})
	}
	return
}

func (sr *SearcherV2) searchIndices(localWhiteList []string) []string {
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

func anyValidStringSlice(slices ...[]string) []string {
	for _, slice := range slices {
		if len(slice) > 0 {
			return slice
		}
	}
	return nil
}
