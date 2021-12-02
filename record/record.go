package record

import (
	"fmt"
	"time"
)

// Record is a model that wraps arbitrary data with Columbus' context
type Record struct {
	Urn         string                 `json:"urn"`
	Name        string                 `json:"name"`
	Service     string                 `json:"service"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Labels      map[string]string      `json:"labels"`
	Tags        []string               `json:"tags"`
	Upstreams   []LineageRecord        `json:"upstreams"`
	Downstreams []LineageRecord        `json:"downstreams"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type LineageRecord struct {
	Urn  string `json:"urn"`
	Type string `json:"type"`
}

type ErrNoSuchRecord struct {
	RecordID string
}

func (err ErrNoSuchRecord) Error() string {
	return fmt.Sprintf("no such record: %q", err.RecordID)
}
