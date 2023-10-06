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

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/goto/compass/core/asset"
	"github.com/goto/salt/log"
	"github.com/newrelic/go-agent/v3/integrations/nrelasticsearch-v7"
	"github.com/olivere/elastic/v7"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	// name of the search index
	defaultSearchIndex = "universe"
)

type Config struct {
	Brokers        string        `mapstructure:"brokers" default:"http://localhost:9200"`
	RequestTimeout time.Duration `mapstructure:"request_timeout" default:"10s"`
}

type searchHit struct {
	Index     string                 `json:"_index"`
	Source    asset.Asset            `json:"_source"`
	HighLight map[string]interface{} `json:"highlight"`
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

type groupResponse struct {
	Aggregations struct {
		CompositeAggregations struct {
			Buckets []aggregationBucket `json:"buckets"`
		} `json:"composite-group"`
	} `json:"aggregations"`
}

type aggregationBucket struct {
	Key      map[string]string `json:"key"`
	DocCount int               `json:"doc_count"`
	Hits     struct {
		Hits struct {
			Hits []groupHits `json:"hits"`
		} `json:"hits"`
	} `json:"hits"`
}

type groupHits struct {
	Source asset.Asset `json:"_source"`
}

// extract error reason from an elasticsearch response
// returns the raw message in case it fails
func errorCodeAndReason(res *esapi.Response) (code, reason string) {
	var (
		r struct {
			Error struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		}
		cp bytes.Buffer
	)
	err := json.NewDecoder(io.TeeReader(res.Body, &cp)).Decode(&r)
	if err != nil {
		return "unknown", fmt.Sprintf("raw response: %s", cp.String())
	}

	return r.Error.Type, r.Error.Reason
}

type Client struct {
	client *elasticsearch.Client
	logger log.Logger

	clientLatency metric.Int64Histogram
}

func NewClient(logger log.Logger, config Config, opts ...ClientOption) (*Client, error) {
	clientLatency, err := otel.Meter("github.com/goto/compass/internal/store/elasticsearch").
		Int64Histogram("compass.es.client.duration")
	if err != nil {
		otel.Handle(err)
	}

	c := &Client{
		logger:        logger,
		clientLatency: clientLatency,
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
	defer drainBody(res)
	if res.IsError() {
		return "", errors.New(res.Status())
	}
	info := struct {
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

func (c *Client) CreateIdx(ctx context.Context, discoveryOp, indexName string) (err error) {
	defer func(start time.Time) {
		const op = "create_index"
		c.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: discoveryOp,
			start:       start,
			err:         err,
		})
	}(time.Now())

	indexSettings := buildTypeIndexSettings()
	res, err := c.client.Indices.Create(
		indexName,
		c.client.Indices.Create.WithBody(strings.NewReader(indexSettings)),
		c.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return asset.DiscoveryError{
			Op:    "CreateIdx",
			Index: indexName,
			Err:   fmt.Errorf("create index '%s': %w", indexName, err),
		}
	}
	defer drainBody(res)
	if res.IsError() {
		code, reason := errorCodeAndReason(res)
		return asset.DiscoveryError{
			Op:     "CreateIdx",
			Index:  indexName,
			ESCode: code,
			Err:    fmt.Errorf("create index '%s': %s", indexName, reason),
		}
	}
	return nil
}

func buildTypeIndexSettings() string {
	return fmt.Sprintf(indexSettingsTemplate, serviceIndexMapping, defaultSearchIndex)
}

// checks for the existence of an index
func (c *Client) indexExists(ctx context.Context, discoveryOp, name string) (exists bool, err error) {
	defer func(start time.Time) {
		const op = "index_exists"
		c.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: discoveryOp,
			start:       start,
			err:         err,
		})
	}(time.Now())

	res, err := c.client.Indices.Exists(
		[]string{name},
		c.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("check index exists: %w", err)
	}
	defer drainBody(res)
	return res.StatusCode == 200, nil
}

type instrumentParams struct {
	op          string
	discoveryOp string
	start       time.Time
	err         error
}

func (c *Client) instrumentOp(ctx context.Context, params instrumentParams) {
	statusCode := "ok"
	if params.err != nil {
		statusCode = "unknown"
		var de asset.DiscoveryError
		if errors.As(params.err, &de) && de.ESCode != "" {
			statusCode = de.ESCode
		}
	}

	c.clientLatency.Record(
		ctx, time.Since(params.start).Milliseconds(), metric.WithAttributes(
			attribute.String("es.operation", params.op),
			attribute.String("es.status_code", statusCode),
			attribute.String("compass.discovery_operation", params.discoveryOp),
		),
	)
}

// drainBody drains and closes the response body to avoid the following
// gotcha:
// http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/index.html#close_http_resp_body
func drainBody(resp *esapi.Response) {
	if resp == nil {
		return
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
