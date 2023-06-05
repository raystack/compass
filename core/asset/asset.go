package asset

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname AssetRepository --filename asset_repository.go --output=./mocks
import (
	"context"
	"time"

	"github.com/goto/compass/core/user"
	"github.com/r3labs/diff/v2"
)

type Repository interface {
	GetAll(context.Context, Filter) ([]Asset, error)
	GetCount(context.Context, Filter) (int, error)
	GetByID(ctx context.Context, id string) (Asset, error)
	GetByURN(ctx context.Context, urn string) (Asset, error)
	GetVersionHistory(ctx context.Context, flt Filter, id string) ([]Asset, error)
	GetByVersionWithID(ctx context.Context, id, version string) (Asset, error)
	GetByVersionWithURN(ctx context.Context, urn, version string) (Asset, error)
	GetTypes(ctx context.Context, flt Filter) (map[Type]int, error)
	Upsert(ctx context.Context, ast *Asset) (string, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteByURN(ctx context.Context, urn string) error
	AddProbe(ctx context.Context, assetURN string, probe *Probe) error
	GetProbes(ctx context.Context, assetURN string) ([]Probe, error)
	GetProbesWithFilter(ctx context.Context, flt ProbesFilter) (map[string][]Probe, error)
}

// Asset is a model that wraps arbitrary data with Compass' context
type Asset struct {
	ID          string                 `json:"id" diff:"-"`
	URN         string                 `json:"urn" diff:"-"`
	Type        Type                   `json:"type" diff:"-"`
	Service     string                 `json:"service" diff:"-"`
	Name        string                 `json:"name" diff:"name"`
	Description string                 `json:"description" diff:"description"`
	Data        map[string]interface{} `json:"data" diff:"data"`
	URL         string                 `json:"url" diff:"url"`
	Labels      map[string]string      `json:"labels" diff:"labels"`
	Owners      []user.User            `json:"owners,omitempty" diff:"owners"`
	CreatedAt   time.Time              `json:"created_at" diff:"-"`
	UpdatedAt   time.Time              `json:"updated_at" diff:"-"`
	Version     string                 `json:"version" diff:"-"`
	UpdatedBy   user.User              `json:"updated_by" diff:"-"`
	Changelog   diff.Changelog         `json:"changelog,omitempty" diff:"-"`
	Probes      []Probe                `json:"probes,omitempty"`
}

// Diff returns nil changelog with nil error if equal
// returns wrapped r3labs/diff Changelog struct with nil error if not equal
func (a *Asset) Diff(otherAsset *Asset) (diff.Changelog, error) {
	return diff.Diff(a, otherAsset, diff.DiscardComplexOrigin(), diff.AllowTypeMismatch(true))
}

// Patch appends asset with data from map. It mutates the asset itself.
// It is using json annotation of the struct to patch the correct keys
func (a *Asset) Patch(patchData map[string]interface{}) {
	patchAsset(a, patchData)
}
