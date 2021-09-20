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
	recordSearcher models.RecordV1Searcher
	typeRepo       models.TypeRepository
	log            logrus.FieldLogger
}

func NewSearchHandler(log logrus.FieldLogger, searcher models.RecordV1Searcher, repo models.TypeRepository) *SearchHandler {
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

		rv := newRecordV1View(result.RecordV1)

		description, _ := getStringFromGenericMap(result.RecordV1, recordType.Fields.Description)
		res := SearchResponse{
			ID:             rv.GetString(recordType.Fields.ID),
			Title:          rv.GetString(recordType.Fields.Title),
			Description:    description,
			Type:           recordType.Name,
			Classification: string(recordType.Classification),
			Labels:         make(map[string]string),
		}

		if err := rv.Error(); err != nil {
			handler.log.
				WithField("record", result.RecordV1).
				Errorf("dropping record from search result: missing mandatory field: %v", err)
			continue
		}

		for _, label := range recordType.Fields.Labels {
			value, err := getStringFromGenericMap(result.RecordV1, label)
			if err != nil {
				continue
			}
			res.Labels[label] = value
		}

		response = append(response, res)
	}
	return
}

// recordView is a helper for querying record fields.
// It provides a fail-through interface for obtaining
// string fields from a record. If an error is encountered,
// all subsequent GetString operations will return immediately
// with an empty string, while the Error method will return
// the error that was encountered
type recordView struct {
	err      error
	RecordV1 models.RecordV1
}

func newRecordV1View(record models.RecordV1) *recordView {
	return &recordView{RecordV1: record}
}

func (view *recordView) GetString(name string) string {
	if view.err != nil {
		return ""
	}
	var val string
	val, view.err = getStringFromGenericMap(view.RecordV1, name)
	if view.err != nil {
		return ""
	}
	return val
}

func (view *recordView) Error() error {
	return view.err
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

func getStringFromGenericMap(m map[string]interface{}, key string) (string, error) {
	val, exists := m[key]
	if !exists {
		return "", fmt.Errorf("no such key: %q", key)
	}
	stringVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("not a string field: %q", key)
	}
	return stringVal, nil
}

func filterConfigFromValues(values url.Values) map[string][]string {
	var filter = make(map[string][]string)
	for key, fields := range values {
		// filters are of form "filter.{field}", apart from "filter.type", which is used
		// for building the type whitelist.
		if !strings.HasPrefix(key, filterPrefix) || strings.EqualFold(key, whiteListQueryParamKey) {
			continue
		}
		filterKey := strings.TrimPrefix(key, filterPrefix)
		filter[filterKey] = fields
	}
	return filter
}

func parseTypeWhiteList(values url.Values) (types []string) {
	for _, typ := range values[whiteListQueryParamKey] {
		typList := strings.Split(typ, ",")
		types = append(types, typList...)
	}
	return
}
