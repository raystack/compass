package handler

//go:generate mockery --name=NamespaceService -r --case underscore --with-expecter --structname NamespaceService --filename namespace_service.go --output=./mocks

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/compass/core/namespace"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
)

type NamespaceService interface {
	Create(ctx context.Context, ns *namespace.Namespace) (string, error)
	Update(ctx context.Context, ns *namespace.Namespace) error
	GetByID(ctx context.Context, id uuid.UUID) (*namespace.Namespace, error)
	GetByName(ctx context.Context, name string) (*namespace.Namespace, error)
	List(ctx context.Context) ([]*namespace.Namespace, error)
}

func (server *Handler) ListNamespaces(ctx context.Context, req *connect.Request[compassv1beta1.ListNamespacesRequest]) (*connect.Response[compassv1beta1.ListNamespacesResponse], error) {
	namespaces, err := server.namespaceService.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	var protoNamespaces []*compassv1beta1.Namespace
	for _, ns := range namespaces {
		protoNamespace, err := namespaceToProto(ns)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		protoNamespaces = append(protoNamespaces, protoNamespace)
	}
	return connect.NewResponse(&compassv1beta1.ListNamespacesResponse{
		Namespaces: protoNamespaces,
	}), nil
}

func (server *Handler) GetNamespace(ctx context.Context, req *connect.Request[compassv1beta1.GetNamespaceRequest]) (*connect.Response[compassv1beta1.GetNamespaceResponse], error) {
	if req.Msg.GetUrn() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("urn is required"))
	}

	var ns *namespace.Namespace
	if nsID, err := uuid.Parse(req.Msg.GetUrn()); err == nil {
		ns, err = server.namespaceService.GetByID(ctx, nsID)
		if err != nil {
			if errors.Is(err, namespace.ErrNotFound) {
				return nil, connect.NewError(connect.CodeNotFound, err)
			}
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	} else {
		ns, err = server.namespaceService.GetByName(ctx, req.Msg.GetUrn())
		if err != nil {
			if errors.Is(err, namespace.ErrNotFound) {
				return nil, connect.NewError(connect.CodeNotFound, err)
			}
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	protoNamespace, err := namespaceToProto(ns)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&compassv1beta1.GetNamespaceResponse{
		Namespace: protoNamespace,
	}), nil
}

func (server *Handler) CreateNamespace(ctx context.Context, req *connect.Request[compassv1beta1.CreateNamespaceRequest]) (*connect.Response[compassv1beta1.CreateNamespaceResponse], error) {
	var metadata map[string]interface{}
	if req.Msg.GetMetadata() != nil {
		metadata = req.Msg.GetMetadata().AsMap()
	}
	namespaceID := uuid.New()
	if id, err := uuid.Parse(req.Msg.GetId()); err == nil {
		namespaceID = id
	}
	ns := &namespace.Namespace{
		ID:       namespaceID,
		Name:     req.Msg.GetName(),
		State:    namespace.State(req.Msg.GetState()),
		Metadata: metadata,
	}
	nsID, err := server.namespaceService.Create(ctx, ns)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&compassv1beta1.CreateNamespaceResponse{
		Id: nsID,
	}), nil
}

func (server *Handler) UpdateNamespace(ctx context.Context, req *connect.Request[compassv1beta1.UpdateNamespaceRequest]) (*connect.Response[compassv1beta1.UpdateNamespaceResponse], error) {
	if req.Msg.GetUrn() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("urn is required"))
	}

	var nsID uuid.UUID
	var nsName string
	if id, err := uuid.Parse(req.Msg.GetUrn()); err == nil {
		nsID = id
	} else {
		nsName = req.Msg.GetUrn()
	}

	var metadata map[string]interface{}
	if req.Msg.GetMetadata() != nil {
		metadata = req.Msg.GetMetadata().AsMap()
	}
	ns := &namespace.Namespace{
		ID:       nsID,
		Name:     nsName,
		State:    namespace.State(req.Msg.GetState()),
		Metadata: metadata,
	}

	if err := server.namespaceService.Update(ctx, ns); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&compassv1beta1.UpdateNamespaceResponse{}), nil
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
