package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/record"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

var (
	defaultMaxResults = 200
	defaultMinScore   = 0.01
	allIndexList      = []string{
		string(record.TypeTable),
		string(record.TypeDashboard),
		string(record.TypeJob),
		string(record.TypeTopic),
	}
)

type SearcherConfig struct {
	Client *elasticsearch.Client
}

// Searcher is an implementation of record.RecordSearcher
type Searcher struct {
	cli *elasticsearch.Client
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
func (sr *Searcher) Search(ctx context.Context, cfg discovery.SearchConfig) (results []record.Record, err error) {
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

	results, err = sr.toRecords(response.Hits.Hits)
	if err != nil {
		err = errors.Wrap(err, "error building search results")
		return
	}

	return
}

func (sr *Searcher) buildIndices(cfg discovery.SearchConfig) []string {
	var indices []string
	if len(cfg.TypeWhiteList) > 0 {
		for _, index := range cfg.TypeWhiteList {
			indices = append(indices, string(index))
		}
	} else {
		indices = allIndexList
	}

	return indices
}

func (sr *Searcher) buildQuery(ctx context.Context, cfg discovery.SearchConfig, indices []string) (io.Reader, error) {
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

func (sr *Searcher) buildFilterQueries(filters map[string][]string) (filterQueries []elastic.Query) {
	for key, rawValues := range filters {
		if len(rawValues) < 1 {
			continue
		}

		var values []interface{}
		for _, rawVal := range rawValues {
			values = append(values, rawVal)
		}

		filterQueries = append(
			filterQueries,
			elastic.NewTermsQuery(key, values...),
		)
	}
	return
}

func (sr *Searcher) toRecords(hits []searchHit) (results []record.Record, err error) {
	for _, hit := range hits {
		r := hit.Source
		r.Type = record.Type(hit.Index)
		results = append(results, r)
	}
	return
}
