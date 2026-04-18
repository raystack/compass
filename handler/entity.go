package handler

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/compass/internal/middleware"
	"github.com/raystack/compass/core/entity"
	"github.com/raystack/compass/core/namespace"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EntityServiceV2 defines entity operations for the handler.
type EntityServiceV2 interface {
	Upsert(ctx context.Context, ns *namespace.Namespace, ent *entity.Entity) (string, error)
	GetByURN(ctx context.Context, ns *namespace.Namespace, urn string) (entity.Entity, error)
	GetByID(ctx context.Context, id string) (entity.Entity, error)
	GetAll(ctx context.Context, ns *namespace.Namespace, flt entity.Filter) ([]entity.Entity, int, error)
	GetTypes(ctx context.Context, ns *namespace.Namespace) (map[entity.Type]int, error)
	Delete(ctx context.Context, ns *namespace.Namespace, urn string) error
	Search(ctx context.Context, cfg entity.SearchConfig) ([]entity.SearchResult, error)
	Suggest(ctx context.Context, ns *namespace.Namespace, text string, limit int) ([]string, error)
	GetContext(ctx context.Context, ns *namespace.Namespace, urn string, depth int) (*entity.ContextGraph, error)
	GetImpact(ctx context.Context, ns *namespace.Namespace, urn string, depth int) ([]entity.Edge, error)
	AssembleContext(ctx context.Context, ns *namespace.Namespace, req entity.AssemblyRequest) (*entity.AssembledContext, error)
}

// EdgeServiceV2 defines edge operations for the handler.
type EdgeServiceV2 interface {
	Upsert(ctx context.Context, ns *namespace.Namespace, e *entity.Edge) error
	GetBySource(ctx context.Context, ns *namespace.Namespace, urn string, filter entity.EdgeFilter) ([]entity.Edge, error)
	GetByTarget(ctx context.Context, ns *namespace.Namespace, urn string, filter entity.EdgeFilter) ([]entity.Edge, error)
	Delete(ctx context.Context, ns *namespace.Namespace, sourceURN, targetURN, edgeType string) error
}

func (server *Handler) GetAllEntities(ctx context.Context, req *connect.Request[compassv1beta1.GetAllEntitiesRequest]) (*connect.Response[compassv1beta1.GetAllEntitiesResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	flt := entity.Filter{
		Size:   int(req.Msg.GetSize()),
		Offset: int(req.Msg.GetOffset()),
		Query:  req.Msg.GetQ(),
	}
	if types := req.Msg.GetTypes(); types != "" {
		for _, t := range strings.Split(types, ",") {
			flt.Types = append(flt.Types, entity.Type(strings.TrimSpace(t)))
		}
	}
	if src := req.Msg.GetSource(); src != "" {
		flt.Source = src
	}

	entities, total, err := server.entityService.GetAll(ctx, ns, flt)
	if err != nil {
		return nil, internalServerError(ctx, "error getting entities", err)
	}

	data := make([]*compassv1beta1.Entity, len(entities))
	for i, e := range entities {
		data[i] = entityToProto(e)
	}

	return connect.NewResponse(&compassv1beta1.GetAllEntitiesResponse{
		Data:  data,
		Total: uint32(total),
	}), nil
}

func (server *Handler) GetEntityByID(ctx context.Context, req *connect.Request[compassv1beta1.GetEntityByIDRequest]) (*connect.Response[compassv1beta1.GetEntityByIDResponse], error) {
	ent, err := server.entityService.GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, internalServerError(ctx, "error getting entity", err)
	}
	return connect.NewResponse(&compassv1beta1.GetEntityByIDResponse{
		Data: entityToProto(ent),
	}), nil
}

func (server *Handler) UpsertEntity(ctx context.Context, req *connect.Request[compassv1beta1.UpsertEntityRequest]) (*connect.Response[compassv1beta1.UpsertEntityResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	ent := &entity.Entity{
		URN:         req.Msg.GetUrn(),
		Type:        entity.Type(req.Msg.GetType()),
		Name:        req.Msg.GetName(),
		Description: req.Msg.GetDescription(),
		Source:      req.Msg.GetSource(),
	}
	if req.Msg.GetProperties() != nil {
		ent.Properties = req.Msg.GetProperties().AsMap()
	}

	id, err := server.entityService.Upsert(ctx, ns, ent)
	if err != nil {
		return nil, internalServerError(ctx, "error upserting entity", err)
	}

	return connect.NewResponse(&compassv1beta1.UpsertEntityResponse{Id: id}), nil
}

func (server *Handler) DeleteEntity(ctx context.Context, req *connect.Request[compassv1beta1.DeleteEntityRequest]) (*connect.Response[compassv1beta1.DeleteEntityResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	if err := server.entityService.Delete(ctx, ns, req.Msg.GetUrn()); err != nil {
		return nil, internalServerError(ctx, "error deleting entity", err)
	}
	return connect.NewResponse(&compassv1beta1.DeleteEntityResponse{}), nil
}

func (server *Handler) SearchEntities(ctx context.Context, req *connect.Request[compassv1beta1.SearchEntitiesRequest]) (*connect.Response[compassv1beta1.SearchEntitiesResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	cfg := entity.SearchConfig{
		Text:       req.Msg.GetText(),
		MaxResults: int(req.Msg.GetSize()),
		Offset:     int(req.Msg.GetOffset()),
		Mode:       entity.SearchMode(req.Msg.GetMode()),
		Namespace:  ns,
	}
	if types := req.Msg.GetTypes(); types != "" {
		cfg.Filters = map[string][]string{"type": strings.Split(types, ",")}
	}
	if src := req.Msg.GetSource(); src != "" {
		if cfg.Filters == nil {
			cfg.Filters = make(map[string][]string)
		}
		cfg.Filters["source"] = []string{src}
	}

	results, err := server.entityService.Search(ctx, cfg)
	if err != nil {
		return nil, internalServerError(ctx, "error searching entities", err)
	}

	data := make([]*compassv1beta1.Entity, len(results))
	for i, r := range results {
		data[i] = &compassv1beta1.Entity{
			Id:          r.ID,
			Urn:         r.URN,
			Type:        r.Type,
			Name:        r.Name,
			Source:      r.Source,
			Description: r.Description,
		}
	}
	return connect.NewResponse(&compassv1beta1.SearchEntitiesResponse{Data: data}), nil
}

func (server *Handler) SuggestEntities(ctx context.Context, req *connect.Request[compassv1beta1.SuggestEntitiesRequest]) (*connect.Response[compassv1beta1.SuggestEntitiesResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	suggestions, err := server.entityService.Suggest(ctx, ns, req.Msg.GetText(), 5)
	if err != nil {
		return nil, internalServerError(ctx, "error suggesting entities", err)
	}
	return connect.NewResponse(&compassv1beta1.SuggestEntitiesResponse{Data: suggestions}), nil
}

func (server *Handler) GetEntityTypes(ctx context.Context, _ *connect.Request[compassv1beta1.GetEntityTypesRequest]) (*connect.Response[compassv1beta1.GetEntityTypesResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	types, err := server.entityService.GetTypes(ctx, ns)
	if err != nil {
		return nil, internalServerError(ctx, "error getting entity types", err)
	}

	data := make([]*compassv1beta1.Type, 0, len(types))
	for t, count := range types {
		data = append(data, &compassv1beta1.Type{Name: t.String(), Count: uint32(count)})
	}
	return connect.NewResponse(&compassv1beta1.GetEntityTypesResponse{Data: data}), nil
}

func (server *Handler) GetEntityContext(ctx context.Context, req *connect.Request[compassv1beta1.GetEntityContextRequest]) (*connect.Response[compassv1beta1.GetEntityContextResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	cg, err := server.entityService.GetContext(ctx, ns, req.Msg.GetUrn(), int(req.Msg.GetDepth()))
	if err != nil {
		return nil, internalServerError(ctx, "error getting entity context", err)
	}

	edges := make([]*compassv1beta1.Edge, len(cg.Edges))
	for i, e := range cg.Edges {
		edges[i] = edgeToProto(e)
	}
	related := make([]*compassv1beta1.Entity, len(cg.Related))
	for i, r := range cg.Related {
		related[i] = entityToProto(r)
	}

	return connect.NewResponse(&compassv1beta1.GetEntityContextResponse{
		Entity:  entityToProto(cg.Entity),
		Edges:   edges,
		Related: related,
	}), nil
}

func (server *Handler) GetEntityImpact(ctx context.Context, req *connect.Request[compassv1beta1.GetEntityImpactRequest]) (*connect.Response[compassv1beta1.GetEntityImpactResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	impactEdges, err := server.entityService.GetImpact(ctx, ns, req.Msg.GetUrn(), int(req.Msg.GetDepth()))
	if err != nil {
		return nil, internalServerError(ctx, "error analyzing impact", err)
	}

	edges := make([]*compassv1beta1.Edge, len(impactEdges))
	for i, e := range impactEdges {
		edges[i] = edgeToProto(e)
	}

	return connect.NewResponse(&compassv1beta1.GetEntityImpactResponse{Edges: edges}), nil
}

func (server *Handler) UpsertEdge(ctx context.Context, req *connect.Request[compassv1beta1.UpsertEdgeRequest]) (*connect.Response[compassv1beta1.UpsertEdgeResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	e := &entity.Edge{
		SourceURN: req.Msg.GetSourceUrn(),
		TargetURN: req.Msg.GetTargetUrn(),
		Type:      req.Msg.GetType(),
		Source:    req.Msg.GetSource(),
	}
	if req.Msg.GetProperties() != nil {
		e.Properties = req.Msg.GetProperties().AsMap()
	}

	if err := server.edgeService.Upsert(ctx, ns, e); err != nil {
		return nil, internalServerError(ctx, "error upserting edge", err)
	}
	return connect.NewResponse(&compassv1beta1.UpsertEdgeResponse{Id: e.ID}), nil
}

func (server *Handler) GetEdges(ctx context.Context, req *connect.Request[compassv1beta1.GetEdgesRequest]) (*connect.Response[compassv1beta1.GetEdgesResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	filter := entity.EdgeFilter{Current: req.Msg.GetCurrentOnly()}
	if t := req.Msg.GetType(); t != "" {
		filter.Types = []string{t}
	}

	var allEdges []entity.Edge
	dir := req.Msg.GetDirection()
	if dir == "" || dir == "both" || dir == "outgoing" {
		edges, err := server.edgeService.GetBySource(ctx, ns, req.Msg.GetUrn(), filter)
		if err != nil {
			return nil, internalServerError(ctx, "error getting outgoing edges", err)
		}
		allEdges = append(allEdges, edges...)
	}
	if dir == "" || dir == "both" || dir == "incoming" {
		edges, err := server.edgeService.GetByTarget(ctx, ns, req.Msg.GetUrn(), filter)
		if err != nil {
			return nil, internalServerError(ctx, "error getting incoming edges", err)
		}
		allEdges = append(allEdges, edges...)
	}

	data := make([]*compassv1beta1.Edge, len(allEdges))
	for i, e := range allEdges {
		data[i] = edgeToProto(e)
	}
	return connect.NewResponse(&compassv1beta1.GetEdgesResponse{Data: data}), nil
}

func (server *Handler) DeleteEdge(ctx context.Context, req *connect.Request[compassv1beta1.DeleteEdgeRequest]) (*connect.Response[compassv1beta1.DeleteEdgeResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)

	if err := server.edgeService.Delete(ctx, ns, req.Msg.GetSourceUrn(), req.Msg.GetTargetUrn(), req.Msg.GetType()); err != nil {
		return nil, internalServerError(ctx, "error deleting edge", err)
	}
	return connect.NewResponse(&compassv1beta1.DeleteEdgeResponse{}), nil
}

// Proto conversion helpers

func entityToProto(e entity.Entity) *compassv1beta1.Entity {
	pb := &compassv1beta1.Entity{
		Id:          e.ID,
		Urn:         e.URN,
		Type:        string(e.Type),
		Name:        e.Name,
		Description: e.Description,
		Source:      e.Source,
		ValidFrom:   timestamppb.New(e.ValidFrom),
		CreatedAt:   timestamppb.New(e.CreatedAt),
		UpdatedAt:   timestamppb.New(e.UpdatedAt),
	}
	if e.ValidTo != nil {
		pb.ValidTo = timestamppb.New(*e.ValidTo)
	}
	if len(e.Properties) > 0 {
		pb.Properties, _ = structpb.NewStruct(e.Properties)
	}
	return pb
}

func edgeToProto(e entity.Edge) *compassv1beta1.Edge {
	pb := &compassv1beta1.Edge{
		Id:        e.ID,
		SourceUrn: e.SourceURN,
		TargetUrn: e.TargetURN,
		Type:      e.Type,
		Source:    e.Source,
		ValidFrom: timestamppb.New(e.ValidFrom),
		CreatedAt: timestamppb.New(e.CreatedAt),
	}
	if e.ValidTo != nil {
		pb.ValidTo = timestamppb.New(*e.ValidTo)
	}
	if len(e.Properties) > 0 {
		pb.Properties, _ = structpb.NewStruct(e.Properties)
	}
	return pb
}
