package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/odpf/columbus/discovery"
	"github.com/odpf/columbus/record"
)

var (
	filterPrefix           = "filter."
	whiteListQueryParamKey = "filter.type"
)

type SearchHandler struct {
	discoveryService *discovery.Service
	log              logrus.FieldLogger
}

func NewSearchHandler(log logrus.FieldLogger, discoveryService *discovery.Service) *SearchHandler {
	handler := &SearchHandler{
		discoveryService: discoveryService,
		log:              log,
	}

	return handler
}

func (handler *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg, err := handler.buildSearchCfg(r.URL.Query())
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	results, err := handler.discoveryService.Search(ctx, cfg)
	if err != nil {
		handler.log.Errorf("error searching records: %w", err)
		writeJSONError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	response, err := handler.toSearchResponse(ctx, results)
	if err != nil {
		handler.log.Errorf("error mapping search results: %w", err)
		writeJSONError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	// if the search response is empty, instead of returning
	// a 'null' return an empty list
	// this can happen where there are no viable search results,
	if response == nil {
		response = make([]SearchResponse, 0)
	}
	writeJSON(w, http.StatusOK, response)
}

func (handler *SearchHandler) buildSearchCfg(params url.Values) (cfg discovery.SearchConfig, err error) {
	text := strings.TrimSpace(params.Get("text"))
	if text == "" {
		err = fmt.Errorf("'text' must be specified")
		return
	}
	cfg.Text = text
	cfg.MaxResults, _ = strconv.Atoi(params.Get("size"))
	cfg.Filters = filterConfigFromValues(params)
	cfg.TypeWhiteList, err = parseTypeWhiteList(params)
	if err != nil {
		return
	}

	return
}

func (handler *SearchHandler) toSearchResponse(ctx context.Context, records []record.Record) (response []SearchResponse, err error) {
	for _, r := range records {
		res := SearchResponse{
			ID:          r.Urn,
			Title:       r.Name,
			Description: r.Description,
			Labels:      r.Labels,
			Service:     r.Service,
			Type:        r.Type.String(),
		}

		response = append(response, res)
	}
	return
}

func parseTypeWhiteList(values url.Values) (types []record.Type, err error) {
	for _, commaSeparatedTypes := range values[whiteListQueryParamKey] {
		for _, ts := range strings.Split(commaSeparatedTypes, ",") {
			var t record.Type
			t, err = validateType(ts)
			if err != nil {
				return
			}
			types = append(types, t)
		}
	}
	return
}
