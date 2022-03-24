package v1beta1

import (
	"context"
	"strings"

	compassv1beta1 "github.com/odpf/columbus/api/proto/odpf/compass/v1beta1"
	"github.com/odpf/columbus/discovery"
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

	results, err := h.DiscoveryService.Search(ctx, cfg)
	if err != nil {
		return nil, internalServerError(h.Logger, "error searching records")
	}

	assetsPB := []*compassv1beta1.Asset{}
	for _, sr := range results {
		assetPB, err := sr.ToAsset().ToProto()
		if err != nil {
			return nil, internalServerError(h.Logger, err.Error())
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
		Text:          text,
		MaxResults:    int(req.GetSize()),
		Filters:       filterConfigFromValues(req.GetFilter()),
		RankBy:        req.GetRankby(),
		Queries:       queryConfigFromValues(req.GetQuery()),
		TypeWhiteList: parseTypeWhiteList(req.GetFilter()),
	}

	suggestions, err := h.DiscoveryService.Suggest(ctx, cfg)
	if err != nil {
		return nil, internalServerError(h.Logger, "error building suggestions")
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

func queryConfigFromValues(queryMap map[string]string) map[string]string {
	var query = make(map[string]string)
	if len(queryMap) > 0 {
		query = queryMap
	}
	return query
}

func parseTypeWhiteList(fltMap map[string]string) (types []string) {
	if val, ok := fltMap[whiteListQueryParamKey]; ok {
		types = append(types, strings.Split(val, ",")...)
	}
	return
}
