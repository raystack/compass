package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/asset"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
)

func (server *APIServer) GetAllTypes(ctx context.Context, req *connect.Request[compassv1beta1.GetAllTypesRequest]) (*connect.Response[compassv1beta1.GetAllTypesResponse], error) {

	flt, err := asset.NewFilterBuilder().
		Types(req.Msg.GetTypes()).
		Services(req.Msg.GetServices()).
		Q(req.Msg.GetQ()).
		QFields(req.Msg.GetQFields()).
		Data(req.Msg.GetData()).
		Build()
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s", bodyParserErrorMsg(err)))
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

	return connect.NewResponse(&compassv1beta1.GetAllTypesResponse{
		Data: results,
	}), nil
}
