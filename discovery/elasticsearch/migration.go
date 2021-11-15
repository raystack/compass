package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
)

// used as body to create index requests
// aliases the index to defaultSearchIndex
// and sets up the camelcase analyzer
var indexSettingsTemplate = `{
	"mappings": %s,
	"aliases": {
		%q: {}
	},
	"settings": {
		"analysis": {
			"analyzer": {
				"default": {
					"type": "pattern",
					"pattern": "([^\\p{L}\\d]+)|(?<=\\D)(?=\\d)|(?<=\\d)(?=\\D)|(?<=[\\p{L}&&[^\\p{Lu}]])(?=\\p{Lu})|(?<=\\p{Lu})(?=\\p{Lu}[\\p{L}&&[^\\p{Lu}]])"
				}
			}
		}
	}
}`

func Migrate(ctx context.Context, client *elasticsearch.Client) error {
	for _, index := range allIndexList {
		exists, err := indexExists(ctx, client, index)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if err := createIndex(ctx, client, index); err != nil {
			return err
		}
	}

	return nil
}

// checks for the existence of an index
func indexExists(ctx context.Context, cli *elasticsearch.Client, name string) (bool, error) {
	res, err := cli.Indices.Exists(
		[]string{name},
		cli.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("error checking index existence: %w", elasticSearchError(err))
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

func createIndex(ctx context.Context, client *elasticsearch.Client, index string) error {
	indexSettings, err := createIndexSettings(index)
	if err != nil {
		return fmt.Errorf("error building index settings: %v", err)
	}
	res, err := client.Indices.Create(
		index,
		client.Indices.Create.WithBody(strings.NewReader(indexSettings)),
		client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return elasticSearchError(err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error creating index %q: %s", index, errorReasonFromResponse(res))
	}
	return nil
}

func createIndexSettings(index string) (string, error) {
	mappings, err := createIndexMapping(index)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(indexSettingsTemplate, mappings, "universe"), nil
}

type mappingType struct {
	Properties map[string]interface{} `json:"properties"`
}

// generates the mapping payload for a given type
// this is used for configuring boost
func createIndexMapping(index string) (string, error) {
	mapping := mappingType{
		Properties: make(map[string]interface{}),
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(mapping); err != nil {
		return "{}", fmt.Errorf("error encoding mapping to JSON: %v", err)
	}
	return buf.String(), nil
}
