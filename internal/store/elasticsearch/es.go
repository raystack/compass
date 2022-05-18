package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/odpf/compass/core/asset"
	"github.com/olivere/elastic/v7"
)

const (
	// name of the search index
	defaultSearchIndex = "universe"
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
	*elasticsearch.Client
}

func Migrate(ctx context.Context, cli *elasticsearch.Client, assetType asset.Type) error {
	// checking for the existence of index before adding the metadata entry
	idxExists, err := indexExists(ctx, cli, assetType.String())
	if err != nil {
		return fmt.Errorf("error checking index existence: %w", err)
	}

	// update/create the index
	if idxExists {
		err = updateIdx(ctx, cli, assetType)
		if err != nil {
			err = fmt.Errorf("error updating index: %w", err)
		}
	} else {
		err = createIdx(ctx, cli, assetType)
		if err != nil {
			err = fmt.Errorf("error creating index: %w", err)
		}
	}

	return err
}

func createIdx(ctx context.Context, cli *elasticsearch.Client, assetType asset.Type) error {
	indexSettings := buildTypeIndexSettings()
	res, err := cli.Indices.Create(
		assetType.String(),
		cli.Indices.Create.WithBody(strings.NewReader(indexSettings)),
		cli.Indices.Create.WithContext(ctx),
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

func updateIdx(ctx context.Context, cli *elasticsearch.Client, assetType asset.Type) error {
	res, err := cli.Indices.PutMapping(
		strings.NewReader(typeIndexMapping),
		cli.Indices.PutMapping.WithIndex(assetType.String()),
		cli.Indices.PutMapping.WithContext(ctx),
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
