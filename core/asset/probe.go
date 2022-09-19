package asset

import (
	"time"
)

// Probe represents a single asset's probe
type Probe struct {
	ID           string                 `json:"id"`
	AssetURN     string                 `json:"asset_urn"`
	Status       string                 `json:"status"`
	StatusReason string                 `json:"status_reason"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}
