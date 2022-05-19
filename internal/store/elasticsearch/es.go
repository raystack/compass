package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/newrelic/go-agent/v3/integrations/nrelasticsearch-v7"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/salt/log"
	"github.com/olivere/elastic/v7"
)

const (
	// name of the search index
	defaultSearchIndex = "universe"
)

type Config struct {
	Brokers string `mapstructure:"brokers" default:"http://localhost:9200"`
}

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
}

type esIndex struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
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

// helper for decorating unsuccesful invocations of the es REST API
// (transport errors)
func elasticSearchError(err error) error {
	return fmt.Errorf("elasticsearch error: %w", err)
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
		// uncomment below code to debug request and response to elasticsearch
		// Logger: &estransport.ColorLogger{
		//	Output:             os.Stdout,
		//	EnableRequestBody:  true,
		//	EnableResponseBody: true,
		// },
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

func (c *Client) Migrate(ctx context.Context, assetType asset.Type) error {
	// checking for the existence of index before adding the metadata entry
	idxExists, err := c.indexExists(ctx, assetType.String())
	if err != nil {
		return fmt.Errorf("error checking index existence: %w", err)
	}

	// update/create the index
	if idxExists {
		c.logger.Info("index already exist, updating it instead")
		if err = c.updateIdx(ctx, assetType); err != nil {
			return fmt.Errorf("error updating index: %w", err)
		}
		return nil
	}

	if err = c.createIdx(ctx, assetType); err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	return nil
}

func (c *Client) createIdx(ctx context.Context, assetType asset.Type) error {
	indexSettings := buildTypeIndexSettings()
	res, err := c.client.Indices.Create(
		assetType.String(),
		c.client.Indices.Create.WithBody(strings.NewReader(indexSettings)),
		c.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error creating index %q: %s", assetType, errorReasonFromResponse(res))
	}
	return nil
}

func (c *Client) updateIdx(ctx context.Context, assetType asset.Type) error {
	res, err := c.client.Indices.PutMapping(
		strings.NewReader(typeIndexMapping),
		c.client.Indices.PutMapping.WithIndex(assetType.String()),
		c.client.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error updating index %q: %s", assetType, errorReasonFromResponse(res))
	}
	return nil
}

func buildTypeIndexSettings() string {
	return fmt.Sprintf(indexSettingsTemplate, typeIndexMapping, defaultSearchIndex)
}

// checks for the existence of an index
func (c *Client) indexExists(ctx context.Context, name string) (bool, error) {
	res, err := c.client.Indices.Exists(
		[]string{name},
		c.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("indexExists: %w", elasticSearchError(err))
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}
