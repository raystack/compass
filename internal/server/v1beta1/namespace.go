package handlersv1beta1

//go:generate mockery --name=NamespaceService -r --case underscore --with-expecter --structname NamespaceService --filename namespace_service.go --output=./mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/namespace"
	compassv1beta1 "github.com/raystack/compass/proto/raystack/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type NamespaceService interface {
	Create(ctx context.Context, ns *namespace.Namespace) (string, error)
	Update(ctx context.Context, ns *namespace.Namespace) error
	GetByID(ctx context.Context, id uuid.UUID) (*namespace.Namespace, error)
	GetByName(ctx context.Context, name string) (*namespace.Namespace, error)
	List(ctx context.Context) ([]*namespace.Namespace, error)
}

func (server *APIServer) ListNamespaces(ctx context.Context, req *compassv1beta1.ListNamespacesRequest) (*compassv1beta1.ListNamespacesResponse, error) {
	namespaces, err := server.namespaceService.List(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var protoNamespaces []*compassv1beta1.Namespace
	for _, ns := range namespaces {
		protoNamespace, err := namespaceToProto(ns)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		protoNamespaces = append(protoNamespaces, protoNamespace)
	}
	return &compassv1beta1.ListNamespacesResponse{
		Namespaces: protoNamespaces,
	}, nil
}

func (server *APIServer) GetNamespace(ctx context.Context, req *compassv1beta1.GetNamespaceRequest) (*compassv1beta1.GetNamespaceResponse, error) {
	var ns *namespace.Namespace
	if nsID, err := uuid.Parse(req.GetUrn()); err == nil {
		if ns, err = server.namespaceService.GetByID(ctx, nsID); err != nil {
			return nil, status.Error(codes.NotFound, err.Error())
		}
	} else {
		if ns, err = server.namespaceService.GetByName(ctx, req.GetUrn()); err != nil {
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}

	protoNamespace, err := namespaceToProto(ns)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &compassv1beta1.GetNamespaceResponse{
		Namespace: protoNamespace,
	}, nil
}

func (server *APIServer) CreateNamespace(ctx context.Context, req *compassv1beta1.CreateNamespaceRequest) (*compassv1beta1.CreateNamespaceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var metadata map[string]interface{}
	if req.GetMetadata() != nil {
		metadata = req.GetMetadata().AsMap()
	}
	namespaceID := uuid.New()
	if id, err := uuid.Parse(req.GetId()); err == nil {
		namespaceID = id
	}
	ns := &namespace.Namespace{
		ID:       namespaceID,
		Name:     req.GetName(),
		State:    namespace.State(req.GetState()),
		Metadata: metadata,
	}
	nsID, err := server.namespaceService.Create(ctx, ns)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &compassv1beta1.CreateNamespaceResponse{
		Id: nsID,
	}, nil
}

func (server *APIServer) UpdateNamespace(ctx context.Context, req *compassv1beta1.UpdateNamespaceRequest) (*compassv1beta1.UpdateNamespaceResponse, error) {
	var nsID uuid.UUID
	var nsName string
	if id, err := uuid.Parse(req.GetUrn()); err == nil {
		nsID = id
	} else {
		nsName = req.GetUrn()
	}

	var metadata map[string]interface{}
	if req.GetMetadata() != nil {
		metadata = req.GetMetadata().AsMap()
	}
	ns := &namespace.Namespace{
		ID:       nsID,
		Name:     nsName,
		State:    namespace.State(req.GetState()),
		Metadata: metadata,
	}

	if err := server.namespaceService.Update(ctx, ns); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &compassv1beta1.UpdateNamespaceResponse{}, nil
}

func namespaceToProto(ns *namespace.Namespace) (*compassv1beta1.Namespace, error) {
	meta, err := structpb.NewStruct(ns.Metadata)
	if err != nil {
		return nil, err
	}
	return &compassv1beta1.Namespace{
		Id:       ns.ID.String(),
		Name:     ns.Name,
		State:    ns.State.String(),
		Metadata: meta,
	}, nil
}
