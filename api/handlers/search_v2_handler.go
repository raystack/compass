package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/odpf/columbus/models"
)

type SearchV2Handler struct {
	recordSearcher models.RecordV2Searcher
	typeRepo       models.TypeRepository
	log            logrus.FieldLogger
}

func NewSearchV2Handler(log logrus.FieldLogger, searcher models.RecordV2Searcher, repo models.TypeRepository) *SearchV2Handler {
	handler := &SearchV2Handler{
		recordSearcher: searcher,
		typeRepo:       repo,
		log:            log,
	}

	return handler
}

func (handler *SearchV2Handler) Search(w http.ResponseWriter, r *http.Request) {
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

func (handler *SearchV2Handler) buildSearchCfg(params url.Values) (cfg models.SearchConfig, err error) {
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

func (handler *SearchV2Handler) toSearchResponse(ctx context.Context, results []models.SearchResultV2) (response []SearchResponse, err error) {
	typeRepo := newCachingTypeRepo(handler.typeRepo)
	for _, result := range results {
		recordType, err := typeRepo.GetByName(ctx, result.TypeName)
		if err != nil {
			return nil, fmt.Errorf("typeRepository.GetByName: %q: %v", result.TypeName, err)
		}

		res := SearchResponse{
			ID:             result.RecordV2.Urn,
			Title:          result.RecordV2.Name,
			Description:    result.RecordV2.Description,
			Labels:         result.RecordV2.Labels,
			Type:           recordType.Name,
			Classification: string(recordType.Classification),
		}

		response = append(response, res)
	}
	return
}
