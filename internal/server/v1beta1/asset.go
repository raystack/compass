package handlersv1beta1

//go:generate mockery --name=AssetService -r --case underscore --with-expecter --structname AssetService --filename asset_service.go --output=./mocks
import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/core/asset"
	"github.com/odpf/compass/core/star"
	"github.com/odpf/compass/core/user"
	"github.com/r3labs/diff/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AssetService interface {
	GetAllAssets(context.Context, asset.Filter, bool) ([]asset.Asset, uint32, error)
	GetAssetByID(ctx context.Context, id string) (asset.Asset, error)
	GetAssetByURN(ctx context.Context, urn string, typ asset.Type, service string) (asset.Asset, error)
	GetAssetByVersion(ctx context.Context, id string, version string) (asset.Asset, error)
	GetAssetVersionHistory(ctx context.Context, flt asset.Filter, id string) ([]asset.Asset, error)
	UpsertPatchAsset(context.Context, *asset.Asset, []asset.LineageNode, []asset.LineageNode) (string, error)
	DeleteAsset(context.Context, string) error

	GetLineage(ctx context.Context, node asset.LineageNode) (asset.LineageGraph, error)
	GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error)

	SearchAssets(ctx context.Context, cfg asset.SearchConfig) (results []asset.SearchResult, err error)
	SuggestAssets(ctx context.Context, cfg asset.SearchConfig) (suggestions []string, err error)
}

func (server *APIServer) GetAllAssets(ctx context.Context, req *compassv1beta1.GetAllAssetsRequest) (*compassv1beta1.GetAllAssetsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	flt, err := server.buildAssetsFilter(
		req.GetTypes(),
		req.GetServices(),
		req.GetQ(),
		req.GetQFields(),
		int(req.GetSize()),
		int(req.GetOffset()),
		req.GetSort(),
		req.GetDirection(),
		req.GetData())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}

	assets, totalCount, err := server.assetService.GetAllAssets(ctx, flt, req.GetWithTotal())
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	assetsProto := []*compassv1beta1.Asset{}
	for _, a := range assets {
		ap, err := assetToProto(a, false)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		assetsProto = append(assetsProto, ap)
	}

	response := &compassv1beta1.GetAllAssetsResponse{
		Data: assetsProto,
	}

	if req.GetWithTotal() {
		response.Total = totalCount
	}

	return response, nil
}

func (server *APIServer) buildAssetsFilter(
	types string,
	services string,
	q string,
	qFields string,
	size int,
	offset int,
	sortBy string,
	sortDirection string,
	data map[string]string,
) (asset.Filter, error) {

	flt := asset.Filter{
		Size:          size,
		Offset:        offset,
		SortBy:        sortBy,
		SortDirection: sortDirection,
		Query:         q,
		Data:          data,
	}

	if types != "" {
		typs := strings.Split(types, ",")
		for _, typeVal := range typs {
			flt.Types = append(flt.Types, asset.Type(typeVal))
		}
	}
	if services != "" {
		flt.Services = strings.Split(services, ",")
	}

	if qFields != "" {
		flt.QueryFields = strings.Split(qFields, ",")
	}

	if err := flt.Validate(); err != nil {
		return asset.Filter{}, err
	}

	return flt, nil
}

// func (server *APIServer) buildGetAllAssetsFilter(req *compassv1beta1.GetAllAssetsRequest) (asset.Filter, error) {

// 	flt := asset.Filter{
// 		Size:          int(req.GetSize()),
// 		Offset:        int(req.GetOffset()),
// 		SortBy:        req.GetSort(),
// 		SortDirection: req.GetDirection(),
// 		Query:         req.GetQ(),
// 		Data:          req.GetData(),
// 	}

// 	if req.GetTypes() != "" {
// 		typs := strings.Split(req.GetTypes(), ",")
// 		for _, typeVal := range typs {
// 			flt.Types = append(flt.Types, asset.Type(typeVal))
// 		}
// 	}
// 	if req.GetServices() != "" {
// 		flt.Services = strings.Split(req.GetServices(), ",")
// 	}

// 	if req.GetQFields() != "" {
// 		flt.QueryFields = strings.Split(req.GetQFields(), ",")
// 	}

// 	if err := flt.Validate(); err != nil {
// 		return asset.Filter{}, err
// 	}

// 	return flt, nil
// }

func (server *APIServer) GetAssetByID(ctx context.Context, req *compassv1beta1.GetAssetByIDRequest) (*compassv1beta1.GetAssetByIDResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	ast, err := server.assetService.GetAssetByID(ctx, req.GetId())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	astProto, err := assetToProto(ast, false)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetAssetByIDResponse{
		Data: astProto,
	}, nil
}

func (server *APIServer) GetAssetStargazers(ctx context.Context, req *compassv1beta1.GetAssetStargazersRequest) (*compassv1beta1.GetAssetStargazersResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	users, err := server.starService.GetStargazers(ctx, star.Filter{
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
		return nil, internalServerError(server.logger, err.Error())
	}

	usersPB := []*compassv1beta1.User{}
	for _, us := range users {
		usersPB = append(usersPB, userToProto(us))
	}

	return &compassv1beta1.GetAssetStargazersResponse{
		Data: usersPB,
	}, nil
}

func (server *APIServer) GetAssetVersionHistory(ctx context.Context, req *compassv1beta1.GetAssetVersionHistoryRequest) (*compassv1beta1.GetAssetVersionHistoryResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	assetVersions, err := server.assetService.GetAssetVersionHistory(ctx, asset.Filter{
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
		return nil, internalServerError(server.logger, err.Error())
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, av := range assetVersions {
		avPB, err := assetToProto(av, true)
		if err != nil {
			return nil, internalServerError(server.logger, err.Error())
		}
		assetsPB = append(assetsPB, avPB)
	}

	return &compassv1beta1.GetAssetVersionHistoryResponse{
		Data: assetsPB,
	}, nil
}

func (server *APIServer) GetAssetByVersion(ctx context.Context, req *compassv1beta1.GetAssetByVersionRequest) (*compassv1beta1.GetAssetByVersionResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := asset.ParseVersion(req.GetVersion()); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	ast, err := server.assetService.GetAssetByVersion(ctx, req.GetId(), req.GetVersion())
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	assetPB, err := assetToProto(ast, true)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.GetAssetByVersionResponse{
		Data: assetPB,
	}, nil
}

func (server *APIServer) UpsertPatchAsset(ctx context.Context, req *compassv1beta1.UpsertPatchAssetRequest) (*compassv1beta1.UpsertPatchAssetResponse, error) {
	userID, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	baseAsset := req.GetAsset()
	if baseAsset == nil {
		return nil, status.Error(codes.InvalidArgument, "asset cannot be empty")
	}

	urn, typ, service, err := server.validatePatchAsset(baseAsset)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ast, err := server.assetService.GetAssetByURN(ctx, urn, asset.Type(typ), service)
	if err != nil && !errors.As(err, &asset.NotFoundError{}) {
		return nil, internalServerError(server.logger, err.Error())
	}

	patchAssetMap := decodePatchAssetToMap(baseAsset)
	ast.Patch(patchAssetMap)

	if err := server.validateAsset(ast); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ast.UpdatedBy.ID = userID
	upstreams := []asset.LineageNode{}
	for _, pb := range req.GetUpstreams() {
		upstreams = append(upstreams, lineageNodeFromProto(pb))
	}
	downstreams := []asset.LineageNode{}
	for _, pb := range req.GetDownstreams() {
		downstreams = append(downstreams, lineageNodeFromProto(pb))
	}

	assetID, err := server.assetService.UpsertPatchAsset(ctx, &ast, upstreams, downstreams)
	if err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.UpsertPatchAssetResponse{
		Id: assetID,
	}, nil
}

func (server *APIServer) DeleteAsset(ctx context.Context, req *compassv1beta1.DeleteAssetRequest) (*compassv1beta1.DeleteAssetResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := server.assetService.DeleteAsset(ctx, req.GetId()); err != nil {
		if errors.As(err, new(asset.InvalidError)) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.As(err, new(asset.NotFoundError)) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.DeleteAssetResponse{}, nil
}

func (server *APIServer) validateAsset(ast asset.Asset) error {
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

func (server *APIServer) validatePatchAsset(ast *compassv1beta1.UpsertPatchAssetRequest_BaseAsset) (urn, typ, service string, err error) {
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

// assetToProto transforms struct to proto
func assetToProto(a asset.Asset, withChangelog bool) (assetPB *compassv1beta1.Asset, err error) {
	var data *structpb.Struct
	if len(a.Data) > 0 {
		data, err = structpb.NewStruct(a.Data)
		if err != nil {
			return
		}
	}

	var labels *structpb.Struct
	if len(a.Labels) > 0 {
		labelsMapInterface := make(map[string]interface{}, len(a.Labels))
		for k, v := range a.Labels {
			labelsMapInterface[k] = v
		}
		labels, err = structpb.NewStruct(labelsMapInterface)
		if err != nil {
			return
		}
	}

	owners := []*compassv1beta1.User{}
	for _, o := range a.Owners {
		owners = append(owners, userToProto(o))
	}

	var changelogProto []*compassv1beta1.Change
	if withChangelog {
		changelogProto, err = changelogToProto(a.Changelog)
		if err != nil {
			return nil, err
		}
	}

	var createdAtPB *timestamppb.Timestamp
	if !a.CreatedAt.IsZero() {
		createdAtPB = timestamppb.New(a.CreatedAt)
	}

	var updatedAtPB *timestamppb.Timestamp
	if !a.UpdatedAt.IsZero() {
		updatedAtPB = timestamppb.New(a.UpdatedAt)
	}

	assetPB = &compassv1beta1.Asset{
		Id:          a.ID,
		Urn:         a.URN,
		Type:        string(a.Type),
		Service:     a.Service,
		Name:        a.Name,
		Description: a.Description,
		Data:        data,
		Labels:      labels,
		Owners:      owners,
		Version:     a.Version,
		UpdatedBy:   userToProto(a.UpdatedBy),
		Changelog:   changelogProto,
		CreatedAt:   createdAtPB,
		UpdatedAt:   updatedAtPB,
	}
	return
}

// changelogToProto transforms changelog struct to proto
func changelogToProto(cl diff.Changelog) ([]*compassv1beta1.Change, error) {
	if len(cl) == 0 {
		return nil, nil
	}
	var protoChanges []*compassv1beta1.Change
	for _, ch := range cl {
		chProto, err := diffChangeToProto(ch)
		if err != nil {
			return nil, err
		}

		protoChanges = append(protoChanges, chProto)
	}
	return protoChanges, nil
}

func diffChangeToProto(dc diff.Change) (*compassv1beta1.Change, error) {
	from, err := structpb.NewValue(dc.From)
	if err != nil {
		return nil, err
	}
	to, err := structpb.NewValue(dc.To)
	if err != nil {
		return nil, err
	}

	return &compassv1beta1.Change{
		Type: dc.Type,
		Path: dc.Path,
		From: from,
		To:   to,
	}, nil
}

// assetFromProto transforms proto to struct
// changelog is not populated by user, it should always be processed and coming from the server
func assetFromProto(pb *compassv1beta1.Asset) asset.Asset {
	var assetOwners []user.User
	for _, op := range pb.GetOwners() {
		if op == nil {
			continue
		}
		assetOwners = append(assetOwners, userFromProto(op))
	}

	var labels map[string]string
	if pb.GetLabels() != nil {
		labels = make(map[string]string)
		for key, value := range pb.GetLabels().AsMap() {
			strKey := fmt.Sprintf("%v", key)
			strValue := fmt.Sprintf("%v", value)
			labels[strKey] = strValue
		}
	}

	var dataValue map[string]interface{}
	if pb.GetData() != nil {
		dataValue = pb.GetData().AsMap()
	}

	var createdAt time.Time
	if pb.GetCreatedAt() != nil {
		createdAt = pb.GetCreatedAt().AsTime()
	}

	var updatedAt time.Time
	if pb.GetUpdatedAt() != nil {
		updatedAt = pb.GetUpdatedAt().AsTime()
	}

	var updatedBy user.User
	if pb.GetUpdatedBy() != nil {
		updatedBy = userFromProto(pb.GetUpdatedBy())
	}

	var clog diff.Changelog
	if len(pb.GetChangelog()) > 0 {
		for _, cg := range pb.GetChangelog() {
			if cg == nil {
				continue
			}
			clog = append(clog, diffChangeFromProto(cg))
		}
	}

	return asset.Asset{
		ID:          pb.GetId(),
		URN:         pb.GetUrn(),
		Type:        asset.Type(pb.GetType()),
		Service:     pb.GetService(),
		Name:        pb.GetName(),
		Description: pb.GetDescription(),
		Data:        dataValue,
		Labels:      labels,
		Owners:      assetOwners,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Version:     pb.GetVersion(),
		Changelog:   clog,
		UpdatedBy:   updatedBy,
	}
}

// diffChangeFromProto converts Change proto to diff.Change
func diffChangeFromProto(pb *compassv1beta1.Change) diff.Change {
	var fromItf interface{}
	if pb.GetFrom() != nil {
		fromItf = pb.GetFrom().AsInterface()
	}

	var toItf interface{}
	if pb.GetTo() != nil {
		toItf = pb.GetTo().AsInterface()
	}

	return diff.Change{
		Type: pb.GetType(),
		Path: pb.GetPath(),
		From: fromItf,
		To:   toItf,
	}
}
