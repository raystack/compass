package asset

import (
	"context"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
)

//go:generate mockery --name=LineageRepository -r --case underscore --with-expecter --structname LineageRepository --filename lineage_repository.go --output=./mocks
type LineageRepository interface {
	GetGraph(ctx context.Context, node Node) (Graph, error)
	Upsert(ctx context.Context, node Node, upstreams, downstreams []Node) error
}

type Graph []Edge

type Edge struct {
	// Source represents source's node ID
	Source string `json:"source"`

	// Target represents target's node ID
	Target string `json:"target"`

	// Prop is a map containing extra information about the edge
	Prop map[string]interface{} `json:"prop"`
}

func (e Edge) ToProto() (*compassv1beta1.LineageEdge, error) {
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

func NewEdgeFromProto(pb *compassv1beta1.LineageEdge) Edge {
	var propVal map[string]interface{}
	propPB := pb.GetProp()
	if propPB != nil {
		propVal = propPB.AsMap()
	}

	return Edge{
		Source: pb.GetSource(),
		Target: pb.GetTarget(),
		Prop:   propVal,
	}
}

type Node struct {
	URN     string `json:"urn"`
	Type    Type   `json:"type"`
	Service string `json:"service"`
}

func (n Node) ToProto() *compassv1beta1.LineageNode {
	return &compassv1beta1.LineageNode{
		Urn:     n.URN,
		Type:    string(n.Type),
		Service: n.Service,
	}
}

func NewNodeFromProto(proto *compassv1beta1.LineageNode) Node {
	return Node{
		URN:     proto.GetUrn(),
		Type:    Type(proto.GetType()),
		Service: proto.GetService(),
	}
}
