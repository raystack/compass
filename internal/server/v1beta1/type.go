package handlersv1beta1

import (
	"context"
	"fmt"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/asset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *APIServer) GetAllTypes(ctx context.Context, req *compassv1beta1.GetAllTypesRequest) (*compassv1beta1.GetAllTypesResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	flt, err := server.buildAssetsFilter(
		req.GetTypes(),
		req.GetServices(),
		req.GetQ(),
		req.GetQFields(),
		0,
		0,
		"",
		"",
		req.GetData())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	typesNameMap, err := server.assetService.GetTypes(ctx, flt)
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
