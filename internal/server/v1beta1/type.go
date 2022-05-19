package handlersv1beta1

import (
	"context"
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/asset"
)

func (server *APIServer) GetAllTypes(ctx context.Context, req *compassv1beta1.GetAllTypesRequest) (*compassv1beta1.GetAllTypesResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	typesNameMap, err := server.assetService.GetTypes(ctx)
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error fetching types: %s", err.Error()))
	}

	results := []*compassv1beta1.Type{}
	for _, typName := range asset.AllSupportedTypes {
		count := typesNameMap[typName]
		results = append(results, &compassv1beta1.Type{
			Name:  typName.String(),
			Count: uint32(count),
		})
	}

	return &compassv1beta1.GetAllTypesResponse{
		Data: results,
	}, nil
}
