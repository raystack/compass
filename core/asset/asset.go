package asset

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname AssetRepository --filename asset_repository.go --output=./mocks
import (
	"context"
	"fmt"
	"time"

	"github.com/r3labs/diff/v3"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/core/user"
)

type Repository interface {
	GetAll(context.Context, Filter) ([]Asset, error)
	GetCount(context.Context, Filter) (int, error)
	GetByID(ctx context.Context, id string) (Asset, error)
	GetByURN(ctx context.Context, urn string) (Asset, error)
	GetVersionHistory(ctx context.Context, flt Filter, id string) ([]Asset, error)
	GetByVersionWithID(ctx context.Context, id string, version string) (Asset, error)
	GetByVersionWithURN(ctx context.Context, urn string, version string) (Asset, error)
	GetTypes(ctx context.Context, flt Filter) (map[Type]int, error)
	Upsert(ctx context.Context, ns *namespace.Namespace, ast *Asset) (string, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteByURN(ctx context.Context, urn string) error
	SoftDeleteByID(ctx context.Context, id string) (string, error)
	SoftDeleteByURN(ctx context.Context, urn string) (string, error)
	GetCountByIsDeleted(ctx context.Context, isDeleted bool) (int, error)
	HardDeleteByURNs(ctx context.Context, urns []string) error
	AddProbe(ctx context.Context, ns *namespace.Namespace, assetURN string, probe *Probe) error
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
	IsDeleted   bool                   `json:"is_deleted" diff:"is_deleted"`
	CreatedAt   time.Time              `json:"created_at" diff:"-"`
	UpdatedAt   time.Time              `json:"updated_at" diff:"-"`
	RefreshedAt *time.Time             `json:"refreshed_at,omitempty" diff:"-"`
	Version     string                 `json:"version" diff:"-"`
	UpdatedBy   user.User              `json:"updated_by" diff:"-"`
	Changelog   diff.Changelog         `json:"changelog,omitempty" diff:"-"`
	Probes      []Probe                `json:"probes,omitempty"`
}

var ErrAssetAlreadyDeleted = fmt.Errorf("asset already deleted")

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
