package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/raystack/compass/core/document"
	"github.com/raystack/compass/core/namespace"
	"github.com/raystack/compass/internal/middleware"
)

// DocumentService defines document operations for the handler.
type DocumentService interface {
	Upsert(ctx context.Context, ns *namespace.Namespace, doc *document.Document) (string, error)
	GetByID(ctx context.Context, id string) (document.Document, error)
	GetByEntityURN(ctx context.Context, ns *namespace.Namespace, entityURN string) ([]document.Document, error)
	GetAll(ctx context.Context, ns *namespace.Namespace, filter document.Filter) ([]document.Document, error)
	Delete(ctx context.Context, ns *namespace.Namespace, id string) error
}

// DocumentHandler handles HTTP requests for document CRUD.
type DocumentHandler struct {
	service DocumentService
}

func NewDocumentHandler(service DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

// RegisterRoutes registers document HTTP routes on the mux.
func (h *DocumentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/documents", h.upsert)
	mux.HandleFunc("GET /v1/documents", h.list)
	mux.HandleFunc("GET /v1/documents/{id}", h.get)
	mux.HandleFunc("DELETE /v1/documents/{id}", h.delete)
	mux.HandleFunc("GET /v1/entities/{urn}/documents", h.getByEntity)
}

func (h *DocumentHandler) upsert(w http.ResponseWriter, r *http.Request) {
	ns := middleware.FetchNamespaceFromContext(r.Context())

	var req struct {
		EntityURN  string                 `json:"entity_urn"`
		Title      string                 `json:"title"`
		Body       string                 `json:"body"`
		Format     string                 `json:"format,omitempty"`
		Source     string                 `json:"source,omitempty"`
		SourceID   string                 `json:"source_id,omitempty"`
		Properties map[string]interface{} `json:"properties,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.EntityURN == "" || req.Title == "" || req.Body == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "entity_urn, title, and body are required"})
		return
	}

	doc := &document.Document{
		EntityURN:  req.EntityURN,
		Title:      req.Title,
		Body:       req.Body,
		Format:     req.Format,
		Source:     req.Source,
		SourceID:   req.SourceID,
		Properties: req.Properties,
	}

	id, err := h.service.Upsert(r.Context(), ns, doc)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

func (h *DocumentHandler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	doc, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *DocumentHandler) list(w http.ResponseWriter, r *http.Request) {
	ns := middleware.FetchNamespaceFromContext(r.Context())

	filter := document.Filter{
		EntityURN: r.URL.Query().Get("entity_urn"),
		Source:    r.URL.Query().Get("source"),
	}

	docs, err := h.service.GetAll(r.Context(), ns, filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": docs})
}

func (h *DocumentHandler) delete(w http.ResponseWriter, r *http.Request) {
	ns := middleware.FetchNamespaceFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	if err := h.service.Delete(r.Context(), ns, id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *DocumentHandler) getByEntity(w http.ResponseWriter, r *http.Request) {
	ns := middleware.FetchNamespaceFromContext(r.Context())
	urn := r.PathValue("urn")
	if urn == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "urn is required"})
		return
	}

	docs, err := h.service.GetByEntityURN(r.Context(), ns, urn)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": docs})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
