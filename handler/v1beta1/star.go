package handlersv1beta1

//go:generate mockery --name=StarService -r --case underscore --with-expecter --structname StarService --filename star_service.go --output=./mocks
import (
	"context"
	"github.com/raystack/compass/core/namespace"

	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/core/star"
	"github.com/raystack/compass/core/user"
)

type StarService interface {
	GetStarredAssetsByUserID(context.Context, star.Filter, string) ([]asset.Asset, error)
	GetStarredAssetByUserID(context.Context, string, string) (asset.Asset, error)
	GetStargazers(context.Context, star.Filter, string) ([]user.User, error)
	Stars(context.Context, *namespace.Namespace, string, string) (string, error)
	Unstars(context.Context, string, string) error
}
