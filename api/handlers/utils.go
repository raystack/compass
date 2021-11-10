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
			for _, v := range strings.Split(value, ",") {
				filterValues = append(filterValues, v)
			}
		}

		filterKey := strings.TrimPrefix(key, filterPrefix)
		filter[filterKey] = filterValues
	}
	return filter
}
