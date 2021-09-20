package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/odpf/columbus/models"
)

// used as a utility for generating request payload
// since github.com/olivere/elastic generates the
// <Q> in {"query": <Q>}
type searchQuery struct {
	Query    interface{} `json:"query"`
	MinScore interface{} `json:"min_score"`
}

type searchHit struct {
	Index  string          `json:"_index"`
	Source models.RecordV1 `json:"_source"`
}

type searchResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []searchHit `json:"hits"`
	} `json:"hits"`
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

// checks for the existence of an index
func indexExists(ctx context.Context, cli *elasticsearch.Client, name string) (bool, error) {
	res, err := cli.Indices.Exists(
		[]string{name},
		cli.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("indexExists: %w", elasticSearchError(err))
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}

// helper for decorating unsuccesful invocations of the es REST API
// (transport errors)
func elasticSearchError(err error) error {
	return fmt.Errorf("elasticsearch error: %w", err)
}
