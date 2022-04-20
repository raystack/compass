package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"strings"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/asset"
	"github.com/odpf/compass/lineage"
	"github.com/odpf/compass/star"
	"github.com/odpf/compass/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) GetAllAssets(ctx context.Context, req *compassv1beta1.GetAllAssetsRequest) (*compassv1beta1.GetAllAssetsResponse, error) {

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	flt, err := h.buildGetAllAssetsFilter(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	assets, err := h.AssetRepository.GetAll(ctx, flt)
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	assetsProto := []*compassv1beta1.Asset{}
	for _, a := range assets {
		ap, err := a.ToProto(false)
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
		}
		assetsProto = append(assetsProto, ap)
	}

	response := &compassv1beta1.GetAllAssetsResponse{
		Data: assetsProto,
	}

	if req.GetWithTotal() {
		total, err := h.AssetRepository.GetCount(ctx, flt)
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
		}
		response.Total = uint32(total)
	}

	return response, nil
}

func (h *Handler) buildGetAllAssetsFilter(req *compassv1beta1.GetAllAssetsRequest) (asset.Filter, error) {

	flt := asset.Filter{
		Size:          int(req.GetSize()),
		Offset:        int(req.GetOffset()),
		SortBy:        req.GetSort(),
		SortDirection: req.GetDirection(),
		Query:         req.GetQ(),
		Data:          req.GetData(),
	}

	if req.GetTypes() != "" {
		typs := strings.Split(req.GetTypes(), ",")
		for _, typeVal := range typs {
			flt.Types = append(flt.Types, asset.Type(typeVal))
		}
	}
	if req.GetServices() != "" {
		flt.Services = strings.Split(req.GetServices(), ",")
	}

	if req.GetQFields() != "" {
		flt.QueryFields = strings.Split(req.GetQFields(), ",")
	}

	if err := flt.Validate(); err != nil {
		return asset.Filter{}, err
	}

	return flt, nil
}

func (h *Handler) GetAssetByID(ctx context.Context, req *compassv1beta1.GetAssetByIDRequest) (*compassv1beta1.GetAssetByIDResponse, error) {
	ast, err := h.AssetRepository.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(h.Logger, err.Error())
	}

	astProto, err := ast.ToProto(false)
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.GetAssetByIDResponse{
		Data: astProto,
	}, nil
}

func (h *Handler) UpsertAsset(ctx context.Context, req *compassv1beta1.UpsertAssetRequest) (*compassv1beta1.UpsertAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	baseAsset := req.GetAsset()
	if baseAsset == nil {
		return nil, status.Error(codes.InvalidArgument, "asset cannot be empty")
	}

	ast := asset.Asset{
		URN:         baseAsset.GetUrn(),
		Type:        asset.Type(baseAsset.GetType()),
		Name:        baseAsset.GetName(),
		Service:     baseAsset.GetService(),
		Description: baseAsset.GetDescription(),
	}
	ast.AssignDataFromProto(baseAsset.GetData())
	ast.AssignLabelsFromProto(baseAsset.GetLabels())

	if err := h.validateAsset(ast); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ast.UpdatedBy.ID = userID
	assetID, err := h.AssetRepository.Upsert(ctx, &ast)
	if errors.As(err, new(asset.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	ast.ID = assetID
	if err := h.DiscoveryRepository.Upsert(ctx, ast); err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	if err := h.saveLineage(ctx, ast, req.GetUpstreams(), req.GetDownstreams()); err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.UpsertAssetResponse{
		Id: assetID,
	}, nil
}

func (h *Handler) UpsertPatchAsset(ctx context.Context, req *compassv1beta1.UpsertPatchAssetRequest) (*compassv1beta1.UpsertPatchAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	baseAsset := req.GetAsset()
	if baseAsset == nil {
		return nil, status.Error(codes.InvalidArgument, "asset cannot be empty")
	}

	urn, typ, service, err := h.validatePatchAsset(baseAsset)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ast, err := h.AssetRepository.Find(ctx, urn, asset.Type(typ), service)
	if err != nil && !errors.As(err, &asset.NotFoundError{}) {
		return nil, internalServerError(h.Logger, err.Error())
	}

	patchAssetMap := decodePatchAssetToMap(baseAsset)
	ast.Patch(patchAssetMap)

	if err := h.validateAsset(ast); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ast.UpdatedBy.ID = userID
	assetID, err := h.AssetRepository.Upsert(ctx, &ast)
	if errors.As(err, new(asset.InvalidError)) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	ast.ID = assetID
	if err := h.DiscoveryRepository.Upsert(ctx, ast); err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	if err := h.saveLineage(ctx, ast, req.GetUpstreams(), req.GetDownstreams()); err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.UpsertPatchAssetResponse{
		Id: ast.ID,
	}, nil
}

func (h *Handler) DeleteAsset(ctx context.Context, req *compassv1beta1.DeleteAssetRequest) (*compassv1beta1.DeleteAssetResponse, error) {
	userID := user.FromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, errMissingUserInfo.Error())
	}

	if err := h.AssetRepository.Delete(ctx, req.GetId()); err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(h.Logger, err.Error())
	}

	if err := h.DiscoveryRepository.Delete(ctx, req.GetId()); err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.DeleteAssetResponse{}, nil
}

func (h *Handler) GetAssetStargazers(ctx context.Context, req *compassv1beta1.GetAssetStargazersRequest) (*compassv1beta1.GetAssetStargazersResponse, error) {

	users, err := h.StarRepository.GetStargazers(ctx, star.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}, req.GetId())
	if err != nil {
		if errors.Is(err, star.ErrEmptyUserID) || errors.Is(err, star.ErrEmptyAssetID) || errors.As(err, new(star.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(star.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(h.Logger, err.Error())
	}

	usersPB := []*compassv1beta1.User{}
	for _, us := range users {
		usersPB = append(usersPB, us.ToProto())
	}

	return &compassv1beta1.GetAssetStargazersResponse{
		Data: usersPB,
	}, nil
}

func (h *Handler) GetAssetVersionHistory(ctx context.Context, req *compassv1beta1.GetAssetVersionHistoryRequest) (*compassv1beta1.GetAssetVersionHistoryResponse, error) {
	assetVersions, err := h.AssetRepository.GetVersionHistory(ctx, asset.Filter{
		Size:   int(req.GetSize()),
		Offset: int(req.GetOffset()),
	}, req.GetId())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(h.Logger, err.Error())
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, av := range assetVersions {
		avPB, err := av.ToProto(true)
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
		}
		assetsPB = append(assetsPB, avPB)
	}

	return &compassv1beta1.GetAssetVersionHistoryResponse{
		Data: assetsPB,
	}, nil
}

func (h *Handler) GetAssetByVersion(ctx context.Context, req *compassv1beta1.GetAssetByVersionRequest) (*compassv1beta1.GetAssetByVersionResponse, error) {
	if _, err := asset.ParseVersion(req.GetVersion()); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	ast, err := h.AssetRepository.GetByVersion(ctx, req.GetId(), req.GetVersion())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(h.Logger, err.Error())
	}

	assetPB, err := ast.ToProto(true)
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.GetAssetByVersionResponse{
		Data: assetPB,
	}, nil
}

func (h *Handler) validateAsset(ast asset.Asset) error {
	if ast.URN == "" {
		return fmt.Errorf("urn is required")
	}
	if ast.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !ast.Type.IsValid() {
		return fmt.Errorf("type is invalid")
	}
	if ast.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ast.Data == nil {
		return fmt.Errorf("data is required")
	}
	if ast.Service == "" {
		return fmt.Errorf("service is required")
	}

	return nil
}

func (h *Handler) validatePatchAsset(ast *compassv1beta1.UpsertPatchAssetRequest_BaseAsset) (urn, typ, service string, err error) {
	if urn = ast.GetUrn(); urn == "" {
		err = fmt.Errorf("urn is required and can't be empty")
		return
	}

	if typ = ast.GetType(); typ == "" {
		err = fmt.Errorf("type is required and can't be empty")
		return
	}

	if !asset.Type(typ).IsValid() {
		err = fmt.Errorf("type is invalid")
		return
	}

	if service = ast.GetService(); service == "" {
		err = fmt.Errorf("service is required and can't be empty")
		return
	}

	return
}

func (h *Handler) saveLineage(ctx context.Context, ast asset.Asset, upstreamsPB, downstreamsPB []*compassv1beta1.LineageNode) error {
	upstreams := []lineage.Node{}
	for _, pb := range upstreamsPB {
		upstreams = append(upstreams, lineage.NewNodeFromProto(pb))
	}
	downstreams := []lineage.Node{}
	for _, pb := range downstreamsPB {
		downstreams = append(downstreams, lineage.NewNodeFromProto(pb))
	}
	node := lineage.Node{
		URN:     ast.URN,
		Type:    ast.Type,
		Service: ast.Service,
	}

	return h.LineageRepository.Upsert(ctx, node, upstreams, downstreams)
}

func decodePatchAssetToMap(pb *compassv1beta1.UpsertPatchAssetRequest_BaseAsset) map[string]interface{} {
	if pb == nil {
		return nil
	}
	m := map[string]interface{}{}
	m["urn"] = pb.GetUrn()
	m["type"] = pb.GetType()
	m["service"] = pb.GetService()
	if pb.GetName() != nil {
		m["name"] = pb.GetName().Value
	}
	if pb.GetDescription() != nil {
		m["description"] = pb.GetDescription().Value
	}
	if pb.GetData() != nil {
		m["data"] = pb.GetData().AsMap()
	}
	if pb.GetLabels() != nil {
		m["labels"] = pb.GetLabels().AsMap()
	}
	if len(pb.GetOwners()) > 0 {
		ownersMap := []map[string]interface{}{}
		ownersPB := pb.GetOwners()
		for _, ownerPB := range ownersPB {
			ownerMap := map[string]interface{}{}
			if ownerPB.GetId() != "" {
				ownerMap["id"] = ownerPB.GetId()
			}
			if ownerPB.GetUuid() != "" {
				ownerMap["uuid"] = ownerPB.GetUuid()
			}
			if ownerPB.GetEmail() != "" {
				ownerMap["email"] = ownerPB.GetEmail()
			}
			if ownerPB.GetProvider() != "" {
				ownerMap["provider"] = ownerPB.GetProvider()
			}
			ownersMap = append(ownersMap, ownerMap)
		}
		m["owners"] = ownersMap
	}

	return m
}
