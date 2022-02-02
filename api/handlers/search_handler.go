package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/odpf/columbus/discovery"
	"github.com/odpf/salt/log"
)

var (
	filterPrefix           = "filter."
	whiteListQueryParamKey = "filter.type"

	queryPrefix = "query."
)

type SearchHandler struct {
	discoveryService *discovery.Service
	logger           log.Logger
}

func NewSearchHandler(logger log.Logger, discoveryService *discovery.Service) *SearchHandler {
	handler := &SearchHandler{
		discoveryService: discoveryService,
		logger:           logger,
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
		handler.logger.Error("error searching assets", "error", err.Error())
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
		handler.logger.Error("error building suggestions", "error", err.Error())
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
	cfg.Queries = queryConfigFromValues(params)
	cfg.TypeWhiteList, err = parseTypeWhiteList(params)
	return
}

func parseTypeWhiteList(values url.Values) (types []string, err error) {
	for _, commaSeparatedTypes := range values[whiteListQueryParamKey] {
		types = append(types, strings.Split(commaSeparatedTypes, ",")...)
	}
	return
}
