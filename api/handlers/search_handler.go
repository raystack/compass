package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/odpf/columbus/discovery"
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
		handler.log.Errorf("error searching records: %s", err.Error())
		writeJSONError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (handler *SearchHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg, err := handler.buildSearchCfg(r.URL.Query())
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	suggestions, err := handler.discoveryService.Suggest(ctx, cfg)
	if err != nil {
		handler.log.Errorf("error building suggestions: %s", err.Error())
		writeJSONError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	writeJSON(w, http.StatusOK, SuggestResponse{
		Suggestions: suggestions,
	})
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
	cfg.RankBy = params.Get("rankby")
	cfg.TypeWhiteList, err = parseTypeWhiteList(params)
	return
}

func parseTypeWhiteList(values url.Values) (types []string, err error) {
	for _, commaSeparatedTypes := range values[whiteListQueryParamKey] {
		for _, ts := range strings.Split(commaSeparatedTypes, ",") {
			types = append(types, ts)
		}
	}
	return
}
