package handlersv1beta1

import (
	"context"
	"fmt"
	"github.com/raystack/compass/pkg/grpc_interceptor"
	"strings"

	"github.com/raystack/compass/core/asset"
	compassv1beta1 "github.com/raystack/compass/proto/raystack/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *APIServer) SearchAssets(ctx context.Context, req *compassv1beta1.SearchAssetsRequest) (*compassv1beta1.SearchAssetsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}
	ns := grpc_interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	text := strings.TrimSpace(req.GetText())
	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "'text' must be specified")
	}

	cfg := asset.SearchConfig{
		Text:       text,
		MaxResults: int(req.GetSize()),
		Filters:    filterConfigFromValues(req.GetFilter()),
		RankBy:     req.GetRankby(),
		Queries:    req.GetQuery(),
		Namespace:  ns,
	}

	results, err := server.assetService.SearchAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("error searching asset: %s", err.Error()))
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, sr := range results {
		assetPB, err := assetToProto(sr.ToAsset(), false)
		if err != nil {
			return nil, internalServerError(server.logger, fmt.Sprintf("error converting assets to proto: %s", err.Error()))
		}
		assetsPB = append(assetsPB, assetPB)
	}

	return &compassv1beta1.SearchAssetsResponse{
		Data: assetsPB,
	}, nil
}

func (server *APIServer) SuggestAssets(ctx context.Context, req *compassv1beta1.SuggestAssetsRequest) (*compassv1beta1.SuggestAssetsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, bodyParserErrorMsg(err))
	}
	ns := grpc_interceptor.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	text := strings.TrimSpace(req.GetText())
	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "'text' must be specified")
	}

	cfg := asset.SearchConfig{
		Text:      text,
		Namespace: ns,
	}
	suggestions, err := server.assetService.SuggestAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(server.logger, err.Error())
	}

	return &compassv1beta1.SuggestAssetsResponse{
		Data: suggestions,
	}, nil
}

func filterConfigFromValues(fltMap map[string]string) map[string][]string {
	var filter = make(map[string][]string)
	for key, value := range fltMap {
		var filterValues []string
		filterValues = append(filterValues, strings.Split(value, ",")...)

		filter[key] = filterValues
	}
	return filter
}
