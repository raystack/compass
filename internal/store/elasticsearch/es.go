package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/google/uuid"
	"github.com/newrelic/go-agent/v3/integrations/nrelasticsearch-v7"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/namespace"
	"github.com/odpf/salt/log"
	"github.com/olivere/elastic/v7"
	"io"
	"strings"
)

const (
	// name of the search index
	defaultSearchIndexAlias = "universe"

	// DefaultSharedIndexName use single shared index for small tenants
	DefaultSharedIndexName = "compass-idx-default"

	// DefaultShardCountPerIndex provides shard count in an index. For large scale, it should be at least 12
	DefaultShardCountPerIndex = 6

	// DedicatedTenantIndexPrefix is used to avoid index conflicts with system managed indices
	DedicatedTenantIndexPrefix = "compass-idx"

	// TenantIndexAliasPrefix is used to avoid index/alias conflicts, convention is same for
	// dedicated and shared tenants
	TenantIndexAliasPrefix = "compass-alias"
)

type Config struct {
	Brokers string `mapstructure:"brokers" default:"http://localhost:9200"`
}

type Route struct {
	// Index could be shared or a dedicated, if dedicated it will be identified by namespace name
	// for most cases it will be `DefaultSharedIndexName`
	Index string `json:"index"`
	// ReadKey route search query to respective shards
	// for most cases it will be namespace id
	ReadKey string `json:"read_key"`
	// WriteKey route index query to respective shards
	// for most cases it will be namespace id
	WriteKey string `json:"write_key"`
	// FilterKey finds set of documents uniquely for a tenant
	// for most cases it will be namespace id
	FilterKey string `json:"filter_key"`
}

var (
	DefaultRoute = &Route{
		Index:     DefaultSharedIndexName,
		ReadKey:   uuid.Nil.String(),
		WriteKey:  uuid.Nil.String(),
		FilterKey: uuid.Nil.String(),
	}
)

// used as a utility for generating request payload
// since github.com/olivere/elastic generates the
// <Q> in {"query": <Q>}
type searchQuery struct {
	Query    interface{} `json:"query"`
	MinScore float32     `json:"min_score"`
}

type searchHit struct {
	Index  string      `json:"_index"`
	Source asset.Asset `json:"_source"`
}

type aggregationBucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}

type searchResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Total elastic.TotalHits `json:"total"`
		Hits  []searchHit       `json:"hits"`
	} `json:"hits"`
	Suggest map[string][]struct {
		Text    string                           `json:"text"`
		Offset  int                              `json:"offset"`
		Length  float32                          `json:"length"`
		Options []elastic.SearchSuggestionOption `json:"options"`
	} `json:"suggest"`
	Aggregations struct {
		AggregationName struct {
			DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
			SumOtherDocCount        int `json:"sum_other_doc_count"`
			Buckets                 []aggregationBucket
		} `json:"aggregation_name"`
	} `json:"aggregations"`
}

// extract error reason from an elasticsearch response
// returns the raw message in case it fails
func errorReasonFromResponse(res *esapi.Response) string {
	var (
		response struct {
			Error struct {
				Reason string `json:"reason"`
			} `json:"error"`
		}
		copy bytes.Buffer
	)
	reader := io.TeeReader(res.Body, &copy)
	err := json.NewDecoder(reader).Decode(&response)
	if err != nil {
		return fmt.Sprintf("raw response = %s", copy.String())
	}
	return response.Error.Reason
}

type Client struct {
	client *elasticsearch.Client
	logger log.Logger
}

func NewClient(logger log.Logger, config Config, opts ...ClientOption) (*Client, error) {
	c := &Client{
		logger: logger,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.client != nil {
		return c, nil
	}

	brokers := strings.Split(config.Brokers, ",")
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: brokers,
		Transport: nrelasticsearch.NewRoundTripper(nil),
		//uncomment below code to debug request and response to elasticsearch
		//Logger: &estransport.ColorLogger{
		//	Output:             os.Stdout,
		//	EnableRequestBody:  true,
		//	EnableResponseBody: true,
		//},
		// Retry on 429 TooManyRequests statuses as well
		RetryOnStatus: []int{502, 503, 504, 429},
		MaxRetries:    3,
	})
	if err != nil {
		return nil, err
	}
	c.client = esClient

	return c, nil
}

func (c *Client) Init() (string, error) {
	res, err := c.client.Info()
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		return "", errors.New(res.Status())
	}
	var info = struct {
		ClusterName string `json:"cluster_name"`
		Version     struct {
			Number string `json:"number"`
		} `json:"version"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&info)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%q (server version %s)", info.ClusterName, info.Version.Number), nil
}

func (c *Client) CreateIndex(ctx context.Context, name string, shardCount int) error {
	indexSettings := buildIndexSettings(shardCount)

	res, err := c.client.Indices.Create(
		name,
		c.client.Indices.Create.WithBody(strings.NewReader(indexSettings)),
		c.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return asset.DiscoveryError{Err: err}
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error creating index %q: %s", name, errorReasonFromResponse(res))
	}
	return nil
}

func buildIndexSettings(shardCount int) string {
	return fmt.Sprintf(indexSettingsTemplate, serviceIndexMapping, defaultSearchIndexAlias, shardCount)
}

func (c *Client) CreateIdxAlias(ctx context.Context, ns *namespace.Namespace) error {
	var indexSettings string
	if ns.State == namespace.SharedState {
		indexSettings = buildIndexAliasSettings(indexSharedAliasSettingsTemplate, ns)
	} else {
		indexSettings = buildIndexAliasSettings(indexDedicatedAliasSettingsTemplate, ns)
	}
	res, err := esapi.IndicesUpdateAliasesRequest{
		Body: strings.NewReader(indexSettings),
	}.Do(ctx, c.client)
	if err != nil {
		return asset.DiscoveryError{Err: err}
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error creating index alias %q: %s", ns.Name, errorReasonFromResponse(res))
	}
	return nil
}

func BuildAliasNameFromNamespace(ns *namespace.Namespace) string {
	return fmt.Sprintf("%s-%s", TenantIndexAliasPrefix, ns.Name)
}

func BuildIndexNameFromNamespace(ns *namespace.Namespace) string {
	index := DefaultSharedIndexName
	if ns.State == namespace.DedicatedState {
		index = fmt.Sprintf("%s-%s", DedicatedTenantIndexPrefix, ns.Name)
	}
	return index
}

func buildRouteFromNamespace(ns *namespace.Namespace) Route {
	return Route{
		Index: BuildIndexNameFromNamespace(ns),
		// instead of id, we could use namespace name as well for read/write key
		// not sure which will work better
		ReadKey:   ns.ID.String(),
		WriteKey:  ns.ID.String(),
		FilterKey: ns.ID.String(),
	}
}

func buildIndexAliasSettings(template string, ns *namespace.Namespace) string {
	aliasName := BuildAliasNameFromNamespace(ns)
	route := buildRouteFromNamespace(ns)
	return strings.NewReplacer(
		"{{alias_name}}", aliasName,
		"{{index_name}}", route.Index,
		"{{filter_id}}", route.FilterKey,
		"{{write_id}}", route.WriteKey,
		"{{read_id}}", route.ReadKey,
	).Replace(template)
}

// IndexExists checks for the existence of an index
func (c *Client) IndexExists(ctx context.Context, ns *namespace.Namespace) (bool, error) {
	indexName := BuildIndexNameFromNamespace(ns)
	res, err := c.client.Indices.Exists(
		[]string{indexName},
		c.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("indexExists: %w", err)
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}
