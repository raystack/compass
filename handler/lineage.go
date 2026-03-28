package handler

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/internal/middleware"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
)

func (server *Handler) GetGraph(ctx context.Context, req *connect.Request[compassv1beta1.GetGraphRequest]) (*connect.Response[compassv1beta1.GetGraphResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	direction := asset.LineageDirection(req.Msg.GetDirection())
	if !direction.IsValid() {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid direction value"))
	}

	// Default to true for backward compatibility
	withAttributes := true
	if req.Msg.WithAttributes != nil {
		withAttributes = *req.Msg.WithAttributes
	}

	lineage, err := server.assetService.GetLineage(ctx, req.Msg.GetUrn(), asset.LineageQuery{
		Level:          int(req.Msg.GetLevel()),
		Direction:      direction,
		WithAttributes: withAttributes,
		IncludeDeleted: req.Msg.GetIncludeDeleted(),
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

	return connect.NewResponse(&compassv1beta1.GetGraphResponse{
		Data:      edges,
		NodeAttrs: nodeAttrs,
	}), nil
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
