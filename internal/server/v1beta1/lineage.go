package handlersv1beta1

import (
	"context"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/asset"
)

func (server *APIServer) GetGraph(ctx context.Context, req *compassv1beta1.GetGraphRequest) (*compassv1beta1.GetGraphResponse, error) {
	graph, err := server.assetService.GetLineage(ctx, asset.Node{URN: req.GetUrn()})
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	graphPB := []*compassv1beta1.LineageEdge{}
	for _, edge := range graph {
		edgePB, err := edge.ToProto()
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		graphPB = append(graphPB, edgePB)
	}

	return &compassv1beta1.GetGraphResponse{
		Data: graphPB,
	}, nil
}
