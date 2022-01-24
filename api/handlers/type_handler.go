package handlers

import (
	"net/http"

	"github.com/odpf/columbus/asset"
	"github.com/odpf/columbus/discovery"
	"github.com/odpf/salt/log"
)

// TypeHandler exposes a REST interface to types
type TypeHandler struct {
	typeRepo discovery.TypeRepository
	logger   log.Logger
}

func NewTypeHandler(logger log.Logger, er discovery.TypeRepository) *TypeHandler {
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
	for _, typName := range asset.AllSupportedTypes {
		count, _ := typesNameMap[typName]
		results = append(results, TypeWithCount{
			Name:  typName.String(),
			Count: count,
		})
	}

	writeJSON(w, http.StatusOK, results)
}
