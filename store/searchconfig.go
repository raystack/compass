package store

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/odpf/columbus/models"
)

var (
	defaultMaxResults = 200
	defaultMinScore   = 0.01
)

type SearcherConfig struct {
	Client              *elasticsearch.Client
	TypeRepo            models.TypeRepository
	TypeWhiteList       []string
	CachedTypesDuration int
}
