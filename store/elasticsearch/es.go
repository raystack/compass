package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/odpf/columbus/record"
	"github.com/olivere/elastic/v7"
)

// used as a utility for generating request payload
// since github.com/olivere/elastic generates the
// <Q> in {"query": <Q>}
type searchQuery struct {
	Query    interface{} `json:"query"`
	MinScore float32     `json:"min_score"`
}

type searchHit struct {
	Index  string        `json:"_index"`
	Source record.Record `json:"_source"`
}

type searchResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []searchHit `json:"hits"`
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
