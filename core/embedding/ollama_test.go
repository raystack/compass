package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllama_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req ollamaEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "nomic-embed-text" {
			t.Errorf("unexpected model: %s", req.Model)
		}
		if len(req.Input) != 1 {
			t.Errorf("expected 1 input, got %d", len(req.Input))
		}

		resp := ollamaEmbedResponse{
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	ollama := NewOllama(OllamaConfig{Host: server.URL, Model: "nomic-embed-text"})
	vec, err := ollama.Embed(context.Background(), "test text")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(vec) != 3 {
		t.Fatalf("expected 3 dimensions, got %d", len(vec))
	}
	if vec[0] != 0.1 {
		t.Errorf("expected vec[0]=0.1, got %f", vec[0])
	}
}

func TestOllama_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaEmbedRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		embeddings := make([][]float64, len(req.Input))
		for i := range req.Input {
			embeddings[i] = []float64{float64(i), 0.5, 1.0}
		}
		_ = json.NewEncoder(w).Encode(ollamaEmbedResponse{Embeddings: embeddings})
	}))
	defer server.Close()

	ollama := NewOllama(OllamaConfig{Host: server.URL})
	vectors, err := ollama.EmbedBatch(context.Background(), []string{"text1", "text2", "text3"})
	if err != nil {
		t.Fatalf("EmbedBatch failed: %v", err)
	}
	if len(vectors) != 3 {
		t.Fatalf("expected 3 vectors, got %d", len(vectors))
	}
}

func TestOllama_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("model not found"))
	}))
	defer server.Close()

	ollama := NewOllama(OllamaConfig{Host: server.URL})
	_, err := ollama.Embed(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOllama_Name(t *testing.T) {
	o := NewOllama(OllamaConfig{Model: "nomic-embed-text"})
	if o.Name() != "ollama/nomic-embed-text" {
		t.Errorf("unexpected name: %s", o.Name())
	}
}
