package v1beta1

import (
	"context"
	"fmt"
	"strings"

	compassv1beta1 "github.com/odpf/compass/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/compass/discovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	whiteListQueryParamKey = "type"
)

func (h *Handler) SearchAssets(ctx context.Context, req *compassv1beta1.SearchAssetsRequest) (*compassv1beta1.SearchAssetsResponse, error) {
	text := strings.TrimSpace(req.GetText())
	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "'text' must be specified")
	}

	cfg := discovery.SearchConfig{
		Text:          text,
		MaxResults:    int(req.GetSize()),
		Filters:       filterConfigFromValues(req.GetFilter()),
		RankBy:        req.GetRankby(),
		Queries:       req.GetQuery(),
		TypeWhiteList: parseTypeWhiteList(req.GetFilter()),
	}

	results, err := h.DiscoveryRepository.Search(ctx, cfg)
	if err != nil {
		return nil, internalServerError(h.Logger, fmt.Sprintf("error searching asset: %s", err.Error()))
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, sr := range results {
		assetPB, err := sr.ToAsset().ToProto(false)
		if err != nil {
			return nil, internalServerError(h.Logger, fmt.Sprintf("error converting assets to proto: %s", err.Error()))
		}
		assetsPB = append(assetsPB, assetPB)
	}

	return &compassv1beta1.SearchAssetsResponse{
		Data: assetsPB,
	}, nil
}

func (h *Handler) SuggestAssets(ctx context.Context, req *compassv1beta1.SuggestAssetsRequest) (*compassv1beta1.SuggestAssetsResponse, error) {
	text := strings.TrimSpace(req.GetText())
	if text == "" {
		return nil, status.Error(codes.InvalidArgument, "'text' must be specified")
	}

	cfg := discovery.SearchConfig{
		Text: text,
	}

	suggestions, err := h.DiscoveryRepository.Suggest(ctx, cfg)
	if err != nil {
		return nil, internalServerError(h.Logger, err.Error())
	}

	return &compassv1beta1.SuggestAssetsResponse{
		Data: suggestions,
	}, nil
}

func filterConfigFromValues(fltMap map[string]string) map[string][]string {
	var filter = make(map[string][]string)
	for key, value := range fltMap {
		// filters are of form "filter[{field}]", apart from "filter[type]", which is used
		// for building the type whitelist.
		if key == whiteListQueryParamKey {
			continue
		}

		var filterValues []string
		filterValues = append(filterValues, strings.Split(value, ",")...)

		filter[key] = filterValues
	}
	return filter
}

func parseTypeWhiteList(fltMap map[string]string) (types []string) {
	if val, ok := fltMap[whiteListQueryParamKey]; ok {
		types = append(types, strings.Split(val, ",")...)
	}
	return
}
