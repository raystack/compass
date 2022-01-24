package handlers

import (
	"fmt"
	"github.com/odpf/salt/log"
	"net/http"
	"strings"

	"github.com/odpf/columbus/record"
)

// TypeHandler exposes a REST interface to types
type TypeHandler struct {
	typeRepo record.TypeRepository
	logger   log.Logger
}

func NewTypeHandler(logger log.Logger, er record.TypeRepository) *TypeHandler {
	h := &TypeHandler{
		typeRepo: er,
		logger:   logger,
	}

	return h
}

func (h *TypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	typesNameMap, err := h.typeRepo.GetAll(r.Context())
	if err != nil {
		internalServerError(w, h.logger, "error fetching types")
		return
	}

	type TypeWithCount struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	results := []TypeWithCount{}
	for _, typName := range record.AllSupportedTypes {
		count := typesNameMap[typName]
		results = append(results, TypeWithCount{
			Name:  typName.String(),
			Count: count,
		})
	}

	writeJSON(w, http.StatusOK, results)
}
