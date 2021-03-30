package store

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/odpf/columbus/models"
)

type mappingType struct {
	Properties map[string]interface{} `json:"properties"`
}

var keywordField = map[string]interface{}{
	"keyword": map[string]interface{}{
		"type":         "keyword",
		"ignore_above": 256,
	},
}

// generates the mapping payload for a given type
// this is used for configuring boost
func createIndexMapping(e models.Type) (string, error) {
	mapping := mappingType{
		Properties: make(map[string]interface{}),
	}
	for field, boost := range e.Boost {
		mapping.Properties[field] = map[string]interface{}{
			"type":   "text",
			"boost":  boost,
			"fields": keywordField,
		}
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(mapping); err != nil {
		return "{}", fmt.Errorf("error encoding mapping to JSON: %v", err)
	}
	return buf.String(), nil
}
