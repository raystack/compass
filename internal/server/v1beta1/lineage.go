package handlersv1beta1

import (
	"context"
	"fmt"

	"github.com/goto/compass/core/asset"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (server *APIServer) GetGraph(ctx context.Context, req *compassv1beta1.GetGraphRequest) (*compassv1beta1.GetGraphResponse, error) {
	_, err := server.ValidateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	direction := asset.LineageDirection(req.GetDirection())
	if !direction.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid direction value")
	}

	lineage, err := server.assetService.GetLineage(ctx, req.GetUrn(), asset.LineageQuery{
		Level:     int(req.GetLevel()),
		Direction: direction,
	})
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	edges := make([]*compassv1beta1.LineageEdge, 0, len(lineage.Edges))
	for _, edge := range lineage.Edges {
		edgePB, err := lineageEdgeToProto(edge)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		edges = append(edges, edgePB)
	}

	nodeAttrs := make(map[string]*compassv1beta1.GetGraphResponse_NodeAttributes, len(lineage.NodeAttrs))
	for urn, attrs := range lineage.NodeAttrs {
		probesInfo, err := probesInfoToProto(attrs.Probes)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}

		nodeAttrs[urn] = &compassv1beta1.GetGraphResponse_NodeAttributes{
			Probes: probesInfo,
		}
	}

	return &compassv1beta1.GetGraphResponse{
		Data:      edges,
		NodeAttrs: nodeAttrs,
	}, nil
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

func probesInfoToProto(probes asset.ProbesInfo) (*compassv1beta1.GetGraphResponse_ProbesInfo, error) {
	latest, err := probeToProto(probes.Latest)
	if err != nil {
		return nil, fmt.Errorf("convert probe to proto representation: %w", err)
	}

	return &compassv1beta1.GetGraphResponse_ProbesInfo{
		Latest: latest,
	}, nil
}
