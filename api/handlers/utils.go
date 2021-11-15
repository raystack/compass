package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/odpf/columbus/record"
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

func validateType(t string) (record.Type, error) {
	for _, validType := range record.TypeList {
		if string(validType) == t {
			return validType, nil
		}
	}

	return record.Type(t), fmt.Errorf("invalid type \"%s\"", t)
}
