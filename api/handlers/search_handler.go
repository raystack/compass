package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/odpf/columbus/models"
)

var (
	filterPrefix           = "filter."
	whiteListQueryParamKey = "filter.type"
)

type SearchHandler struct {
	recordSearcher models.RecordSearcher
	typeRepo       models.TypeRepository
	log            logrus.FieldLogger
}

func NewSearchHandler(log logrus.FieldLogger, searcher models.RecordSearcher, repo models.TypeRepository) *SearchHandler {
	handler := &SearchHandler{
		recordSearcher: searcher,
		typeRepo:       repo,
		log:            log,
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
	results, err := handler.recordSearcher.Search(ctx, cfg)
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

func (handler *SearchHandler) buildSearchCfg(params url.Values) (cfg models.SearchConfig, err error) {
	text := strings.TrimSpace(params.Get("text"))
	if text == "" {
		err = fmt.Errorf("'text' must be specified")
		return
	}
	cfg.Text = text
	cfg.MaxResults, _ = strconv.Atoi(params.Get("size"))
	cfg.Filters = filterConfigFromValues(params)
	cfg.TypeWhiteList = parseTypeWhiteList(params)
	return
}

func (handler *SearchHandler) toSearchResponse(ctx context.Context, results []models.SearchResult) (response []SearchResponse, err error) {
	typeRepo := newCachingTypeRepo(handler.typeRepo)
	for _, result := range results {
		recordType, err := typeRepo.GetByName(ctx, result.TypeName)
		if err != nil {
			return nil, fmt.Errorf("typeRepository.GetByName: %q: %v", result.TypeName, err)
		}

		res := SearchResponse{
			Type:        recordType.Name,
			ID:          result.Record.Urn,
			Title:       result.Record.Name,
			Description: result.Record.Description,
			Labels:      result.Record.Labels,
			Service:     result.Record.Service,
		}

		response = append(response, res)
	}
	return
}

// cachingTypeRepo is a decorator over a models.TypeRepository
// that caches results of previous read-only operations
type cachingTypeRepo struct {
	mu    sync.Mutex
	cache map[string]models.Type
	repo  models.TypeRepository
}

func (decorator *cachingTypeRepo) CreateOrReplace(ctx context.Context, ent models.Type) error {
	panic("not implemented")
}

func (decorator *cachingTypeRepo) GetByName(ctx context.Context, name string) (models.Type, error) {
	ent, exists := decorator.cache[name]
	if exists {
		return ent, nil
	}

	decorator.mu.Lock()
	defer decorator.mu.Unlock()
	ent, err := decorator.repo.GetByName(ctx, name)
	if err != nil {
		return ent, err
	}
	decorator.cache[ent.Name] = ent
	return ent, nil
}

func newCachingTypeRepo(repo models.TypeRepository) *cachingTypeRepo {
	return &cachingTypeRepo{
		repo:  repo,
		cache: make(map[string]models.Type),
	}
}

func parseTypeWhiteList(values url.Values) (types []string) {
	for _, typ := range values[whiteListQueryParamKey] {
		typList := strings.Split(typ, ",")
		types = append(types, typList...)
	}
	return
}
