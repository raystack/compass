package asset

import "time"

type ProbesFilter struct {
	AssetURNs []string
	MaxRows   int
	NewerThan time.Time
	OlderThan time.Time
}
