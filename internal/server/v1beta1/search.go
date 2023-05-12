package handlersv1beta1

import (
	"context"
	"fmt"
	"strings"

	"github.com/goto/compass/core/asset"
	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *APIServer) SearchAssets(ctx context.Context, req *compassv1beta1.SearchAssetsRequest) (*compassv1beta1.SearchAssetsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
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

func (server *APIServer) GroupAssets(ctx context.Context, req *compassv1beta1.GroupAssetsRequest) (*compassv1beta1.GroupAssetsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("group assets: %w", err)
	}

	if len(req.GetGroupby()) == 0 || req.GetGroupby()[0] == "" {
		return nil, status.Error(codes.InvalidArgument, "'group_by' must be specified")
	}

	cfg := asset.GroupConfig{
		GroupBy:        req.GetGroupby(),
		Filters:        filterConfigFromValues(req.GetFilter()),
		IncludedFields: req.GetIncludeFields(),
		Size:           int(req.GetSize()),
	}

	results, err := server.assetService.GroupAssets(ctx, cfg)
	if err != nil {
		return nil, internalServerError(server.logger, fmt.Sprintf("group asset: %s", err))
	}

	groupInfoArr := make([]*compassv1beta1.AssetGroup, len(results))
	for i, gr := range results {
		assetsPB := make([]*compassv1beta1.Asset, len(gr.Assets))
		for j, as := range gr.Assets {
			assetPB, err := assetToProto(as, false)
			if err != nil {
				return nil, internalServerError(server.logger, fmt.Sprintf("convert asset to proto: %s", err))
			}
			assetsPB[j] = assetPB
		}

		fields := make([]*compassv1beta1.GroupField, len(gr.Fields))
		for j, gf := range gr.Fields {
			fields[j] = &compassv1beta1.GroupField{
				GroupKey:   gf.Name,
				GroupValue: gf.Value,
			}
		}
		groupInfoArr[i] = &compassv1beta1.AssetGroup{
			GroupFields: fields,
			Assets:      assetsPB,
		}
	}

	return &compassv1beta1.GroupAssetsResponse{
		AssetGroups: groupInfoArr,
	}, nil
}

func (server *APIServer) SuggestAssets(ctx context.Context, req *compassv1beta1.SuggestAssetsRequest) (*compassv1beta1.SuggestAssetsResponse, error) {
	_, err := server.validateUserInCtx(ctx)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(req.GetText())
	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "'text' must be specified")
	}

	cfg := asset.SearchConfig{
		Text: text,
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
