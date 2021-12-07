package elasticsearch

import "time"

const (
	defaultScrollTimeout   = 30 * time.Second
	defaultScrollBatchSize = 1000
)
