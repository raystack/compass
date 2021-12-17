package handlers

import (
	"net/url"
	"strings"
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
