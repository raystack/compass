package asset

//go:generate mockery --name=DiscoveryRepository -r --case underscore --with-expecter --structname DiscoveryRepository --filename discovery_repository.go --output=./mocks
import (
	"context"
	"github.com/odpf/compass/core/namespace"
	"github.com/odpf/compass/core/validator"
)

type DiscoveryRepository interface {
	Upsert(ctx context.Context, ns *namespace.Namespace, ast *Asset) error
	DeleteByID(ctx context.Context, ns *namespace.Namespace, assetID string) error
	DeleteByURN(ctx context.Context, ns *namespace.Namespace, assetURN string) error
	Search(ctx context.Context, cfg SearchConfig) (results []SearchResult, err error)
	Suggest(ctx context.Context, cfg SearchConfig) (suggestions []string, err error)
}

// SearchFilter is a filter intended to be used as a search
// criteria for operations involving asset search
type SearchFilter = map[string][]string

// SearchConfig represents a search query along
// with any corresponding filter(s)
type SearchConfig struct {
	// Text to search for
	Text string `validate:"required"`

	// Filters specifies document level values to look for.
	// Multiple values can be specified for a single key
	Filters SearchFilter

	// Number of relevant results to return
	MaxResults int

	// RankBy is a param to rank based on a specific parameter
	RankBy string

	// Queries is a param to search a resource based on asset's fields
	Queries map[string]string

	// Namespace under which assets are partitioned. *Required*
	Namespace *namespace.Namespace `validate:"required"`
}

func (s SearchConfig) Validate() error {
	return validator.ValidateStruct(s)
}

// SearchResult represents an item/result in a list of search results
type SearchResult struct {
	ID          string                 `json:"id"`
	URN         string                 `json:"urn"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Service     string                 `json:"service"`
	Description string                 `json:"description"`
	Labels      map[string]string      `json:"labels"`
	Data        map[string]interface{} `json:"data"`
}

// ToAsset returns search result as asset
func (sr SearchResult) ToAsset() Asset {
	return Asset{
		ID:          sr.ID,
		URN:         sr.URN,
		Name:        sr.Title,
		Type:        Type(sr.Type),
		Service:     sr.Service,
		Description: sr.Description,
		Labels:      sr.Labels,
		Data:        sr.Data,
	}
}
