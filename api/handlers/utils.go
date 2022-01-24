package handlers

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/star"
	"github.com/odpf/salt/log"
)

func filterConfigFromValues(querystring url.Values) map[string][]string {
	var filter = make(map[string][]string)
	for key, values := range querystring {
		// filters are of form "filter.{field}", apart from "filter.type", which is used
		// for building the type whitelist.
		if !strings.HasPrefix(key, filterPrefix) || strings.EqualFold(key, whiteListQueryParamKey) {
			continue
		}

		var filterValues []string
		for _, value := range values {
			filterValues = append(filterValues, strings.Split(value, ",")...)
		}

		filterKey := strings.TrimPrefix(key, filterPrefix)
		filter[filterKey] = filterValues
	}
	return filter
}

func queryConfigFromValues(querystring url.Values) map[string]string {
	var query = make(map[string]string)
	for key, values := range querystring {
		// filters are of form "query.{field}"
		if !strings.HasPrefix(key, queryPrefix) {
			continue
		}

		queryKey := strings.TrimPrefix(key, queryPrefix)
		query[queryKey] = values[0] // cannot have duplicate query key, always get the first one
	}
	return query
}

func buildStarConfig(logger log.Logger, query url.Values) star.Config {
	var offset, size int
	var err error
	sizeString := query.Get("size")
	if sizeString != "" {
		size, err = strconv.Atoi(sizeString)
		if err != nil {
			logger.Warn("can't parse \"size\" query params")
		}
	}
	offsetString := query.Get("offset")
	if offsetString != "" {
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			logger.Warn("can't parse \"offset\" query params")
		}
	}
	return star.Config{Offset: offset, Size: size}
}

func buildStarFromPath(pathParams map[string]string) *star.Star {
	return &star.Star{
		Asset: asset.Asset{
			Type: asset.Type(pathParams["asset_type"]),
			URN:  pathParams["asset_urn"],
		},
	}
}
