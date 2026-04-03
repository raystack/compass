package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAI_Embed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var req openaiEmbeddingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "text-embedding-3-small" {
			t.Errorf("unexpected model: %s", req.Model)
		}

		resp := openaiEmbeddingsResponse{
			Data: []openaiEmbeddingData{
				{Embedding: []float64{0.1, 0.2, 0.3}, Index: 0},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	openai := NewOpenAI(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	vec, err := openai.Embed(context.Background(), "test text")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(vec) != 3 {
		t.Fatalf("expected 3 dimensions, got %d", len(vec))
	}
}

func TestOpenAI_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openaiEmbeddingsRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		data := make([]openaiEmbeddingData, len(req.Input))
		for i := range req.Input {
			data[i] = openaiEmbeddingData{Embedding: []float64{float64(i), 0.5}, Index: i}
		}
		_ = json.NewEncoder(w).Encode(openaiEmbeddingsResponse{Data: data})
	}))
	defer server.Close()

	openai := NewOpenAI(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	vectors, err := openai.EmbedBatch(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("EmbedBatch failed: %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vectors))
	}
}

func TestOpenAI_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid api key"}}`))
	}))
	defer server.Close()

	openai := NewOpenAI(OpenAIConfig{APIKey: "bad-key", BaseURL: server.URL})
	_, err := openai.Embed(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOpenAI_Name(t *testing.T) {
	o := NewOpenAI(OpenAIConfig{Model: "text-embedding-3-small"})
	if o.Name() != "openai/text-embedding-3-small" {
		t.Errorf("unexpected name: %s", o.Name())
	}
}
