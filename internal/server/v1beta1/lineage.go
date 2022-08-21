package handlersv1beta1

import (
	"context"

	"github.com/odpf/compass/core/asset"
	compassv1beta1 "github.com/odpf/compass/proto/odpf/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (server *APIServer) GetGraph(ctx context.Context, req *compassv1beta1.GetGraphRequest) (*compassv1beta1.GetGraphResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	direction := asset.LineageDirection(req.GetDirection())
	if !direction.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid direction value")
	}

	graph, err := server.assetService.GetLineage(ctx, asset.LineageNode{URN: req.GetUrn()}, asset.LineageQuery{
		Level:     int(req.GetLevel()),
		Direction: direction,
	})
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	graphPB := []*compassv1beta1.LineageEdge{}
	for _, edge := range graph {
		edgePB, err := lineageEdgeToProto(edge)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		graphPB = append(graphPB, edgePB)
	}

	return &compassv1beta1.GetGraphResponse{
		Data: graphPB,
	}, nil
}

func lineageNodeFromProto(proto *compassv1beta1.LineageNode) asset.LineageNode {
	return asset.LineageNode{
		URN:     proto.GetUrn(),
		Type:    asset.Type(proto.GetType()),
		Service: proto.GetService(),
	}
}

func lineageEdgeToProto(e asset.LineageEdge) (*compassv1beta1.LineageEdge, error) {
	var (
		propPB *structpb.Struct
		err    error
	)

	if len(e.Prop) > 0 {
		propPB, err = structpb.NewStruct(e.Prop)
		if err != nil {
			return nil, err
		}
	}
	return &compassv1beta1.LineageEdge{
		Source: e.Source,
		Target: e.Target,
		Prop:   propPB,
	}, nil
}
