package handler

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/compass/core/asset"
	"github.com/raystack/compass/internal/middleware"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
)

func (server *Handler) SearchAssets(ctx context.Context, req *connect.Request[compassv1beta1.SearchAssetsRequest]) (*connect.Response[compassv1beta1.SearchAssetsResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	cfg := asset.SearchConfig{
		Text:          strings.TrimSpace(req.Msg.GetText()),
		MaxResults:    int(req.Msg.GetSize()),
		Offset:        int(req.Msg.GetOffset()),
		Filters:       filterConfigFromValues(req.Msg.GetFilter()),
		RankBy:        req.Msg.GetRankby(),
		Queries:       req.Msg.GetQuery(),
		IncludeFields: req.Msg.GetIncludeFields(),
		Flags:         getSearchFlagsFromProto(req.Msg.GetFlags()),
		Namespace:     ns,
	}

	results, err := server.assetService.SearchAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(ctx, "error searching asset", err)
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, sr := range results {
		assetPB, err := assetToProto(sr.ToAsset(), false)
		if err != nil {
			return nil, internalServerError(ctx, "error converting assets to proto", err)
		}
		assetsPB = append(assetsPB, assetPB)
	}

	return connect.NewResponse(&compassv1beta1.SearchAssetsResponse{
		Data: assetsPB,
	}), nil
}

func (server *Handler) SuggestAssets(ctx context.Context, req *connect.Request[compassv1beta1.SuggestAssetsRequest]) (*connect.Response[compassv1beta1.SuggestAssetsResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	text := strings.TrimSpace(req.Msg.GetText())
	if text == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("'text' must be specified"))
	}

	cfg := asset.SearchConfig{
		Text:      text,
		Namespace: ns,
	}
	suggestions, err := server.assetService.SuggestAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(ctx, "internal error", err)
	}

	return connect.NewResponse(&compassv1beta1.SuggestAssetsResponse{
		Data: suggestions,
	}), nil
}

func (server *Handler) GroupAssets(ctx context.Context, req *connect.Request[compassv1beta1.GroupAssetsRequest]) (*connect.Response[compassv1beta1.GroupAssetsResponse], error) {
	ns := middleware.FetchNamespaceFromContext(ctx)
	if _, err := server.validateUserInCtx(ctx, ns); err != nil {
		return nil, err
	}

	if len(req.Msg.GetGroupby()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("groupby must be specified"))
	}

	cfg := asset.GroupConfig{
		GroupBy:       req.Msg.GetGroupby(),
		Filters:       filterConfigFromValues(req.Msg.GetFilter()),
		IncludeFields: req.Msg.GetIncludeFields(),
		Size:          int(req.Msg.GetSize()),
		Namespace:     ns,
	}

	results, err := server.assetService.GroupAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(ctx, "error grouping assets", err)
	}

	var groups []*compassv1beta1.AssetGroup
	for _, gr := range results {
		var fields []*compassv1beta1.GroupField
		for _, f := range gr.Fields {
			fields = append(fields, &compassv1beta1.GroupField{
				GroupKey:   f.Key,
				GroupValue: f.Value,
			})
		}
		var assets []*compassv1beta1.Asset
		for _, a := range gr.Assets {
			ap, err := assetToProto(a, false)
			if err != nil {
				return nil, internalServerError(ctx, "error converting asset to proto", err)
			}
			assets = append(assets, ap)
		}
		groups = append(groups, &compassv1beta1.AssetGroup{
			GroupFields: fields,
			Assets:      assets,
		})
	}

	return connect.NewResponse(&compassv1beta1.GroupAssetsResponse{
		AssetGroups: groups,
	}), nil
}

func getSearchFlagsFromProto(flags *compassv1beta1.SearchFlags) asset.SearchFlags {
	if flags == nil {
		return asset.SearchFlags{}
	}
	return asset.SearchFlags{
		DisableFuzzy:    flags.GetDisableFuzzy(),
		EnableHighlight: flags.GetEnableHighlight(),
		IsColumnSearch:  flags.GetIsColumnSearch(),
	}
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
